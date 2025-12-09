package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

type DiffAnalyzer struct{
	executor gitexec.Executor
}

func NewDiffAnalyzer() *DiffAnalyzer {
	return &DiffAnalyzer{
		executor: gitexec.NewExecutor(),
	}
}

// NewDiffAnalyzerWithExecutor creates a DiffAnalyzer with a custom executor (for testing)
func NewDiffAnalyzerWithExecutor(executor gitexec.Executor) *DiffAnalyzer {
	return &DiffAnalyzer{
		executor: executor,
	}
}

func (d *DiffAnalyzer) GetDiff(fromCommit, toCommit string) (string, error) {
	output, err := d.executor.Run("diff", fromCommit, toCommit)
	if err != nil {
		return "", fmt.Errorf("failed to get git diff: %w", err)
	}
	return output, nil
}

func (d *DiffAnalyzer) GetCommitAuthor(commit string) (string, error) {
	output, err := d.executor.Run("show", "-s", "--format=%an", commit)
	if err != nil {
		return "", fmt.Errorf("failed to get commit author: %w", err)
	}
	return output, nil
}

func (d *DiffAnalyzer) GetLatestCommit() (string, error) {
	output, err := d.executor.Run("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit: %w", err)
	}
	return output, nil
}

func (d *DiffAnalyzer) GetCommitDiff(commit string) (string, error) {
	output, err := d.executor.Run("show", commit)
	if err != nil {
		return "", fmt.Errorf("failed to get commit diff: %w", err)
	}
	return output, nil
}

func (d *DiffAnalyzer) GetCurrentBranch() (string, error) {
	// Check if we're in a git repository first
	if !d.IsGitRepository() {
		return "", fmt.Errorf("not in a git repository")
	}

	output, err := d.executor.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		// Handle common error cases
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return "", fmt.Errorf("git command failed (exit code %d): %s",
				exitError.ExitCode(), strings.TrimSpace(string(exitError.Stderr)))
		}
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := output

	// Handle special cases
	if branch == "" {
		return "", fmt.Errorf("empty branch name returned from git")
	}

	// Handle detached HEAD state
	if branch == "HEAD" {
		return d.handleDetachedHead()
	}

	// Normalize branch name (remove any problematic characters)
	branch = normalizeBranchName(branch)

	return branch, nil
}

// handleDetachedHead returns a branch name for detached HEAD state
func (d *DiffAnalyzer) handleDetachedHead() (string, error) {
	hashOutput, hashErr := d.executor.Run("rev-parse", "--short", "HEAD")
	if hashErr != nil {
		return "detached-HEAD", nil
	}
	return fmt.Sprintf("detached-HEAD@%s", hashOutput), nil
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
	_, err := d.executor.Run("rev-parse", "--git-dir")
	return err == nil
}
