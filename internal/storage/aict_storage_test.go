package storage

import (
	"os"
	"path/filepath"
	"strings"
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

func TestSaveCheckpointCorruptedJSONL(t *testing.T) {
	// JSONL形式: 破損行があってもSaveCheckpointは追記でき、LoadCheckpointsは破損行をスキップする
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

	// 破損行を含むJSONLファイルを作成
	checkpointsDir := filepath.Join(gitDir, "aict", "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}
	corruptedData := []byte("{invalid json content\n")
	if err := os.WriteFile(filepath.Join(checkpointsDir, "latest.json"), corruptedData, 0644); err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	// JSONL方式: 破損行があっても追記は成功する
	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "Test",
		Type:      tracker.AuthorTypeHuman,
	}

	err = store.SaveCheckpoint(checkpoint)
	if err != nil {
		t.Fatalf("SaveCheckpoint should succeed even with corrupted lines: %v", err)
	}

	// LoadCheckpointsは破損行をスキップし、有効なチェックポイントのみ返す
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints failed: %v", err)
	}
	if len(checkpoints) != 1 {
		t.Errorf("Expected 1 valid checkpoint (corrupted line skipped), got %d", len(checkpoints))
	}
	if len(checkpoints) > 0 && checkpoints[0].Author != "Test" {
		t.Errorf("Expected author Test, got %s", checkpoints[0].Author)
	}
}

func TestSaveCheckpointCorruptedJSONArray(t *testing.T) {
	// 旧JSON配列形式が破損している場合、マイグレーションでエラーを返す
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

	// 破損したJSON配列ファイルを作成（'['で始まるが不正なJSON）
	checkpointsDir := filepath.Join(gitDir, "aict", "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}
	corruptedData := []byte("[invalid json array")
	if err := os.WriteFile(filepath.Join(checkpointsDir, "latest.json"), corruptedData, 0644); err != nil {
		t.Fatalf("Failed to write corrupted file: %v", err)
	}

	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "Test",
		Type:      tracker.AuthorTypeHuman,
	}

	err = store.SaveCheckpoint(checkpoint)
	if err == nil {
		t.Error("SaveCheckpoint should return error for corrupted JSON array")
	}
}

func TestLoadCheckpointsJSONArrayBackwardCompat(t *testing.T) {
	// 旧JSON配列形式のファイルが正しく読み込めることを確認
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

	// 旧JSON配列形式のファイルを直接作成
	checkpointsDir := filepath.Join(gitDir, "aict", "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}

	jsonArray := `[
  {"timestamp":"2025-01-01T00:00:00Z","author":"human","type":"human","changes":{"a.go":{"added":5,"deleted":0}}},
  {"timestamp":"2025-01-01T01:00:00Z","author":"claude","type":"ai","changes":{"b.go":{"added":10,"deleted":2}}}
]`
	if err := os.WriteFile(filepath.Join(checkpointsDir, "latest.json"), []byte(jsonArray), 0644); err != nil {
		t.Fatalf("Failed to write JSON array file: %v", err)
	}

	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints failed for JSON array: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints from JSON array, got %d", len(checkpoints))
	}
	if len(checkpoints) > 0 && checkpoints[0].Author != "human" {
		t.Errorf("Expected first author human, got %s", checkpoints[0].Author)
	}
}

func TestSaveCheckpointMigratesJSONArray(t *testing.T) {
	// 旧JSON配列ファイルにSaveCheckpointするとJSONL形式にマイグレーションされる
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

	// 旧JSON配列形式のファイルを作成
	checkpointsDir := filepath.Join(gitDir, "aict", "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to create checkpoints dir: %v", err)
	}

	jsonArray := `[{"timestamp":"2025-01-01T00:00:00Z","author":"existing","type":"human","changes":{}}]`
	checkpointsFile := filepath.Join(checkpointsDir, "latest.json")
	if err := os.WriteFile(checkpointsFile, []byte(jsonArray), 0644); err != nil {
		t.Fatalf("Failed to write JSON array file: %v", err)
	}

	// 新しいチェックポイントを追記
	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "new",
		Type:      tracker.AuthorTypeAI,
	}

	if err := store.SaveCheckpoint(checkpoint); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// 全チェックポイントが読み込めることを確認（旧1件 + 新1件）
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints failed: %v", err)
	}

	if len(checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints (1 migrated + 1 new), got %d", len(checkpoints))
	}

	// ファイルがJSONL形式になっていることを確認（'['で始まらない）
	data, err := os.ReadFile(checkpointsFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if len(data) > 0 && data[0] == '[' {
		t.Error("File should be in JSONL format after migration, but starts with '['")
	}
}

func TestSaveCheckpointJSONLFormat(t *testing.T) {
	// JSONL形式で保存されることを確認
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

	// 2件のチェックポイントを保存
	for _, author := range []string{"human", "ai"} {
		cp := &tracker.CheckpointV2{
			Timestamp: time.Now(),
			Author:    author,
			Type:      tracker.AuthorTypeHuman,
		}
		if err := store.SaveCheckpoint(cp); err != nil {
			t.Fatalf("SaveCheckpoint failed for %s: %v", author, err)
		}
	}

	// ファイル内容がJSONL形式であることを確認
	checkpointsFile := filepath.Join(gitDir, "aict", "checkpoints", "latest.json")
	data, err := os.ReadFile(checkpointsFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("Expected 2 JSONL lines, got %d", lines)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *tracker.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go"},
				DefaultAuthor:      "dev",
			},
			wantErr: false,
		},
		{
			name: "zero percentage is valid",
			cfg: &tracker.Config{
				TargetAIPercentage: 0,
				TrackedExtensions:  []string{".go"},
				DefaultAuthor:      "dev",
			},
			wantErr: false,
		},
		{
			name: "100 percentage is valid",
			cfg: &tracker.Config{
				TargetAIPercentage: 100,
				TrackedExtensions:  []string{".go"},
				DefaultAuthor:      "dev",
			},
			wantErr: false,
		},
		{
			name: "negative percentage",
			cfg: &tracker.Config{
				TargetAIPercentage: -1,
				TrackedExtensions:  []string{".go"},
				DefaultAuthor:      "dev",
			},
			wantErr: true,
			errMsg:  "target_ai_percentage",
		},
		{
			name: "over 100 percentage",
			cfg: &tracker.Config{
				TargetAIPercentage: 101,
				TrackedExtensions:  []string{".go"},
				DefaultAuthor:      "dev",
			},
			wantErr: true,
			errMsg:  "target_ai_percentage",
		},
		{
			name: "empty tracked extensions",
			cfg: &tracker.Config{
				TargetAIPercentage: 80,
				TrackedExtensions:  []string{},
				DefaultAuthor:      "dev",
			},
			wantErr: true,
			errMsg:  "tracked_extensions",
		},
		{
			name: "empty default author",
			cfg: &tracker.Config{
				TargetAIPercentage: 80,
				TrackedExtensions:  []string{".go"},
				DefaultAuthor:      "",
			},
			wantErr: true,
			errMsg:  "default_author",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message %q should contain %q", err.Error(), tt.errMsg)
				}
			}
		})
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
