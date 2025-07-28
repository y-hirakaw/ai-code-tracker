package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HookManager はGit hooks と Claude Code hooks の管理を提供する
type HookManager struct {
	gitRepo string
}

// NewHookManager は新しいHookManagerインスタンスを作成する
func NewHookManager(gitRepo string) *HookManager {
	return &HookManager{
		gitRepo: gitRepo,
	}
}

// ClaudeCodeHook はClaude Code hooks設定を表す
type ClaudeCodeHook struct {
	Matcher string `json:"matcher"`
	Hooks   []Hook `json:"hooks"`
}

// Hook は個別のhook定義を表す
type Hook struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// ClaudeCodeHooksConfig はClaude Code hooks設定全体を表す
type ClaudeCodeHooksConfig struct {
	Hooks map[string][]ClaudeCodeHook `json:"hooks"`
}

// SetupGitHooks はGit hooksを自動設定する
func (hm *HookManager) SetupGitHooks() error {
	hooksDir := filepath.Join(hm.gitRepo, ".git", "hooks")
	
	// hooksディレクトリが存在するか確認
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf("Git hooksディレクトリが存在しません: %s", hooksDir)
	}

	// post-commit hookを設定
	postCommitPath := filepath.Join(hooksDir, "post-commit")
	postCommitContent, err := hm.generatePostCommitHook()
	if err != nil {
		return fmt.Errorf("post-commit hookの生成に失敗しました: %w", err)
	}

	// 既存のpost-commit hookがある場合、バックアップを作成
	if _, err := os.Stat(postCommitPath); err == nil {
		backupPath := postCommitPath + ".backup"
		if err := os.Rename(postCommitPath, backupPath); err != nil {
			return fmt.Errorf("既存のpost-commit hookのバックアップに失敗しました: %w", err)
		}
		fmt.Printf("既存のpost-commit hookを %s にバックアップしました\n", backupPath)
	}

	// post-commit hookを書き込み
	err = os.WriteFile(postCommitPath, []byte(postCommitContent), 0755)
	if err != nil {
		return fmt.Errorf("post-commit hookの書き込みに失敗しました: %w", err)
	}

	fmt.Printf("Git post-commit hook を設定しました: %s\n", postCommitPath)
	return nil
}

