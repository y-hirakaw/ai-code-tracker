package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/hooks"
	"github.com/y-hirakaw/ai-code-tracker/internal/utils"
)

// SetupHandler はsetupコマンドを処理する
type SetupHandler struct{}

// NewSetupHandler は新しいSetupHandlerを作成する
func NewSetupHandler() *SetupHandler {
	return &SetupHandler{}
}

// Handle はsetupコマンドを実行する
func (h *SetupHandler) Handle(args []string) error {
	var (
		gitHooksOnly    = false
		claudeHooksOnly = false
		removeHooks     = false
		showStatus      = false
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--git-hooks":
			gitHooksOnly = true
		case "--claude-hooks":
			claudeHooksOnly = true
		case "--remove":
			removeHooks = true
		case "--status":
			showStatus = true
		}
	}

	// 現在のディレクトリを取得
	currentDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	// HookManagerを初期化
	hookManager := hooks.NewHookManager(currentDir)

	// Gitリポジトリの検証
	if err := hookManager.ValidateGitRepo(); err != nil {
		return errors.WrapError(err, errors.ErrorTypeGit, "git_validation_failed")
	}

	// ステータス表示
	if showStatus {
		return h.showHookStatus(hookManager)
	}

	// hooks削除
	if removeHooks {
		return h.removeHooks(hookManager, gitHooksOnly, claudeHooksOnly)
	}

	// Claude Code hooks状態チェック（Claude hooksが設定対象の場合のみ）
	if !gitHooksOnly {
		if shouldShowGuide, message := h.checkClaudeHooksStatus(hookManager); shouldShowGuide {
			fmt.Println(message)
			fmt.Println("\nClaude Code hooksの設定を続行しますか？")
			fmt.Print("(y/N): ")
			
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("設定をキャンセルしました。")
				return nil
			}
		}
	}

	// 権限チェック
	if err := hookManager.CheckPermissions(); err != nil {
		return errors.WrapError(err, errors.ErrorTypeSecurity, "permission_check_failed")
	}

	// hooks設定
	return h.setupHooks(hookManager, gitHooksOnly, claudeHooksOnly)
}

// setupHooks はhooksを設定する
func (h *SetupHandler) setupHooks(hookManager *hooks.HookManager, gitOnly, claudeOnly bool) error {
	fmt.Println("=== AI Code Tracker Hooks 設定 ===")

	// 既存のhooksをバックアップ
	if err := hookManager.BackupExistingHooks(); err != nil {
		fmt.Printf("警告: 既存hooksのバックアップに失敗しました: %v\n", err)
	}

	// Git hooks設定
	if !claudeOnly {
		fmt.Println("📁 Git hooks を設定中...")
		if err := hookManager.SetupGitHooks(); err != nil {
			return errors.WrapError(err, errors.ErrorTypeConfig, "git_hooks_setup_failed")
		}
		fmt.Println("✅ Git hooks を設定しました")
	}

	// Claude Code hooks設定
	if !gitOnly {
		fmt.Println("\n🤖 Claude Code hooks を設定中...")
		if err := hookManager.SetupClaudeCodeHooks(); err != nil {
			return errors.WrapError(err, errors.ErrorTypeConfig, "claude_hooks_setup_failed")
		}
		fmt.Println("✅ Claude Code hooks を設定しました")
	}

	fmt.Println("\n🎉 Hooks設定が完了しました！")
	fmt.Println("\n次のステップ:")
	if !gitOnly {
		homeDir, _ := os.UserHomeDir()
		settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
		fmt.Printf("1. 設定ファイル: %s\n", settingsPath)
		fmt.Println("2. Claude Codeを再起動してhooksを有効化")
		
		// 旧ファイルが存在する場合は削除を推奨
		oldHooksPath := filepath.Join(homeDir, ".claude", "hooks-aict.json")
		if _, err := os.Stat(oldHooksPath); err == nil {
			fmt.Printf("\n💡 旧設定ファイルの削除を推奨: rm %s\n", oldHooksPath)
		}
	}
	if !claudeOnly {
		fmt.Println("3. Gitでコミットを行うと自動的にトラッキングが開始されます")
	}

	return nil
}

// removeHooks はhooksを削除する
func (h *SetupHandler) removeHooks(hookManager *hooks.HookManager, gitOnly, claudeOnly bool) error {
	fmt.Println("=== AI Code Tracker Hooks 削除 ===")

	// Git hooks削除
	if !claudeOnly {
		fmt.Println("📁 Git hooks を削除中...")
		if err := hookManager.RemoveGitHooks(); err != nil {
			return errors.WrapError(err, errors.ErrorTypeConfig, "git_hooks_removal_failed")
		}
		fmt.Println("✅ Git hooks を削除しました")
	}

	// Claude Code hooks削除
	if !gitOnly {
		fmt.Println("\n🤖 Claude Code hooks を削除中...")
		if err := hookManager.RemoveClaudeCodeHooks(); err != nil {
			return errors.WrapError(err, errors.ErrorTypeConfig, "claude_hooks_removal_failed")
		}
		fmt.Println("✅ Claude Code hooks を削除しました")
	}

	fmt.Println("\n🎉 Hooks削除が完了しました！")
	return nil
}

