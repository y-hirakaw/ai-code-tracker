package main

import (
	"os"
	"os/exec"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := testutil.TempGitRepo(t)

	// Create initial commit
	content := `package main

func hello() string {
	return "hello"
}
`
	testutil.CreateTestFile(t, tmpDir, "test.go", content)
	testutil.GitCommit(t, tmpDir, "Initial commit")

	cleanup := func() {
		// No cleanup needed - t.TempDir() handles it
	}

	return tmpDir, cleanup
}

func TestGetDetailedDiff(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	tests := []struct {
		name        string
		setup       func() string // Returns filepath
		wantAdded   int
		wantDeleted int
		wantErr     bool
	}{
		{
			name: "New file",
			setup: func() string {
				filepath := "new.go"
				content := `package main

func new() {
	println("new")
}
`
				os.WriteFile(filepath, []byte(content), 0644)
				return filepath
			},
			wantAdded:   5, // TrimSpace removes trailing newline, so 5 lines instead of 6
			wantDeleted: 0,
			wantErr:     false,
		},
		{
			name: "Modified file - added lines",
			setup: func() string {
				filepath := "test.go"
				content := `package main

func hello() string {
	return "hello"
}

func world() string {
	return "world"
}
`
				os.WriteFile(filepath, []byte(content), 0644)
				return filepath
			},
			wantAdded:   4,
			wantDeleted: 0,
			wantErr:     false,
		},
		{
			name: "Modified file - deleted lines",
			setup: func() string {
				filepath := "test.go"
				content := `package main

func hello() string {
	return "hello"
}
`
				// First, add lines
				newContent := `package main

func hello() string {
	return "hello"
}

func temp() {
}
`
				os.WriteFile(filepath, []byte(newContent), 0644)
				cmd := exec.Command("git", "add", filepath)
				cmd.Run()
				cmd = exec.Command("git", "commit", "-m", "Add temp function")
				cmd.Run()

				// Then delete
				os.WriteFile(filepath, []byte(content), 0644)
				return filepath
			},
			wantAdded:   0,
			wantDeleted: 3,
			wantErr:     false,
		},
		{
			name: "Modified file - content changed",
			setup: func() string {
				// First commit current state
				filepath := "test.go"
				cmd := exec.Command("git", "add", filepath)
				cmd.Run()
				cmd = exec.Command("git", "commit", "-m", "Checkpoint")
				cmd.Run()

				// Then modify
				content := `package main

func hello() string {
	return "HELLO"
}
`
				os.WriteFile(filepath, []byte(content), 0644)
				return filepath
			},
			wantAdded:   1, // One line changed
			wantDeleted: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := tt.setup()

			added, deleted, lineRanges, err := getDetailedDiff(filepath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if added != tt.wantAdded {
				t.Errorf("Added lines = %d, want %d", added, tt.wantAdded)
			}

			if deleted != tt.wantDeleted {
				t.Errorf("Deleted lines = %d, want %d", deleted, tt.wantDeleted)
			}

			if added > 0 && len(lineRanges) == 0 {
				t.Error("Expected line ranges when lines were added")
			}
		})
	}
}

func TestGetLineRangesFromDiff(t *testing.T) {
	tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	tests := []struct {
		name      string
		setup     func() string
		wantCount int
		wantErr   bool
	}{
		{
			name: "Single line addition",
			setup: func() string {
				filepath := "test.go"
				content := `package main

func hello() string {
	return "hello"
}

func added() {}
`
				os.WriteFile(filepath, []byte(content), 0644)
				return filepath
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "New file (no diff available)",
			setup: func() string {
				filepath := "multi.go"
				content := `package main

func a() {}

func b() {}

func c() {}
`
				os.WriteFile(filepath, []byte(content), 0644)
				return filepath
			},
			wantCount: 0, // New files have no git diff
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := tt.setup()

			ranges, err := getLineRangesFromDiff(filepath)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.wantCount == 0 {
				// For new files, we expect no ranges from git diff
				if len(ranges) != 0 {
					t.Errorf("Line ranges count = %d, want 0 for new files", len(ranges))
				}
			} else if len(ranges) < tt.wantCount {
				t.Errorf("Line ranges count = %d, want at least %d", len(ranges), tt.wantCount)
			}
		})
	}
}

