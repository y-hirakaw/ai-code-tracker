package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestAIctStorage(t *testing.T) {
	// Create temporary .git directory
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldDir)

	// Create storage
	store, err := NewAIctStorage()
	if err != nil {
		t.Fatalf("NewAIctStorage failed: %v", err)
	}

	// Test SaveCheckpoint
	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "Test",
		Type:      tracker.AuthorTypeHuman,
		Changes: map[string]tracker.Change{
			"test.go": {Added: 10, Deleted: 2, Lines: [][]int{{1, 10}}},
		},
	}

	if err := store.SaveCheckpoint(checkpoint); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Test LoadCheckpoints
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints failed: %v", err)
	}

	if len(checkpoints) != 1 {
		t.Errorf("Expected 1 checkpoint, got %d", len(checkpoints))
	}

	if checkpoints[0].Author != "Test" {
		t.Errorf("Expected author Test, got %s", checkpoints[0].Author)
	}

	// Test multiple checkpoints
	checkpoint2 := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "AI",
		Type:      tracker.AuthorTypeAI,
		Metadata:  map[string]string{"model": "test-model"},
		Changes: map[string]tracker.Change{
			"main.go": {Added: 50, Deleted: 5, Lines: [][]int{{21, 70}}},
		},
	}

	if err := store.SaveCheckpoint(checkpoint2); err != nil {
		t.Fatalf("SaveCheckpoint (second) failed: %v", err)
	}

	checkpoints, err = store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints (second) failed: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints, got %d", len(checkpoints))
	}

	// Test ClearCheckpoints
	if err := store.ClearCheckpoints(); err != nil {
		t.Fatalf("ClearCheckpoints failed: %v", err)
	}

	checkpoints, _ = store.LoadCheckpoints()
	if len(checkpoints) != 0 {
		t.Errorf("Expected 0 checkpoints after clear, got %d", len(checkpoints))
	}
}

func TestSaveCheckpointCorruptedFile(t *testing.T) {
	// 破損したJSONファイルがある場合、SaveCheckpointがエラーを返すことを確認
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldDir)

	store, err := NewAIctStorage()
	if err != nil {
		t.Fatalf("NewAIctStorage failed: %v", err)
	}

	// 破損したJSONファイルを作成
	checkpointsDir := filepath.Join(gitDir, "aict", "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}
	corruptedData := []byte("{invalid json content")
	if err := os.WriteFile(filepath.Join(checkpointsDir, "latest.json"), corruptedData, 0644); err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	// SaveCheckpointがエラーを返すことを確認（旧実装ではエラーが無視されていた）
	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "Test",
		Type:      tracker.AuthorTypeHuman,
	}

	err = store.SaveCheckpoint(checkpoint)
	if err == nil {
		t.Error("SaveCheckpoint should return error for corrupted file")
	}
}

func TestSaveCheckpointAtomicWrite(t *testing.T) {
	// アトミック書き込みで一時ファイルが残らないことを確認
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldDir)

	store, err := NewAIctStorage()
	if err != nil {
		t.Fatalf("NewAIctStorage failed: %v", err)
	}

	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "Test",
		Type:      tracker.AuthorTypeHuman,
	}

	if err := store.SaveCheckpoint(checkpoint); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// 一時ファイルが残っていないことを確認
	checkpointsDir := filepath.Join(gitDir, "aict", "checkpoints")
	entries, err := os.ReadDir(checkpointsDir)
	if err != nil {
		t.Fatalf("Failed to read checkpoints dir: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("Temporary file should not remain: %s", entry.Name())
		}
	}
}

func TestAIctStorageConfig(t *testing.T) {
	// Create temporary .git directory
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(oldDir)

	// Create storage
	store, err := NewAIctStorage()
	if err != nil {
		t.Fatalf("NewAIctStorage failed: %v", err)
	}

	// Test SaveConfig
	cfg := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".py"},
		ExcludePatterns:    []string{"*_test.go"},
		DefaultAuthor:      "TestUser",
		AIAgents:           []string{"Claude Code", "Cursor"},
	}

	if err := store.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Test LoadConfig
	loadedCfg, err := store.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loadedCfg.DefaultAuthor != "TestUser" {
		t.Errorf("Expected default author TestUser, got %s", loadedCfg.DefaultAuthor)
	}

	if len(loadedCfg.AIAgents) != 2 {
		t.Errorf("Expected 2 AI agents, got %d", len(loadedCfg.AIAgents))
	}
}
