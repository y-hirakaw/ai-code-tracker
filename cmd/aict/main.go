package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/git"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/templates"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

const version = "0.6.1"

var defaultBaseDir = ".ai_code_tracking"

// exitFunc is used to mock os.Exit in tests
var exitFunc = os.Exit

func main() {
	if len(os.Args) < 2 {
		printUsage()
		exitFunc(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		handleInit()
	case "track":
		handleTrack()
	case "report":
		handleReportWithOptions()
	case "setup-hooks":
		handleSetupHooks()
	case "reset":
		if err := handleReset(); err != nil {
			fmt.Printf("Error: %v\n", err)
			exitFunc(1)
		}
	case "version", "--version", "-v":
		fmt.Printf("AI Code Tracker (aict) version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	case "config":
		handleConfig()
	case "mark-ai-edit":
		handleMarkAIEdit()
	case "snapshot":
		handleSnapshot()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		exitFunc(1)
	}
}

func handleInit() {
	baseDir := defaultBaseDir

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Error creating tracking directory: %v\n", err)
		exitFunc(1)
	}

	metricsStorage := storage.NewMetricsStorage(baseDir)
	config := storage.GetDefaultConfig()

	if err := metricsStorage.SaveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		exitFunc(1)
	}

	// Initialize metrics
	initialMetrics := &tracker.AnalysisResult{
		TotalLines:  0,
		AILines:     0,
		HumanLines:  0,
		Percentage:  0.0,
		LastUpdated: time.Now(),
	}

	if err := metricsStorage.SaveMetrics(initialMetrics); err != nil {
		fmt.Printf("Error initializing metrics: %v\n", err)
		exitFunc(1)
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
	fmt.Println("✓ Metrics initialized for tracking")
	fmt.Println()
	fmt.Println("Next step:")
	fmt.Println("Run 'aict setup-hooks' to enable automatic tracking with Claude Code and Git")
	fmt.Println()
	fmt.Println("Now you can start tracking AI vs Human code contributions.")
}

func handleTrack() {
	fs := flag.NewFlagSet("track", flag.ExitOnError)
	author := fs.String("author", "", "Author of the checkpoint (required)")
	fs.Parse(os.Args[2:])

	if *author == "" {
		fmt.Println("Error: -author flag is required")
		fmt.Println("Usage: aict track -author <author_name>")
		exitFunc(1)
	}

	baseDir := defaultBaseDir

	// Check if initialized
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Printf("Error: AI Code Tracker not initialized. Run 'aict init' first.\n")
		exitFunc(1)
	}

	// Record checkpoint using new JSONL format
	recorder := tracker.NewCheckpointRecorder(baseDir)
	if err := recorder.RecordCheckpoint(*author); err != nil {
		fmt.Printf("Error recording checkpoint: %v\n", err)
		exitFunc(1)
	}

	// Update metrics
	if err := updateMetricsFromRecords(baseDir); err != nil {
		fmt.Printf("Error updating metrics: %v\n", err)
		exitFunc(1)
	}

	fmt.Printf("✓ Checkpoint recorded for author: %s\n", *author)
}

// updateMetricsFromRecords updates metrics based on JSONL records
func updateMetricsFromRecords(baseDir string) error {
	// Load configuration
	metricsStorage := storage.NewMetricsStorage(baseDir)
	config, err := metricsStorage.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Read all records
	recorder := tracker.NewCheckpointRecorder(baseDir)
	records, err := recorder.ReadAllRecords()
	if err != nil {
		return fmt.Errorf("error reading records: %w", err)
	}

	// Analyze records
	analyzer := tracker.NewAnalyzer(config)
	result, err := analyzer.AnalyzeRecords(records)
	if err != nil {
		return fmt.Errorf("error analyzing records: %w", err)
	}

	// Update total lines
	result.TotalLines = result.AILines + result.HumanLines

	// Save updated metrics
	return metricsStorage.SaveMetrics(result)
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
		exitFunc(1)
	}

	// Setup Claude Code hooks
	if err := setupClaudeHooks(); err != nil {
		fmt.Printf("Error setting up Claude Code hooks: %v\n", err)
		exitFunc(1)
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

	// Create Pre-commit hook
	preCommitHookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(preCommitHookPath, []byte(templates.PreCommitHook), 0755); err != nil {
		return err
	}
	// Create Post-commit hook
	postCommitHookPath := filepath.Join(hooksDir, "post-commit")
	if err := os.WriteFile(postCommitHookPath, []byte(templates.PostCommitHook), 0755); err != nil {
		return err
	}

	return nil
}

