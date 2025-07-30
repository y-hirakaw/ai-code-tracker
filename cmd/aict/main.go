package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	case "reset":
		if err := handleReset(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
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

	// Create baseline checkpoint from current codebase
	checkpointMgr := tracker.NewCheckpointManager(baseDir)
	baselineCheckpoint, err := checkpointMgr.CreateCheckpoint("baseline", config.TrackedExtensions)
	if err != nil {
		fmt.Printf("Error creating baseline checkpoint: %v\n", err)
		os.Exit(1)
	}
	
	if err := checkpointMgr.SaveCheckpoint(baselineCheckpoint); err != nil {
		fmt.Printf("Error saving baseline checkpoint: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize metrics with baseline
	baselineLines := countTotalLines(baselineCheckpoint)
	initialMetrics := &tracker.AnalysisResult{
		TotalLines:    baselineLines,
		BaselineLines: baselineLines,
		AILines:       0,
		HumanLines:    0,
		Percentage:    0.0,
		LastUpdated:   time.Now(),
	}
	
	if err := metricsStorage.SaveMetrics(initialMetrics); err != nil {
		fmt.Printf("Error initializing metrics: %v\n", err)
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
	fmt.Printf("✓ Baseline checkpoint created (%d lines)\n", baselineLines)
	fmt.Println("✓ Metrics initialized for tracking changes from baseline")
	fmt.Println()
	fmt.Println("Next step:")
	fmt.Println("Run 'aict setup-hooks' to enable automatic tracking with Claude Code and Git")
	fmt.Println()
	fmt.Println("From now on, only code changes from this baseline will be tracked.")
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

	// Find baseline checkpoint and previous checkpoint
	var baselineCheckpoint *tracker.Checkpoint
	var previousCheckpoint *tracker.Checkpoint
	
	for i := len(allCheckpoints) - 1; i >= 0; i-- {
		if allCheckpoints[i].ID != checkpoint.ID {
			if allCheckpoints[i].Author == "baseline" {
				baselineCheckpoint = allCheckpoints[i]
			} else if previousCheckpoint == nil {
				previousCheckpoint = allCheckpoints[i]
			}
		}
	}

	if baselineCheckpoint == nil {
		fmt.Println("Error: No baseline checkpoint found. Please run 'aict init' first.")
		os.Exit(1)
	}

	if previousCheckpoint == nil {
		// First tracking after baseline - compare against baseline
		previousCheckpoint = baselineCheckpoint
	}

	// Analyze diff between previous and current checkpoint
	analyzer := tracker.NewAnalyzer(config)
	result, err := analyzer.AnalyzeCheckpoints(previousCheckpoint, checkpoint)
	if err != nil {
		fmt.Printf("Error analyzing checkpoints: %v\n", err)
		os.Exit(1)
	}

	// Add changes to current metrics
	currentMetrics.AILines += result.AILines
	currentMetrics.HumanLines += result.HumanLines
	
	// Update total lines (baseline + changes)
	currentMetrics.TotalLines = countTotalLines(checkpoint)
	currentMetrics.LastUpdated = result.LastUpdated

	// Calculate percentage based on added lines only (excluding baseline)
	addedLines := currentMetrics.AILines + currentMetrics.HumanLines
	if addedLines > 0 {
		currentMetrics.Percentage = float64(currentMetrics.AILines) / float64(addedLines) * 100
	} else {
		currentMetrics.Percentage = 0.0
	}

	if err := metricsStorage.SaveMetrics(currentMetrics); err != nil {
		fmt.Printf("Error saving metrics: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Checkpoint saved and metrics updated!")
	analyzer2 := tracker.NewAnalyzer(config)
	fmt.Println(analyzer2.GenerateReport(currentMetrics))
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
		fmt.Print("Do you want to merge AI Code Tracker functionality? (y/N): ")
		
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response != "y" && response != "yes" {
			fmt.Println("Git hook setup cancelled. Please manually integrate the AI Code Tracker hook.")
			return fmt.Errorf("user cancelled Git hook setup")
		}
		
		// Merge with existing hook
		if err := mergeGitHook(hookSource, hookDest); err != nil {
			return err
		}
		fmt.Println("✓ Git post-commit hook merged with existing hook")
	} else {
		// No existing hook, just copy
		if err := copyFile(hookSource, hookDest); err != nil {
			fmt.Println("Make sure to run 'aict init' first to create hook files.")
			return err
		}
		fmt.Println("✓ Git post-commit hook installed")
	}
	
	// Make it executable
	if err := os.Chmod(hookDest, 0755); err != nil {
		fmt.Printf("Warning: Could not make post-commit hook executable: %v\n", err)
	}
	
	return nil
}

func setupClaudeHooks() error {
	claudeDir := ".claude"
	settingsPath := filepath.Join(claudeDir, "settings.json")
	
	// Check if Claude settings already exist
	if _, err := os.Stat(settingsPath); err == nil {
		fmt.Printf("Warning: Claude settings already exist at %s\n", settingsPath)
		fmt.Print("Do you want to merge AI Code Tracker hooks? (y/N): ")
		
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		if response != "y" && response != "yes" {
			fmt.Println("Claude hook setup cancelled. Please manually add the following hooks:")
			fmt.Println(templates.ClaudeSettingsJSON)
			return nil
		}
		
		// Merge with existing settings
		if err := mergeClaudeSettings(settingsPath); err != nil {
			return err
		}
		fmt.Println("✓ Claude Code hooks merged with existing settings")
	} else {
		// No existing settings, create new
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			return err
		}
		
		if err := os.WriteFile(settingsPath, []byte(templates.ClaudeSettingsJSON), 0644); err != nil {
			return err
		}
		fmt.Println("✓ Claude Code hook configuration created")
	}
	
	fmt.Println("✓ Hook scripts are available in .ai_code_tracking/hooks/")
	return nil
}

func mergeGitHook(hookSource, hookDest string) error {
	// Read existing hook
	existingContent, err := os.ReadFile(hookDest)
	if err != nil {
		return err
	}
	
	// Read AI Code Tracker hook
	aictContent, err := os.ReadFile(hookSource)
	if err != nil {
		return err
	}
	
	// Create merged content
	mergedContent := string(existingContent) + "\n\n# AI Code Tracker\n" + string(aictContent)
	
	// Write merged hook
	return os.WriteFile(hookDest, []byte(mergedContent), 0755)
}

func mergeClaudeSettings(settingsPath string) error {
	// Read existing settings
	existingContent, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}
	
	var existingSettings map[string]interface{}
	if err := json.Unmarshal(existingContent, &existingSettings); err != nil {
		return fmt.Errorf("failed to parse existing settings: %v", err)
	}
	
	// Parse AI Code Tracker settings
	var aictSettings map[string]interface{}
	if err := json.Unmarshal([]byte(templates.ClaudeSettingsJSON), &aictSettings); err != nil {
		return fmt.Errorf("failed to parse AICT settings: %v", err)
	}
	
	// Merge hooks
	existingHooks, hasHooks := existingSettings["hooks"].([]interface{})
	if !hasHooks {
		existingHooks = []interface{}{}
	}
	
	aictHooks := aictSettings["hooks"].([]interface{})
	mergedHooks := append(existingHooks, aictHooks...)
	existingSettings["hooks"] = mergedHooks
	
	// Write merged settings
	mergedContent, err := json.MarshalIndent(existingSettings, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(settingsPath, mergedContent, 0644)
}

func handleReset() error {
	baseDir := defaultBaseDir
	
	fmt.Print("This will reset all tracking metrics to zero and set current codebase as baseline. Continue? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("Reset cancelled.")
		return nil
	}
	
	metricsStorage := storage.NewMetricsStorage(baseDir)
	
	// Reset metrics to zero
	resetMetrics := &tracker.AnalysisResult{
		TotalLines:    0,
		BaselineLines: 0,
		AILines:       0,
		HumanLines:    0,
		Percentage:    0.0,
		LastUpdated:   time.Now(),
	}
	
	if err := metricsStorage.SaveMetrics(resetMetrics); err != nil {
		return fmt.Errorf("error resetting metrics: %v", err)
	}
	
	// Clear all checkpoints
	checkpointsDir := filepath.Join(baseDir, "checkpoints")
	if err := os.RemoveAll(checkpointsDir); err != nil {
		return fmt.Errorf("error clearing checkpoints: %v", err)
	}
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return fmt.Errorf("error recreating checkpoints directory: %v", err)
	}
	
	fmt.Println("✓ Metrics reset to zero")
	fmt.Println("✓ All checkpoints cleared")
	fmt.Println()
	fmt.Println("AI Code Tracker has been reset.")
	fmt.Println("Next step: Run 'aict init' to create a new baseline from current codebase.")
	
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
	fmt.Println("  aict reset                   Reset metrics to start tracking from current codebase state")
}

func getGitUserName() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}