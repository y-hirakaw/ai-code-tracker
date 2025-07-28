package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/hooks"
	"github.com/ai-code-tracker/aict/internal/utils"
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
		hooksPath := filepath.Join(homeDir, ".claude", "hooks-aict.json")
		fmt.Printf("1. 環境変数を設定: export CLAUDE_HOOKS_CONFIG=%s\n", hooksPath)
		fmt.Println("2. Claude Codeを再起動してhooksを有効化")
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

		if envVarSet, ok := claudeHooks["env_var_set"].(bool); ok {
			if envVarSet {
				fmt.Println("  ✅ 環境変数設定済み")
			} else {
				fmt.Println("  ❌ 環境変数未設定")
				if path, ok := claudeHooks["path"].(string); ok {
					fmt.Printf("  💡 実行してください: export CLAUDE_HOOKS_CONFIG=%s\n", path)
				}
			}
		}

		if backup, ok := claudeHooks["backup"].(bool); ok && backup {
			fmt.Println("  💾 バックアップあり")
		}
	}

	return nil
}