package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/git"
	"github.com/y-hirakaw/ai-code-tracker/internal/interfaces"
	"github.com/y-hirakaw/ai-code-tracker/internal/security"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
	"github.com/y-hirakaw/ai-code-tracker/internal/validation"
)

// HandlerConfig contains dependencies for handlers
type HandlerConfig struct {
	BaseDir        string
	Storage        interfaces.Storage
	MetricsStorage interfaces.MetricsStorage
	GitAnalyzer    *git.ContextAwareDiffAnalyzer
	Validator      *validation.ConfigValidator
	SafeOps        *security.SafeFileOperations
}

// NewHandlerConfig creates a new handler configuration with all dependencies
func NewHandlerConfig(baseDir string) (*HandlerConfig, error) {
	// Create safe file operations
	safeOps, err := security.NewSafeFileOperations(baseDir)
	if err != nil {
		return nil, errors.NewConfigError("NewHandlerConfig", "failed to create safe operations", err)
	}
	
	// Create storage instances
	jsonStorage, err := storage.NewJSONStorageV2(baseDir)
	if err != nil {
		return nil, err
	}
	
	metricsStorage := storage.NewMetricsStorageV2(jsonStorage)
	
	// Create Git analyzer with timeout
	gitAnalyzer := git.NewContextAwareDiffAnalyzer(30 * time.Second)
	
	// Create validator
	validator := validation.NewConfigValidator()
	
	return &HandlerConfig{
		BaseDir:        baseDir,
		Storage:        jsonStorage,
		MetricsStorage: metricsStorage,
		GitAnalyzer:    gitAnalyzer,
		Validator:      validator,
		SafeOps:        safeOps,
	}, nil
}

// InitHandler handles the init command with improved error handling
func (h *HandlerConfig) InitHandler() error {
	fmt.Println("Initializing AI Code Tracker...")
	
	// Create base directory
	if err := os.MkdirAll(h.BaseDir, 0755); err != nil {
		return errors.NewStorageError("InitHandler", h.BaseDir, err)
	}
	
	// Load or create configuration
	config, err := h.MetricsStorage.LoadConfig()
	if err != nil {
		return err
	}
	
	// Validate configuration
	if err := h.Validator.Validate(config); err != nil {
		return err
	}
	
	// Save validated configuration
	if err := h.MetricsStorage.SaveConfig(config); err != nil {
		return err
	}
	
	// Create necessary subdirectories
	dirs := []string{"checkpoints", "hooks", "metrics", "metrics/archive"}
	for _, dir := range dirs {
		dirPath := filepath.Join(h.BaseDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return errors.NewStorageError("InitHandler", dirPath, err)
		}
	}
	
	// Create hook files
	if err := h.createHookFiles(); err != nil {
		return err
	}
	
	// Initialize metrics
	initialMetrics := &tracker.AnalysisResult{
		TotalLines:  0,
		AILines:     0,
		HumanLines:  0,
		Percentage:  0,
		LastUpdated: time.Now(),
	}
	
	if err := h.MetricsStorage.SaveMetrics(initialMetrics); err != nil {
		return err
	}
	
	// Record initial checkpoint
	recorder := tracker.NewCheckpointRecorder(h.BaseDir)
	if err := recorder.RecordHumanCheckpoint(getGitUserName()); err != nil {
		fmt.Printf("Warning: Failed to record initial checkpoint: %v\n", err)
	}
	
	fmt.Println("✓ Configuration created")
	fmt.Println("✓ Hook files generated")
	fmt.Println("✓ Initial checkpoint recorded")
	fmt.Println()
	fmt.Println("AI Code Tracker initialized successfully!")
	fmt.Println("Next steps:")
	fmt.Println("1. Run 'aict setup-hooks' to configure Git and Claude hooks")
	fmt.Println("2. Start coding with Claude to track AI-generated code")
	
	return nil
}

