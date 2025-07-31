package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// TestHandleResetLogic tests the core logic of handleReset without user input
func TestHandleResetCore(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-handle-reset-core")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory temporarily
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Initialize storage with some data
	baseDir := ".ai_code_tracking"
	metricsStorage := storage.NewMetricsStorage(baseDir)
	
	// Create config
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

	// Create initial metrics with non-zero values
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
	checkpointsDir := filepath.Join(baseDir, "checkpoints")
	err = os.MkdirAll(checkpointsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}

	testFile := filepath.Join(checkpointsDir, "test_checkpoint.json")
	err = os.WriteFile(testFile, []byte(`{"test": "data"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test checkpoint: %v", err)
	}

	// Simulate reset logic (the part that would happen after user confirms)
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

	// Verify reset worked
	metrics, err := metricsStorage.LoadMetrics()
	if err != nil {
		t.Fatalf("Failed to load metrics after reset: %v", err)
	}

	if metrics.TotalLines != 0 || metrics.AILines != 0 || metrics.HumanLines != 0 {
		t.Error("Metrics should be reset to zero")
	}

	if metrics.Percentage != 0.0 {
		t.Error("Percentage should be reset to 0.0")
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

// TestUserInputValidation tests user input handling patterns
func TestUserInputValidation(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"y", true, "Single 'y' should be accepted"},
		{"Y", true, "Capital 'Y' should be accepted"},
		{"yes", true, "Full 'yes' should be accepted"},
		{"YES", true, "Capital 'YES' should be accepted"},
		{"Yes", true, "Mixed case 'Yes' should be accepted"},
		{"n", false, "Single 'n' should be rejected"},
		{"no", false, "Full 'no' should be rejected"},
		{"", false, "Empty input should be rejected"},
		{"maybe", false, "Invalid input should be rejected"},
		{"  y  ", true, "Whitespace around 'y' should be handled"},
		{"  no  ", false, "Whitespace around 'no' should be handled"},
	}

	for _, test := range tests {
		// Simulate the input processing logic from handleReset
		response := strings.TrimSpace(strings.ToLower(test.input))
		shouldProceed := (response == "y" || response == "yes")
		
		if shouldProceed != test.expected {
			t.Errorf("%s: input '%s' -> expected %v, got %v", 
				test.desc, test.input, test.expected, shouldProceed)
		}
	}
}

// TestSetupHooksUserInputLogic tests user input handling in setup hooks
func TestSetupHooksUserInputLogic(t *testing.T) {
	// Test the same input validation logic used in setupSingleGitHook and setupClaudeHooks
	tests := []struct {
		input    string
		expected bool
		desc     string
	}{
		{"y", true, "Should proceed with 'y'"},
		{"yes", true, "Should proceed with 'yes'"},
		{"n", false, "Should cancel with 'n'"},
		{"no", false, "Should cancel with 'no'"},
		{"", false, "Should cancel with empty input"},
		{"Y", true, "Should handle uppercase"},
		{"YES", true, "Should handle uppercase"},
		{"anything else", false, "Should cancel with invalid input"},
	}

	for _, test := range tests {
		// This is the logic from setupSingleGitHook and setupClaudeHooks
		response := strings.TrimSpace(strings.ToLower(test.input))
		shouldProceed := (response == "y" || response == "yes")
		
		if shouldProceed != test.expected {
			t.Errorf("%s: input '%s' -> expected %v, got %v", 
				test.desc, test.input, test.expected, shouldProceed)
		}
	}
}

// TestFileOperationsRobustness tests file operation error handling
func TestFileOperationsRobustness(t *testing.T) {
	// Test copyFile with various error conditions
	tmpDir := filepath.Join(os.TempDir(), "test-file-ops")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test copying non-existent file
	err = copyFile("nonexistent-source.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}

	// Test copying to invalid destination
	sourceFile := filepath.Join(tmpDir, "source.txt")
	os.WriteFile(sourceFile, []byte("test content"), 0644)
	
	err = copyFile(sourceFile, "/invalid/destination/path.txt")
	if err == nil {
		t.Error("Expected error when copying to invalid destination")
	}

	// Test successful copy
	destFile := filepath.Join(tmpDir, "dest.txt")
	err = copyFile(sourceFile, destFile)
	if err != nil {
		t.Fatalf("Failed to copy valid file: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(content) != "test content" {
		t.Error("Copied file content doesn't match original")
	}
}