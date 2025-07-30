package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/y-hirakawa/ai-code-tracker/internal/git"
	"github.com/y-hirakawa/ai-code-tracker/internal/storage"
	"github.com/y-hirakawa/ai-code-tracker/internal/templates"
	"github.com/y-hirakawa/ai-code-tracker/internal/tracker"
)

const (
	defaultBaseDir = ".ai_code_tracking"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		handleInit()
	case "track":
		handleTrack()
	case "report":
		handleReport()
	case "setup-hooks":
		handleSetupHooks()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleInit() {
	baseDir := defaultBaseDir
	
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Error creating tracking directory: %v\n", err)
		os.Exit(1)
	}

	metricsStorage := storage.NewMetricsStorage(baseDir)
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs"},
		ExcludePatterns:    []string{"*_test.go", "*.test.js", "*.spec.ts", "*_generated.go"},
		AuthorMappings:     make(map[string]string),
	}

	if err := metricsStorage.SaveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	gitAnalyzer := git.NewDiffAnalyzer()
	if gitAnalyzer.IsGitRepository() {
		userName := getGitUserName()
		if userName != "" {
			config.AuthorMappings[userName] = "human"
			metricsStorage.SaveConfig(config)
		}
	}

	// Create hook scripts only
	if err := createHookFiles(baseDir); err != nil {
		fmt.Printf("Warning: Could not create hook files: %v\n", err)
	} else {
		fmt.Println("✓ Hook scripts created in .ai_code_tracking/hooks/")
	}

	fmt.Println("AI Code Tracker initialized successfully!")
	fmt.Printf("Configuration saved to %s/config.json\n", baseDir)
	fmt.Println("Run 'aict setup-hooks' to enable automatic tracking with Claude Code and Git.")
}

