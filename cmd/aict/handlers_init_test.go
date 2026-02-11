package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
)

func TestHandleInitV2_CreatesConfig(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleInitV2()
	if err != nil {
		t.Fatalf("handleInitV2() error = %v", err)
	}

	// config.json が作成されていることを確認
	configPath := filepath.Join(tmpDir, ".git", "aict", "config.json")
	testutil.AssertFileExists(t, configPath)
}

func TestHandleInitV2_ConfigValues(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	err := handleInitV2()
	if err != nil {
		t.Fatalf("handleInitV2() error = %v", err)
	}

	// loadStorageAndConfig で設定を検証
	_, cfg, err := loadStorageAndConfig()
	if err != nil {
		t.Fatalf("loadStorageAndConfig() error = %v", err)
	}

	// TempGitRepo は user.name を "Test User" に設定
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

	// 2回実行してもエラーにならない
	if err := handleInitV2(); err != nil {
		t.Fatalf("first handleInitV2() error = %v", err)
	}
	if err := handleInitV2(); err != nil {
		t.Fatalf("second handleInitV2() error = %v", err)
	}
}
