package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type DiffAnalyzer struct{}

func NewDiffAnalyzer() *DiffAnalyzer {
	return &DiffAnalyzer{}
}

func (d *DiffAnalyzer) GetDiff(fromCommit, toCommit string) (string, error) {
	cmd := exec.Command("git", "diff", fromCommit, toCommit)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	return string(output), nil
}

func (d *DiffAnalyzer) GetCommitAuthor(commit string) (string, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%an", commit)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit author: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DiffAnalyzer) GetLatestCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DiffAnalyzer) GetCommitDiff(commit string) (string, error) {
	cmd := exec.Command("git", "show", commit)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit diff: %w", err)
	}
	return string(output), nil
}

func (d *DiffAnalyzer) GetCurrentBranch() (string, error) {
	// Check if we're in a git repository first
	if !d.IsGitRepository() {
		return "", fmt.Errorf("not in a git repository")
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Handle common error cases
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", fmt.Errorf("git command failed (exit code %d): %s", 
				exitError.ExitCode(), strings.TrimSpace(string(exitError.Stderr)))
		}
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	
	// Handle special cases
	if branch == "" {
		return "", fmt.Errorf("empty branch name returned from git")
	}
	
	// Handle detached HEAD state
	if branch == "HEAD" {
		// Try to get commit hash instead
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
		hashOutput, hashErr := cmd.Output()
		if hashErr != nil {
			return "detached-HEAD", nil
		}
		return fmt.Sprintf("detached-HEAD@%s", strings.TrimSpace(string(hashOutput))), nil
	}
	
	// Normalize branch name (remove any problematic characters)
	branch = normalizeBranchName(branch)
	
	return branch, nil
}

// normalizeBranchName cleans up branch names for consistent storage
func normalizeBranchName(branch string) string {
	// Remove any leading/trailing whitespace
	branch = strings.TrimSpace(branch)
	
	// Handle remote branch references (origin/feature/test -> feature/test)
	if strings.Contains(branch, "/") {
		parts := strings.Split(branch, "/")
		if len(parts) >= 2 && (parts[0] == "origin" || parts[0] == "upstream") {
			// Join remaining parts
			branch = strings.Join(parts[1:], "/")
		}
	}
	
	return branch
}

func (d *DiffAnalyzer) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
