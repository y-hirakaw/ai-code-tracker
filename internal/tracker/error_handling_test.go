package tracker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestAnalyzerErrorHandling tests error conditions in the analyzer
func TestAnalyzerErrorHandling(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go"},
		ExcludePatterns:    []string{},
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	// Test IsAIAuthor with empty config
	isAI := analyzer.IsAIAuthor("test-author")
	// Should work without crashing
	_ = isAI

	// Test with empty before checkpoint
	after := &Checkpoint{
		ID:        "after",
		Timestamp: time.Now(),
		Author:    "test",
		Files:     make(map[string]FileContent),
	}

	// Test with valid checkpoints but empty data
	before := &Checkpoint{
		ID:        "before",
		Timestamp: time.Now().Add(-1 * time.Hour),
		Author:    "human",
		Files:     make(map[string]FileContent),
	}

	_, err := analyzer.AnalyzeCheckpoints(before, after)
	if err != nil {
		t.Errorf("Should handle empty checkpoints gracefully: %v", err)
	}
}

// TestCheckpointManagerErrorHandling tests error conditions in checkpoint management
func TestCheckpointManagerErrorHandling(t *testing.T) {
	// Test LoadCheckpoint with non-existent file
	cm := NewCheckpointManager("/tmp")
	_, err := cm.LoadCheckpoint("nonexistent.json")
	if err == nil {
		t.Error("Expected error when loading non-existent checkpoint")
	}

	// Test with read-only directory (skip if running as root)
	if os.Getuid() != 0 {
		// Create read-only directory for testing
		tmpDir := filepath.Join(os.TempDir(), "readonly-test")
		os.MkdirAll(tmpDir, 0444)
		defer os.RemoveAll(tmpDir)

		checkpoint := &Checkpoint{
			ID:     "test",
			Author: "test",
			Files:  make(map[string]FileContent),
		}

		cmReadOnly := NewCheckpointManager(tmpDir)
		err = cmReadOnly.SaveCheckpoint(checkpoint)
		if err == nil {
			t.Error("Expected error when saving to read-only directory")
		}
	}
}

// TestCheckpointRecorderErrorHandling tests error conditions in checkpoint recording
func TestCheckpointRecorderErrorHandling(t *testing.T) {
	// Test with invalid base directory
	recorder := NewCheckpointRecorder("/invalid/path")

	// Test RecordCheckpoint with invalid directory
	err := recorder.RecordCheckpoint("test")
	if err == nil {
		t.Error("Expected error when recording checkpoint to invalid directory")
	}

	// Test ReadAllRecords with invalid directory
	_, err = recorder.ReadAllRecords()
	if err != nil && !os.IsNotExist(err) {
		// This is OK - file doesn't exist returns empty slice
		t.Errorf("Unexpected error reading from invalid directory: %v", err)
	}
}

// TestFileHandlingEdgeCases tests edge cases in file handling
func TestFileHandlingEdgeCases(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-file-edge-cases")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCheckpointManager(tmpDir)

	// Test with empty file
	emptyFile := filepath.Join(tmpDir, "empty.go")
	err = os.WriteFile(emptyFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	lines, err := cm.readFileLines(emptyFile)
	if err != nil {
		t.Fatalf("Failed to read empty file: %v", err)
	}

	if len(lines) != 1 || lines[0] != "" {
		t.Error("Empty file should return single empty line")
	}

	// Test with file containing only newlines
	newlineFile := filepath.Join(tmpDir, "newlines.go")
	err = os.WriteFile(newlineFile, []byte("\n\n\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create newlines file: %v", err)
	}

	lines, err = cm.readFileLines(newlineFile)
	if err != nil {
		t.Fatalf("Failed to read newlines file: %v", err)
	}

	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (3 empty + 1 EOF), got %d", len(lines))
	}
}