// TrackHandler handles the track command with context support
func (h *HandlerConfig) TrackHandler() error {
	// Check if initialized
	if !h.Storage.Exists("config.json") {
		return errors.NewValidationError("TrackHandler", "config", "AI Code Tracker not initialized. Run 'aict init' first.")
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	// Check if in git repository
	isGitRepo, err := h.GitAnalyzer.IsGitRepositoryWithContext(ctx)
	if err != nil {
		return err
	}
	
	var commitHash string
	if isGitRepo {
		commitHash, err = h.GitAnalyzer.GetLatestCommitWithContext(ctx)
		if err != nil {
			// Non-fatal: continue without commit hash
			fmt.Printf("Warning: Could not get latest commit: %v\n", err)
		}
	}
	
	// Record checkpoint
	recorder := tracker.NewCheckpointRecorder(h.BaseDir)
	checkpoint, err := recorder.RecordHumanCheckpointWithCommit(getGitUserName(), commitHash)
	if err != nil {
		return errors.NewStorageError("TrackHandler", "checkpoint", err)
	}
	
	fmt.Printf("✓ Checkpoint recorded: %s\n", checkpoint.ID)
	if commitHash != "" {
		fmt.Printf("  Commit: %s\n", commitHash)
	}
	
	// Update metrics from records
	if err := h.updateMetricsFromRecords(); err != nil {
		return err
	}
	
	return nil
}

// ResetHandler handles the reset command with safety checks
func (h *HandlerConfig) ResetHandler() error {
	// Reset metrics
	resetMetrics := &tracker.AnalysisResult{
		TotalLines:  0,
		AILines:     0,
		HumanLines:  0,
		Percentage:  0.0,
		LastUpdated: time.Now(),
	}
	
	if err := h.MetricsStorage.SaveMetrics(resetMetrics); err != nil {
		return errors.NewStorageError("ResetHandler", "metrics", err)
	}
	
	// Clear checkpoints directory safely
	checkpointsDir := filepath.Join(h.BaseDir, "checkpoints")
	if err := h.SafeOps.SafeRemoveAll(checkpointsDir); err != nil {
		return err
	}
	
	// Recreate checkpoints directory
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return errors.NewStorageError("ResetHandler", checkpointsDir, err)
	}
	
	return nil
}

// createHookFiles creates hook files with security checks
func (h *HandlerConfig) createHookFiles() error {
	hooks := map[string]string{
		"hooks/pre-tool-use.sh":  templates.PreToolUseHook,
		"hooks/post-tool-use.sh": templates.PostToolUseHook,
		"hooks/post-commit":      templates.PostCommitHook,
	}
	
	for filename, content := range hooks {
		filePath := filepath.Join(h.BaseDir, filename)
		
		// Validate path
		validPath, err := h.SafeOps.ValidatePath(filePath)
		if err != nil {
			return err
		}
		
		// Write hook file
		if err := os.WriteFile(validPath, []byte(content), 0755); err != nil {
			return errors.NewStorageError("createHookFiles", filename, err)
		}
	}
	
	return nil
}

// updateMetricsFromRecords updates metrics with improved error handling
func (h *HandlerConfig) updateMetricsFromRecords() error {
	// Load configuration
	config, err := h.MetricsStorage.LoadConfig()
	if err != nil {
		return err
	}
	
	// Create analyzer
	analyzer := tracker.NewAnalyzer(config)
	
	// Read all records
	recorder := tracker.NewCheckpointRecorder(h.BaseDir)
	records, err := recorder.ReadRecords()
	if err != nil {
		return errors.NewStorageError("updateMetricsFromRecords", "records", err)
	}
	
	if len(records) == 0 {
		fmt.Println("No checkpoints found.")
		return nil
	}
	
	// Analyze records
	result, err := analyzer.AnalyzeRecords(records)
	if err != nil {
		return errors.NewAnalysisError("updateMetricsFromRecords", "analysis failed", err)
	}
	
	// Save updated metrics
	if err := h.MetricsStorage.SaveMetrics(result); err != nil {
		return err
	}
	
	// Generate and display report
	report, err := analyzer.GenerateReportFromRecords(records, 0)
	if err != nil {
		return errors.NewAnalysisError("updateMetricsFromRecords", "report generation failed", err)
	}
	
	fmt.Println(report)
	
	return nil
}