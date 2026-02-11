package gitnotes

import (
	"encoding/json"
	"fmt"
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
			if args[3] == "commit1" {
				return string(json1), nil
			}
			if args[3] == "commit2" {
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
