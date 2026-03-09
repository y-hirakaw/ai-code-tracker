package storage

import (
	"fmt"
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


func createTestStorage(t *testing.T) (*AIctStorage, func()) {
	t.Helper()
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
	store, err := NewAIctStorage()
	if err != nil {
		os.Chdir(oldDir)
		t.Fatalf("NewAIctStorage failed: %v", err)
	}
	return store, func() { os.Chdir(oldDir) }
}

func TestRemoveConsumedCheckpoints(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	now := time.Now()
	// 3つのチェックポイントを保存（異なるBaseCommitで独立）
	for i, offset := range []time.Duration{-3 * time.Minute, -2 * time.Minute, -1 * time.Minute} {
		cp := &tracker.CheckpointV2{
			Timestamp:  now.Add(offset),
			Author:     fmt.Sprintf("author%d", i),
			Type:       tracker.AuthorTypeHuman,
			BaseCommit: fmt.Sprintf("abc%d", i),
		}
		if err := store.SaveCheckpoint(cp); err != nil {
			t.Fatalf("SaveCheckpoint failed: %v", err)
		}
	}

	// LoadCheckpoints で JSON round-trip 後の timestamp を使う（monotonic clock 除去済み）
	loaded, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints failed: %v", err)
	}

	// author0 と author2 を消費済みとしてマーク → author1 だけ残る
	consumed := map[time.Time]bool{
		loaded[0].Timestamp: true,
		loaded[2].Timestamp: true,
	}
	if err := store.RemoveConsumedCheckpoints(consumed); err != nil {
		t.Fatalf("RemoveConsumedCheckpoints failed: %v", err)
	}

	remaining, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints failed: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("Expected 1 remaining checkpoint, got %d", len(remaining))
	}
	if remaining[0].Author != "author1" {
		t.Errorf("Expected remaining author 'author1', got %q", remaining[0].Author)
	}
}

func TestRemoveConsumedCheckpoints_Empty(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	cp := &tracker.CheckpointV2{Timestamp: time.Now(), Author: "test", Type: tracker.AuthorTypeHuman}
	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// 空のtimestampセット → 何も消えない
	if err := store.RemoveConsumedCheckpoints(map[time.Time]bool{}); err != nil {
		t.Fatalf("RemoveConsumedCheckpoints failed: %v", err)
	}

	remaining, _ := store.LoadCheckpoints()
	if len(remaining) != 1 {
		t.Errorf("Expected 1 checkpoint (nothing consumed), got %d", len(remaining))
	}
}

func TestPurgeExpiredCheckpoints(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	// 期限切れ（25時間前）と有効（1分前）のチェックポイント
	expired := &tracker.CheckpointV2{
		Timestamp: time.Now().Add(-25 * time.Hour),
		Author:    "old",
		Type:      tracker.AuthorTypeAI,
	}
	valid := &tracker.CheckpointV2{
		Timestamp: time.Now().Add(-1 * time.Minute),
		Author:    "new",
		Type:      tracker.AuthorTypeAI,
	}
	if err := store.SaveCheckpoint(expired); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}
	if err := store.SaveCheckpoint(valid); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	if err := store.PurgeExpiredCheckpoints(); err != nil {
		t.Fatalf("PurgeExpiredCheckpoints failed: %v", err)
	}

	remaining, _ := store.LoadCheckpoints()
	if len(remaining) != 1 {
		t.Fatalf("Expected 1 remaining (valid), got %d", len(remaining))
	}
	if remaining[0].Author != "new" {
		t.Errorf("Expected remaining author 'new', got %q", remaining[0].Author)
	}
}

func TestPurgeExpiredCheckpoints_AllValid(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	cp := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    "recent",
		Type:      tracker.AuthorTypeHuman,
	}
	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	if err := store.PurgeExpiredCheckpoints(); err != nil {
		t.Fatalf("PurgeExpiredCheckpoints failed: %v", err)
	}

	remaining, _ := store.LoadCheckpoints()
	if len(remaining) != 1 {
		t.Errorf("Expected 1 checkpoint (all valid), got %d", len(remaining))
	}
}

