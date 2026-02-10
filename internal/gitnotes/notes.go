package gitnotes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

const (
	// AuthorshipNotesRef is the SPEC.md準拠 git notes reference
	AuthorshipNotesRef = "refs/aict/authorship"
)

// NotesManager handles git notes operations
type NotesManager struct {
	executor gitexec.Executor
}

// NewNotesManager creates a new NotesManager
func NewNotesManager() *NotesManager {
	return &NotesManager{
		executor: gitexec.NewExecutor(),
	}
}

// NewNotesManagerWithExecutor creates a NotesManager with a custom executor (for testing)
func NewNotesManagerWithExecutor(executor gitexec.Executor) *NotesManager {
	return &NotesManager{
		executor: executor,
	}
}

// GetCurrentCommit returns the current HEAD commit hash
func GetCurrentCommit() (string, error) {
	executor := gitexec.NewExecutor()
	output, err := executor.Run("rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return output, nil
}

// isNoteNotFound checks if the error indicates a missing note (not a real git error)
func isNoteNotFound(err error) bool {
	return strings.Contains(err.Error(), "no note found")
}

// SPEC.md準拠: Authorship Log操作

// AddAuthorshipLog adds an AuthorshipLog to Git notes
func (nm *NotesManager) AddAuthorshipLog(log *tracker.AuthorshipLog) error {
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal authorship log: %w", err)
	}

	// refs/aict/authorship/ に保存
	_, err = nm.executor.Run("notes", "--ref="+AuthorshipNotesRef, "add", "-f", "-m", string(data), log.Commit)
	if err != nil {
		return fmt.Errorf("failed to add authorship log: %w", err)
	}

	return nil
}

// GetAuthorshipLog retrieves an AuthorshipLog from Git notes
func (nm *NotesManager) GetAuthorshipLog(commitHash string) (*tracker.AuthorshipLog, error) {
	output, err := nm.executor.Run("notes", "--ref="+AuthorshipNotesRef, "show", commitHash)
	if err != nil {
		if isNoteNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get authorship log for %s: %w", commitHash, err)
	}

	var log tracker.AuthorshipLog
	if err := json.Unmarshal([]byte(output), &log); err != nil {
		return nil, fmt.Errorf("failed to parse authorship log: %w", err)
	}

	return &log, nil
}

// ListAuthorshipLogs lists all commits that have Authorship Logs
func (nm *NotesManager) ListAuthorshipLogs() (map[string]*tracker.AuthorshipLog, error) {
	output, err := nm.executor.Run("notes", "--ref="+AuthorshipNotesRef, "list")
	if err != nil {
		// No notes exist yet
		return make(map[string]*tracker.AuthorshipLog), nil
	}

	logs := make(map[string]*tracker.AuthorshipLog)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format: "noteHash commitHash"
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}

		commitHash := parts[1]
		log, err := nm.GetAuthorshipLog(commitHash)
		if err != nil {
			continue
		}

		if log != nil {
			logs[commitHash] = log
		}
	}

	return logs, nil
}
