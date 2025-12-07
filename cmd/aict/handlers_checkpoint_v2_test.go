package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// setupTestRepo creates a temporary git repository for testing
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "aict-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git
	cmds := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if err := cmd.Run(); err != nil {
			cleanup()
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func hello() string {
	return "hello"
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		cleanup()
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.go")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		cleanup()
		t.Fatalf("Failed to git commit: %v", err)
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
			wantAdded:   6,
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
