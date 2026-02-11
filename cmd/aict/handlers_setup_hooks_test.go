package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/templates"
)

func TestCreateClaudeHooks(t *testing.T) {
	// Create a temp directory to serve as the hooks directory
	hooksDir := t.TempDir()

	// Call createClaudeHooks
	err := createClaudeHooks(hooksDir)
	if err != nil {
		t.Fatalf("createClaudeHooks() error = %v", err)
	}

	// Verify pre-tool-use.sh
	t.Run("pre-tool-use.sh created with correct content", func(t *testing.T) {
		preHookPath := filepath.Join(hooksDir, "pre-tool-use.sh")
		content, err := os.ReadFile(preHookPath)
		if err != nil {
			t.Fatalf("failed to read pre-tool-use.sh: %v", err)
		}
		if string(content) != templates.PreToolUseHook {
			t.Errorf("pre-tool-use.sh content mismatch\ngot length: %d\nwant length: %d", len(content), len(templates.PreToolUseHook))
		}
	})

	// Verify post-tool-use.sh
	t.Run("post-tool-use.sh created with correct content", func(t *testing.T) {
		postHookPath := filepath.Join(hooksDir, "post-tool-use.sh")
		content, err := os.ReadFile(postHookPath)
		if err != nil {
			t.Fatalf("failed to read post-tool-use.sh: %v", err)
		}
		if string(content) != templates.PostToolUseHook {
			t.Errorf("post-tool-use.sh content mismatch\ngot length: %d\nwant length: %d", len(content), len(templates.PostToolUseHook))
		}
	})

	// Verify executable permissions
	t.Run("hooks have executable permissions", func(t *testing.T) {
		for _, name := range []string{"pre-tool-use.sh", "post-tool-use.sh"} {
			hookPath := filepath.Join(hooksDir, name)
			info, err := os.Stat(hookPath)
			if err != nil {
				t.Fatalf("failed to stat %s: %v", name, err)
			}
			perm := info.Mode().Perm()
			// Check that owner execute bit is set (0755 -> 0o100 bit)
			if perm&0100 == 0 {
				t.Errorf("%s: expected executable permission, got %o", name, perm)
			}
		}
	})
}

func TestSetupPostCommitHook_NewHook(t *testing.T) {
	// Create a temp directory structure simulating a git repository
	repoRoot := t.TempDir()
	gitDir := filepath.Join(repoRoot, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Call setupPostCommitHook (no existing hook, so no stdin prompt)
	err := setupPostCommitHook(repoRoot)
	if err != nil {
		t.Fatalf("setupPostCommitHook() error = %v", err)
	}

	// Verify post-commit hook was created
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "post-commit")

	t.Run("post-commit hook file exists", func(t *testing.T) {
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			t.Fatal("post-commit hook was not created")
		}
	})

	t.Run("post-commit hook has correct content", func(t *testing.T) {
		content, err := os.ReadFile(hookPath)
		if err != nil {
			t.Fatalf("failed to read post-commit hook: %v", err)
		}
		if string(content) != templates.PostCommitHook {
			t.Errorf("post-commit hook content mismatch\ngot length: %d\nwant length: %d", len(content), len(templates.PostCommitHook))
		}
	})

	t.Run("post-commit hook has executable permissions", func(t *testing.T) {
		info, err := os.Stat(hookPath)
		if err != nil {
			t.Fatalf("failed to stat post-commit hook: %v", err)
		}
		perm := info.Mode().Perm()
		if perm&0100 == 0 {
			t.Errorf("post-commit hook: expected executable permission, got %o", perm)
		}
	})

	t.Run("hooks directory was auto-created", func(t *testing.T) {
		hooksDir := filepath.Join(repoRoot, ".git", "hooks")
		info, err := os.Stat(hooksDir)
		if err != nil {
			t.Fatalf("hooks directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("hooks path is not a directory")
		}
	})
}

func TestSetupClaudeSettings_NewSettings(t *testing.T) {
	// Create a temp directory simulating a repo root
	repoRoot := t.TempDir()

	// Call setupClaudeSettings (no existing file, so no stdin prompt)
	err := setupClaudeSettings(repoRoot)
	if err != nil {
		t.Fatalf("setupClaudeSettings() error = %v", err)
	}

	settingsPath := filepath.Join(repoRoot, ".claude", "settings.json")

	t.Run("settings.json file exists", func(t *testing.T) {
		if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
			t.Fatal("settings.json was not created")
		}
	})

	t.Run("settings.json has correct content", func(t *testing.T) {
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("failed to read settings.json: %v", err)
		}
		if string(content) != templates.ClaudeSettingsJSON {
			t.Errorf("settings.json content mismatch\ngot length: %d\nwant length: %d", len(content), len(templates.ClaudeSettingsJSON))
		}
	})

	t.Run(".claude directory was auto-created", func(t *testing.T) {
		claudeDir := filepath.Join(repoRoot, ".claude")
		info, err := os.Stat(claudeDir)
		if err != nil {
			t.Fatalf(".claude directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error(".claude path is not a directory")
		}
	})

	t.Run("settings.json has regular file permissions", func(t *testing.T) {
		info, err := os.Stat(settingsPath)
		if err != nil {
			t.Fatalf("failed to stat settings.json: %v", err)
		}
		perm := info.Mode().Perm()
		// settings.json is written with 0644
		if perm&0200 == 0 {
			t.Errorf("settings.json: expected writable permission, got %o", perm)
		}
	})
}
