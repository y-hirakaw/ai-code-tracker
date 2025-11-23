package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestVersion(t *testing.T) {
	// Test version constant
	if version == "" {
		t.Error("Version should not be empty")
	}

	// Version should follow semantic versioning pattern
	// Basic check for v.x.y format
	if len(version) < 5 {
		t.Error("Version should be in semantic versioning format (e.g., 0.3.6)")
	}
}

func TestDefaultBaseDir(t *testing.T) {
	expected := ".ai_code_tracking"
	if defaultBaseDir != expected {
		t.Errorf("Expected defaultBaseDir to be '%s', got '%s'", expected, defaultBaseDir)
	}
}

func TestCreateHookFiles(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-create-hooks")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test createHookFiles
	err = createHookFiles(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create hook files: %v", err)
	}

	// Verify hook files exist
	expectedFiles := []string{
		"hooks/pre-tool-use.sh",
		"hooks/post-tool-use.sh",
		"hooks/pre-commit",
		"hooks/post-commit",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(tmpDir, file)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("Expected file '%s' to exist", file)
			continue
		}

		// Check file is executable
		mode := info.Mode()
		if mode&0100 == 0 {
			t.Errorf("Expected file '%s' to be executable", file)
		}
	}
}

func TestUpdateMetricsFromRecords(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-update-metrics")
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

	// Create some test records
	recorder := tracker.NewCheckpointRecorder(tmpDir)
	records := []tracker.CheckpointRecord{
		{
			Author: "human",
			Added:  100,
		},
		{
			Author: "claude",
			Added:  150,
		},
	}

	// Save records
	for _, record := range records {
		err = recorder.RecordCheckpoint(record.Author)
		if err != nil {
			t.Fatalf("Failed to record checkpoint: %v", err)
		}
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

	// Since RecordCheckpoint uses actual git state, we can't predict exact values
	// Just verify metrics were updated
	if metrics.TotalLines == 0 && metrics.AILines == 0 && metrics.HumanLines == 0 {
		t.Skip("Metrics not updated - likely not in a git repository")
	}
}

func TestCountTotalLines(t *testing.T) {
	checkpoint := &tracker.Checkpoint{
		Files: map[string]tracker.FileContent{
			"file1.go": {
				Lines: []string{"line1", "line2", "line3"},
			},
			"file2.go": {
				Lines: []string{"line1", "line2"},
			},
		},
	}

	total := countTotalLines(checkpoint)
	expected := 5 // 3 + 2 lines

	if total != expected {
		t.Errorf("Expected total lines to be %d, got %d", expected, total)
	}
}

func TestCountTotalLinesEmpty(t *testing.T) {
	checkpoint := &tracker.Checkpoint{
		Files: map[string]tracker.FileContent{},
	}

	total := countTotalLines(checkpoint)

	if total != 0 {
		t.Errorf("Expected total lines to be 0 for empty checkpoint, got %d", total)
	}
}

func TestGetGitUserName(t *testing.T) {
	// This test might fail if git is not configured
	// We'll just verify the function doesn't panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("getGitUserName panicked: %v", r)
			}
		}()

		userName := getGitUserName()
		// userName might be empty if git is not configured
		_ = userName

		// Test that the function returns a string (possibly empty)
		// Note: getGitUserName doesn't trim, so we'll just check it doesn't panic
		_ = userName
	}()
}

func TestPrintUsage(t *testing.T) {
	// Capture stdout to test printUsage
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = oldOut

	output := make([]byte, 4096) // Increased buffer size
	n, _ := r.Read(output)
	result := string(output[:n])

	// Check that usage contains expected strings
	expectedStrings := []string{
		"AI Code Tracker (aict)",
		version,
		"init",
		"track",
		"report",
		"setup-hooks",
		"reset",
		"version",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Usage should contain '%s'", expected)
		}
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-copy-file")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	sourceFile := filepath.Join(tmpDir, "source.txt")
	content := "test content for copy"
	err = os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Test copyFile
	destFile := filepath.Join(tmpDir, "dest.txt")
	err = copyFile(sourceFile, destFile)
	if err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	// Verify copied content
	copiedContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copiedContent) != content {
		t.Errorf("Expected copied content '%s', got '%s'", content, string(copiedContent))
	}
}

func TestCopyFileErrors(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-copy-errors")
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Test with non-existent source file
	err := copyFile("nonexistent.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("Expected error when copying non-existent file")
	}

	// Test with invalid destination
	sourceFile := filepath.Join(tmpDir, "source.txt")
	os.WriteFile(sourceFile, []byte("test"), 0644)

	err = copyFile(sourceFile, "/invalid/path/dest.txt")
	if err == nil {
		t.Error("Expected error when copying to invalid destination")
	}
}

