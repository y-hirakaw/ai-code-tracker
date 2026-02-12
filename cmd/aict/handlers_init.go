package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// stdinReader is used to read user input (replaceable for testing)
var stdinReader = bufio.NewReader(os.Stdin)

// handleInitV2 handles SPEC.md準拠の新しい初期化処理
func handleInitV2() error {
	return handleInitV2WithOptions(false)
}

func handleInitV2WithOptions(withHooks bool) error {
	// .git/aict/ ディレクトリを作成
	store, err := storage.NewAIctStorage()
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	// デフォルト設定を作成
	gitUserName := getGitUserName()
	if gitUserName == "" {
		gitUserName = "Developer"
	}

	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions: []string{
			".go", ".py", ".js", ".ts", ".java",
			".cpp", ".c", ".h", ".rs", ".rb",
			".php", ".swift", ".kt", ".cs",
		},
		ExcludePatterns: []string{
			"*_test.go",
			"*_generated.go",
			"vendor/*",
			"node_modules/*",
			"*.min.js",
		},
		DefaultAuthor: gitUserName,
		AIAgents: []string{
			"Claude Code",
			"Claude",
			"GitHub Copilot",
			"ChatGPT",
			"Cursor",
		},
	}

	// 設定を保存
	if err := store.SaveConfig(config); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println("✓ AI Code Tracker initialized successfully!")
	fmt.Printf("✓ Configuration saved to .git/aict/config.json\n")
	fmt.Printf("✓ Default author: %s\n", config.DefaultAuthor)
	fmt.Printf("✓ Target AI percentage: %.0f%%\n", config.TargetAIPercentage)
	fmt.Println()

	// hooks設定の判定
	setupHooks := withHooks
	if !withHooks {
		fmt.Print("Set up hooks for automatic tracking? (Y/n): ")
		response, _ := stdinReader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		setupHooks = response == "" || response == "y" || response == "yes"
	}

	if setupHooks {
		fmt.Println()
		if err := handleSetupHooksV2(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: hook setup failed: %v\n", err)
			fmt.Println("You can set up hooks later with 'aict setup-hooks'")
		}
	} else {
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Run 'aict setup-hooks' to enable automatic tracking")
		fmt.Println("  2. Run 'aict checkpoint' to record code changes")
		fmt.Println("  3. Commit your changes with git")
		fmt.Println("  4. Run 'aict report --range <range>' to view statistics")
	}
	return nil
}