func TestGetFileList(t *testing.T) {
	tests := []struct {
		name    string
		changes map[string]tracker.Change
		want    int
	}{
		{
			name:    "empty map",
			changes: map[string]tracker.Change{},
			want:    0,
		},
		{
			name: "single file",
			changes: map[string]tracker.Change{
				"main.go": {Added: 10},
			},
			want: 1,
		},
		{
			name: "multiple files",
			changes: map[string]tracker.Change{
				"main.go":  {Added: 10},
				"utils.go": {Added: 5, Deleted: 2},
				"api.go":   {Deleted: 3},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileList(tt.changes)
			if len(result) != tt.want {
				t.Errorf("getFileList() returned %d files, want %d", len(result), tt.want)
			}
			// 全てのキーが結果に含まれることを確認
			resultMap := make(map[string]bool)
			for _, f := range result {
				resultMap[f] = true
			}
			for key := range tt.changes {
				if !resultMap[key] {
					t.Errorf("getFileList() missing file: %s", key)
				}
			}
		})
	}
}

func TestDetectChangesFromSnapshot_MixedChanges(t *testing.T) {
	lastCheckpoint := &tracker.CheckpointV2{
		Snapshot: map[string]tracker.FileSnapshot{
			"unchanged.go":  {Hash: "aaa", Lines: 10},
			"modified.go":   {Hash: "bbb", Lines: 20},
			"deleted.go":    {Hash: "ccc", Lines: 15},
		},
	}
	currentSnapshot := map[string]tracker.FileSnapshot{
		"unchanged.go":  {Hash: "aaa", Lines: 10}, // no change
		"modified.go":   {Hash: "ddd", Lines: 25}, // modified (hash changed)
		"new.go":        {Hash: "eee", Lines: 8},  // new file
	}

	changes, err := detectChangesFromSnapshot(lastCheckpoint, currentSnapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// unchanged.go should not be in changes
	if _, exists := changes["unchanged.go"]; exists {
		t.Error("unchanged.go should NOT be in changes")
	}

	// modified.go should be in changes
	if _, exists := changes["modified.go"]; !exists {
		t.Error("modified.go should be in changes")
	}

	// deleted.go should be in changes
	if ch, exists := changes["deleted.go"]; !exists {
		t.Error("deleted.go should be in changes")
	} else if ch.Deleted != 15 {
		t.Errorf("deleted.go Deleted = %d, want 15", ch.Deleted)
	}

	// new.go should be in changes
	if ch, exists := changes["new.go"]; !exists {
		t.Error("new.go should be in changes")
	} else if ch.Added != 8 {
		t.Errorf("new.go Added = %d, want 8", ch.Added)
	}

	// Total: 3 changes (modified, deleted, new)
	if len(changes) != 3 {
		t.Errorf("total changes = %d, want 3", len(changes))
	}
}

func TestDetectChangesFromSnapshot(t *testing.T) {
	tests := []struct {
		name            string
		lastCheckpoint  *tracker.CheckpointV2
		currentSnapshot map[string]tracker.FileSnapshot
		wantChanges     int
	}{
		{
			name:           "No previous checkpoint",
			lastCheckpoint: nil,
			currentSnapshot: map[string]tracker.FileSnapshot{
				"test.go": {Hash: "abc123", Lines: 10},
			},
			wantChanges: 0,
		},
		{
			name: "New file added",
			lastCheckpoint: &tracker.CheckpointV2{
				Snapshot: map[string]tracker.FileSnapshot{},
			},
			currentSnapshot: map[string]tracker.FileSnapshot{
				"new.go": {Hash: "def456", Lines: 20},
			},
			wantChanges: 1,
		},
		{
			name: "File modified",
			lastCheckpoint: &tracker.CheckpointV2{
				Snapshot: map[string]tracker.FileSnapshot{
					"test.go": {Hash: "abc123", Lines: 10},
				},
			},
			currentSnapshot: map[string]tracker.FileSnapshot{
				"test.go": {Hash: "xyz789", Lines: 15},
			},
			wantChanges: 1,
		},
		{
			name: "File deleted",
			lastCheckpoint: &tracker.CheckpointV2{
				Snapshot: map[string]tracker.FileSnapshot{
					"old.go": {Hash: "abc123", Lines: 10},
				},
			},
			currentSnapshot: map[string]tracker.FileSnapshot{},
			wantChanges:     1,
		},
		{
			name: "No changes",
			lastCheckpoint: &tracker.CheckpointV2{
				Snapshot: map[string]tracker.FileSnapshot{
					"test.go": {Hash: "abc123", Lines: 10},
				},
			},
			currentSnapshot: map[string]tracker.FileSnapshot{
				"test.go": {Hash: "abc123", Lines: 10},
			},
			wantChanges: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := detectChangesFromSnapshot(tt.lastCheckpoint, tt.currentSnapshot)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(changes) != tt.wantChanges {
				t.Errorf("Changes count = %d, want %d", len(changes), tt.wantChanges)
			}
		})
	}
}