func TestRewriteCheckpoints(t *testing.T) {
	store, cleanup := createTestStorage(t)
	defer cleanup()

	// 3つ保存して2つに書き換え
	for i := 0; i < 3; i++ {
		cp := &tracker.CheckpointV2{
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Author:    "test",
			Type:      tracker.AuthorTypeHuman,
		}
		if err := store.SaveCheckpoint(cp); err != nil {
			t.Fatalf("SaveCheckpoint failed: %v", err)
		}
	}

	all, _ := store.LoadCheckpoints()
	if len(all) != 3 {
		t.Fatalf("Expected 3 checkpoints, got %d", len(all))
	}

	// 最初の2つだけ書き戻す
	if err := store.rewriteCheckpoints(all[:2]); err != nil {
		t.Fatalf("rewriteCheckpoints failed: %v", err)
	}

	remaining, _ := store.LoadCheckpoints()
	if len(remaining) != 2 {
		t.Errorf("Expected 2 checkpoints after rewrite, got %d", len(remaining))
	}
}

func TestGetAictDir(t *testing.T) {
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

	dir := store.GetAictDir()
	if dir == "" {
		t.Error("GetAictDir() returned empty string")
	}

	// Should end with /aict or \aict (path separator dependent)
	if !strings.HasSuffix(dir, "aict") {
		t.Errorf("GetAictDir() = %q, should end with 'aict'", dir)
	}

	// Should contain .git/aict path components
	if !strings.Contains(dir, filepath.Join(".git", "aict")) {
		t.Errorf("GetAictDir() = %q, should contain '.git/aict'", dir)
	}

	// Directory should actually exist on disk
	info, err := os.Stat(dir)
	if err != nil {
		t.Errorf("GetAictDir() directory does not exist: %v", err)
	} else if !info.IsDir() {
		t.Errorf("GetAictDir() path is not a directory")
	}
}

func TestExpandConsumedByBaseCommit_BasicPairing(t *testing.T) {
	now := time.Now()
	// Developer baseline (empty Changes) + AI edit のペア
	cpBaseline := &tracker.CheckpointV2{
		Timestamp:  now.Add(-2 * time.Minute),
		Author:     "Developer",
		Type:       tracker.AuthorTypeHuman,
		BaseCommit: "abc123",
		Snapshot:   map[string]tracker.FileSnapshot{"main.go": {Hash: "h1", Lines: 10}},
	}
	cpAI := &tracker.CheckpointV2{
		Timestamp:  now.Add(-1 * time.Minute),
		Author:     "Claude",
		Type:       tracker.AuthorTypeAI,
		BaseCommit: "abc123",
		Changes:    map[string]tracker.Change{"main.go": {Added: 5}},
		Snapshot:   map[string]tracker.FileSnapshot{"main.go": {Hash: "h2", Lines: 15}},
	}
	checkpoints := []*tracker.CheckpointV2{cpBaseline, cpAI}

	// AI CPだけ消費済み
	consumed := map[time.Time]bool{cpAI.Timestamp: true}
	ExpandConsumedByBaseCommit(checkpoints, consumed)

	// Developer baselineもペアで消費されるべき
	if !consumed[cpBaseline.Timestamp] {
		t.Error("Developer baseline should be consumed as pair with AI checkpoint")
	}
}