// SetupClaudeCodeHooks はClaude Code hooksを設定する（settings.json版）
func (hm *HookManager) SetupClaudeCodeHooks() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// .claudeディレクトリを作成
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("Claudeディレクトリの作成に失敗: %w", err)
	}

	// 現在の設定を読み込み（存在する場合）
	var currentSettings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &currentSettings); err != nil {
			currentSettings = make(map[string]interface{})
		}
	} else {
		currentSettings = make(map[string]interface{})
	}

	// AICT hooks設定を作成（Claude CLI標準形式）
	aictHooks := map[string][]ClaudeCodeHook{
		"PreToolUse": {
			{
				Matcher: "Edit|Write|MultiEdit",
				Hooks: []Hook{
					{
						Type:    "command",
						Command: `echo '{"decision": "approve"}'`,
					},
				},
			},
		},
		"PostToolUse": {
			{
				Matcher: "Edit|Write|MultiEdit",
				Hooks: []Hook{
					{
						Type: "command", 
						Command: `bash -c 'INPUT=$(cat); FILE=$(echo "$INPUT" | jq -r ".tool_input.path // .tool_input.file_path // empty"); if [ -n "$FILE" ]; then aict track --ai --author "Claude Code" --model "claude-sonnet-4" --files "$FILE" --message "Claude Code automated edit" 2>/dev/null || true; fi; echo "{\"continue\": true}"'`,
					},
				},
			},
		},
		"Stop": {
			{
				Matcher: "*",
				Hooks: []Hook{
					{
						Type: "command",
						Command: `bash -c 'STATS=$(aict stats 2>/dev/null | head -3 || echo "No stats available"); echo "{\"continue\": true, \"userMessage\": \"📊 AICT Session: $STATS\"}" 2>/dev/null || echo "{\"continue\": true}"'`,
					},
				},
			},
		},
	}

	// 既存のhooks設定とマージ
	mergedHooks := hm.mergeHooksConfig(currentSettings, aictHooks)
	
	// 大文字小文字が混在するキーを統一（Claude CLI標準形式に合わせる）
	normalizedHooks := make(map[string][]ClaudeCodeHook)
	for hookType, hooks := range mergedHooks {
		switch strings.ToLower(hookType) {
		case "pretooluse":
			// 既存のPreToolUseと統合
			if existing, exists := normalizedHooks["PreToolUse"]; exists {
				normalizedHooks["PreToolUse"] = append(existing, hooks...)
			} else {
				normalizedHooks["PreToolUse"] = hooks
			}
		case "posttooluse":
			// 既存のPostToolUseと統合
			if existing, exists := normalizedHooks["PostToolUse"]; exists {
				normalizedHooks["PostToolUse"] = append(existing, hooks...)
			} else {
				normalizedHooks["PostToolUse"] = hooks
			}
		case "stop":
			// 既存のStopと統合
			if existing, exists := normalizedHooks["Stop"]; exists {
				normalizedHooks["Stop"] = append(existing, hooks...)
			} else {
				normalizedHooks["Stop"] = hooks
			}
		default:
			normalizedHooks[hookType] = hooks
		}
	}
	
	currentSettings["hooks"] = normalizedHooks

	// JSON形式で保存
	data, err := json.MarshalIndent(currentSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("設定のJSON変換に失敗: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("設定ファイルの保存に失敗: %w", err)
	}

	return nil
}

// mergeHooksConfig は既存のhooks設定とAICT hooksをマージする
func (hm *HookManager) mergeHooksConfig(currentSettings map[string]interface{}, aictHooks map[string][]ClaudeCodeHook) map[string][]ClaudeCodeHook {
	result := make(map[string][]ClaudeCodeHook)
	
	// 既存のhooks設定を取得
	if existingHooks, exists := currentSettings["hooks"]; exists {
		if hooksMap, ok := existingHooks.(map[string]interface{}); ok {
			// 各フック種別を処理（大文字小文字を統一）
			for hookType, hooksList := range hooksMap {
				// キーを小文字に統一
				normalizedHookType := strings.ToLower(hookType)
				if hooksArray, ok := hooksList.([]interface{}); ok {
					var mergedHooks []ClaudeCodeHook
					
					// 既存のhooksからAICT関連以外を保持
					for _, hook := range hooksArray {
						if hookItem, ok := hook.(map[string]interface{}); ok {
							claudeHook := ClaudeCodeHook{}
							
							// matcher設定
							if matcher, exists := hookItem["matcher"]; exists {
								if matcherStr, ok := matcher.(string); ok {
									claudeHook.Matcher = matcherStr
								}
							}
							
							// hooks配列を処理
							if hooksField, exists := hookItem["hooks"]; exists {
								if hooksDef, ok := hooksField.([]interface{}); ok {
									var hooks []Hook
									keepHook := true
									
									for _, h := range hooksDef {
										if hookDef, ok := h.(map[string]interface{}); ok {
											hook := Hook{}
											if hookType, exists := hookDef["type"]; exists {
												if typeStr, ok := hookType.(string); ok {
													hook.Type = typeStr
												}
											}
											if command, exists := hookDef["command"]; exists {
												if cmdStr, ok := command.(string); ok {
													hook.Command = cmdStr
													// AICT関連のhookかチェック（より広範囲に判定）
													if strings.Contains(cmdStr, "aict ") || 
													   strings.Contains(cmdStr, `'{"decision": "approve"}'`) ||
													   strings.Contains(cmdStr, "AICT Session") ||
													   (normalizedHookType == "notification" && cmdStr == "exit 0") {
														keepHook = false
														break
													}
												}
											}
											hooks = append(hooks, hook)
										}
									}
									
									if keepHook {
										claudeHook.Hooks = hooks
										mergedHooks = append(mergedHooks, claudeHook)
									}
								}
							}
						}
					}
					
					// 空の配列は結果に含めない
					if len(mergedHooks) > 0 {
						result[normalizedHookType] = mergedHooks
					}
				}
			}
		}
	}
	
	// AICT hooksを追加（大文字小文字を考慮してマージ）
	for hookType, aictHooksList := range aictHooks {
		// 既存のhooksと統合
		existing := result[hookType]
		result[hookType] = append(existing, aictHooksList...)
	}
	
	return result
}

