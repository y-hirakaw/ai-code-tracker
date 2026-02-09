package gitnotes

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

const (
	// NotesRef is the git notes reference for AICT (legacy)
	NotesRef = "refs/notes/aict"

	// AuthorshipNotesRef is the SPEC.md準拠 git notes reference
	AuthorshipNotesRef = "refs/aict/authorship"
)

// AIEditNote represents AI editing information stored in git notes
type AIEditNote struct {
	Timestamp time.Time         `json:"timestamp"`
	Tool      string            `json:"tool"`      // "claude", "copilot", etc.
	Files     map[string][]int  `json:"files"`     // filepath -> line numbers
	Commit    string            `json:"commit"`    // commit hash when note was created
}

// NotesManager handles git notes operations
type NotesManager struct {
	ref      string
	executor gitexec.Executor
}

// NewNotesManager creates a new NotesManager
func NewNotesManager() *NotesManager {
	return &NotesManager{
		ref:      NotesRef,
		executor: gitexec.NewExecutor(),
	}
}

// NewNotesManagerWithExecutor creates a NotesManager with a custom executor (for testing)
func NewNotesManagerWithExecutor(executor gitexec.Executor) *NotesManager {
	return &NotesManager{
		ref:      NotesRef,
		executor: executor,
	}
}

// AddNote adds a new git note for the current HEAD
func (nm *NotesManager) AddNote(note *AIEditNote) error {
	// Serialize note to JSON
	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal note: %w", err)
	}

	// Add git note
	_, err = nm.executor.Run("notes", "--ref="+nm.ref, "add", "-f", "-m", string(data), "HEAD")
	if err != nil {
		return fmt.Errorf("failed to add git note: %w", err)
	}

	return nil
}

// GetNote retrieves the git note for a specific commit
func (nm *NotesManager) GetNote(commitHash string) (*AIEditNote, error) {
	output, err := nm.executor.Run("notes", "--ref="+nm.ref, "show", commitHash)
	if err != nil {
		if isNoteNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get note for %s: %w", commitHash, err)
	}

	var note AIEditNote
	if err := json.Unmarshal([]byte(output), &note); err != nil {
		return nil, fmt.Errorf("failed to parse note: %w", err)
	}

	return &note, nil
}

// ListNotes lists all commits that have AICT notes
func (nm *NotesManager) ListNotes() (map[string]*AIEditNote, error) {
	output, err := nm.executor.Run("notes", "--ref="+nm.ref, "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	notes := make(map[string]*AIEditNote)
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
		note, err := nm.GetNote(commitHash)
		if err != nil {
			continue
		}

		if note != nil {
			notes[commitHash] = note
		}
	}

	return notes, nil
}

// RemoveNote removes the git note for a specific commit
func (nm *NotesManager) RemoveNote(commitHash string) error {
	_, err := nm.executor.Run("notes", "--ref="+nm.ref, "remove", commitHash)
	// Ignore error if note doesn't exist
	return err
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
