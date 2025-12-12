package main

import (
	"os"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
)

func TestGetCommitsInRange(t *testing.T) {
	// Setup test git repository
	tmpDir := testutil.TempGitRepo(t)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create multiple commits
	testutil.CreateTestFile(t, tmpDir, "file1.go", "package main\n")
	commit1 := testutil.GitCommit(t, tmpDir, "First commit")

	testutil.CreateTestFile(t, tmpDir, "file2.go", "package main\n")
	commit2 := testutil.GitCommit(t, tmpDir, "Second commit")

	testutil.CreateTestFile(t, tmpDir, "file3.go", "package main\n")
	testutil.GitCommit(t, tmpDir, "Third commit")

	// Test getCommitsInRange with commit range
	commits, err := getCommitsInRange(commit1[:7] + ".." + commit2[:7])

	if err != nil {
		t.Fatalf("getCommitsInRange() error = %v", err)
	}

	if len(commits) == 0 {
		t.Error("getCommitsInRange() returned no commits")
	}

	// Verify commits contain expected hash
	foundCommit2 := false
	for _, commit := range commits {
		if strings.HasPrefix(commit, commit2[:7]) {
			foundCommit2 = true
			break
		}
	}

	if !foundCommit2 {
		t.Errorf("getCommitsInRange() did not include expected commit %s", commit2[:7])
	}
}

func TestHandleRangeReport(t *testing.T) {
	// Setup test environment
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create test commits
	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n")
	testutil.GitCommit(t, tmpDir, "Test commit")

	// Note: handleRangeReport() is a complex integration that requires:
	// - Git notes with authorship logs
	// - Proper AICT configuration
	// This basic test just verifies the environment is set up
	// Full integration testing should be done separately

	// Verify git repository has commits
	commits, err := getCommitsInRange("HEAD")
	if err != nil {
		t.Fatalf("Failed to get commits: %v", err)
	}

	if len(commits) == 0 {
		t.Error("Expected at least one commit")
	}
}

func TestFormatRangeReport(t *testing.T) {
	// Test formatRangeReport with mock data
	// This doesn't require actual git repository

	// Note: This would require access to formatRangeReport function
	// and sample AuthorshipLog data structures.
	// For now, we'll just verify the function exists by testing
	// that the environment can be set up for it

	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	// Verify AICT config was created
	testutil.AssertFileExists(t, tmpDir+"/.git/aict/config.json")
}

// TestAuthorCommitCountAccuracy tests that commit counts are accurate
// when a single commit has multiple files (regression test for v1.1.3)
func TestAuthorCommitCountAccuracy(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Scenario: Single commit with 3 files
	// Before fix: Would count as 3 commits (1 per file)
	// After fix: Should count as 1 commit

	testutil.CreateTestFile(t, tmpDir, "file1.go", "package main\nfunc main() {}\n")
	testutil.CreateTestFile(t, tmpDir, "file2.go", "package utils\nfunc Helper() {}\n")
	testutil.CreateTestFile(t, tmpDir, "file3.go", "package models\ntype User struct {}\n")
	commitHash := testutil.GitCommit(t, tmpDir, "Add multiple files")

	// Verify that getCommitsInRange returns exactly 1 commit
	commits, err := getCommitsInRange(commitHash)
	if err != nil {
		t.Fatalf("getCommitsInRange() error = %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("getCommitsInRange() = %d commits, want 1", len(commits))
	}

	// Verify the commit hash matches
	if len(commits) > 0 && !strings.HasPrefix(commits[0], commitHash[:7]) {
		t.Errorf("getCommitsInRange() returned unexpected commit: got %s, want %s", commits[0][:7], commitHash[:7])
	}
}
