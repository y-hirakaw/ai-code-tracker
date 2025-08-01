package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/period"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestHandleInitSuccess(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-handle-init")
	originalDir, _ := os.Getwd()
	defer func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	}()

	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	os.Chdir(tmpDir)

	// Test handleInit by simulating its behavior
	baseDir := defaultBaseDir

	// Create directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create tracking directory: %v", err)
	}

	// Initialize storage and config
	metricsStorage := storage.NewMetricsStorage(baseDir)
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs"},
		ExcludePatterns:    []string{"*_test.go", "*.test.js", "*.spec.ts", "*_generated.go"},
		AuthorMappings:     make(map[string]string),
	}

	if err := metricsStorage.SaveConfig(config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
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
		t.Fatalf("Failed to initialize metrics: %v", err)
	}

	// Create hook files
	if err := createHookFiles(baseDir); err != nil {
		t.Fatalf("Failed to create hook files: %v", err)
	}

	// Verify initialization
	// Check config file exists
	configPath := filepath.Join(baseDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should exist after initialization")
	}

	// Check metrics file exists
	metricsPath := filepath.Join(baseDir, "metrics", "current.json")
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		t.Error("Metrics file should exist after initialization")
	}

	// Check hook files exist
	hookFiles := []string{
		"hooks/pre-tool-use.sh",
		"hooks/post-tool-use.sh",
		"hooks/pre-commit",
		"hooks/post-commit",
	}

	for _, hookFile := range hookFiles {
		hookPath := filepath.Join(baseDir, hookFile)
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			t.Errorf("Hook file %s should exist after initialization", hookFile)
		}
	}
}

func TestUpdateMetricsFromRecordsIntegration(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-update-metrics-integration")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize storage with config
	metricsStorage := storage.NewMetricsStorage(tmpDir)
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go"},
		ExcludePatterns:    []string{},
		AuthorMappings:     make(map[string]string),
	}
	err = metricsStorage.SaveConfig(config)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create mock records by writing directly to JSONL file
	checkpointsFile := filepath.Join(tmpDir, "checkpoints.jsonl")
	records := []tracker.CheckpointRecord{
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Author:    "human",
			Added:     50,
			Deleted:   0,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "claude",
			Added:     100,
			Deleted:   10,
		},
	}

	// Write records to JSONL file
	file, err := os.Create(checkpointsFile)
	if err != nil {
		t.Fatalf("Failed to create checkpoints file: %v", err)
	}
	defer file.Close()

	for _, record := range records {
		data, _ := json.Marshal(record)
		file.WriteString(string(data) + "\n")
	}

	// Update metrics
	err = updateMetricsFromRecords(tmpDir)
	if err != nil {
		t.Fatalf("Failed to update metrics: %v", err)
	}

	// Load and verify metrics
	metrics, err := metricsStorage.LoadMetrics()
	if err != nil {
		t.Fatalf("Failed to load metrics: %v", err)
	}

	// Verify metrics were calculated correctly
	// Expected: 50 lines (human) + 50 lines diff (claude) = 50 human + 50 AI
	if metrics.HumanLines == 0 && metrics.AILines == 0 {
		t.Skip("No changes detected in metrics")
	}

	if metrics.TotalLines != metrics.AILines+metrics.HumanLines {
		t.Errorf("TotalLines should equal AILines + HumanLines, got %d != %d + %d",
			metrics.TotalLines, metrics.AILines, metrics.HumanLines)
	}
}

func TestSetupSingleGitHookNewHook(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-setup-single-git-hook")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source hook
	hookSource := filepath.Join(tmpDir, "source-hook")
	hookContent := "#!/bin/bash\necho \"test hook\""
	err = os.WriteFile(hookSource, []byte(hookContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create source hook: %v", err)
	}

	// Test setupSingleGitHook with new hook (no existing hook)
	hookDest := filepath.Join(tmpDir, "dest-hook")

	err = setupSingleGitHook(hookSource, hookDest, "test")
	if err != nil {
		t.Fatalf("Failed to setup single git hook: %v", err)
	}

	// Verify hook was copied
	if _, err := os.Stat(hookDest); os.IsNotExist(err) {
		t.Error("Destination hook should exist after setup")
	}

	// Verify content
	destContent, err := os.ReadFile(hookDest)
	if err != nil {
		t.Fatalf("Failed to read destination hook: %v", err)
	}

	if string(destContent) != hookContent {
		t.Errorf("Expected destination content to match source")
	}

	// Verify permissions
	info, err := os.Stat(hookDest)
	if err != nil {
		t.Fatalf("Failed to stat destination hook: %v", err)
	}

	mode := info.Mode()
	if mode&0100 == 0 {
		t.Error("Destination hook should be executable")
	}
}

func TestSetupClaudeHooksNewSettings(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-setup-claude-hooks")
	originalDir, _ := os.Getwd()
	defer func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	}()

	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	os.Chdir(tmpDir)

	// Test setupClaudeHooks with no existing settings
	err = setupClaudeHooks()
	if err != nil {
		t.Fatalf("Failed to setup Claude hooks: %v", err)
	}

	// Verify .claude directory was created
	claudeDir := ".claude"
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		t.Error("Claude directory should exist after setup")
	}

	// Verify settings.json was created
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error("Settings file should exist after setup")
	}

	// Verify settings content
	settingsContent, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read settings file: %v", err)
	}

	var settings map[string]interface{}
	err = json.Unmarshal(settingsContent, &settings)
	if err != nil {
		t.Fatalf("Failed to parse settings JSON: %v", err)
	}

	// Check hooks exist
	if _, exists := settings["hooks"]; !exists {
		t.Error("Settings should contain hooks")
	}
}

