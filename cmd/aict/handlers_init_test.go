package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
)

// setStdinReader replaces stdinReader for testing and returns a cleanup function
func setStdinReader(input string) func() {
	original := stdinReader
	stdinReader = bufio.NewReader(strings.NewReader(input))
	return func() { stdinReader = original }
}

func TestHandleInitV2_CreatesConfig(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	defer setStdinReader("n\n")()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleInitV2()
	if err != nil {
		t.Fatalf("handleInitV2() error = %v", err)
	}

	configPath := filepath.Join(tmpDir, ".git", "aict", "config.json")
	testutil.AssertFileExists(t, configPath)
}

func TestHandleInitV2_ConfigValues(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	defer setStdinReader("n\n")()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleInitV2()
	if err != nil {
		t.Fatalf("handleInitV2() error = %v", err)
	}

	_, cfg, err := loadStorageAndConfig()
	if err != nil {
		t.Fatalf("loadStorageAndConfig() error = %v", err)
	}

	if cfg.DefaultAuthor != "Test User" {
		t.Errorf("DefaultAuthor = %q, want %q", cfg.DefaultAuthor, "Test User")
	}
	if cfg.TargetAIPercentage != 80.0 {
		t.Errorf("TargetAIPercentage = %v, want 80.0", cfg.TargetAIPercentage)
	}
	if len(cfg.TrackedExtensions) == 0 {
		t.Error("TrackedExtensions should not be empty")
	}
	if len(cfg.AIAgents) == 0 {
		t.Error("AIAgents should not be empty")
	}
}

func TestHandleInitV2_Idempotent(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	defer setStdinReader("n\nn\n")()
	if err := handleInitV2(); err != nil {
		t.Fatalf("first handleInitV2() error = %v", err)
	}
	stdinReader = bufio.NewReader(strings.NewReader("n\n"))
	if err := handleInitV2(); err != nil {
		t.Fatalf("second handleInitV2() error = %v", err)
	}
}

func TestHandleInitV2_InteractiveYes(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	defer setStdinReader("y\n")()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleInitV2()
	if err != nil {
		t.Fatalf("handleInitV2() error = %v", err)
	}

	// hooks が作成されている
	preHookPath := filepath.Join(tmpDir, ".git", "aict", "hooks", "pre-tool-use.sh")
	testutil.AssertFileExists(t, preHookPath)
}

func TestHandleInitV2_InteractiveDefaultIsYes(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	defer setStdinReader("\n")()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleInitV2()
	if err != nil {
		t.Fatalf("handleInitV2() error = %v", err)
	}

	preHookPath := filepath.Join(tmpDir, ".git", "aict", "hooks", "pre-tool-use.sh")
	testutil.AssertFileExists(t, preHookPath)
}

func TestHandleInitV2WithOptions_WithHooks(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// --with-hooks: stdinを読まずにhooksを設定
	err := handleInitV2WithOptions(true)
	if err != nil {
		t.Fatalf("handleInitV2WithOptions(true) error = %v", err)
	}

	configPath := filepath.Join(tmpDir, ".git", "aict", "config.json")
	testutil.AssertFileExists(t, configPath)

	preHookPath := filepath.Join(tmpDir, ".git", "aict", "hooks", "pre-tool-use.sh")
	testutil.AssertFileExists(t, preHookPath)
	postHookPath := filepath.Join(tmpDir, ".git", "aict", "hooks", "post-tool-use.sh")
	testutil.AssertFileExists(t, postHookPath)
}