// generatePostCommitHook はpost-commit hook スクリプトを生成する
func (hm *HookManager) generatePostCommitHook() (string, error) {
	// aictコマンドのパスを取得
	aictPath, err := exec.LookPath("aict")
	if err != nil {
		// PATH上にない場合は相対パスを使用
		aictPath = "aict"
	}

	postCommitScript := fmt.Sprintf(`#!/bin/bash
# AI Code Tracker - Git post-commit hook
# 自動生成されたファイル - 手動編集しないでください

# デバッグモード（環境変数で制御）
if [ "$ACT_DEBUG" = "1" ]; then
    set -x
    exec 2>>.git/ai-tracker/hook-debug.log
fi

# AI Code Trackerがインストールされているか確認
if ! command -v %s &> /dev/null; then
    exit 0
fi

# プロジェクトがAICTで初期化されているか確認
if [ ! -d ".git/ai-tracker" ]; then
    exit 0
fi

# 重複トラッキング防止機能
check_duplicate() {
    local LOCKFILE=".git/ai-tracker/.commit-lock"
    local CURRENT_TIME=$(date +%%s)
    
    if [ -f "$LOCKFILE" ]; then
        local LOCK_TIME=$(cat "$LOCKFILE" 2>/dev/null || echo 0)
        local TIME_DIFF=$((CURRENT_TIME - LOCK_TIME))
        
        if [ $TIME_DIFF -lt 5 ]; then
            [ "$ACT_DEBUG" = "1" ] && echo "[AICT] Skipping duplicate tracking (${TIME_DIFF}s since last)"
            exit 0
        fi
    fi
    
    echo "$CURRENT_TIME" > "$LOCKFILE"
}

# メイン処理
main() {
    check_duplicate
    
    # コミット情報を取得
    local COMMIT_SHA=$(git rev-parse HEAD)
    local COMMIT_MSG=$(git log -1 --pretty=%%B)
    local COMMIT_AUTHOR=$(git log -1 --pretty=%%an)
    local COMMIT_EMAIL=$(git log -1 --pretty=%%ae)
    
    # Claude Codeのコミットパターンを検出
    local IS_CLAUDE=false
    if [[ "$COMMIT_MSG" =~ "🤖 Generated with [Claude Code]" ]] || \
       [[ "$COMMIT_MSG" =~ "Co-Authored-By: Claude" ]] || \
       [[ "$COMMIT_AUTHOR" =~ ^(Claude|claude) ]] || \
       [[ "$COMMIT_EMAIL" =~ claude ]] || \
       [[ "$COMMIT_EMAIL" =~ anthropic ]]; then
        IS_CLAUDE=true
    fi
    
    # トラッキング実行
    local TRACK_ARGS=(
        "--commit-hash" "$COMMIT_SHA"
        "--message" "$(echo "$COMMIT_MSG" | head -1 | cut -c1-100)"
        "--quiet"
    )
    
    if [ "$IS_CLAUDE" = true ]; then
        # AIコミットとして記録
        %s track --ai --author "Claude Code" --model "claude-sonnet-4" "${TRACK_ARGS[@]}" 2>/dev/null || true
    else
        # 人間のコミットとして記録
        %s track --author "$COMMIT_AUTHOR" "${TRACK_ARGS[@]}" 2>/dev/null || true
    fi
    
    # 統計表示（オプション）
    if [ "$ACT_SHOW_STATS" = "1" ]; then
        echo "───────────────────────────────────────"
        %s stats --format summary 2>/dev/null || true
        echo "───────────────────────────────────────"
    fi
}

# エラーハンドリング
trap 'rm -f .git/ai-tracker/.commit-lock' EXIT

# メイン処理実行
main

exit 0
`, aictPath, aictPath, aictPath, aictPath)

	return postCommitScript, nil
}

