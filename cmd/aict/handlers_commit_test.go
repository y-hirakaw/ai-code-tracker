package main

import (
	"os"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestGetLatestCommitHash(t *testing.T) {
	// Setup test git repository
	tmpDir := testutil.TempGitRepo(t)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create a commit
	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n")
	testutil.GitCommit(t, tmpDir, "Test commit")

	// Test getLatestCommitHash
	hash, err := getLatestCommitHash()

	if err != nil {
		t.Fatalf("getLatestCommitHash() error = %v", err)
	}

	if hash == "" {
		t.Error("getLatestCommitHash() returned empty hash")
	}

	// Verify hash format (40 hex characters)
	if len(hash) != 40 {
		t.Errorf("getLatestCommitHash() hash length = %d, want 40", len(hash))
	}

	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("getLatestCommitHash() invalid character in hash: %c", c)
			break
		}
	}
}

func TestHandleCommit_EnvironmentSetup(t *testing.T) {
	// このテストはhandleCommit()の統合テストの前提条件（環境セットアップ）を検証する
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n\nfunc main() {}\n")
	testutil.GitCommit(t, tmpDir, "Initial commit")

	hash, err := getLatestCommitHash()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}

	if hash == "" {
		t.Error("Expected commit hash after git commit")
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		fpath    string
		pattern  string
		expected bool
	}{
		// サフィックスワイルドカード（*で始まるパターン）
		{"suffix match _test.go", "handlers_commit_test.go", "*_test.go", true},
		{"suffix match nested", "internal/foo/bar_test.go", "*_test.go", true},
		{"suffix no match", "handlers_commit.go", "*_test.go", false},

		// プレフィックスワイルドカード（*で終わるパターン）
		{"prefix match vendor", "vendor/lib/foo.go", "vendor/*", true},
		{"prefix match node_modules", "node_modules/pkg/index.js", "node_modules/*", true},
		{"prefix no match", "src/main.go", "vendor/*", false},

		// 完全一致
		{"exact match", "Makefile", "Makefile", true},
		{"exact no match", "makefile", "Makefile", false},

		// 空文字列
		{"empty pattern", "foo.go", "", false},
		{"empty fpath", "", "*_test.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.fpath, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.fpath, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsTrackedFile(t *testing.T) {
	cfg := &tracker.Config{
		TrackedExtensions: []string{".go", ".py", ".js"},
		ExcludePatterns:   []string{"*_test.go", "vendor/*"},
	}

	tests := []struct {
		name     string
		fpath    string
		expected bool
	}{
		// 追跡対象
		{"go file tracked", "main.go", true},
		{"py file tracked", "script.py", true},
		{"js file tracked", "app.js", true},
		{"nested go file", "internal/pkg/handler.go", true},

		// 除外パターンに該当
		{"test file excluded", "main_test.go", false},
		{"nested test excluded", "pkg/handler_test.go", false},
		{"vendor excluded", "vendor/lib/foo.go", false},

		// 追跡対象外の拡張子
		{"md not tracked", "README.md", false},
		{"txt not tracked", "notes.txt", false},
		{"yaml not tracked", "config.yaml", false},
		{"no extension", "Makefile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTrackedFile(tt.fpath, cfg)
			if result != tt.expected {
				t.Errorf("isTrackedFile(%q) = %v, want %v", tt.fpath, result, tt.expected)
			}
		})
	}
}

func TestIsTrackedFile_EmptyConfig(t *testing.T) {
	cfg := &tracker.Config{
		TrackedExtensions: []string{},
		ExcludePatterns:   []string{},
	}

	if isTrackedFile("main.go", cfg) {
		t.Error("isTrackedFile should return false when no extensions are configured")
	}
}

func TestBuildAuthorshipMap(t *testing.T) {
	now := time.Now()

	cpHuman := &tracker.CheckpointV2{
		Timestamp: now.Add(-2 * time.Minute),
		Author:    "human",
		Type:      tracker.AuthorTypeHuman,
		Changes: map[string]tracker.Change{
			"main.go":  {Added: 10, Deleted: 0},
			"utils.go": {Added: 5, Deleted: 2},
		},
	}

	cpAI := &tracker.CheckpointV2{
		Timestamp: now.Add(-1 * time.Minute),
		Author:    "claude",
		Type:      tracker.AuthorTypeAI,
		Changes: map[string]tracker.Change{
			"main.go":    {Added: 20, Deleted: 5},
			"handler.go": {Added: 30, Deleted: 0},
		},
	}

	changedFiles := map[string]bool{
		"main.go":    true,
		"utils.go":   true,
		"handler.go": true,
		"other.go":   true, // チェックポイントに存在しないファイル
	}

	result := buildAuthorshipMap([]*tracker.CheckpointV2{cpHuman, cpAI}, changedFiles)

	// main.go は最後のチェックポイント（cpAI）が優先される
	if cp, exists := result["main.go"]; !exists {
		t.Error("main.go should be in authorship map")
	} else if cp.Author != "claude" {
		t.Errorf("main.go author = %q, want %q (last checkpoint wins)", cp.Author, "claude")
	}

	// utils.go は cpHuman のみ
	if cp, exists := result["utils.go"]; !exists {
		t.Error("utils.go should be in authorship map")
	} else if cp.Author != "human" {
		t.Errorf("utils.go author = %q, want %q", cp.Author, "human")
	}

	// handler.go は cpAI のみ
	if cp, exists := result["handler.go"]; !exists {
		t.Error("handler.go should be in authorship map")
	} else if cp.Author != "claude" {
		t.Errorf("handler.go author = %q, want %q", cp.Author, "claude")
	}

	// other.go はどのチェックポイントにも存在しない
	if _, exists := result["other.go"]; exists {
		t.Error("other.go should NOT be in authorship map (no checkpoint)")
	}
}

