package tracker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCheckpointManager(t *testing.T) {
	baseDir := "/tmp/test-checkpoints"
	cm := NewCheckpointManager(baseDir)

	if cm.baseDir != baseDir {
		t.Errorf("Expected baseDir to be '%s', got '%s'", baseDir, cm.baseDir)
	}
}

func TestGenerateID(t *testing.T) {
	cm := NewCheckpointManager("/tmp")
	id1 := cm.generateID()
	time.Sleep(10 * time.Millisecond)
	id2 := cm.generateID()

	if len(id1) != 8 {
		t.Errorf("Expected ID length to be 8, got %d", len(id1))
	}

	if id1 == id2 {
		t.Errorf("Expected different IDs, got the same: %s", id1)
	}
}

func TestReadFileLines(t *testing.T) {
	// Create temporary test file
	tmpFile := filepath.Join(os.TempDir(), "test_file.txt")
	content := "line1\nline2\nline3"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	cm := NewCheckpointManager("/tmp")
	lines, err := cm.readFileLines(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read file lines: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	expectedLines := []string{"line1", "line2", "line3"}
	for i, line := range lines {
		if line != expectedLines[i] {
			t.Errorf("Expected line %d to be '%s', got '%s'", i, expectedLines[i], line)
		}
	}
}

func TestSaveAndLoadCheckpoint(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-checkpoints")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCheckpointManager(tmpDir)

	// Create checkpoint
	checkpoint := &Checkpoint{
		ID:        "test123",
		Timestamp: time.Now(),
		Author:    "test-author",
		Files: map[string]FileContent{
			"file1.go": {
				Path:  "file1.go",
				Lines: []string{"package main", "func main() {}"},
			},
		},
		NumstatData: map[string][2]int{
			"file1.go": {10, 5},
		},
	}

	// Save checkpoint
	err = cm.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Load checkpoint
	filename := filepath.Join(tmpDir, "checkpoints", "test-author_test123.json")
	loaded, err := cm.LoadCheckpoint(filename)
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	// Verify loaded data
	if loaded.ID != checkpoint.ID {
		t.Errorf("Expected ID '%s', got '%s'", checkpoint.ID, loaded.ID)
	}

	if loaded.Author != checkpoint.Author {
		t.Errorf("Expected Author '%s', got '%s'", checkpoint.Author, loaded.Author)
	}

	if len(loaded.Files) != len(checkpoint.Files) {
		t.Errorf("Expected %d files, got %d", len(checkpoint.Files), len(loaded.Files))
	}

	loadedFile, exists := loaded.Files["file1.go"]
	if !exists {
		t.Error("Expected file1.go to exist in loaded checkpoint")
	} else {
		if loadedFile.Path != "file1.go" {
			t.Errorf("Expected file path 'file1.go', got '%s'", loadedFile.Path)
		}
		if len(loadedFile.Lines) != 2 {
			t.Errorf("Expected 2 lines, got %d", len(loadedFile.Lines))
		}
	}

	// Verify numstat data
	numstat, exists := loaded.NumstatData["file1.go"]
	if !exists {
		t.Error("Expected numstat data for file1.go")
	} else {
		if numstat[0] != 10 || numstat[1] != 5 {
			t.Errorf("Expected numstat [10, 5], got %v", numstat)
		}
	}
}

func TestGetLatestCheckpoints(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-latest-checkpoints")
	checkpointDir := filepath.Join(tmpDir, "checkpoints")
	err := os.MkdirAll(checkpointDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCheckpointManager(tmpDir)

	// Create multiple checkpoints
	for i := 0; i < 5; i++ {
		checkpoint := &Checkpoint{
			ID:        string(rune('a' + i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Author:    "test-author",
			Files:     make(map[string]FileContent),
		}

		data, _ := json.MarshalIndent(checkpoint, "", "  ")
		filename := filepath.Join(checkpointDir, checkpoint.Author+"_"+checkpoint.ID+".json")
		os.WriteFile(filename, data, 0644)
	}

	// Test getting latest checkpoints
	checkpoints, err := cm.GetLatestCheckpoints("test-author", 3)
	if err != nil {
		t.Fatalf("Failed to get latest checkpoints: %v", err)
	}

	if len(checkpoints) != 3 {
		t.Errorf("Expected 3 checkpoints, got %d", len(checkpoints))
	}

	// Test getting all checkpoints
	allCheckpoints, err := cm.GetLatestCheckpoints("test-author", 0)
	if err != nil {
		t.Fatalf("Failed to get all checkpoints: %v", err)
	}

	if len(allCheckpoints) != 5 {
		t.Errorf("Expected 5 checkpoints, got %d", len(allCheckpoints))
	}

	// Test wildcard author
	wildcardCheckpoints, err := cm.GetLatestCheckpoints("*", 0)
	if err != nil {
		t.Fatalf("Failed to get wildcard checkpoints: %v", err)
	}

	if len(wildcardCheckpoints) != 5 {
		t.Errorf("Expected 5 checkpoints with wildcard, got %d", len(wildcardCheckpoints))
	}
}

func TestScanCodeFiles(t *testing.T) {
	// Create temporary directory structure
	tmpDir := filepath.Join(os.TempDir(), "test-scan-files")
	err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := map[string]string{
		"file1.go":        "package main\nfunc main() {}",
		"file2.js":        "console.log('hello');",
		"file3.txt":       "should not be included",
		"subdir/file4.go": "package sub\nfunc Test() {}",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		err := os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Create checkpoint manager and scan files
	cm := NewCheckpointManager(tmpDir)
	checkpoint := &Checkpoint{
		ID:     "test",
		Author: "test",
		Files:  make(map[string]FileContent),
	}

	extensions := []string{".go", ".js"}
	err = cm.scanCodeFiles(tmpDir, extensions, checkpoint)
	if err != nil {
		t.Fatalf("Failed to scan code files: %v", err)
	}

	// Verify results
	if len(checkpoint.Files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(checkpoint.Files))
	}

	// Check specific files
	expectedFiles := []string{
		filepath.Join(tmpDir, "file1.go"),
		filepath.Join(tmpDir, "file2.js"),
		filepath.Join(tmpDir, "subdir/file4.go"),
	}

	for _, expectedPath := range expectedFiles {
		if _, exists := checkpoint.Files[expectedPath]; !exists {
			t.Errorf("Expected file %s to be included", expectedPath)
		}
	}

	// Ensure .txt file is not included
	txtPath := filepath.Join(tmpDir, "file3.txt")
	if _, exists := checkpoint.Files[txtPath]; exists {
		t.Errorf("Did not expect file %s to be included", txtPath)
	}
}
