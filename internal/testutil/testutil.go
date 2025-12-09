package testutil

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// TempGitRepo creates a temporary git repository for testing
// Uses t.TempDir() for automatic cleanup
func TempGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to setup git repo: %v", err)
		}
	}

	return dir
}

// InitAICT initializes AICT configuration in a directory
func InitAICT(t *testing.T, dir string) string {
	t.Helper()
	aictDir := filepath.Join(dir, ".git", "aict")
	checkpointsDir := filepath.Join(aictDir, "checkpoints")

	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		t.Fatalf("Failed to create AICT dir: %v", err)
	}

	// Create minimal config
	config := map[string]interface{}{
		"target_ai_percentage": 80,
		"tracked_extensions":   []string{".go", ".py", ".js", ".ts"},
		"exclude_patterns":     []string{"*_test.go", "vendor/*"},
		"default_author":       "human",
		"ai_agents":            []string{"Claude", "AI"},
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(aictDir, "config.json")
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return aictDir
}

// CreateTestFile creates a file with specified content in the directory
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	filePath := filepath.Join(dir, filename)

	// Create parent directories if needed
	parentDir := filepath.Dir(filePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", filename, err)
	}

	return filePath
}

// GitCommit creates a git commit in the specified directory
func GitCommit(t *testing.T, dir, message string) string {
	t.Helper()

	// Stage all files
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to git add: %v", err)
	}

	// Commit
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = dir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to git commit: %v", err)
	}

	// Get commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = dir
	output, err := hashCmd.Output()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}

	return string(output)[:7] // Return short hash
}

// CreateTestCheckpoint creates a test checkpoint with specified parameters
func CreateTestCheckpoint(author string, authorType tracker.AuthorType) *tracker.CheckpointV2 {
	return &tracker.CheckpointV2{
		Author:   author,
		Type:     authorType,
		Metadata: make(map[string]string),
		Changes:  make(map[string]tracker.Change),
		Snapshot: make(map[string]tracker.FileSnapshot),
	}
}

// AssertError asserts that an error occurred
func AssertError(t *testing.T, err error, context string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", context)
	}
}

// AssertNoError asserts that no error occurred
func AssertNoError(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", context, err)
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, got, want interface{}, context string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", context, got, want)
	}
}

// AssertFileExists asserts that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("Expected file to not exist: %s", path)
	}
}