// generateClaudeCodeHooksConfig はClaude Code hooks設定を生成する
func (hm *HookManager) generateClaudeCodeHooksConfig() (*ClaudeCodeHooksConfig, error) {
	// aictコマンドのパスを取得
	aictPath, err := exec.LookPath("aict")
	if err != nil {
		aictPath = "aict"
	}

	config := &ClaudeCodeHooksConfig{
		Hooks: map[string][]ClaudeCodeHook{
			"preToolUse": {
				{
					Matcher: "Edit|Write|MultiEdit",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: fmt.Sprintf(`bash -c 'INPUT=$(cat); FILE=$(echo "$INPUT" | jq -r ".tool_input.path // .tool_input.file_path // empty"); if [ -n "$FILE" ] && [ -f "$FILE" ]; then %s track --quiet --pre-edit --files "$FILE" 2>/dev/null; fi; echo "{\"decision\": \"approve\"}"'`, aictPath),
						},
					},
				},
			},
			"postToolUse": {
				{
					Matcher: "Edit|Write|MultiEdit",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: fmt.Sprintf(`bash -c 'INPUT=$(cat); FILE=$(echo "$INPUT" | jq -r ".tool_input.path // .tool_input.file_path // empty"); MODEL=$(echo "$INPUT" | jq -r ".metadata.model // \"claude-sonnet-4\""); if [ -n "$FILE" ]; then %s track --quiet --ai --author "Claude Code" --model "$MODEL" --files "$FILE" 2>/dev/null || true; fi; echo "{\"continue\": true}"'`, aictPath),
						},
					},
				},
				{
					Matcher: "Bash",
					Hooks: []Hook{
						{
							Type:    "command",
							Command: fmt.Sprintf(`bash -c 'INPUT=$(cat); CMD=$(echo "$INPUT" | jq -r ".tool_input.command // empty"); MODEL=$(echo "$INPUT" | jq -r ".metadata.model // \"claude-sonnet-4\""); %s track --quiet --ai --author "Claude Code" --model "$MODEL" --command "$CMD" 2>/dev/null || true; echo "{\"continue\": true}"'`, aictPath),
						},
					},
				},
			},
			"stop": {
				{
					Hooks: []Hook{
						{
							Type:    "command",
							Command: fmt.Sprintf(`bash -c 'STATS=$(%s stats --format json --since $(date -d "1 hour ago" +%%Y-%%m-%%d) 2>/dev/null || echo "{}"); if [ "$STATS" != "{}" ]; then AI_EVENTS=$(echo "$STATS" | jq -r ".ai_events // 0"); HUMAN_EVENTS=$(echo "$STATS" | jq -r ".human_events // 0"); TOTAL=$((AI_EVENTS + HUMAN_EVENTS)); if [ $TOTAL -gt 0 ]; then PERCENT=$((AI_EVENTS * 100 / TOTAL)); echo "{\"continue\": true, \"userMessage\": \"📊 Session: AI: $AI_EVENTS events ($PERCENT%%), Human: $HUMAN_EVENTS events\"}"; else echo "{\"continue\": true}"; fi; else echo "{\"continue\": true}"; fi'`, aictPath),
						},
					},
				},
			},
			"notification": {
				{
					Hooks: []Hook{
						{
							Type:    "command",
							Command: fmt.Sprintf(`bash -c 'INPUT=$(cat); MSG=$(echo "$INPUT" | jq -r ".message // empty"); if [[ "$MSG" == *"idle"* ]] || [[ "$MSG" == *"permission"* ]]; then %s track --quiet --checkpoint "session" 2>/dev/null || true; fi; exit 0'`, aictPath),
						},
					},
				},
			},
		},
	}

	return config, nil
}

