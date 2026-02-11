package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/templates"
)

// handleSetupHooksV2 handles SPEC.md準拠のhookセットアップ
func handleSetupHooksV2() error {
	fmt.Println("Setting up AI Code Tracker hooks (SPEC.md)...")

	// Gitリポジトリのルートディレクトリを取得
	executor := newExecutor()
	repoRoot, err := executor.Run("rev-parse", "--show-toplevel")
	if err != nil {
		return fmt.Errorf("failed to get repository root (are you in a git repo?): %w", err)
	}

	// .git ディレクトリの絶対パスを決定
	// 注: 将来的には git rev-parse --git-dir を使うのがより堅牢だが、
	// 現状の構成では .git がルート直下にあると仮定
	gitDir := filepath.Join(repoRoot, ".git")

	// .git/aict/hooks/ ディレクトリを作成
	aictHooksDir := filepath.Join(gitDir, "aict", "hooks")
	if err := os.MkdirAll(aictHooksDir, 0755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	// Claude Code hooksを作成
	if err := createClaudeHooks(aictHooksDir); err != nil {
		return fmt.Errorf("creating Claude Code hooks: %w", err)
	}

	// Git post-commit hookを作成
	if err := setupPostCommitHook(repoRoot); err != nil {
		return fmt.Errorf("setting up post-commit hook: %w", err)
	}

	// .claude/settings.json を更新
	if err := setupClaudeSettings(repoRoot); err != nil {
		return fmt.Errorf("setting up Claude Code settings: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Hook setup complete!")
	fmt.Println()
	fmt.Println("Hooks created:")
	fmt.Printf("  - %s/pre-tool-use.sh  (records human checkpoint)\n", aictHooksDir)
	fmt.Printf("  - %s/post-tool-use.sh (records AI checkpoint)\n", aictHooksDir)
	fmt.Printf("  - %s/hooks/post-commit           (generates Authorship Log)\n", gitDir)
	fmt.Println()
	fmt.Println("Claude Code will now automatically track AI vs Human contributions.")
	return nil
}

func createClaudeHooks(hooksDir string) error {
	// pre-tool-use.sh
	preHookPath := filepath.Join(hooksDir, "pre-tool-use.sh")
	if err := os.WriteFile(preHookPath, []byte(templates.PreToolUseHook), 0755); err != nil {
		return fmt.Errorf("failed to create pre-tool-use.sh: %w", err)
	}

	// post-tool-use.sh
	postHookPath := filepath.Join(hooksDir, "post-tool-use.sh")
	if err := os.WriteFile(postHookPath, []byte(templates.PostToolUseHook), 0755); err != nil {
		return fmt.Errorf("failed to create post-tool-use.sh: %w", err)
	}

	fmt.Println("✓ Claude Code hooks created")
	return nil
}

func setupPostCommitHook(repoRoot string) error {
	// post-commit hookを.git/hooks/にコピー
	gitHooksDir := filepath.Join(repoRoot, ".git", "hooks")
	gitHookPath := filepath.Join(gitHooksDir, "post-commit")

	// .git/hooks/ディレクトリがなければ作成
	if err := os.MkdirAll(gitHooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create .git/hooks directory: %w", err)
	}

	// 既存のpost-commit hookをチェック
	if _, err := os.Stat(gitHookPath); err == nil {
		fmt.Printf("Warning: Git post-commit hook already exists at %s\n", gitHookPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Post-commit hook setup cancelled.")
			fmt.Println("Please manually add the following to your post-commit hook:")
			fmt.Println("  aict commit")
			return nil
		}
	}

	// post-commit hookを作成
	if err := os.WriteFile(gitHookPath, []byte(templates.PostCommitHook), 0755); err != nil {
		return fmt.Errorf("failed to create post-commit hook: %w", err)
	}

	fmt.Println("✓ Git post-commit hook installed")
	return nil
}

func setupClaudeSettings(repoRoot string) error {
	settingsDir := filepath.Join(repoRoot, ".claude")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	settingsPath := filepath.Join(settingsDir, "settings.json")

	// 既存のsettings.jsonをチェック
	if _, err := os.Stat(settingsPath); err == nil {
		fmt.Printf("Warning: Claude Code settings already exist at %s\n", settingsPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Claude Code settings setup cancelled.")
			fmt.Println("Please manually add hook configuration to .claude/settings.json")
			return nil
		}
	}

	// settings.jsonを作成
	if err := os.WriteFile(settingsPath, []byte(templates.ClaudeSettingsJSON), 0644); err != nil {
		return fmt.Errorf("failed to create settings.json: %w", err)
	}

	fmt.Println("✓ Claude Code settings configured")
	return nil
}
