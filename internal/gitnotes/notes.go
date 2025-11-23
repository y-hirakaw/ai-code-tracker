package gitnotes

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	// NotesRef is the git notes reference for AICT
	NotesRef = "refs/notes/aict"
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
	ref string
}

// NewNotesManager creates a new NotesManager
func NewNotesManager() *NotesManager {
	return &NotesManager{
		ref: NotesRef,
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
	cmd := exec.Command("git", "notes", "--ref="+nm.ref, "add", "-f", "-m", string(data), "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add git note: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetNote retrieves the git note for a specific commit
func (nm *NotesManager) GetNote(commitHash string) (*AIEditNote, error) {
	cmd := exec.Command("git", "notes", "--ref="+nm.ref, "show", commitHash)
	output, err := cmd.Output()
	if err != nil {
		// No note exists for this commit
		return nil, nil
	}

	var note AIEditNote
	if err := json.Unmarshal(output, &note); err != nil {
		return nil, fmt.Errorf("failed to parse note: %w", err)
	}

	return &note, nil
}

// ListNotes lists all commits that have AICT notes
func (nm *NotesManager) ListNotes() (map[string]*AIEditNote, error) {
	cmd := exec.Command("git", "notes", "--ref="+nm.ref, "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	notes := make(map[string]*AIEditNote)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

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
	cmd := exec.Command("git", "notes", "--ref="+nm.ref, "remove", commitHash)
	_, err := cmd.CombinedOutput()
	// Ignore error if note doesn't exist
	return err
}

// GetCurrentCommit returns the current HEAD commit hash
func GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
