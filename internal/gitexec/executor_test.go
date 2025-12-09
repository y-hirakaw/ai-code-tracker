package gitexec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setupGitRepo creates a temporary git repository for testing
func setupGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repository
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize git repo: %v", err)
	}

	// Configure git user
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test User").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@example.com").Run()

	cleanup := func() {
		// t.TempDir() handles cleanup automatically
	}

	return tmpDir, cleanup
}

// createFileAndCommit creates a file and commits it
func createFileAndCommit(t *testing.T, dir, filename, content, message string) {
	t.Helper()

	// Create file
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Git add
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	// Git commit
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}
}

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()
	if executor == nil {
		t.Error("NewExecutor() returned nil")
	}

	_, ok := executor.(*RealExecutor)
	if !ok {
		t.Error("NewExecutor() did not return *RealExecutor")
	}
}

func TestRealExecutor_Run(t *testing.T) {
	// Setup test git repository
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create a test file and commit
	createFileAndCommit(t, tmpDir, "test.txt", "test content\n", "Test commit")

	executor := NewExecutor()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "git status",
			args:    []string{"status", "--short"},
			wantErr: false,
			check:   func(output string) bool { return true }, // Any output is fine
		},
		{
			name:    "git rev-parse HEAD",
			args:    []string{"rev-parse", "HEAD"},
			wantErr: false,
			check: func(output string) bool {
				return len(output) == 40 // SHA-1 hash length
			},
		},
		{
			name:    "git log format",
			args:    []string{"log", "-1", "--format=%s"},
			wantErr: false,
			check: func(output string) bool {
				return output == "Test commit"
			},
		},
		{
			name:    "invalid git command",
			args:    []string{"invalid-command"},
			wantErr: true,
			check:   func(output string) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executor.Run(tt.args...)

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

			if !tt.check(output) {
				t.Errorf("Output check failed for %q: %s", tt.name, output)
			}
		})
	}
}

func TestRealExecutor_RunInDir(t *testing.T) {
	// Setup test git repository
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Create a test file and commit
	createFileAndCommit(t, tmpDir, "test.txt", "test content\n", "Test commit")

	executor := NewExecutor()

	tests := []struct {
		name    string
		dir     string
		args    []string
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "git status in specific dir",
			dir:     tmpDir,
			args:    []string{"status", "--short"},
			wantErr: false,
			check:   func(output string) bool { return true },
		},
		{
			name:    "git rev-parse HEAD in specific dir",
			dir:     tmpDir,
			args:    []string{"rev-parse", "HEAD"},
			wantErr: false,
			check: func(output string) bool {
				return len(output) == 40
			},
		},
		{
			name:    "invalid directory",
			dir:     "/nonexistent/directory",
			args:    []string{"status"},
			wantErr: true,
			check:   func(output string) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := executor.RunInDir(tt.dir, tt.args...)

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

			if !tt.check(output) {
				t.Errorf("Output check failed for %q: %s", tt.name, output)
			}
		})
	}
}

func TestMockExecutor(t *testing.T) {
	mock := NewMockExecutor()

	// Set up mock behavior
	mock.RunFunc = func(args ...string) (string, error) {
		if args[0] == "rev-parse" && args[1] == "HEAD" {
			return "abc123def456", nil
		}
		return "", fmt.Errorf("unknown command")
	}

	// Test Run
	output, err := mock.Run("rev-parse", "HEAD")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output != "abc123def456" {
		t.Errorf("Expected 'abc123def456', got %q", output)
	}

	// Verify call log
	calls := mock.GetCalls("Run")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].Args[0] != "rev-parse" {
		t.Errorf("Expected first arg 'rev-parse', got %q", calls[0].Args[0])
	}
}

func TestMockExecutor_RunInDir(t *testing.T) {
	mock := NewMockExecutor()

	// Set up mock behavior
	mock.RunInDirFunc = func(dir string, args ...string) (string, error) {
		if dir == "/test/dir" && args[0] == "status" {
			return "clean", nil
		}
		return "", fmt.Errorf("unknown command in dir %s", dir)
	}

	// Test RunInDir
	output, err := mock.RunInDir("/test/dir", "status")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output != "clean" {
		t.Errorf("Expected 'clean', got %q", output)
	}

	// Verify call log
	calls := mock.GetCalls("RunInDir")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(calls))
	}

	if calls[0].Dir != "/test/dir" {
		t.Errorf("Expected dir '/test/dir', got %q", calls[0].Dir)
	}
}

func TestMockExecutor_Reset(t *testing.T) {
	mock := NewMockExecutor()

	// Make some calls
	mock.Run("status")
	mock.Run("log")

	if len(mock.CallLog) != 2 {
		t.Errorf("Expected 2 calls before reset, got %d", len(mock.CallLog))
	}

	// Reset
	mock.Reset()

	if len(mock.CallLog) != 0 {
		t.Errorf("Expected 0 calls after reset, got %d", len(mock.CallLog))
	}
}
