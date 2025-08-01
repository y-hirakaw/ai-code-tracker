package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDiffAnalyzer(t *testing.T) {
	analyzer := NewDiffAnalyzer()
	if analyzer == nil {
		t.Error("Expected NewDiffAnalyzer to return non-nil analyzer")
	}
}

func TestIsGitRepository(t *testing.T) {
	analyzer := NewDiffAnalyzer()

	// Test in a non-git directory
	tmpDir := filepath.Join(os.TempDir(), "test-non-git")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to non-git directory
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	if analyzer.IsGitRepository() {
		t.Error("Expected IsGitRepository to return false in non-git directory")
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	err = cmd.Run()
	if err != nil {
		t.Skip("Git not available, skipping git repository test")
		return
	}

	if !analyzer.IsGitRepository() {
		t.Error("Expected IsGitRepository to return true in git directory")
	}
}

func setupTestGitRepo(t *testing.T) (string, func()) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-git-repo")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	err = cmd.Run()
	if err != nil {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
		t.Skip("Git not available, skipping test")
	}

	// Configure git user
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("initial content\n"), 0644)
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "Initial commit").Run()

	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestGetLatestCommit(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	commit, err := analyzer.GetLatestCommit()
	if err != nil {
		t.Fatalf("Failed to get latest commit: %v", err)
	}

	// Verify commit hash format (40 characters)
	if len(commit) != 40 {
		t.Errorf("Expected commit hash to be 40 characters, got %d", len(commit))
	}

	// Verify it's hexadecimal
	for _, c := range commit {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("Invalid character in commit hash: %c", c)
		}
	}
}

func TestGetCurrentBranch(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	branch, err := analyzer.GetCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	// Default branch should be "master" or "main"
	if branch != "master" && branch != "main" {
		t.Errorf("Expected branch to be 'master' or 'main', got '%s'", branch)
	}

	// Create and switch to new branch
	exec.Command("git", "checkout", "-b", "test-branch").Run()

	branch, err = analyzer.GetCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch after switch: %v", err)
	}

	if branch != "test-branch" {
		t.Errorf("Expected branch to be 'test-branch', got '%s'", branch)
	}
}

func TestGetCommitAuthor(t *testing.T) {
	_, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	// Get latest commit
	commit, _ := analyzer.GetLatestCommit()

	author, err := analyzer.GetCommitAuthor(commit)
	if err != nil {
		t.Fatalf("Failed to get commit author: %v", err)
	}

	if author != "Test User" {
		t.Errorf("Expected author to be 'Test User', got '%s'", author)
	}

	// Test with invalid commit
	_, err = analyzer.GetCommitAuthor("invalid-commit-hash")
	if err == nil {
		t.Error("Expected error for invalid commit hash")
	}
}

func TestGetCommitDiff(t *testing.T) {
	tmpDir, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	// Make a change and commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("initial content\nmodified content\n"), 0644)
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "Modified test.txt").Run()

	// Get latest commit
	commit, _ := analyzer.GetLatestCommit()

	diff, err := analyzer.GetCommitDiff(commit)
	if err != nil {
		t.Fatalf("Failed to get commit diff: %v", err)
	}

	// Verify diff contains expected content
	if !strings.Contains(diff, "Modified test.txt") {
		t.Error("Expected diff to contain commit message")
	}

	if !strings.Contains(diff, "+modified content") {
		t.Error("Expected diff to contain added line")
	}

	if !strings.Contains(diff, "test.txt") {
		t.Error("Expected diff to contain file name")
	}
}

func TestGetDiff(t *testing.T) {
	tmpDir, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	// Get first commit
	firstCommit, _ := analyzer.GetLatestCommit()

	// Make changes and create second commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("initial content\nmodified content\n"), 0644)
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "Second commit").Run()

	secondCommit, _ := analyzer.GetLatestCommit()

	// Get diff between commits
	diff, err := analyzer.GetDiff(firstCommit, secondCommit)
	if err != nil {
		t.Fatalf("Failed to get diff: %v", err)
	}

	// Verify diff content
	if !strings.Contains(diff, "+modified content") {
		t.Error("Expected diff to contain added line")
	}

	if !strings.Contains(diff, "test.txt") {
		t.Error("Expected diff to contain file name")
	}

	// Test with invalid commits
	_, err = analyzer.GetDiff("invalid1", "invalid2")
	if err == nil {
		t.Error("Expected error for invalid commit hashes")
	}
}

func TestGetDiffWithNewFile(t *testing.T) {
	tmpDir, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	// Get first commit
	firstCommit, _ := analyzer.GetLatestCommit()

	// Create new file and commit
	newFile := filepath.Join(tmpDir, "new.txt")
	os.WriteFile(newFile, []byte("new file content\n"), 0644)
	exec.Command("git", "add", "new.txt").Run()
	exec.Command("git", "commit", "-m", "Added new file").Run()

	secondCommit, _ := analyzer.GetLatestCommit()

	// Get diff
	diff, err := analyzer.GetDiff(firstCommit, secondCommit)
	if err != nil {
		t.Fatalf("Failed to get diff with new file: %v", err)
	}

	// Verify new file in diff
	if !strings.Contains(diff, "new.txt") {
		t.Error("Expected diff to contain new file name")
	}

	if !strings.Contains(diff, "+new file content") {
		t.Error("Expected diff to contain new file content")
	}
}

func TestGetDiffWithDeletedFile(t *testing.T) {
	tmpDir, cleanup := setupTestGitRepo(t)
	defer cleanup()

	analyzer := NewDiffAnalyzer()

	// Create additional file
	extraFile := filepath.Join(tmpDir, "extra.txt")
	os.WriteFile(extraFile, []byte("extra content\n"), 0644)
	exec.Command("git", "add", "extra.txt").Run()
	exec.Command("git", "commit", "-m", "Added extra file").Run()

	firstCommit, _ := analyzer.GetLatestCommit()

	// Delete file and commit
	os.Remove(extraFile)
	exec.Command("git", "add", "-A").Run()
	exec.Command("git", "commit", "-m", "Deleted extra file").Run()

	secondCommit, _ := analyzer.GetLatestCommit()

	// Get diff
	diff, err := analyzer.GetDiff(firstCommit, secondCommit)
	if err != nil {
		t.Fatalf("Failed to get diff with deleted file: %v", err)
	}

	// Verify deleted file in diff
	if !strings.Contains(diff, "extra.txt") {
		t.Error("Expected diff to contain deleted file name")
	}

	if !strings.Contains(diff, "-extra content") {
		t.Error("Expected diff to contain deleted content")
	}
}