func setupGitHook() error {
	// Setup pre-commit hook
	preCommitSource := filepath.Join(defaultBaseDir, "hooks", "pre-commit")
	preCommitDest := ".git/hooks/pre-commit"

	if err := setupSingleGitHook(preCommitSource, preCommitDest, "pre-commit"); err != nil {
		return err
	}

	// Setup post-commit hook
	postCommitSource := filepath.Join(defaultBaseDir, "hooks", "post-commit")
	postCommitDest := ".git/hooks/post-commit"

	if err := setupSingleGitHook(postCommitSource, postCommitDest, "post-commit"); err != nil {
		return err
	}

	return nil
}

func setupSingleGitHook(hookSource, hookDest, hookName string) error {
	// Check if Git hook already exists
	if _, err := os.Stat(hookDest); err == nil {
		fmt.Printf("Warning: Git %s hook already exists at %s\n", hookName, hookDest)
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

	// Merge hooks - handle both object and array formats
	aictHooks, ok := aictSettings["hooks"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("AICT settings hooks format is invalid")
	}

	// Check if existing hooks is an object or array
	if existingHooksObj, isObj := existingSettings["hooks"].(map[string]interface{}); isObj {
		// Merge object-style hooks
		for key, value := range aictHooks {
			existingHooksObj[key] = value
		}
	} else {
		// If existing hooks is array or doesn't exist, replace with AICT hooks
		existingSettings["hooks"] = aictHooks
	}

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
		TotalLines:  0,
		AILines:     0,
		HumanLines:  0,
		Percentage:  0.0,
		LastUpdated: time.Now(),
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
	fmt.Printf("AI Code Tracker (aict) v%s - Track AI vs Human code contributions\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aict init                    Initialize tracking in current directory")
	fmt.Println("  aict track -author <name>    Create a checkpoint for the specified author")
	fmt.Println("  aict report [options]        Show tracking metrics")
	fmt.Println("    --since <date>             Show report since date/time")
	fmt.Println("    --from <date> --to <date>  Show report for date range")
	fmt.Println("    --last <duration>          Show report for last N days/weeks/months (e.g., '7d', '2w', '1m')")
	fmt.Println("    --format <format>          Output format: table, graph, json (default: table)")
	fmt.Println("    --branch <name>            Show report for specific branch")
	fmt.Println("    --branch-regex <pattern>   Show report for branches matching regex pattern")
	fmt.Println("    --branch-pattern <pattern> Show report for branches matching glob pattern (e.g., 'feature/*')")
	fmt.Println("    --all-branches             Show summary of all branches")
	fmt.Println("  aict setup-hooks             Setup Claude Code and Git hooks for automatic tracking")
	fmt.Println("  aict mark-ai-edit            Mark current changes as AI-edited (called by hooks)")
	fmt.Println("    --tool <name>              AI tool name (default: claude)")
	fmt.Println("  aict snapshot                Analyze entire codebase using git blame")
	fmt.Println("  aict config                  View and edit configuration settings")
	fmt.Println("  aict reset                   Reset metrics to start tracking from current codebase state")
	fmt.Println("  aict version                 Show version information")
}

func getGitUserName() string {
	cmd := exec.Command("git", "config", "user.name")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

func handleConfig() {
	baseDir := defaultBaseDir
	configPath := filepath.Join(baseDir, "config.json")

	// Check if initialized - create default config if not exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config using MetricsStorage
		metricsStorage := storage.NewMetricsStorage(baseDir)
		_, err := metricsStorage.LoadConfig() // This will auto-create if not exists
		if err != nil {
			fmt.Printf("Error: Failed to create default config: %v\n", err)
			exitFunc(1)
		}
	}

	// Display configuration explanation
	fmt.Println("AI Code Tracker Configuration")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Configuration Fields:")
	fmt.Println("  target_ai_percentage   - Target percentage of AI-generated code (default: 80.0)")
	fmt.Println("  tracked_extensions     - File extensions to track (e.g., .go, .py, .js)")
	fmt.Println("  exclude_patterns       - Patterns to exclude from tracking (e.g., *_generated.go)")
	fmt.Println("  author_mappings        - Map git usernames to 'human' or 'ai' categories")
	fmt.Println()
	fmt.Printf("Opening config file: %s\n", configPath)
	fmt.Println()

	// Determine editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Fallback editors by platform
		switch runtime.GOOS {
		case "windows":
			editor = "notepad"
		case "darwin":
			// Try common macOS editors
			if _, err := exec.LookPath("code"); err == nil {
				editor = "code"
			} else if _, err := exec.LookPath("vim"); err == nil {
				editor = "vim"
			} else {
				editor = "open -e" // TextEdit
			}
		default:
			// Linux/Unix
			if _, err := exec.LookPath("vim"); err == nil {
				editor = "vim"
			} else if _, err := exec.LookPath("nano"); err == nil {
				editor = "nano"
			} else {
				editor = "vi"
			}
		}
	}

	// Open editor
	var cmd *exec.Cmd
	if strings.Contains(editor, " ") {
		// Handle commands with arguments (like "open -e")
		parts := strings.Fields(editor)
		cmd = exec.Command(parts[0], append(parts[1:], configPath)...)
	} else {
		cmd = exec.Command(editor, configPath)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error opening editor: %v\n", err)
		fmt.Printf("You can manually edit the file at: %s\n", configPath)
		exitFunc(1)
	}
}

