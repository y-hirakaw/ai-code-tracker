package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewJSONStorage(t *testing.T) {
	baseDir := "/tmp/test-storage"
	storage := NewJSONStorage(baseDir)

	if storage.baseDir != baseDir {
		t.Errorf("Expected baseDir to be '%s', got '%s'", baseDir, storage.baseDir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-json-storage")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Test data structure
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
		Items []string `json:"items"`
	}

	// Save data
	original := TestData{
		Name:  "test",
		Value: 42,
		Items: []string{"item1", "item2", "item3"},
	}

	err = storage.Save("test.json", original)
	if err != nil {
		t.Fatalf("Failed to save data: %v", err)
	}

	// Load data
	var loaded TestData
	err = storage.Load("test.json", &loaded)
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Verify loaded data
	if loaded.Name != original.Name {
		t.Errorf("Expected name '%s', got '%s'", original.Name, loaded.Name)
	}

	if loaded.Value != original.Value {
		t.Errorf("Expected value %d, got %d", original.Value, loaded.Value)
	}

	if len(loaded.Items) != len(original.Items) {
		t.Errorf("Expected %d items, got %d", len(original.Items), len(loaded.Items))
	}

	for i, item := range loaded.Items {
		if item != original.Items[i] {
			t.Errorf("Item %d: expected '%s', got '%s'", i, original.Items[i], item)
		}
	}
}

func TestSaveWithSubdirectory(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-subdirectory")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Save data in subdirectory
	data := map[string]string{"key": "value"}
	err = storage.Save("subdir/nested/test.json", data)
	if err != nil {
		t.Fatalf("Failed to save data in subdirectory: %v", err)
	}

	// Verify file exists
	filePath := filepath.Join(tmpDir, "subdir/nested/test.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to exist in subdirectory")
	}
}

func TestExists(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-exists")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Test non-existent file
	if storage.Exists("nonexistent.json") {
		t.Error("Expected Exists to return false for non-existent file")
	}

	// Create file
	data := map[string]string{"test": "data"}
	storage.Save("exists.json", data)

	// Test existing file
	if !storage.Exists("exists.json") {
		t.Error("Expected Exists to return true for existing file")
	}
}

func TestDelete(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-delete")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Create file
	data := map[string]string{"delete": "me"}
	storage.Save("delete.json", data)

	// Verify file exists
	if !storage.Exists("delete.json") {
		t.Fatal("Expected file to exist before deletion")
	}

	// Delete file
	err = storage.Delete("delete.json")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify file is deleted
	if storage.Exists("delete.json") {
		t.Error("Expected file to be deleted")
	}
}

func TestList(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-list")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Create test files
	files := []string{
		"file1.json",
		"file2.json",
		"subdir/file3.json",
		"other.txt",
	}

	for _, file := range files {
		storage.Save(file, map[string]string{"file": file})
	}

	// Test listing all json files
	matches, err := storage.List("*.json")
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches for *.json, got %d", len(matches))
	}

	// Test listing with subdirectory pattern
	matches, err = storage.List("**/*.json")
	if err != nil {
		t.Fatalf("Failed to list files with pattern: %v", err)
	}

	// Note: ** patterns might not work as expected with filepath.Glob
	// Test specific subdirectory
	matches, err = storage.List("subdir/*.json")
	if err != nil {
		t.Fatalf("Failed to list subdirectory files: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 match in subdir, got %d", len(matches))
	}

	if matches[0] != "subdir/file3.json" {
		t.Errorf("Expected 'subdir/file3.json', got '%s'", matches[0])
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-load-error")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Try to load non-existent file
	var data map[string]string
	err = storage.Load("nonexistent.json", &data)
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestSaveInvalidData(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-save-error")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	// Try to save a channel (not JSON serializable)
	ch := make(chan int)
	err = storage.Save("invalid.json", ch)
	if err == nil {
		t.Error("Expected error when saving non-serializable data")
	}
}