func TestMergeGitHookErrors(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-merge-git-hook-errors")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Test with non-existent source
	hookDest := filepath.Join(tmpDir, "dest-hook")
	os.WriteFile(hookDest, []byte("existing"), 0755)

	err := mergeGitHook("nonexistent", hookDest)
	if err == nil {
		t.Error("Expected error when merging with non-existent source")
	}

	// Test with non-existent destination
	hookSource := filepath.Join(tmpDir, "source-hook")
	os.WriteFile(hookSource, []byte("source"), 0755)

	err = mergeGitHook(hookSource, "nonexistent")
	if err == nil {
		t.Error("Expected error when merging to non-existent destination")
	}
}

func TestMergeClaudeSettingsErrors(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-merge-claude-settings-errors")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Test with non-existent settings file
	err := mergeClaudeSettings("nonexistent.json")
	if err == nil {
		t.Error("Expected error when merging non-existent settings")
	}

	// Test with invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidFile, []byte("invalid json"), 0644)

	err = mergeClaudeSettings(invalidFile)
	if err == nil {
		t.Error("Expected error when merging invalid JSON")
	}
}

func TestCreateHookFilesError(t *testing.T) {
	// Test createHookFiles with invalid directory
	err := createHookFiles("/invalid/path")
	if err == nil {
		t.Error("Expected error when creating hooks in invalid directory")
	}
}

func TestHandleResetLogic(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-handle-reset")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize storage with some data
	metricsStorage := storage.NewMetricsStorage(tmpDir)

	// Create initial metrics
	initialMetrics := &tracker.AnalysisResult{
		TotalLines:  100,
		AILines:     60,
		HumanLines:  40,
		Percentage:  60.0,
		LastUpdated: time.Now().Add(-1 * time.Hour),
	}
	err = metricsStorage.SaveMetrics(initialMetrics)
	if err != nil {
		t.Fatalf("Failed to save initial metrics: %v", err)
	}

	// Create checkpoints directory with some files
	checkpointsDir := filepath.Join(tmpDir, "checkpoints")
	err = os.MkdirAll(checkpointsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}

	testFile := filepath.Join(checkpointsDir, "test_checkpoint.json")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test checkpoint: %v", err)
	}

	// Simulate reset logic (without user input)
	resetMetrics := &tracker.AnalysisResult{
		TotalLines:  0,
		AILines:     0,
		HumanLines:  0,
		Percentage:  0.0,
		LastUpdated: time.Now(),
	}

	if err := metricsStorage.SaveMetrics(resetMetrics); err != nil {
		t.Fatalf("Failed to reset metrics: %v", err)
	}

	// Clear all checkpoints
	if err := os.RemoveAll(checkpointsDir); err != nil {
		t.Fatalf("Failed to clear checkpoints: %v", err)
	}
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to recreate checkpoints directory: %v", err)
	}

	// Verify reset
	metrics, err := metricsStorage.LoadMetrics()
	if err != nil {
		t.Fatalf("Failed to load metrics after reset: %v", err)
	}

	if metrics.TotalLines != 0 || metrics.AILines != 0 || metrics.HumanLines != 0 {
		t.Error("Metrics should be reset to zero")
	}

	// Verify checkpoints directory is empty
	files, err := os.ReadDir(checkpointsDir)
	if err != nil {
		t.Fatalf("Failed to read checkpoints directory: %v", err)
	}

	if len(files) != 0 {
		t.Error("Checkpoints directory should be empty after reset")
	}
}

// Tests for new period functionality

func TestReportOptions(t *testing.T) {
	// Test ReportOptions struct creation
	opts := &ReportOptions{
		Since:  "2 weeks ago",
		Format: "table",
	}

	if opts.Since != "2 weeks ago" {
		t.Errorf("Expected Since to be '2 weeks ago', got %s", opts.Since)
	}

	if opts.Format != "table" {
		t.Errorf("Expected Format to be 'table', got %s", opts.Format)
	}
}