func handleTrack() {
	fs := flag.NewFlagSet("track", flag.ExitOnError)
	author := fs.String("author", "", "Author of the checkpoint (required)")
	fs.Parse(os.Args[2:])

	if *author == "" {
		fmt.Println("Error: -author flag is required")
		fmt.Println("Usage: aict track -author <author_name>")
		os.Exit(1)
	}

	baseDir := defaultBaseDir
	metricsStorage := storage.NewMetricsStorage(baseDir)
	
	config, err := metricsStorage.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	checkpointMgr := tracker.NewCheckpointManager(baseDir)
	checkpoint, err := checkpointMgr.CreateCheckpoint(*author, config.TrackedExtensions)
	if err != nil {
		fmt.Printf("Error creating checkpoint: %v\n", err)
		os.Exit(1)
	}

	if err := checkpointMgr.SaveCheckpoint(checkpoint); err != nil {
		fmt.Printf("Error saving checkpoint: %v\n", err)
		os.Exit(1)
	}

	// Load current metrics
	currentMetrics, err := metricsStorage.LoadMetrics()
	if err != nil {
		fmt.Printf("Error loading metrics: %v\n", err)
		os.Exit(1)
	}

	// Get all checkpoints to find the previous one
	allCheckpoints, err := checkpointMgr.GetLatestCheckpoints("*", 100)
	if err != nil {
		fmt.Printf("Error getting checkpoints: %v\n", err)
		os.Exit(1)
	}

	// Find the last checkpoint before this one
	var previousCheckpoint *tracker.Checkpoint
	for i := len(allCheckpoints) - 1; i >= 0; i-- {
		if allCheckpoints[i].ID != checkpoint.ID {
			previousCheckpoint = allCheckpoints[i]
			break
		}
	}

	if previousCheckpoint == nil {
		// First checkpoint - all lines are from this author
		if analyzer := tracker.NewAnalyzer(config); analyzer.IsAIAuthor(*author) {
			currentMetrics.AILines = countTotalLines(checkpoint)
			currentMetrics.TotalLines = currentMetrics.AILines
		} else {
			currentMetrics.HumanLines = countTotalLines(checkpoint)
			currentMetrics.TotalLines = currentMetrics.HumanLines
		}
		currentMetrics.LastUpdated = checkpoint.Timestamp
	} else {
		// Analyze diff between previous and current checkpoint
		analyzer := tracker.NewAnalyzer(config)
		result, err := analyzer.AnalyzeCheckpoints(previousCheckpoint, checkpoint)
		if err != nil {
			fmt.Printf("Error analyzing checkpoints: %v\n", err)
			os.Exit(1)
		}

		// Add to current metrics
		currentMetrics.AILines += result.AILines
		currentMetrics.HumanLines += result.HumanLines
		// Recalculate total lines from the latest checkpoint
		currentMetrics.TotalLines = countTotalLines(checkpoint)
		// Ensure human lines = total - AI lines
		currentMetrics.HumanLines = currentMetrics.TotalLines - currentMetrics.AILines
		currentMetrics.LastUpdated = result.LastUpdated
	}

	if currentMetrics.TotalLines > 0 {
		currentMetrics.Percentage = float64(currentMetrics.AILines) / float64(currentMetrics.TotalLines) * 100
	}

	if err := metricsStorage.SaveMetrics(currentMetrics); err != nil {
		fmt.Printf("Error saving metrics: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Checkpoint saved and metrics updated!")
	analyzer := tracker.NewAnalyzer(config)
	fmt.Println(analyzer.GenerateReport(currentMetrics))
}

func handleReport() {
	baseDir := defaultBaseDir
	metricsStorage := storage.NewMetricsStorage(baseDir)
	
	config, err := metricsStorage.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	metrics, err := metricsStorage.LoadMetrics()
	if err != nil {
		fmt.Printf("Error loading metrics: %v\n", err)
		os.Exit(1)
	}

	analyzer := tracker.NewAnalyzer(config)
	fmt.Println(analyzer.GenerateReport(metrics))
}

func countTotalLines(checkpoint *tracker.Checkpoint) int {
	total := 0
	for _, file := range checkpoint.Files {
		total += len(file.Lines)
	}
	return total
}

func handleSetupHooks() {
	fmt.Println("Setting up AI Code Tracker hooks...")
	
	// Setup Git post-commit hook
	if err := setupGitHook(); err != nil {
		fmt.Printf("Error setting up Git post-commit hook: %v\n", err)
		os.Exit(1)
	}
	
	// Setup Claude Code hooks
	if err := setupClaudeHooks(); err != nil {
		fmt.Printf("Error setting up Claude Code hooks: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println()
	fmt.Println("✓ Hook setup complete! Claude Code will now automatically track AI vs Human contributions.")
}

func createHookFiles(baseDir string) error {
	hooksDir := filepath.Join(baseDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return err
	}

	// Create PreToolUse hook
	preHookPath := filepath.Join(hooksDir, "pre-tool-use.sh")
	if err := os.WriteFile(preHookPath, []byte(templates.PreToolUseHook), 0755); err != nil {
		return err
	}

	// Create PostToolUse hook
	postHookPath := filepath.Join(hooksDir, "post-tool-use.sh")
	if err := os.WriteFile(postHookPath, []byte(templates.PostToolUseHook), 0755); err != nil {
		return err
	}

	// Create Post-commit hook
	commitHookPath := filepath.Join(hooksDir, "post-commit")
	if err := os.WriteFile(commitHookPath, []byte(templates.PostCommitHook), 0755); err != nil {
		return err
	}

	return nil
}

func setupGitHook() error {
	// Copy Git post-commit hook from .ai_code_tracking/hooks/
	hookSource := filepath.Join(defaultBaseDir, "hooks", "post-commit")
	hookDest := ".git/hooks/post-commit"
	
	// Check if Git post-commit hook already exists
	if _, err := os.Stat(hookDest); err == nil {
		fmt.Printf("Warning: Git post-commit hook already exists at %s\n", hookDest)
		fmt.Println("Please manually integrate the AI Code Tracker hook or backup existing hook.")
		return fmt.Errorf("existing Git hook found")
	}
	
	if err := copyFile(hookSource, hookDest); err != nil {
		fmt.Println("Make sure to run 'aict init' first to create hook files.")
		return err
	}
	
	// Make it executable
	if err := os.Chmod(hookDest, 0755); err != nil {
		fmt.Printf("Warning: Could not make post-commit hook executable: %v\n", err)
	} else {
		fmt.Println("✓ Git post-commit hook installed")
	}
	
	return nil
}

func setupClaudeHooks() error {
	claudeDir := ".claude"
	settingsPath := filepath.Join(claudeDir, "settings.json")
	
	// Check if Claude settings already exist
	if _, err := os.Stat(settingsPath); err == nil {
		fmt.Printf("Warning: Claude settings already exist at %s\n", settingsPath)
		fmt.Println("Please manually add AI Code Tracker hooks to your existing settings.")
		fmt.Println("Add the following hooks to your .claude/settings.json:")
		fmt.Println(templates.ClaudeSettingsJSON)
		return nil
	}
	
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	
	if err := os.WriteFile(settingsPath, []byte(templates.ClaudeSettingsJSON), 0644); err != nil {
		return err
	}
	
	fmt.Println("✓ Claude Code hook configuration created")
	fmt.Println("✓ Hook scripts are available in .ai_code_tracking/hooks/")
	
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

func printUsage() {
	fmt.Println("AI Code Tracker (aict) - Track AI vs Human code contributions")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aict init                    Initialize tracking in current directory")
	fmt.Println("  aict track -author <name>    Create a checkpoint for the specified author")
	fmt.Println("  aict report                  Show current tracking metrics")
	fmt.Println("  aict setup-hooks             Setup Claude Code and Git hooks for automatic tracking")
}

func getGitUserName() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}