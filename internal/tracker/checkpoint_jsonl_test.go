package tracker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCheckpointRecorder(t *testing.T) {
	baseDir := "/tmp/test-recorder"
	recorder := NewCheckpointRecorder(baseDir)

	if recorder.baseDir != baseDir {
		t.Errorf("Expected baseDir to be '%s', got '%s'", baseDir, recorder.baseDir)
	}
}

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		filepath string
		expected string
	}{
		{"main.go", ".go"},
		{"src/app.js", ".js"},
		{"path/to/file.py", ".py"},
		{"no_extension", ""},
		{"ends_with_dot.", ""},
		{".hidden", ".hidden"},
		{"file.tar.gz", ".gz"},
	}

	for _, test := range tests {
		result := getFileExtension(test.filepath)
		if result != test.expected {
			t.Errorf("getFileExtension(%s) = %s, expected %s", test.filepath, result, test.expected)
		}
	}
}

func TestShouldTrackFileRecorder(t *testing.T) {
	config := &Config{
		TrackedExtensions: []string{".go", ".js", ".py"},
		ExcludePatterns:   []string{"*_test.go", "vendor/*", "*.min.js"},
	}

	recorder := NewCheckpointRecorder("/tmp")

	tests := []struct {
		filepath string
		expected bool
	}{
		{"main.go", true},
		{"app.js", true},
		{"script.py", true},
		{"main_test.go", false}, // Matches exclude pattern
		{"lib.go", true},        // Doesn't match exclude pattern on basename
		{"app.min.js", false},   // Matches exclude pattern
		{"README.md", false},    // Not tracked extension
		{"config.json", false},  // Not tracked extension
	}

	for _, test := range tests {
		result := recorder.shouldTrackFile(test.filepath, config)
		if result != test.expected {
			t.Errorf("shouldTrackFile(%s) = %v, expected %v", test.filepath, result, test.expected)
		}
	}
}

func TestAppendAndReadRecords(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-jsonl-records")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	recorder := NewCheckpointRecorder(tmpDir)

	// Test appending records
	records := []CheckpointRecord{
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Author:    "human",
			Commit:    "abc123",
			Added:     10,
			Deleted:   5,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "claude",
			Commit:    "def456",
			Added:     20,
			Deleted:   3,
		},
		{
			Timestamp: time.Now(),
			Author:    "human",
			Commit:    "ghi789",
			Added:     15,
			Deleted:   8,
		},
	}

	// Append records
	for _, record := range records {
		err := recorder.appendRecord(record)
		if err != nil {
			t.Fatalf("Failed to append record: %v", err)
		}
	}

	// Read all records
	readRecords, err := recorder.ReadAllRecords()
	if err != nil {
		t.Fatalf("Failed to read records: %v", err)
	}

	if len(readRecords) != len(records) {
		t.Errorf("Expected %d records, got %d", len(records), len(readRecords))
	}

	// Verify record content
	for i, record := range readRecords {
		if record.Author != records[i].Author {
			t.Errorf("Record %d: expected author '%s', got '%s'", i, records[i].Author, record.Author)
		}
		if record.Commit != records[i].Commit {
			t.Errorf("Record %d: expected commit '%s', got '%s'", i, records[i].Commit, record.Commit)
		}
		if record.Added != records[i].Added {
			t.Errorf("Record %d: expected added %d, got %d", i, records[i].Added, record.Added)
		}
		if record.Deleted != records[i].Deleted {
			t.Errorf("Record %d: expected deleted %d, got %d", i, records[i].Deleted, record.Deleted)
		}
	}
}

func TestGetLatestRecords(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-latest-records")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	recorder := NewCheckpointRecorder(tmpDir)

	// Append 5 records
	for i := 0; i < 5; i++ {
		record := CheckpointRecord{
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Author:    "test",
			Commit:    string(rune('a' + i)),
			Added:     i * 10,
			Deleted:   i,
		}
		recorder.appendRecord(record)
	}

	// Test getting latest 3 records
	latest, err := recorder.GetLatestRecords(3)
	if err != nil {
		t.Fatalf("Failed to get latest records: %v", err)
	}

	if len(latest) != 3 {
		t.Errorf("Expected 3 records, got %d", len(latest))
	}

	// Verify we got the last 3 records
	if latest[0].Commit != "c" {
		t.Errorf("Expected first record commit to be 'c', got '%s'", latest[0].Commit)
	}
	if latest[2].Commit != "e" {
		t.Errorf("Expected last record commit to be 'e', got '%s'", latest[2].Commit)
	}

	// Test getting all records
	all, err := recorder.GetLatestRecords(0)
	if err != nil {
		t.Fatalf("Failed to get all records: %v", err)
	}

	if len(all) != 5 {
		t.Errorf("Expected 5 records, got %d", len(all))
	}

	// Test getting more than available
	tooMany, err := recorder.GetLatestRecords(10)
	if err != nil {
		t.Fatalf("Failed to get records: %v", err)
	}

	if len(tooMany) != 5 {
		t.Errorf("Expected 5 records when asking for 10, got %d", len(tooMany))
	}
}

func TestGetLastRecord(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-last-record")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	recorder := NewCheckpointRecorder(tmpDir)

	// Test with no records
	lastRecord, err := recorder.getLastRecord()
	if err != nil {
		t.Fatalf("Failed to get last record: %v", err)
	}
	if lastRecord != nil {
		t.Error("Expected nil for last record when no records exist")
	}

	// Add a record
	record := CheckpointRecord{
		Timestamp: time.Now(),
		Author:    "test-author",
		Commit:    "test123",
		Added:     50,
		Deleted:   10,
	}
	recorder.appendRecord(record)

	// Get last record
	lastRecord, err = recorder.getLastRecord()
	if err != nil {
		t.Fatalf("Failed to get last record: %v", err)
	}

	if lastRecord == nil {
		t.Fatal("Expected non-nil last record")
	}

	if lastRecord.Author != "test-author" {
		t.Errorf("Expected author 'test-author', got '%s'", lastRecord.Author)
	}
	if lastRecord.Commit != "test123" {
		t.Errorf("Expected commit 'test123', got '%s'", lastRecord.Commit)
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-load-config")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file
	config := &Config{
		TargetAIPercentage: 75.0,
		TrackedExtensions:  []string{".go", ".py"},
		ExcludePatterns:    []string{"*_test.go"},
		AuthorMappings: map[string]string{
			"AI Bot": "ai",
		},
	}

	configData, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(tmpDir, "config.json")
	os.WriteFile(configPath, configData, 0644)

	// Test loading config
	recorder := NewCheckpointRecorder(tmpDir)
	loadedConfig, err := recorder.loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.TargetAIPercentage != 75.0 {
		t.Errorf("Expected TargetAIPercentage 75.0, got %.1f", loadedConfig.TargetAIPercentage)
	}

	if len(loadedConfig.TrackedExtensions) != 2 {
		t.Errorf("Expected 2 tracked extensions, got %d", len(loadedConfig.TrackedExtensions))
	}

	if len(loadedConfig.ExcludePatterns) != 1 {
		t.Errorf("Expected 1 exclude pattern, got %d", len(loadedConfig.ExcludePatterns))
	}

	if len(loadedConfig.AuthorMappings) != 1 {
		t.Errorf("Expected 1 author mapping, got %d", len(loadedConfig.AuthorMappings))
	}
}