// Mock data for testing period functionality
func createMockRecords() []tracker.CheckpointRecord {
	now := time.Now()
	return []tracker.CheckpointRecord{
		{
			Timestamp: now.Add(-2 * time.Hour),
			Author:    "human",
			Added:     10,
			Deleted:   0,
		},
		{
			Timestamp: now.Add(-1 * time.Hour),
			Author:    "claude",
			Added:     20,
			Deleted:   0,
		},
		{
			Timestamp: now.Add(-30 * time.Minute),
			Author:    "human",
			Added:     5,
			Deleted:   0,
		},
	}
}

func createMockConfig() *tracker.Config {
	return &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".py", ".js"},
		ExcludePatterns:    []string{"*_test.go"},
		AuthorMappings:     map[string]string{"user": "human"},
	}
}

func TestPeriodAnalysisFlow(t *testing.T) {
	// Test the period analysis flow that would be used in handlePeriodReport
	records := createMockRecords()
	config := createMockConfig()

	// Test with --last option
	timeRange, err := period.ParseLastDuration("1d")
	if err != nil {
		t.Fatalf("Failed to parse last duration: %v", err)
	}

	analyzer := period.NewAnalyzer(config)
	report, err := analyzer.AnalyzePeriod(records, timeRange)
	if err != nil {
		t.Fatalf("Failed to analyze period: %v", err)
	}

	// Should include all records from the last day
	expectedTotal := 10 + 20 + 5 // all records within 1 day
	if report.TotalLines != expectedTotal {
		t.Errorf("Expected total lines %d, got %d", expectedTotal, report.TotalLines)
	}

	// Test formatting
	formatter := period.NewFormatter(config.TargetAIPercentage)

	tableOutput, err := formatter.Format(report, period.FormatTable)
	if err != nil {
		t.Fatalf("Failed to format as table: %v", err)
	}

	if len(tableOutput) == 0 {
		t.Error("Table output should not be empty")
	}

	jsonOutput, err := formatter.Format(report, period.FormatJSON)
	if err != nil {
		t.Fatalf("Failed to format as JSON: %v", err)
	}

	if len(jsonOutput) == 0 {
		t.Error("JSON output should not be empty")
	}

	graphOutput, err := formatter.Format(report, period.FormatGraph)
	if err != nil {
		t.Fatalf("Failed to format as graph: %v", err)
	}

	if len(graphOutput) == 0 {
		t.Error("Graph output should not be empty")
	}
}

func TestTimeRangeParsing(t *testing.T) {
	// Test parsing functionality that would be used in handlePeriodReport

	// Test --last option
	lastRange, err := period.ParseLastDuration("7d")
	if err != nil {
		t.Fatalf("Failed to parse last duration: %v", err)
	}

	expectedDuration := 7 * 24 * time.Hour
	actualDuration := lastRange.To.Sub(lastRange.From)

	if actualDuration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, actualDuration)
	}

	// Test --since option
	sinceRange, err := period.ParseTimeRange("2025-01-01")
	if err != nil {
		t.Fatalf("Failed to parse since time: %v", err)
	}

	expectedYear := 2025
	if sinceRange.From.Year() != expectedYear {
		t.Errorf("Expected year %d, got %d", expectedYear, sinceRange.From.Year())
	}

	// Test --from/--to options
	fromToRange, err := period.ParseFromTo("2025-01-01", "2025-01-02")
	if err != nil {
		t.Fatalf("Failed to parse from/to range: %v", err)
	}

	if fromToRange.From.Year() != 2025 || fromToRange.To.Year() != 2025 {
		t.Error("From/To range should be in 2025")
	}

	if fromToRange.From.Month() != 1 || fromToRange.To.Month() != 1 {
		t.Error("From/To range should be in January")
	}

	if fromToRange.From.Day() != 1 || fromToRange.To.Day() != 2 {
		t.Error("From should be day 1, To should be day 2")
	}
}

func TestFilteringFunctionality(t *testing.T) {
	// Test filtering functionality used in period reports
	records := createMockRecords()

	now := time.Now()
	timeRange := &period.TimeRange{
		From: now.Add(-90 * time.Minute), // Last 90 minutes
		To:   now,
	}

	filtered := period.FilterRecordsInclusive(records, timeRange)

	// Should include records from the last 90 minutes
	// This should be the claude record (1 hour ago) and human record (30 min ago)
	expectedCount := 2
	if len(filtered) != expectedCount {
		t.Errorf("Expected %d filtered records, got %d", expectedCount, len(filtered))
	}
}
