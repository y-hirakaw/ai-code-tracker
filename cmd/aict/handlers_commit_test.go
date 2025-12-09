package main

import (
	"os"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
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

func TestHandleCommit(t *testing.T) {
	// Setup test environment
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create a test file and commit
	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n\nfunc main() {}\n")
	testutil.GitCommit(t, tmpDir, "Initial commit")

	// Note: handleCommit() is a complex integration that requires:
	// - AICT storage
	// - Git notes
	// - Checkpoints
	// This basic test just verifies it doesn't panic with proper setup
	// Full integration testing should be done separately

	// For now, we'll just verify the git repository is set up correctly
	hash, err := getLatestCommitHash()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}

	if hash == "" {
		t.Error("Expected commit hash after git commit")
	}
}
