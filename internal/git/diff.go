package git

import (
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
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

func (d *DiffAnalyzer) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}
