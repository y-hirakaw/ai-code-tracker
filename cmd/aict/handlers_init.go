package main

import (
	"fmt"
	"os"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// handleInitV2 handles SPEC.md準拠の新しい初期化処理
func handleInitV2() {
	// .git/aict/ ディレクトリを作成
	store, err := storage.NewAIctStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
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
			"GitHub Copilot",
			"ChatGPT",
		},
	}

	// 設定を保存
	if err := store.SaveConfig(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ AI Code Tracker initialized successfully!")
	fmt.Printf("✓ Configuration saved to .git/aict/config.json\n")
	fmt.Printf("✓ Default author: %s\n", config.DefaultAuthor)
	fmt.Printf("✓ Target AI percentage: %.0f%%\n", config.TargetAIPercentage)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'aict checkpoint' to record code changes")
	fmt.Println("  2. Commit your changes with git")
	fmt.Println("  3. Run 'aict commit' to generate Authorship Log")
	fmt.Println("  4. Use 'aict report --range <range>' to view statistics")
}