func TestExpandConsumedByBaseCommit_DoubleStash(t *testing.T) {
	now := time.Now()
	// stashセッション1: main.go を編集
	cp1Dev := &tracker.CheckpointV2{
		Timestamp:  now.Add(-4 * time.Minute),
		Author:     "Developer",
		Type:       tracker.AuthorTypeHuman,
		BaseCommit: "abc123",
		Snapshot:   map[string]tracker.FileSnapshot{"main.go": {Hash: "h1"}},
	}
	cp1AI := &tracker.CheckpointV2{
		Timestamp:  now.Add(-3 * time.Minute),
		Author:     "Claude",
		Type:       tracker.AuthorTypeAI,
		BaseCommit: "abc123",
		Changes:    map[string]tracker.Change{"main.go": {Added: 5}},
		Snapshot:   map[string]tracker.FileSnapshot{"main.go": {Hash: "h2"}},
	}
	// stashセッション2: util.go を編集（同じBaseCommitだが別ファイル）
	cp2Dev := &tracker.CheckpointV2{
		Timestamp:  now.Add(-2 * time.Minute),
		Author:     "Developer",
		Type:       tracker.AuthorTypeHuman,
		BaseCommit: "abc123",
		Snapshot:   map[string]tracker.FileSnapshot{"util.go": {Hash: "h3"}},
	}
	cp2AI := &tracker.CheckpointV2{
		Timestamp:  now.Add(-1 * time.Minute),
		Author:     "Claude",
		Type:       tracker.AuthorTypeAI,
		BaseCommit: "abc123",
		Changes:    map[string]tracker.Change{"util.go": {Added: 3}},
		Snapshot:   map[string]tracker.FileSnapshot{"util.go": {Hash: "h4"}},
	}
	checkpoints := []*tracker.CheckpointV2{cp1Dev, cp1AI, cp2Dev, cp2AI}

	// セッション1のAI CPだけ消費済み
	consumed := map[time.Time]bool{cp1AI.Timestamp: true}
	ExpandConsumedByBaseCommit(checkpoints, consumed)

	// セッション1のDev baselineは消費（main.goが重複）
	if !consumed[cp1Dev.Timestamp] {
		t.Error("Session 1 Developer baseline should be consumed (file overlap)")
	}
	// セッション2は消費されない（util.goはmain.goと重複しない）
	if consumed[cp2Dev.Timestamp] {
		t.Error("Session 2 Developer baseline should NOT be consumed (no file overlap)")
	}
	if consumed[cp2AI.Timestamp] {
		t.Error("Session 2 AI checkpoint should NOT be consumed (no file overlap)")
	}
}

func TestExpandConsumedByBaseCommit_EmptyBaseCommit(t *testing.T) {
	now := time.Now()
	// 初回コミット前のチェックポイント（BaseCommit=""）
	cp1 := &tracker.CheckpointV2{
		Timestamp:  now.Add(-2 * time.Minute),
		Author:     "Developer",
		Type:       tracker.AuthorTypeHuman,
		BaseCommit: "",
		Snapshot:   map[string]tracker.FileSnapshot{"main.go": {Hash: "h1"}},
	}
	cp2 := &tracker.CheckpointV2{
		Timestamp:  now.Add(-1 * time.Minute),
		Author:     "Claude",
		Type:       tracker.AuthorTypeAI,
		BaseCommit: "",
		Changes:    map[string]tracker.Change{"main.go": {Added: 10}},
		Snapshot:   map[string]tracker.FileSnapshot{"main.go": {Hash: "h2"}},
	}
	// 別ファイル（BaseCommit=""だがファイル重複なし）
	cp3 := &tracker.CheckpointV2{
		Timestamp:  now.Add(-30 * time.Second),
		Author:     "Developer",
		Type:       tracker.AuthorTypeHuman,
		BaseCommit: "",
		Snapshot:   map[string]tracker.FileSnapshot{"other.go": {Hash: "h3"}},
	}
	checkpoints := []*tracker.CheckpointV2{cp1, cp2, cp3}

	consumed := map[time.Time]bool{cp2.Timestamp: true}
	ExpandConsumedByBaseCommit(checkpoints, consumed)

	if !consumed[cp1.Timestamp] {
		t.Error("cp1 should be consumed (same BaseCommit + file overlap)")
	}
	if consumed[cp3.Timestamp] {
		t.Error("cp3 should NOT be consumed (same BaseCommit but no file overlap)")
	}
}

func TestExpandConsumedByBaseCommit_NoConsumed(t *testing.T) {
	now := time.Now()
	cp := &tracker.CheckpointV2{
		Timestamp:  now,
		Author:     "Developer",
		Type:       tracker.AuthorTypeHuman,
		BaseCommit: "abc123",
		Changes:    map[string]tracker.Change{"main.go": {Added: 5}},
	}
	checkpoints := []*tracker.CheckpointV2{cp}
	consumed := map[time.Time]bool{}

	ExpandConsumedByBaseCommit(checkpoints, consumed)

	if consumed[cp.Timestamp] {
		t.Error("Nothing should be consumed when consumed set is empty")
	}
}