// checkClaudeHooksStatus はClaude Code hooksの設定状況をチェックし、ガイダンスが必要かどうかを判定する
func (h *SetupHandler) checkClaudeHooksStatus(hookManager *hooks.HookManager) (bool, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return true, "🤖 Claude Code hooks設定状況の確認中..."
	}

	// 新しい設定ファイル（~/.claude/settings.json）をチェック
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	oldHooksPath := filepath.Join(homeDir, ".claude", "hooks-aict.json")
	
	var hasNewSettings, hasOldSettings, hasValidAICTHooks bool
	
	// settings.jsonの確認
	if data, err := os.ReadFile(settingsPath); err == nil {
		hasNewSettings = true
		var settings map[string]interface{}
		if json.Unmarshal(data, &settings) == nil {
			// JSON全体を文字列として確認（aictコマンドが含まれているか）
			if strings.Contains(string(data), "aict track") || strings.Contains(string(data), "aict ") {
				hasValidAICTHooks = true
			}
		}
	}

	// 古いhooks-aict.jsonの確認
	if _, err := os.Stat(oldHooksPath); err == nil {
		hasOldSettings = true
	}

	// 状況に応じたメッセージを生成
	var message string
	needsGuidance := false

	if hasValidAICTHooks {
		message = "✅ ~/.claude/settings.json に AICT hooks が既に設定されています。\n" +
			"🔄 設定を上書きする場合は続行してください。"
		needsGuidance = true
	} else if hasNewSettings && hasOldSettings {
		message = "⚠️  Claude Code hooks設定が重複しています:\n" +
			"   • ~/.claude/settings.json (新形式) - AICT hooks未設定\n" +
			"   • ~/.claude/hooks-aict.json (旧形式) - 存在\n" +
			"\n💡 新形式の ~/.claude/settings.json にAICT hooksを設定します。\n" +
			"   旧ファイルは手動で削除することを推奨します。"
		needsGuidance = true
	} else if hasOldSettings && !hasNewSettings {
		message = "⚠️  旧形式のClaude Code hooks設定が検出されました:\n" +
			"   • ~/.claude/hooks-aict.json (旧形式)\n" +
			"\n💡 新形式の ~/.claude/settings.json を作成してAICT hooksを設定します。\n" +
			"   設定後、旧ファイルは手動で削除することを推奨します。"
		needsGuidance = true
	} else if hasNewSettings && !hasValidAICTHooks {
		message = "📋 ~/.claude/settings.json が存在しますが、AICT hooksは未設定です。\n" +
			"🔧 AICT hooksを追加します。"
		needsGuidance = false // 通常の設定として進行
	} else {
		message = "📋 ~/.claude/settings.json が存在しません。\n" +
			"🆕 新規作成してAICT hooksを設定します。"
		needsGuidance = false // 通常の設定として進行
	}

	return needsGuidance, message
}

// showHookStatus はhooksの設定状況を表示する
func (h *SetupHandler) showHookStatus(hookManager *hooks.HookManager) error {
	fmt.Println("=== AI Code Tracker Hooks 設定状況 ===")

	status, err := hookManager.GetHookStatus()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeConfig, "hook_status_fetch_failed")
	}

	// Git hooks状況
	if gitHooks, ok := status["git_hooks"].(map[string]interface{}); ok {
		fmt.Println("📁 Git Hooks:")
		if installed, ok := gitHooks["installed"].(bool); ok && installed {
			fmt.Println("  ✅ インストール済み")
		} else {
			fmt.Println("  ❌ 未インストール")
		}

		if path, ok := gitHooks["path"].(string); ok {
			fmt.Printf("  📂 パス: %s\n", path)
		}

		if executable, ok := gitHooks["executable"].(bool); ok {
			if executable {
				fmt.Println("  ✅ 実行可能")
			} else {
				fmt.Println("  ❌ 実行権限なし")
			}
		}

		if backup, ok := gitHooks["backup"].(bool); ok && backup {
			fmt.Println("  💾 バックアップあり")
		}
	}

	fmt.Println()

	// Claude Code hooks状況
	if claudeHooks, ok := status["claude_hooks"].(map[string]interface{}); ok {
		fmt.Println("🤖 Claude Code Hooks:")
		if installed, ok := claudeHooks["installed"].(bool); ok && installed {
			fmt.Println("  ✅ インストール済み")
		} else {
			fmt.Println("  ❌ 未インストール")
		}

		if path, ok := claudeHooks["path"].(string); ok {
			fmt.Printf("  📂 パス: %s\n", path)
		}

		if backup, ok := claudeHooks["backup"].(bool); ok && backup {
			fmt.Println("  💾 バックアップあり")
		}
		
		// 旧形式のファイル存在チェック
		homeDir, _ := os.UserHomeDir()
		oldHooksPath := filepath.Join(homeDir, ".claude", "hooks-aict.json")
		if _, err := os.Stat(oldHooksPath); err == nil {
			fmt.Printf("  ⚠️  旧設定ファイルが存在: %s\n", oldHooksPath)
			fmt.Println("     削除を推奨します（新形式のsettings.jsonを使用）")
		}
	}

	return nil
}