func TestBuildAuthorshipMap_EmptyCheckpoints(t *testing.T) {
	changedFiles := map[string]bool{"main.go": true}
	result := buildAuthorshipMap(nil, changedFiles)

	if len(result) != 0 {
		t.Errorf("expected empty map for nil checkpoints, got %d entries", len(result))
	}
}

func TestBuildAuthorshipLogFromDiff(t *testing.T) {
	cfg := &tracker.Config{
		TrackedExtensions: []string{".go", ".py"},
		ExcludePatterns:   []string{"*_test.go"},
		DefaultAuthor:     "default-dev",
	}

	t.Run("checkpoint match", func(t *testing.T) {
		diffMap := map[string]tracker.Change{
			"main.go": {Added: 10, Deleted: 2, Lines: [][]int{{1, 10}}},
		}
		authorMap := map[string]*tracker.CheckpointV2{
			"main.go": {
				Author:   "claude",
				Type:     tracker.AuthorTypeAI,
				Metadata: map[string]string{"model": "opus"},
			},
		}
		changedFiles := map[string]bool{"main.go": true}

		log, err := buildAuthorshipLogFromDiff(diffMap, authorMap, "abc123", changedFiles, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if log.Version != authorship.AuthorshipLogVersion {
			t.Errorf("Version = %q, want %q", log.Version, authorship.AuthorshipLogVersion)
		}
		if log.Commit != "abc123" {
			t.Errorf("Commit = %q, want %q", log.Commit, "abc123")
		}
		fi, exists := log.Files["main.go"]
		if !exists {
			t.Fatal("main.go should be in Files")
		}
		if fi.Authors[0].Name != "claude" {
			t.Errorf("Author = %q, want %q", fi.Authors[0].Name, "claude")
		}
		if fi.Authors[0].Type != tracker.AuthorTypeAI {
			t.Errorf("Type = %q, want %q", fi.Authors[0].Type, tracker.AuthorTypeAI)
		}
	})

	t.Run("default author fallback", func(t *testing.T) {
		diffMap := map[string]tracker.Change{
			"utils.go": {Added: 5, Lines: [][]int{{1, 5}}},
		}
		authorMap := map[string]*tracker.CheckpointV2{}
		changedFiles := map[string]bool{"utils.go": true}

		log, err := buildAuthorshipLogFromDiff(diffMap, authorMap, "def456", changedFiles, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fi := log.Files["utils.go"]
		if fi.Authors[0].Name != "default-dev" {
			t.Errorf("Author = %q, want %q (default author)", fi.Authors[0].Name, "default-dev")
		}
		if fi.Authors[0].Type != tracker.AuthorTypeHuman {
			t.Errorf("Type = %q, want %q", fi.Authors[0].Type, tracker.AuthorTypeHuman)
		}
	})

	t.Run("excluded file filtered", func(t *testing.T) {
		diffMap := map[string]tracker.Change{
			"main_test.go": {Added: 20, Lines: [][]int{{1, 20}}},
		}
		authorMap := map[string]*tracker.CheckpointV2{}
		changedFiles := map[string]bool{"main_test.go": true}

		log, err := buildAuthorshipLogFromDiff(diffMap, authorMap, "ghi789", changedFiles, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(log.Files) != 0 {
			t.Errorf("Files count = %d, want 0 (test file should be excluded)", len(log.Files))
		}
	})

	t.Run("untracked extension filtered", func(t *testing.T) {
		diffMap := map[string]tracker.Change{
			"README.md": {Added: 10, Lines: [][]int{{1, 10}}},
		}
		authorMap := map[string]*tracker.CheckpointV2{}
		changedFiles := map[string]bool{"README.md": true}

		log, err := buildAuthorshipLogFromDiff(diffMap, authorMap, "jkl012", changedFiles, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(log.Files) != 0 {
			t.Errorf("Files count = %d, want 0 (.md not tracked)", len(log.Files))
		}
	})

	t.Run("empty diff", func(t *testing.T) {
		log, err := buildAuthorshipLogFromDiff(map[string]tracker.Change{}, map[string]*tracker.CheckpointV2{}, "xyz", map[string]bool{}, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(log.Files) != 0 {
			t.Errorf("Files count = %d, want 0", len(log.Files))
		}
	})
}

func TestBuildAuthorshipMap_FiltersByChangedFiles(t *testing.T) {
	cp := &tracker.CheckpointV2{
		Author: "human",
		Type:   tracker.AuthorTypeHuman,
		Changes: map[string]tracker.Change{
			"tracked.go":   {Added: 10},
			"untracked.go": {Added: 5},
		},
	}

	// changedFilesに含まれないファイルは除外される
	changedFiles := map[string]bool{"tracked.go": true}
	result := buildAuthorshipMap([]*tracker.CheckpointV2{cp}, changedFiles)

	if _, exists := result["tracked.go"]; !exists {
		t.Error("tracked.go should be in result")
	}
	if _, exists := result["untracked.go"]; exists {
		t.Error("untracked.go should NOT be in result (not in changedFiles)")
	}
}
