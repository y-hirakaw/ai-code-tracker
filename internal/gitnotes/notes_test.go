package gitnotes

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestGetCurrentCommit(t *testing.T) {
	// Note: GetCurrentCommit uses NewExecutor internally, so we can't easily mock it
	// without changing the implementation to accept an executor provider or instance.
	// For now, skipping mock test for this specific function as it's a simple wrapper.
	// Or we could refactor GetCurrentCommit to be a method of NotesManager.
}

func TestAddAuthorshipLog(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	log := &tracker.AuthorshipLog{
		Version:   "1.0.0",
		Commit:    "abc1234",
		Timestamp: time.Now(),
		Files:     make(map[string]tracker.FileInfo),
	}

	mockExec.RunFunc = func(args ...string) (string, error) {
		if args[0] != "notes" || args[2] != "add" {
			return "", fmt.Errorf("unexpected command: %v", args)
		}
		return "", nil
	}

	err := nm.AddAuthorshipLog(log)
	if err != nil {
		t.Fatalf("AddAuthorshipLog failed: %v", err)
	}

	if len(mockExec.CallLog) != 1 {
		t.Errorf("Expected 1 git call, got %d", len(mockExec.CallLog))
	}
}

func TestGetAuthorshipLog(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	expectedLog := &tracker.AuthorshipLog{
		Version: "1.0.0",
		Commit:  "abc1234",
	}
	logJSON, _ := json.Marshal(expectedLog)

	mockExec.RunFunc = func(args ...string) (string, error) {
		if args[0] == "notes" && args[2] == "show" {
			return string(logJSON), nil
		}
		return "", fmt.Errorf("unexpected command")
	}

	log, err := nm.GetAuthorshipLog("abc1234")
	if err != nil {
		t.Fatalf("GetAuthorshipLog failed: %v", err)
	}

	if log.Commit != expectedLog.Commit {
		t.Errorf("Expected commit %s, got %s", expectedLog.Commit, log.Commit)
	}
}

func TestGetAuthorshipLog_NotFound(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("error: no note found for object")
	}

	log, err := nm.GetAuthorshipLog("missing")
	if err != nil {
		t.Fatalf("GetAuthorshipLog should not return error for missing note, got: %v", err)
	}
	if log != nil {
		t.Error("Expected nil log for missing note")
	}
}

func TestParseAuthorshipLogsOutput(t *testing.T) {
	log1 := &tracker.AuthorshipLog{
		Version: "1.0",
		Commit:  "abc123",
		Files: map[string]tracker.FileInfo{
			"file1.go": {Authors: []tracker.AuthorInfo{{Name: "human", Type: "human"}}},
		},
	}
	log2 := &tracker.AuthorshipLog{
		Version: "1.0",
		Commit:  "def456",
		Files: map[string]tracker.FileInfo{
			"file2.go": {Authors: []tracker.AuthorInfo{{Name: "claude", Type: "ai"}}},
		},
	}
	json1, _ := json.MarshalIndent(log1, "", "  ")
	json2, _ := json.MarshalIndent(log2, "", "  ")

	tests := []struct {
		name          string
		input         string
		expectedCount int
		expectedKeys  []string
	}{
		{
			name:          "empty input",
			input:         "",
			expectedCount: 0,
		},
		{
			name:          "single commit with note",
			input:         "__AICT_HASH__abc123\n" + string(json1),
			expectedCount: 1,
			expectedKeys:  []string{"abc123"},
		},
		{
			name:          "multiple commits with notes",
			input:         "__AICT_HASH__abc123\n" + string(json1) + "\n__AICT_HASH__def456\n" + string(json2),
			expectedCount: 2,
			expectedKeys:  []string{"abc123", "def456"},
		},
		{
			name:          "commit without note skipped",
			input:         "__AICT_HASH__abc123\n" + string(json1) + "\n__AICT_HASH__no_note_commit\n\n__AICT_HASH__def456\n" + string(json2),
			expectedCount: 2,
			expectedKeys:  []string{"abc123", "def456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAuthorshipLogsOutput(tt.input)

			if len(result) != tt.expectedCount {
				t.Errorf("got %d logs, want %d", len(result), tt.expectedCount)
			}

			for _, key := range tt.expectedKeys {
				if _, exists := result[key]; !exists {
					t.Errorf("missing key %q in result", key)
				}
			}
		})
	}
}

func TestGetAuthorshipLogsForRange(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	log1 := &tracker.AuthorshipLog{Version: "1.0", Commit: "abc123"}
	json1, _ := json.MarshalIndent(log1, "", "  ")

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "__AICT_HASH__abc123\n" + string(json1), nil
	}

	logs, err := nm.GetAuthorshipLogsForRange("HEAD~5..HEAD")
	if err != nil {
		t.Fatalf("GetAuthorshipLogsForRange() error = %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("got %d logs, want 1", len(logs))
	}

	if log, exists := logs["abc123"]; !exists {
		t.Error("missing key abc123")
	} else if log.Commit != "abc123" {
		t.Errorf("got commit %q, want abc123", log.Commit)
	}

	// git log引数の確認
	calls := mockExec.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}
	if calls[0].Args[0] != "log" {
		t.Errorf("expected 'log' command, got %q", calls[0].Args[0])
	}
}