func handleMarkAIEdit() {
	fs := flag.NewFlagSet("mark-ai-edit", flag.ExitOnError)
	tool := fs.String("tool", "claude", "AI tool name (claude, copilot, etc.)")
	postCommit := fs.Bool("post-commit", false, "Mark the last commit instead of working directory changes")
	fs.Parse(os.Args[2:])

	baseDir := defaultBaseDir

	// Check if initialized
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Printf("Error: AI Code Tracker not initialized. Run 'aict init' first.\n")
		exitFunc(1)
	}

	// Load config to get tracked extensions
	metricsStorage := storage.NewMetricsStorage(baseDir)
	config, err := metricsStorage.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		exitFunc(1)
	}

	// Get current commit
	currentCommit, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		fmt.Printf("Error getting current commit: %v\n", err)
		exitFunc(1)
	}
	commit := strings.TrimSpace(string(currentCommit))

	// Get git diff to find changed files and lines
	var cmd *exec.Cmd
	if *postCommit {
		// For post-commit hook: compare HEAD~1 with HEAD
		cmd = exec.Command("git", "diff", "HEAD~1", "HEAD", "--numstat")
	} else {
		// For post-tool-use hook: compare HEAD with working directory
		cmd = exec.Command("git", "diff", "HEAD", "--numstat")
	}
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running git diff: %v\n", err)
		exitFunc(1)
	}

	// Parse changed files
	changedFiles := make(map[string][]int)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		filePath := strings.Join(parts[2:], " ")

		// Check if file should be tracked
		shouldTrack := false
		for _, ext := range config.TrackedExtensions {
			if strings.HasSuffix(filePath, ext) {
				shouldTrack = true
				break
			}
		}

		if !shouldTrack {
			continue
		}

		// For simplicity in MVP, mark entire file as AI-edited
		// TODO: In future, parse git diff to get exact line numbers
		changedFiles[filePath] = []int{} // Empty array means "all changes in this file"
	}

	if len(changedFiles) == 0 {
		// No tracked files changed, nothing to mark
		return
	}

	// Create git note with AI edit information
	note := struct {
		Timestamp time.Time         `json:"timestamp"`
		Tool      string            `json:"tool"`
		Files     map[string][]int  `json:"files"`
		Commit    string            `json:"commit"`
	}{
		Timestamp: time.Now(),
		Tool:      *tool,
		Files:     changedFiles,
		Commit:    commit,
	}

	noteJSON, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		fmt.Printf("Error creating note: %v\n", err)
		exitFunc(1)
	}

	// Add git note
	cmd = exec.Command("git", "notes", "--ref=aict", "add", "-f", "-m", string(noteJSON), "HEAD")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Error adding git note: %v (output: %s)\n", err, string(output))
		exitFunc(1)
	}

	fmt.Printf("✓ Marked AI edits in %d file(s) for commit %s\n", len(changedFiles), commit[:7])
}

