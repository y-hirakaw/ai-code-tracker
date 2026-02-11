package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestHandleDebug_MissingSubcommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"aict", "debug"}

	err := handleDebug()
	if err == nil {
		t.Fatal("expected error for missing subcommand")
	}
	if !strings.Contains(err.Error(), "debug subcommand required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleDebug_UnknownSubcommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"aict", "debug", "invalid"}

	err := handleDebug()
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
	if !strings.Contains(err.Error(), "unknown debug subcommand") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleDebug_DispatchShow(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	os.Args = []string{"aict", "debug", "show"}
	err := handleDebug()
	if err != nil {
		t.Fatalf("handleDebug() show error = %v", err)
	}
}

func TestHandleDebugShow_NoCheckpoints(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleDebugShow()
	if err != nil {
		t.Fatalf("handleDebugShow() error = %v", err)
	}
}

func TestHandleDebugClean_NoCheckpoints(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleDebugClean()
	if err != nil {
		t.Fatalf("handleDebugClean() error = %v", err)
	}
}

func TestHandleDebugShow_WithCheckpoints(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// チェックポイントを保存
	store, _, err := loadStorageAndConfig()
	if err != nil {
		t.Fatalf("loadStorageAndConfig() error = %v", err)
	}

	cp := testutil.CreateTestCheckpoint("human", "human")
	cp.Changes["test.go"] = tracker.Change{Added: 10, Deleted: 2, Lines: [][]int{{1, 10}}}
	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}

	// チェックポイントが表示されること
	err = handleDebugShow()
	if err != nil {
		t.Fatalf("handleDebugShow() error = %v", err)
	}
}

func TestHandleDebugClean_WithCheckpoints(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// チェックポイントを保存
	store, _, err := loadStorageAndConfig()
	if err != nil {
		t.Fatalf("loadStorageAndConfig() error = %v", err)
	}

	cp := testutil.CreateTestCheckpoint("claude", "ai")
	cp.Changes["main.go"] = tracker.Change{Added: 5}
	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("SaveCheckpoint() error = %v", err)
	}

	// クリーン実行
	err = handleDebugClean()
	if err != nil {
		t.Fatalf("handleDebugClean() error = %v", err)
	}

	// クリーン後はチェックポイントが0件
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		t.Fatalf("LoadCheckpoints() error = %v", err)
	}
	if len(checkpoints) != 0 {
		t.Errorf("expected 0 checkpoints after clean, got %d", len(checkpoints))
	}
}

func TestDisplayCheckpoint(t *testing.T) {
	cp := testutil.CreateTestCheckpoint("human", "human")
	cp.Changes["test.go"] = tracker.Change{Added: 10, Deleted: 2}
	cp.Metadata["message"] = "test checkpoint"

	// panicしないことを確認（出力はstdout）
	displayCheckpoint(1, cp)
}

func TestHandleDebugClearNotes_NoAictRefs(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		if args[0] == "show-ref" {
			// aictに無関係なrefのみ返す
			return "abc123 refs/heads/main\ndef456 refs/heads/feature", nil
		}
		return "", nil
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleDebugClearNotes()
	if err != nil {
		t.Fatalf("handleDebugClearNotes() error = %v", err)
	}

	// update-ref が呼ばれないことを確認
	for _, call := range mock.GetCalls("Run") {
		if len(call.Args) > 0 && call.Args[0] == "update-ref" {
			t.Error("update-ref should not be called when no aict refs exist")
		}
	}
}

func TestHandleDebugClearNotes_WithAictRefs(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		if args[0] == "show-ref" {
			return "abc123 refs/aict/authorship\ndef456 refs/notes/aict\nghi789 refs/heads/feature-aict", nil
		}
		if args[0] == "update-ref" {
			return "", nil
		}
		return "", nil
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleDebugClearNotes()
	if err != nil {
		t.Fatalf("handleDebugClearNotes() error = %v", err)
	}

	// aict refsのみ削除され、ブランチは削除されないことを確認
	updateRefCalls := 0
	for _, call := range mock.GetCalls("Run") {
		if len(call.Args) > 0 && call.Args[0] == "update-ref" {
			updateRefCalls++
		}
	}
	// refs/aict/authorship と refs/notes/aict の2つ（refs/heads/feature-aict はブランチなので除外）
	if updateRefCalls != 2 {
		t.Errorf("expected 2 update-ref calls, got %d", updateRefCalls)
	}
}

func TestHandleDebugClearNotes_ShowRefError(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("fatal: not a git repository")
	}
	newExecutor = func() gitexec.Executor { return mock }

	err := handleDebugClearNotes()
	if err == nil {
		t.Fatal("expected error when show-ref fails")
	}
	if !strings.Contains(err.Error(), "Git refs") {
		t.Errorf("unexpected error: %v", err)
	}
}
