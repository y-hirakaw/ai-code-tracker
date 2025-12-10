package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

// NumstatEntry represents a single file's statistics from git diff --numstat
type NumstatEntry struct {
	Filepath string
	Added    int
	Deleted  int
}

// ParseNumstat parses git diff --numstat output into structured data
// Input format: "added\tdeleted\tfilepath" (one per line)
// Handles binary files (shows "-") and file renames ("path1 => path2")
func ParseNumstat(output string) (map[string][2]int, error) {
	result := make(map[string][2]int)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format: "added\tdeleted\tfilepath"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // Skip binary files which show "-"
		}

		deleted, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		// Handle renames: "path1 => path2" becomes just "path2"
		filepath := strings.Join(parts[2:], " ")
		if idx := strings.Index(filepath, " => "); idx != -1 {
			filepath = filepath[idx+4:]
		}

		result[filepath] = [2]int{added, deleted}
	}

	return result, nil
}

// GetNumstatBetweenCommits runs git diff --numstat between two commits
func GetNumstatBetweenCommits(executor gitexec.Executor, fromCommit, toCommit string) (map[string][2]int, error) {
	output, err := executor.Run("diff", "--numstat", fromCommit, toCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

	return ParseNumstat(output)
}

// GetNumstatFromHead runs git diff --numstat from HEAD (for uncommitted changes)
func GetNumstatFromHead(executor gitexec.Executor) (map[string][2]int, error) {
	output, err := executor.Run("diff", "HEAD", "--numstat")
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

	return ParseNumstat(output)
}
