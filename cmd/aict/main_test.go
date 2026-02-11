package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

func TestDebugf_Enabled(t *testing.T) {
	// debugEnabled を一時的にtrueに設定
	origEnabled := debugEnabled
	defer func() { debugEnabled = origEnabled }()
	debugEnabled = true

	// stderr をキャプチャ
	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	debugf("test message %d", 42)

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("expected debug output when debugEnabled=true")
	}
	if !bytes.Contains([]byte(output), []byte("[DEBUG] test message 42")) {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestDebugf_Disabled(t *testing.T) {
	origEnabled := debugEnabled
	defer func() { debugEnabled = origEnabled }()
	debugEnabled = false

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	debugf("should not appear")

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output != "" {
		t.Errorf("expected no output when debugEnabled=false, got: %q", output)
	}
}

func TestGetGitUserName(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		if args[0] == "config" && args[1] == "user.name" {
			return "Test Developer", nil
		}
		return "", fmt.Errorf("unknown command")
	}
	newExecutor = func() gitexec.Executor { return mock }

	name := getGitUserName()
	if name != "Test Developer" {
		t.Errorf("getGitUserName() = %q, want %q", name, "Test Developer")
	}
}

func TestGetGitUserName_Error(t *testing.T) {
	origExecutor := newExecutor
	defer func() { newExecutor = origExecutor }()

	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("not configured")
	}
	newExecutor = func() gitexec.Executor { return mock }

	name := getGitUserName()
	if name != "" {
		t.Errorf("getGitUserName() = %q, want empty string on error", name)
	}
}

func TestMainCommand_Version(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	os.Args = []string{"aict", "version"}
	exitFunc = func(code int) { /* no-op */ }

	// main() を呼んでpanicしないことを確認
	main()
}

func TestMainCommand_Help(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	os.Args = []string{"aict", "help"}
	exitFunc = func(code int) { /* no-op */ }

	main()
}

// exitPanic is used as a sentinel value to halt main() via panic/recover
type exitPanic int

func TestMainCommand_NoArgs(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	exitCode := -1
	os.Args = []string{"aict"}
	exitFunc = func(code int) {
		exitCode = code
		panic(exitPanic(code)) // halt execution
	}

	func() {
		defer func() { recover() }()
		main()
	}()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestMainCommand_Unknown(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	exitCode := -1
	os.Args = []string{"aict", "nonexistent"}
	exitFunc = func(code int) {
		exitCode = code
		panic(exitPanic(code))
	}

	func() {
		defer func() { recover() }()
		main()
	}()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestMainCommand_SyncError(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	exitCode := -1
	os.Args = []string{"aict", "sync"}
	exitFunc = func(code int) {
		exitCode = code
		panic(exitPanic(code))
	}

	func() {
		defer func() { recover() }()
		main()
	}()

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for sync error, got %d", exitCode)
	}
}

func TestMainCommand_Checkpoint(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	os.Args = []string{"aict", "checkpoint"}
	exitFunc = func(code int) { /* no-op */ }

	// checkpoint without --author uses default, should not panic
	main()
}

func TestMainCommand_VersionFlags(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	exitFunc = func(code int) { /* no-op */ }

	for _, flag := range []string{"--version", "-v"} {
		os.Args = []string{"aict", flag}
		main()
	}
}

func TestMainCommand_HelpFlag(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	exitFunc = func(code int) { /* no-op */ }

	os.Args = []string{"aict", "--help"}
	main()

	os.Args = []string{"aict", "-h"}
	main()
}

func TestMainCommand_DebugError(t *testing.T) {
	origArgs := os.Args
	origExit := exitFunc
	defer func() {
		os.Args = origArgs
		exitFunc = origExit
	}()

	exitCode := -1
	os.Args = []string{"aict", "debug"}
	exitFunc = func(code int) {
		exitCode = code
		panic(exitPanic(code))
	}

	func() {
		defer func() { recover() }()
		main()
	}()

	if exitCode != 1 {
		t.Errorf("expected exit code 1 for debug error, got %d", exitCode)
	}
}
