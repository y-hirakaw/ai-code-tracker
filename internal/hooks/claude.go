package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ※ 型定義はgit.goのClaudeCodeHooksConfig, ClaudeCodeHook, Hookを使用

// SetupClaudeCodeHooksNew はClaude Code hooksを設定する（新実装）
func (m *HookManager) SetupClaudeCodeHooksNew() error {
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

	// AICT hooks設定を作成
	aictHooks := m.createAICTHooksConfig()

	// 既存設定にhooksを追加
	currentSettings["hooks"] = aictHooks

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

// createAICTHooksConfig はAICT用のhooks設定を作成
func (m *HookManager) createAICTHooksConfig() map[string][]ClaudeCodeHook {
	return map[string][]ClaudeCodeHook{
		"preToolUse": {
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
		"postToolUse": {
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
		"stop": {
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
		"notification": {
			{
				Matcher: "*",
				Hooks: []Hook{
					{
						Type:    "command",
						Command: "exit 0",
					},
				},
			},
		},
	}
}

// RemoveClaudeCodeHooksNew はClaude Code hooksを削除する（新実装）
func (m *HookManager) RemoveClaudeCodeHooksNew() error {
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

// GetClaudeHookStatus はClaude Code hooks の設定状況を取得する
func (m *HookManager) GetClaudeHookStatus() map[string]interface{} {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return map[string]interface{}{
			"installed": false,
			"error": "ホームディレクトリ取得エラー",
		}
	}

	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	
	claudeHookStatus := map[string]interface{}{
		"installed": false,
		"path": settingsPath,
		"backup": false,
	}

	// 設定ファイルの存在確認
	if data, err := os.ReadFile(settingsPath); err == nil {
		// JSON全体を文字列として確認（aictコマンドが含まれているか）
		if strings.Contains(string(data), "aict track") || strings.Contains(string(data), "aict ") {
			claudeHookStatus["installed"] = true
		}
	}

	// バックアップファイルの存在確認
	backupPath := settingsPath + ".backup"
	if _, err := os.Stat(backupPath); err == nil {
		claudeHookStatus["backup"] = true
	}

	return claudeHookStatus
}