// RemoveGitHooks はGit hooksを削除する
func (hm *HookManager) RemoveGitHooks() error {
	hooksDir := filepath.Join(hm.gitRepo, ".git", "hooks")
	postCommitPath := filepath.Join(hooksDir, "post-commit")

	// AICT製のhookかどうかチェック
	if content, err := os.ReadFile(postCommitPath); err == nil {
		if strings.Contains(string(content), "AI Code Tracker - Git post-commit hook") {
			// バックアップが存在する場合は復元
			backupPath := postCommitPath + ".backup"
			if _, err := os.Stat(backupPath); err == nil {
				if err := os.Rename(backupPath, postCommitPath); err != nil {
					return fmt.Errorf("post-commit hookの復元に失敗しました: %w", err)
				}
				fmt.Printf("post-commit hookを復元しました: %s\n", postCommitPath)
			} else {
				// バックアップがない場合は削除
				if err := os.Remove(postCommitPath); err != nil {
					return fmt.Errorf("post-commit hookの削除に失敗しました: %w", err)
				}
				fmt.Printf("post-commit hookを削除しました: %s\n", postCommitPath)
			}
		} else {
			fmt.Printf("post-commit hookはAICTによって管理されていません\n")
		}
	} else {
		fmt.Printf("post-commit hookが見つかりません\n")
	}

	return nil
}

// RemoveClaudeCodeHooks はClaude Code hooksを削除する（settings.json版）
func (hm *HookManager) RemoveClaudeCodeHooks() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// 設定ファイルが存在しない場合は何もしない
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return nil
	}

	// 現在の設定を読み込み
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("設定ファイルの解析に失敗: %w", err)
	}

	// hooksを削除
	delete(settings, "hooks")

	// 保存
	newData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("設定のJSON変換に失敗: %w", err)
	}

	if err := os.WriteFile(settingsPath, newData, 0644); err != nil {
		return fmt.Errorf("設定ファイルの保存に失敗: %w", err)
	}

	return nil
}

// ValidateGitRepo はGitリポジトリが有効かどうかを検証する
func (hm *HookManager) ValidateGitRepo() error {
	gitDir := filepath.Join(hm.gitRepo, ".git")
	
	// .gitディレクトリが存在するかチェック
	if info, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("Gitリポジトリではありません: %s", hm.gitRepo)
	} else if !info.IsDir() {
		return fmt.Errorf(".gitがディレクトリではありません: %s", gitDir)
	}

	// git configコマンドで確認
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = hm.gitRepo
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("有効なGitリポジトリではありません: %w", err)
	}

	return nil
}

// GetHookStatus はhooksの設定状況を取得する
func (hm *HookManager) GetHookStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// Git hooks状況
	hooksDir := filepath.Join(hm.gitRepo, ".git", "hooks")
	postCommitPath := filepath.Join(hooksDir, "post-commit")
	
	gitHookStatus := map[string]interface{}{
		"installed": false,
		"path":      postCommitPath,
		"backup":    false,
	}

	if content, err := os.ReadFile(postCommitPath); err == nil {
		if strings.Contains(string(content), "AI Code Tracker") {
			gitHookStatus["installed"] = true
		}
		
		// 実行可能かチェック
		if info, err := os.Stat(postCommitPath); err == nil {
			gitHookStatus["executable"] = info.Mode()&0111 != 0
		}
	}

	// バックアップファイルの存在チェック
	if _, err := os.Stat(postCommitPath + ".backup"); err == nil {
		gitHookStatus["backup"] = true
	}

	status["git_hooks"] = gitHookStatus

	// Claude Code hooks状況（claude.goの機能を使用）
	claudeHookStatus := hm.GetClaudeHookStatus()
	status["claude_hooks"] = claudeHookStatus

	return status, nil
}