func TestListAuthorshipLogs(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	// Mock list output: noteHash commitHash
	mockListOutput := "note123 commit1\nnote456 commit2"

	// Mock logs
	log1 := &tracker.AuthorshipLog{Commit: "commit1", Version: "1.0"}
	log2 := &tracker.AuthorshipLog{Commit: "commit2", Version: "1.0"}
	json1, _ := json.Marshal(log1)
	json2, _ := json.Marshal(log2)

	mockExec.RunFunc = func(args ...string) (string, error) {
		if args[2] == "list" {
			return mockListOutput, nil
		}
		if args[2] == "show" {
			// args[3] is "--" (end of options marker), args[4] is commitHash
			if args[4] == "commit1" {
				return string(json1), nil
			}
			if args[4] == "commit2" {
				return string(json2), nil
			}
		}
		return "", fmt.Errorf("unexpected command: %v", args)
	}

	logs, err := nm.ListAuthorshipLogs()
	if err != nil {
		t.Fatalf("ListAuthorshipLogs failed: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs, got %d", len(logs))
	}
	if logs["commit1"].Commit != "commit1" {
		t.Errorf("Expected commit1 log")
	}
}

func TestParseAuthorshipLogsOutput_MalformedJSON(t *testing.T) {
	// 不正JSONを含むコミットはスキップされ、正常なコミットは処理される
	validLog := &tracker.AuthorshipLog{Version: "1.0", Commit: "good123"}
	validJSON, _ := json.MarshalIndent(validLog, "", "  ")

	input := "__AICT_HASH__bad456\n{invalid json}\n__AICT_HASH__good123\n" + string(validJSON)
	result := parseAuthorshipLogsOutput(input)

	if len(result) != 1 {
		t.Errorf("got %d logs, want 1 (malformed should be skipped)", len(result))
	}
	if _, exists := result["good123"]; !exists {
		t.Error("valid log should be present")
	}
	if _, exists := result["bad456"]; exists {
		t.Error("malformed log should be skipped")
	}
}

func TestGetAuthorshipLogsForRange_GitError(t *testing.T) {
	// git log がエラーを返す場合、空のmapが返される（エラーではない）
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("fatal: bad revision")
	}

	logs, err := nm.GetAuthorshipLogsForRange("invalid..range")
	if err != nil {
		t.Fatalf("should not return error, got: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected empty map, got %d entries", len(logs))
	}
}

func TestAddAuthorshipLog_GitError(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	log := &tracker.AuthorshipLog{
		Version: "1.0.0",
		Commit:  "abc1234",
	}

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("fatal: could not add note")
	}

	err := nm.AddAuthorshipLog(log)
	if err == nil {
		t.Fatal("expected error from AddAuthorshipLog")
	}
	if !strings.Contains(err.Error(), "failed to add authorship log") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAddAuthorshipLog_SecurityArgs(t *testing.T) {
	// "--" がコミットハッシュの前に渡されることを確認（オプション注入防止）
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	log := &tracker.AuthorshipLog{
		Version: "1.0.0",
		Commit:  "abc1234",
	}

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "", nil
	}

	_ = nm.AddAuthorshipLog(log)

	calls := mockExec.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	args := calls[0].Args
	// "--" がコミットハッシュの直前にあることを確認
	foundSeparator := false
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) && args[i+1] == "abc1234" {
			foundSeparator = true
			break
		}
	}
	if !foundSeparator {
		t.Errorf("expected '--' before commit hash in args: %v", args)
	}
}

func TestGetAuthorshipLog_GitError(t *testing.T) {
	// "no note found" 以外のgitエラーはエラーとして返る
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("fatal: not a git repository")
	}

	_, err := nm.GetAuthorshipLog("abc1234")
	if err == nil {
		t.Fatal("expected error for non-'no note found' git error")
	}
	if !strings.Contains(err.Error(), "failed to get authorship log") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestGetAuthorshipLog_InvalidJSON(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "{invalid json}", nil
	}

	_, err := nm.GetAuthorshipLog("abc1234")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "failed to parse authorship log") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestListAuthorshipLogs_NoNotes(t *testing.T) {
	mockExec := gitexec.NewMockExecutor()
	nm := NewNotesManagerWithExecutor(mockExec)

	mockExec.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("error: no ref found")
	}

	logs, err := nm.ListAuthorshipLogs()
	if err != nil {
		t.Fatalf("should not return error, got: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("expected empty map, got %d entries", len(logs))
	}
}

func TestIsNoteNotFound(t *testing.T) {
	tests := []struct {
		errMsg   string
		expected bool
	}{
		{"error: no note found for object abc123", true},
		{"fatal: not a git repository", false},
		{"no note found", true},
		{"something else entirely", false},
	}

	for _, tt := range tests {
		result := isNoteNotFound(fmt.Errorf(tt.errMsg))
		if result != tt.expected {
			t.Errorf("isNoteNotFound(%q) = %v, want %v", tt.errMsg, result, tt.expected)
		}
	}
}
