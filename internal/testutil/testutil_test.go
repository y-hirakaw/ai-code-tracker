package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestTempGitRepo(t *testing.T) {
	dir := TempGitRepo(t)

	// Verify .git directory exists
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Fatalf("Expected .git directory to exist in %s", dir)
	}

	// Verify git config is set
	AssertFileExists(t, filepath.Join(gitDir, "config"))
}

func TestInitAICT(t *testing.T) {
	dir := TempGitRepo(t)
	aictDir := InitAICT(t, dir)

	// Verify AICT directory structure
	AssertFileExists(t, filepath.Join(aictDir, "config.json"))
	AssertFileExists(t, filepath.Join(aictDir, "checkpoints"))
}

func TestCreateTestFile(t *testing.T) {
	dir := t.TempDir()
	content := "package main\n\nfunc main() {}\n"

	filePath := CreateTestFile(t, dir, "test.go", content)

	// Verify file exists and has correct content
	AssertFileExists(t, filePath)

	data, err := os.ReadFile(filePath)
	AssertNoError(t, err, "reading test file")

	if string(data) != content {
		t.Fatalf("Expected content %q, got %q", content, string(data))
	}
}

func TestCreateTestCheckpoint(t *testing.T) {
	checkpoint := CreateTestCheckpoint("TestAuthor", tracker.AuthorTypeHuman)

	AssertEqual(t, checkpoint.Author, "TestAuthor", "checkpoint author")
	AssertEqual(t, checkpoint.Type, tracker.AuthorTypeHuman, "checkpoint type")

	if checkpoint.Metadata == nil {
		t.Fatal("Expected Metadata to be initialized")
	}
	if checkpoint.Changes == nil {
		t.Fatal("Expected Changes to be initialized")
	}
	if checkpoint.Snapshot == nil {
		t.Fatal("Expected Snapshot to be initialized")
	}
}

func TestAssertError(t *testing.T) {
	// This would fail the test if error is nil
	// We just verify the function exists and compiles
	err := os.ErrNotExist
	AssertError(t, err, "test error")
}

func TestAssertNoError(t *testing.T) {
	AssertNoError(t, nil, "no error test")
}

func TestGitCommit(t *testing.T) {
	dir := TempGitRepo(t)

	// Create a test file
	CreateTestFile(t, dir, "test.go", "package main")

	// Commit the file
	hash := GitCommit(t, dir, "Initial commit")

	// Verify commit hash is returned (7 chars)
	if len(hash) != 7 {
		t.Fatalf("Expected commit hash of length 7, got %d: %s", len(hash), hash)
	}
}