// CheckPermissions は必要な権限があるかチェックする
func (hm *HookManager) CheckPermissions() error {
	// .git/hooksディレクトリへの書き込み権限をチェック
	hooksDir := filepath.Join(hm.gitRepo, ".git", "hooks")
	
	// テストファイルを作成して権限をチェック
	testFile := filepath.Join(hooksDir, ".aict-permission-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("Git hooksディレクトリへの書き込み権限がありません: %w", err)
	}
	defer os.Remove(testFile)

	// ホームディレクトリの.claudeへの書き込み権限をチェック
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("Claude設定ディレクトリの作成権限がありません: %w", err)
	}

	testFile2 := filepath.Join(claudeDir, ".aict-permission-test")
	if err := os.WriteFile(testFile2, []byte("test"), 0644); err != nil {
		return fmt.Errorf("Claude設定ディレクトリへの書き込み権限がありません: %w", err)
	}
	defer os.Remove(testFile2)

	return nil
}

// BackupExistingHooks は既存のhooksをバックアップする
func (hm *HookManager) BackupExistingHooks() error {
	hooksDir := filepath.Join(hm.gitRepo, ".git", "hooks")
	
	// 全てのhookファイルをチェック
	hookFiles := []string{"post-commit", "pre-commit", "pre-push", "post-merge"}
	
	for _, hookFile := range hookFiles {
		hookPath := filepath.Join(hooksDir, hookFile)
		backupPath := hookPath + ".aict-backup"
		
		if info, err := os.Stat(hookPath); err == nil && !info.IsDir() {
			// 既存のバックアップがある場合はスキップ
			if _, err := os.Stat(backupPath); err == nil {
				continue
			}
			
			// バックアップを作成
			content, err := os.ReadFile(hookPath)
			if err != nil {
				return fmt.Errorf("%s の読み込みに失敗しました: %w", hookFile, err)
			}
			
			if err := os.WriteFile(backupPath, content, info.Mode()); err != nil {
				return fmt.Errorf("%s のバックアップに失敗しました: %w", hookFile, err)
			}
			
			fmt.Printf("%s をバックアップしました: %s\n", hookFile, backupPath)
		}
	}
	
	return nil
}

// RestoreBackupHooks はバックアップからhooksを復元する
func (hm *HookManager) RestoreBackupHooks() error {
	hooksDir := filepath.Join(hm.gitRepo, ".git", "hooks")
	
	// バックアップファイルを検索
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		return fmt.Errorf("hooksディレクトリの読み込みに失敗しました: %w", err)
	}
	
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".aict-backup") {
			backupPath := filepath.Join(hooksDir, entry.Name())
			originalPath := strings.TrimSuffix(backupPath, ".aict-backup")
			
			// バックアップを復元
			content, err := os.ReadFile(backupPath)
			if err != nil {
				fmt.Printf("バックアップファイルの読み込みに失敗: %s\n", backupPath)
				continue
			}
			
			info, err := entry.Info()
			if err != nil {
				fmt.Printf("ファイル情報の取得に失敗: %s\n", backupPath)
				continue
			}
			
			if err := os.WriteFile(originalPath, content, info.Mode()); err != nil {
				fmt.Printf("ファイルの復元に失敗: %s\n", originalPath)
				continue
			}
			
			// バックアップファイルを削除
			if err := os.Remove(backupPath); err != nil {
				fmt.Printf("バックアップファイルの削除に失敗: %s\n", backupPath)
			}
			
			fmt.Printf("復元しました: %s\n", originalPath)
		}
	}
	
	return nil
}