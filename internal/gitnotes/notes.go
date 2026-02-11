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

	// refs/aict/authorship/ に保存（"--" でオプション終端を明示し、コミットハッシュのオプション注入を防止）
	_, err = nm.executor.Run("notes", "--ref="+AuthorshipNotesRef, "add", "-f", "-m", string(data), "--", log.Commit)
	if err != nil {
		return fmt.Errorf("failed to add authorship log: %w", err)
	}

	return nil
}

// GetAuthorshipLog retrieves an AuthorshipLog from Git notes
func (nm *NotesManager) GetAuthorshipLog(commitHash string) (*tracker.AuthorshipLog, error) {
	output, err := nm.executor.Run("notes", "--ref="+AuthorshipNotesRef, "show", "--", commitHash)
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

// authorshipLogMarker はgit logの出力でコミットを区切るマーカー
const authorshipLogMarker = "__AICT_HASH__"

// GetAuthorshipLogsForRange はコミット範囲内の全Authorship Logを1回のgit呼び出しで取得します。
// git log --notes を使用し、個別に git notes show を呼ぶN+1問題を解消します。
func (nm *NotesManager) GetAuthorshipLogsForRange(rangeSpec string) (map[string]*tracker.AuthorshipLog, error) {
	output, err := nm.executor.Run(
		"log",
		"--no-standard-notes",
		"--notes="+AuthorshipNotesRef,
		"--format="+authorshipLogMarker+"%H%n%N",
		"--end-of-options",
		rangeSpec,
	)
	if err != nil {
		return make(map[string]*tracker.AuthorshipLog), nil
	}

	return parseAuthorshipLogsOutput(output), nil
}

// parseAuthorshipLogsOutput は git log --notes --format の出力をパースします。
func parseAuthorshipLogsOutput(output string) map[string]*tracker.AuthorshipLog {
	logs := make(map[string]*tracker.AuthorshipLog)

	sections := strings.Split(output, authorshipLogMarker)
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		// 最初の行がコミットハッシュ、残りがノート内容（JSON）
		newlineIdx := strings.Index(section, "\n")
		if newlineIdx == -1 {
			continue // ノートなし
		}

		commitHash := strings.TrimSpace(section[:newlineIdx])
		noteContent := strings.TrimSpace(section[newlineIdx+1:])

		if noteContent == "" {
			continue
		}

		var log tracker.AuthorshipLog
		if err := json.Unmarshal([]byte(noteContent), &log); err != nil {
			continue
		}

		logs[commitHash] = &log
	}

	return logs
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
