package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

func TestHandleSync_MissingSubcommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"aict", "sync"}

	err := handleSync()
	if err == nil {
		t.Fatal("expected error for missing subcommand")
	}
	if !strings.Contains(err.Error(), "sync subcommand required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleSync_UnknownSubcommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"aict", "sync", "invalid"}

	err := handleSync()
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown subcommand") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleSyncPush(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		return "", nil
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleSyncPush()
	if err != nil {
		t.Fatalf("handleSyncPush() error = %v", err)
	}

	calls := mock.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("expected 1 git call, got %d", len(calls))
	}
	if calls[0].Args[0] != "push" {
		t.Errorf("expected 'push' command, got %q", calls[0].Args[0])
	}
}

func TestHandleSyncFetch(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		return "", nil
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleSyncFetch()
	if err != nil {
		t.Fatalf("handleSyncFetch() error = %v", err)
	}

	calls := mock.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("expected 1 git call, got %d", len(calls))
	}
	if calls[0].Args[0] != "fetch" {
		t.Errorf("expected 'fetch' command, got %q", calls[0].Args[0])
	}
}

func TestHandleSyncPush_Error(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("remote not found")
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleSyncPush()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "pushing authorship logs") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleSyncFetch_Error(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("remote not found")
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleSyncFetch()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "fetching authorship logs") {
		t.Errorf("unexpected error: %v", err)
	}
}
