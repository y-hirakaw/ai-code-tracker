package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/y-hirakawa/ai-code-tracker/internal/git"
	"github.com/y-hirakawa/ai-code-tracker/internal/storage"
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

	fmt.Println("AI Code Tracker initialized successfully!")
	fmt.Printf("Configuration saved to %s/config.json\n", baseDir)
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
	fmt.Println("Setting up Claude Code hooks and Git post-commit hook...")
	
	// Copy Git post-commit hook
	if err := copyFile("hooks/post-commit", ".git/hooks/post-commit"); err != nil {
		fmt.Printf("Warning: Could not setup Git post-commit hook: %v\n", err)
	} else {
		fmt.Println("✓ Git post-commit hook installed")
	}
	
	// Make it executable
	if err := os.Chmod(".git/hooks/post-commit", 0755); err != nil {
		fmt.Printf("Warning: Could not make post-commit hook executable: %v\n", err)
	}
	
	fmt.Println("✓ Claude Code hooks are configured in .claude/settings.json")
	fmt.Println("✓ Hook scripts are available in hooks/")
	fmt.Println()
	fmt.Println("Hook setup complete! Claude Code will now automatically track AI vs Human contributions.")
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