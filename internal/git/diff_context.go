package git

import (
	"context"
	"os/exec"
	"strings"
	"time"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
)

// ContextAwareDiffAnalyzer extends DiffAnalyzer with context support
type ContextAwareDiffAnalyzer struct {
	*DiffAnalyzer
	timeout time.Duration
}

// NewContextAwareDiffAnalyzer creates a new context-aware diff analyzer
func NewContextAwareDiffAnalyzer(timeout time.Duration) *ContextAwareDiffAnalyzer {
	return &ContextAwareDiffAnalyzer{
		DiffAnalyzer: NewDiffAnalyzer(),
		timeout:      timeout,
	}
}

// GetDiffWithContext retrieves diff with context support
func (d *ContextAwareDiffAnalyzer) GetDiffWithContext(ctx context.Context, fromCommit, toCommit string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", fromCommit, toCommit)
	
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.NewGitError("GetDiffWithContext", "operation timed out", err)
		}
		if ctx.Err() == context.Canceled {
			return "", errors.NewGitError("GetDiffWithContext", "operation was canceled", err)
		}
		return "", errors.NewGitError("GetDiffWithContext", "failed to get diff", err)
	}
	
	return string(output), nil
}

// GetCommitAuthorWithContext retrieves commit author with context support
func (d *ContextAwareDiffAnalyzer) GetCommitAuthorWithContext(ctx context.Context, commit string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "show", "-s", "--format=%an", commit)
	
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", errors.NewGitError("GetCommitAuthorWithContext", "context error", ctx.Err())
		}
		return "", errors.NewGitError("GetCommitAuthorWithContext", "failed to get commit author", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// GetLatestCommitWithContext retrieves the latest commit hash with context support
func (d *ContextAwareDiffAnalyzer) GetLatestCommitWithContext(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", errors.NewGitError("GetLatestCommitWithContext", "context error", ctx.Err())
		}
		return "", errors.NewGitError("GetLatestCommitWithContext", "failed to get latest commit", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// GetCommitDiffWithContext retrieves diff for a specific commit with context support
func (d *ContextAwareDiffAnalyzer) GetCommitDiffWithContext(ctx context.Context, commit string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "show", commit)
	
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", errors.NewGitError("GetCommitDiffWithContext", "context error", ctx.Err())
		}
		return "", errors.NewGitError("GetCommitDiffWithContext", "failed to get commit diff", err)
	}
	
	return string(output), nil
}

// GetCurrentBranchWithContext retrieves the current branch name with context support
func (d *ContextAwareDiffAnalyzer) GetCurrentBranchWithContext(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", errors.NewGitError("GetCurrentBranchWithContext", "context error", ctx.Err())
		}
		return "", errors.NewGitError("GetCurrentBranchWithContext", "failed to get current branch", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// IsGitRepositoryWithContext checks if current directory is a git repository with context support
func (d *ContextAwareDiffAnalyzer) IsGitRepositoryWithContext(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--git-dir")
	
	err := cmd.Run()
	if err != nil {
		if ctx.Err() != nil {
			return false, errors.NewGitError("IsGitRepositoryWithContext", "context error", ctx.Err())
		}
		return false, nil
	}
	
	return true, nil
}

// RunWithTimeout executes a function with the configured timeout
func (d *ContextAwareDiffAnalyzer) RunWithTimeout(fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	
	return fn(ctx)
}