// TestInvalidJSONHandling tests handling of invalid JSON data
func TestInvalidJSONHandling(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-invalid-json")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCheckpointManager(tmpDir)

	// Create invalid JSON file
	invalidFile := filepath.Join(tmpDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	// Test LoadCheckpoint with invalid JSON
	_, err = cm.LoadCheckpoint(invalidFile)
	if err == nil {
		t.Error("Expected error when loading checkpoint with invalid JSON")
	}

	// Test CheckpointRecorder with malformed JSONL
	recorder := NewCheckpointRecorder(tmpDir)
	
	// Create malformed JSONL file
	checkpointsFile := filepath.Join(tmpDir, "checkpoints.jsonl")
	malformedContent := "valid json line\n{invalid json}\nanother valid line\n"
	err = os.WriteFile(checkpointsFile, []byte(malformedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create malformed JSONL file: %v", err)
	}

	// ReadAllRecords should skip malformed lines and continue
	records, err := recorder.ReadAllRecords()
	if err != nil {
		t.Fatalf("Failed to read records with malformed lines: %v", err)
	}

	// Should have processed valid lines and skipped invalid ones
	if len(records) != 0 {
		// The "valid json line" is not actually valid JSON for CheckpointRecord
		t.Logf("Processed %d records from malformed file", len(records))
	}
}

// TestConfigErrorHandling tests configuration-related error handling
func TestConfigErrorHandling(t *testing.T) {
	// Test with invalid config
	invalidConfig := &Config{
		TargetAIPercentage: -10.0, // Invalid percentage
		TrackedExtensions:  []string{},
		ExcludePatterns:    []string{},
		AuthorMappings:     nil,
	}

	analyzer := NewAnalyzer(invalidConfig)

	// Test shouldTrackFile with empty extensions
	tracked := analyzer.shouldTrackFile("test.go")
	if tracked {
		t.Error("File should not be tracked when no extensions are configured")
	}

	// Test IsAIAuthor with nil AuthorMappings
	isAI := analyzer.IsAIAuthor("test")
	if isAI {
		t.Error("Should not identify as AI when no mappings are configured")
	}
}

// TestAnalyzerWithCorruptedData tests analyzer behavior with corrupted data
func TestAnalyzerWithCorruptedData(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go"},
		ExcludePatterns:    []string{},
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	// Test with checkpoint containing nil Files map
	before := &Checkpoint{
		ID:        "before",
		Timestamp: time.Now().Add(-1 * time.Hour),
		Author:    "human",
		Files:     nil, // Nil files map
	}

	after := &Checkpoint{
		ID:        "after",
		Timestamp: time.Now(),
		Author:    "claude",
		Files: map[string]FileContent{
			"test.go": {
				Path:  "test.go",
				Lines: []string{"package main", "func main() {}"},
			},
		},
	}

	// Should handle nil Files map gracefully
	result, err := analyzer.AnalyzeCheckpoints(before, after)
	if err != nil {
		t.Fatalf("Should handle nil Files map gracefully: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil")
	}
}

// TestCheckpointRecorderFilePermissionErrors tests file permission scenarios
func TestCheckpointRecorderFilePermissionErrors(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-permission-errors")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	err = os.MkdirAll(readOnlyDir, 0444) // Read-only permissions
	if err != nil {
		t.Fatalf("Failed to create read-only dir: %v", err)
	}

	recorder := NewCheckpointRecorder(readOnlyDir)

	// Test RecordCheckpoint with read-only directory
	err = recorder.RecordCheckpoint("test")
	if err == nil {
		t.Error("Expected error when writing to read-only directory")
	}
}

// TestGitCommandErrors tests error handling when git commands fail
func TestGitCommandErrors(t *testing.T) {
	// Create temporary directory without git
	tmpDir := filepath.Join(os.TempDir(), "test-git-errors")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to non-git directory
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	cm := NewCheckpointManager(tmpDir)

	// Test CreateCheckpoint in non-git directory
	checkpoint, err := cm.CreateCheckpoint("test", []string{".go"})
	if err != nil {
		t.Fatalf("CreateCheckpoint should not fail in non-git directory: %v", err)
	}

	// Should create checkpoint without commit hash
	if checkpoint.CommitHash != "" {
		t.Error("Commit hash should be empty in non-git directory")
	}

	// NumstatData should be empty or not cause errors
	if checkpoint.NumstatData == nil {
		t.Error("NumstatData should be initialized even in non-git directory")
	}
}

// TestAnalyzeRecordsEdgeCases tests edge cases in record analysis
func TestAnalyzeRecordsEdgeCases(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go"},
		ExcludePatterns:    []string{},
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	// Test with single record
	singleRecord := []CheckpointRecord{
		{
			Timestamp: time.Now(),
			Author:    "test",
			Added:     100,
			Deleted:   50,
		},
	}

	result, err := analyzer.AnalyzeRecords(singleRecord)
	if err != nil {
		t.Fatalf("Should handle single record: %v", err)
	}

	// With single record, no differences to calculate
	if result.AILines != 0 || result.HumanLines != 0 {
		t.Error("Single record should result in zero line changes")
	}

	// Test with records having negative differences
	negativeRecords := []CheckpointRecord{
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "human",
			Added:     100,
			Deleted:   20,
		},
		{
			Timestamp: time.Now(),
			Author:    "claude",
			Added:     50, // Less than previous
			Deleted:   30,
		},
	}

	result, err = analyzer.AnalyzeRecords(negativeRecords)
	if err != nil {
		t.Fatalf("Should handle negative differences: %v", err)
	}

	// Should handle negative differences gracefully (returns 0)
	if result.AILines < 0 || result.HumanLines < 0 {
		t.Error("Should not have negative line counts")
	}
}

// TestSpecialCharacterHandling tests handling of files with special characters
func TestSpecialCharacterHandling(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-special-chars")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCheckpointManager(tmpDir)

	// Create file with special characters in content
	specialFile := filepath.Join(tmpDir, "special.go")
	specialContent := "package main\n// Special chars: 日本語 émojis 🚀\nfunc main() {\n\tprintln(\"Hello, 世界!\")\n}"
	err = os.WriteFile(specialFile, []byte(specialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create special character file: %v", err)
	}

	// Test reading file with special characters
	lines, err := cm.readFileLines(specialFile)
	if err != nil {
		t.Fatalf("Failed to read file with special characters: %v", err)
	}

	// Should preserve special characters
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "日本語") {
		t.Error("Should preserve Japanese characters")
	}
	if !strings.Contains(content, "🚀") {
		t.Error("Should preserve emoji characters")
	}
	if !strings.Contains(content, "世界") {
		t.Error("Should preserve Chinese characters")
	}
}