func handleSnapshot() {
	baseDir := defaultBaseDir

	// Check if initialized
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Printf("Error: AI Code Tracker not initialized. Run 'aict init' first.\n")
		exitFunc(1)
	}

	// Load config
	metricsStorage := storage.NewMetricsStorage(baseDir)
	config, err := metricsStorage.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		exitFunc(1)
	}

	fmt.Println("Analyzing codebase with git blame...")

	// Analyze codebase using git blame combined with git notes

	// Get all tracked files
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error listing files: %v\n", err)
		exitFunc(1)
	}

	aiLines := 0
	humanLines := 0
	filesAnalyzed := 0

	// Cache for git notes to avoid repeated queries
	notesCache := make(map[string]bool) // commitHash -> hasAINote for this file

	lines := strings.Split(string(output), "\n")
	for _, filePath := range lines {
		if filePath == "" {
			continue
		}

		// Check if file should be tracked
		shouldTrack := false
		for _, ext := range config.TrackedExtensions {
			if strings.HasSuffix(filePath, ext) {
				shouldTrack = true
				break
			}
		}

		if !shouldTrack {
			continue
		}

		// Run git blame with commit hash
		blameCmd := exec.Command("git", "blame", "--line-porcelain", filePath)
		blameOutput, err := blameCmd.Output()
		if err != nil {
			// Skip files that can't be blamed
			continue
		}

		// Parse blame output
		scanner := bufio.NewScanner(strings.NewReader(string(blameOutput)))
		var currentCommit string
		var currentAuthor string

		for scanner.Scan() {
			line := scanner.Text()

			// First line of each block contains the commit hash (40 hex chars followed by space and numbers)
			if len(line) >= 40 {
				// Check if first 40 chars are hex (commit hash)
				possibleHash := line[:40]
				isHash := true
				for _, c := range possibleHash {
					if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
						isHash = false
						break
					}
				}
				if isHash {
					currentCommit = possibleHash
					continue
				}
			}

			if strings.HasPrefix(line, "author ") {
				currentAuthor = strings.TrimPrefix(line, "author ")
				continue
			}

			// When we hit the actual code line (starts with tab)
			if strings.HasPrefix(line, "\t") && currentCommit != "" {
				isAI := false

				// Check git notes cache first
				cacheKey := currentCommit + ":" + filePath
				if aiNote, cached := notesCache[cacheKey]; cached {
					isAI = aiNote
				} else {
					// Query git notes for this commit
					notesCmd := exec.Command("git", "notes", "--ref=aict", "show", currentCommit)
					notesOutput, err := notesCmd.Output()

					if err == nil && len(notesOutput) > 0 {
						// Parse JSON to check if this file was AI-edited
						var noteData struct {
							Files map[string][]int `json:"files"`
						}
						if json.Unmarshal(notesOutput, &noteData) == nil {
							if _, exists := noteData.Files[filePath]; exists {
								isAI = true
							}
						}
					}

					// If no git note, fall back to author matching
					if !isAI && currentAuthor != "" {
						authorLower := strings.ToLower(currentAuthor)
						aiAuthors := []string{"claude", "ai", "assistant", "bot", "copilot"}
						for _, aiAuthor := range aiAuthors {
							if strings.Contains(authorLower, aiAuthor) {
								isAI = true
								break
							}
						}

						// Check author mappings
						if !isAI {
							if mapping, exists := config.AuthorMappings[currentAuthor]; exists {
								isAI = strings.ToLower(mapping) == "ai"
							}
						}
					}

					// Cache the result
					notesCache[cacheKey] = isAI
				}

				if isAI {
					aiLines++
				} else {
					humanLines++
				}

				// Reset for next line
				currentCommit = ""
				currentAuthor = ""
			}
		}

		filesAnalyzed++
	}

	totalLines := aiLines + humanLines
	percentage := 0.0
	if totalLines > 0 {
		percentage = float64(aiLines) / float64(totalLines) * 100
	}

	// Save snapshot
	result := &tracker.AnalysisResult{
		TotalLines:  totalLines,
		AILines:     aiLines,
		HumanLines:  humanLines,
		Percentage:  percentage,
		LastUpdated: time.Now(),
	}

	if err := metricsStorage.SaveMetrics(result); err != nil {
		fmt.Printf("Error saving snapshot: %v\n", err)
		exitFunc(1)
	}

	// Generate report
	fmt.Printf("\nCodebase Snapshot\n")
	fmt.Printf("=================\n")
	fmt.Printf("Files analyzed: %d\n", filesAnalyzed)
	fmt.Printf("Total lines: %d\n", totalLines)
	fmt.Printf("  AI lines: %d (%.1f%%)\n", aiLines, percentage)
	fmt.Printf("  Human lines: %d (%.1f%%)\n", humanLines, 100-percentage)
	fmt.Printf("\nTarget: %.1f%% AI code\n", config.TargetAIPercentage)
	fmt.Printf("Progress: %.1f%%\n", percentage/config.TargetAIPercentage*100)
	fmt.Printf("\n✓ Snapshot saved to %s/metrics/current.json\n", baseDir)
}
// Test comment for AI tracking
// Another test
// Success test