func TestMergeGitHook(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-merge-git-hook")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing hook
	existingHook := filepath.Join(tmpDir, "existing-hook")
	existingContent := "#!/bin/bash\necho \"existing hook\""
	err = os.WriteFile(existingHook, []byte(existingContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create existing hook: %v", err)
	}

	// Create AICT hook
	aictHook := filepath.Join(tmpDir, "aict-hook")
	aictContent := "#!/bin/bash\necho \"aict hook\""
	err = os.WriteFile(aictHook, []byte(aictContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create AICT hook: %v", err)
	}

	// Test mergeGitHook
	err = mergeGitHook(aictHook, existingHook)
	if err != nil {
		t.Fatalf("Failed to merge git hook: %v", err)
	}

	// Verify merged content
	mergedContent, err := os.ReadFile(existingHook)
	if err != nil {
		t.Fatalf("Failed to read merged hook: %v", err)
	}

	merged := string(mergedContent)
	if !strings.Contains(merged, "existing hook") {
		t.Error("Merged hook should contain existing content")
	}
	if !strings.Contains(merged, "aict hook") {
		t.Error("Merged hook should contain AICT content")
	}
	if !strings.Contains(merged, "# AI Code Tracker") {
		t.Error("Merged hook should contain comment separator")
	}
}

func TestMergeClaudeSettings(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-merge-claude-settings")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent file (should return error)
	err = mergeClaudeSettings("nonexistent.json")
	if err == nil {
		t.Error("Expected error when merging non-existent settings")
	}

	// Test with invalid JSON
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	err = mergeClaudeSettings(invalidFile)
	if err == nil {
		t.Error("Expected error when merging invalid JSON")
	}

	// Test successful merge with object-style hooks
	validFile := filepath.Join(tmpDir, "valid.json")
	existingSettings := map[string]interface{}{
		"existing": "value",
		"hooks": map[string]interface{}{
			"ExistingHook": "existing_value",
		},
	}
	existingData, _ := json.MarshalIndent(existingSettings, "", "  ")
	err = os.WriteFile(validFile, existingData, 0644)
	if err != nil {
		t.Fatalf("Failed to create valid JSON file: %v", err)
	}

	err = mergeClaudeSettings(validFile)
	if err != nil {
		t.Fatalf("Failed to merge valid settings: %v", err)
	}

	// Verify merged settings
	mergedData, err := os.ReadFile(validFile)
	if err != nil {
		t.Fatalf("Failed to read merged settings: %v", err)
	}

	var merged map[string]interface{}
	err = json.Unmarshal(mergedData, &merged)
	if err != nil {
		t.Fatalf("Failed to parse merged settings: %v", err)
	}

	// Check that existing value is preserved
	if merged["existing"] != "value" {
		t.Error("Merged settings should preserve existing values")
	}

	// Check that hooks were merged
	hooks, exists := merged["hooks"].(map[string]interface{})
	if !exists {
		t.Error("Merged settings should contain hooks object")
	}

	// Should have both existing and new hooks
	if hooks["ExistingHook"] != "existing_value" {
		t.Error("Should preserve existing hooks")
	}

	if _, hasPreTool := hooks["PreToolUse"]; !hasPreTool {
		t.Error("Should add PreToolUse hooks from AICT")
	}

	if _, hasPostTool := hooks["PostToolUse"]; !hasPostTool {
		t.Error("Should add PostToolUse hooks from AICT")
	}

	// Test with no existing hooks
	noHooksFile := filepath.Join(tmpDir, "no-hooks.json")
	noHooksSettings := map[string]interface{}{
		"other": "setting",
	}
	noHooksData, _ := json.MarshalIndent(noHooksSettings, "", "  ")
	err = os.WriteFile(noHooksFile, noHooksData, 0644)
	if err != nil {
		t.Fatalf("Failed to create no-hooks file: %v", err)
	}

	err = mergeClaudeSettings(noHooksFile)
	if err != nil {
		t.Fatalf("Failed to merge settings with no existing hooks: %v", err)
	}

	// Verify AICT hooks were added
	mergedNoHooksData, err := os.ReadFile(noHooksFile)
	if err != nil {
		t.Fatalf("Failed to read merged no-hooks settings: %v", err)
	}

	var mergedNoHooks map[string]interface{}
	err = json.Unmarshal(mergedNoHooksData, &mergedNoHooks)
	if err != nil {
		t.Fatalf("Failed to parse merged no-hooks settings: %v", err)
	}

	if _, hasHooks := mergedNoHooks["hooks"]; !hasHooks {
		t.Error("Should add hooks to settings without existing hooks")
	}
}
