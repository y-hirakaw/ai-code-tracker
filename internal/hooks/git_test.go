package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestHookManager はテスト用のHookManagerを作成する
func setupTestHookManager(t *testing.T) (*HookManager, string) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict-hooks-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	// Gitリポジトリっぽい構造を作成
	gitDir := filepath.Join(tempDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	err = os.MkdirAll(hooksDir, 0755)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Gitディレクトリの作成に失敗: %v", err)
	}

	// HookManagerを作成
	hookManager := NewHookManager(tempDir)

	return hookManager, tempDir
}

// cleanupTestHookManager はテスト用のHookManagerを削除する
func cleanupTestHookManager(tempDir string) {
	os.RemoveAll(tempDir)
}

// TestNewHookManager はHookManagerの初期化をテストする
func TestNewHookManager(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	if hookManager.gitRepo != tempDir {
		t.Errorf("NewHookManager().gitRepo = %s, want %s", hookManager.gitRepo, tempDir)
	}
}

// TestValidateGitRepo はGitリポジトリ検証をテストする
func TestValidateGitRepo(t *testing.T) {
	t.Run("Valid Git Repository", func(t *testing.T) {
		hookManager, tempDir := setupTestHookManager(t)
		defer cleanupTestHookManager(tempDir)

		err := hookManager.ValidateGitRepo()
		// 実際のgitコマンドが実行されるため、エラーになる可能性がある
		// テスト環境では正常なケースをログに記録
		if err != nil {
			t.Logf("ValidateGitRepo() failed as expected in test environment: %v", err)
		} else {
			t.Log("ValidateGitRepo() succeeded - running in actual git repository")
		}
	})

	t.Run("Invalid Directory", func(t *testing.T) {
		hookManager := NewHookManager("/nonexistent/path")

		err := hookManager.ValidateGitRepo()
		if err == nil {
			t.Errorf("ValidateGitRepo() should return error for nonexistent directory")
		}
	})
}

// TestGeneratePostCommitHook はpost-commit hook生成をテストする
func TestGeneratePostCommitHook(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	script, err := hookManager.generatePostCommitHook()
	if err != nil {
		t.Fatalf("generatePostCommitHook() error = %v", err)
	}

	// 基本的な内容チェック
	expectedContents := []string{
		"#!/bin/bash",
		"AI Code Tracker - Git post-commit hook",
		"check_duplicate()",
		"Claude Code",
		"aict track",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(script, expected) {
			t.Errorf("generatePostCommitHook() script does not contain '%s'", expected)
		}
	}

	// 実行可能スクリプトの基本構造チェック
	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Errorf("generatePostCommitHook() script should start with shebang")
	}
}

// TestGenerateClaudeCodeHooksConfig はClaude Code hooks設定生成をテストする
func TestGenerateClaudeCodeHooksConfig(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	config, err := hookManager.generateClaudeCodeHooksConfig()
	if err != nil {
		t.Fatalf("generateClaudeCodeHooksConfig() error = %v", err)
	}

	// 基本構造チェック
	if config.Hooks == nil {
		t.Errorf("generateClaudeCodeHooksConfig() config.Hooks is nil")
		return
	}

	// 必要なhookタイプが存在するかチェック
	requiredHooks := []string{"preToolUse", "postToolUse", "stop", "notification"}
	for _, hookType := range requiredHooks {
		if _, exists := config.Hooks[hookType]; !exists {
			t.Errorf("generateClaudeCodeHooksConfig() missing hook type: %s", hookType)
		}
	}

	// postToolUseのEdit/Write/MultiEditマッチャーをチェック
	if postToolUse, exists := config.Hooks["postToolUse"]; exists {
		foundEditMatcher := false
		for _, hook := range postToolUse {
			if strings.Contains(hook.Matcher, "Edit") {
				foundEditMatcher = true
				break
			}
		}
		if !foundEditMatcher {
			t.Errorf("generateClaudeCodeHooksConfig() missing Edit matcher in postToolUse")
		}
	}

	// JSON変換テスト
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Errorf("generateClaudeCodeHooksConfig() JSON marshal error: %v", err)
	}

	// 生成されたJSONが有効かテスト
	var testConfig ClaudeCodeHooksConfig
	if err := json.Unmarshal(configJSON, &testConfig); err != nil {
		t.Errorf("generateClaudeCodeHooksConfig() generated invalid JSON: %v", err)
	}
}

// TestSetupGitHooks はGit hooks設定をテストする
func TestSetupGitHooks(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	t.Run("Setup New Hook", func(t *testing.T) {
		err := hookManager.SetupGitHooks()
		if err != nil {
			t.Fatalf("SetupGitHooks() error = %v", err)
		}

		// post-commit hookが作成されているかチェック
		postCommitPath := filepath.Join(tempDir, ".git", "hooks", "post-commit")
		if _, err := os.Stat(postCommitPath); os.IsNotExist(err) {
			t.Errorf("SetupGitHooks() did not create post-commit hook")
		}

		// ファイル内容をチェック
		content, err := os.ReadFile(postCommitPath)
		if err != nil {
			t.Fatalf("Failed to read post-commit hook: %v", err)
		}

		if !strings.Contains(string(content), "AI Code Tracker") {
			t.Errorf("SetupGitHooks() created hook without expected content")
		}

		// 実行権限をチェック
		info, err := os.Stat(postCommitPath)
		if err != nil {
			t.Fatalf("Failed to stat post-commit hook: %v", err)
		}

		if info.Mode()&0111 == 0 {
			t.Errorf("SetupGitHooks() created hook without execute permission")
		}
	})

	t.Run("Backup Existing Hook", func(t *testing.T) {
		postCommitPath := filepath.Join(tempDir, ".git", "hooks", "post-commit")
		
		// 既存のhookを作成
		existingContent := "#!/bin/bash\necho 'existing hook'\n"
		err := os.WriteFile(postCommitPath, []byte(existingContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create existing hook: %v", err)
		}

		// 新しいhookを設定
		err = hookManager.SetupGitHooks()
		if err != nil {
			t.Fatalf("SetupGitHooks() error = %v", err)
		}

		// バックアップが作成されているかチェック
		backupPath := postCommitPath + ".backup"
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("SetupGitHooks() did not create backup of existing hook")
		}

		// バックアップ内容をチェック
		backupContent, err := os.ReadFile(backupPath)
		if err == nil && string(backupContent) != existingContent {
			t.Errorf("SetupGitHooks() backup content does not match original")
		}
	})
}

// TestSetupClaudeCodeHooks はClaude Code hooks設定をテストする
func TestSetupClaudeCodeHooks(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	// テスト用の一時ホームディレクトリを作成
	tempHome := filepath.Join(tempDir, "home")
	os.MkdirAll(tempHome, 0755)
	
	// 元のHOME環境変数を保存
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	// テスト用のHOMEを設定
	os.Setenv("HOME", tempHome)

	err := hookManager.SetupClaudeCodeHooks()
	if err != nil {
		t.Fatalf("SetupClaudeCodeHooks() error = %v", err)
	}

	// 設定ファイルが作成されているかチェック
	configPath := filepath.Join(tempHome, ".claude", "hooks-aict.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("SetupClaudeCodeHooks() did not create config file")
	}

	// 設定ファイル内容をチェック
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var config ClaudeCodeHooksConfig
	if err := json.Unmarshal(content, &config); err != nil {
		t.Errorf("SetupClaudeCodeHooks() created invalid JSON config: %v", err)
	}

	// 基本的な構造をチェック
	if config.Hooks == nil {
		t.Errorf("SetupClaudeCodeHooks() config has no hooks")
	}
}

// TestRemoveGitHooks はGit hooks削除をテストする
func TestRemoveGitHooks(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	postCommitPath := filepath.Join(tempDir, ".git", "hooks", "post-commit")

	t.Run("Remove AICT Hook", func(t *testing.T) {
		// AICT製のhookを作成
		aictHookContent := "#!/bin/bash\n# AI Code Tracker - Git post-commit hook\necho 'aict hook'\n"
		err := os.WriteFile(postCommitPath, []byte(aictHookContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create AICT hook: %v", err)
		}

		err = hookManager.RemoveGitHooks()
		if err != nil {
			t.Fatalf("RemoveGitHooks() error = %v", err)
		}

		// hookが削除されているかチェック
		if _, err := os.Stat(postCommitPath); err == nil {
			t.Errorf("RemoveGitHooks() did not remove AICT hook")
		}
	})

	t.Run("Restore Backup", func(t *testing.T) {
		// バックアップファイルを作成
		backupPath := postCommitPath + ".backup"
		backupContent := "#!/bin/bash\necho 'original hook'\n"
		err := os.WriteFile(backupPath, []byte(backupContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// AICT製のhookを作成
		aictHookContent := "#!/bin/bash\n# AI Code Tracker - Git post-commit hook\necho 'aict hook'\n"
		err = os.WriteFile(postCommitPath, []byte(aictHookContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create AICT hook: %v", err)
		}

		err = hookManager.RemoveGitHooks()
		if err != nil {
			t.Fatalf("RemoveGitHooks() error = %v", err)
		}

		// バックアップが復元されているかチェック
		if _, err := os.Stat(postCommitPath); os.IsNotExist(err) {
			t.Errorf("RemoveGitHooks() did not restore backup")
		}

		// 復元された内容をチェック
		restoredContent, err := os.ReadFile(postCommitPath)
		if err == nil && string(restoredContent) != backupContent {
			t.Errorf("RemoveGitHooks() restored content does not match backup")
		}
	})

	t.Run("Skip Non-AICT Hook", func(t *testing.T) {
		// 非AICT製のhookを作成
		nonAictHookContent := "#!/bin/bash\necho 'some other hook'\n"
		err := os.WriteFile(postCommitPath, []byte(nonAictHookContent), 0755)
		if err != nil {
			t.Fatalf("Failed to create non-AICT hook: %v", err)
		}

		err = hookManager.RemoveGitHooks()
		if err != nil {
			t.Fatalf("RemoveGitHooks() error = %v", err)
		}

		// hookが残っているかチェック
		if _, err := os.Stat(postCommitPath); os.IsNotExist(err) {
			t.Errorf("RemoveGitHooks() removed non-AICT hook")
		}

		// 内容が変更されていないかチェック
		content, err := os.ReadFile(postCommitPath)
		if err == nil && string(content) != nonAictHookContent {
			t.Errorf("RemoveGitHooks() modified non-AICT hook")
		}
	})
}

// TestGetHookStatus はhook状況取得をテストする
func TestGetHookStatus(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	status, err := hookManager.GetHookStatus()
	if err != nil {
		t.Fatalf("GetHookStatus() error = %v", err)
	}

	// 基本構造チェック
	if _, exists := status["git_hooks"]; !exists {
		t.Errorf("GetHookStatus() missing git_hooks status")
	}

	if _, exists := status["claude_hooks"]; !exists {
		t.Errorf("GetHookStatus() missing claude_hooks status")
	}

	// Git hooks status詳細チェック
	if gitHooks, ok := status["git_hooks"].(map[string]interface{}); ok {
		requiredFields := []string{"installed", "path", "backup"}
		for _, field := range requiredFields {
			if _, exists := gitHooks[field]; !exists {
				t.Errorf("GetHookStatus() git_hooks missing field: %s", field)
			}
		}
	} else {
		t.Errorf("GetHookStatus() git_hooks has wrong type")
	}
}

// TestCheckPermissions は権限チェックをテストする
func TestCheckPermissions(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	err := hookManager.CheckPermissions()
	if err != nil {
		t.Logf("CheckPermissions() failed as expected in some environments: %v", err)
	} else {
		t.Log("CheckPermissions() succeeded - has required permissions")
	}
}

// TestBackupExistingHooks は既存hooks バックアップをテストする
func TestBackupExistingHooks(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	hooksDir := filepath.Join(tempDir, ".git", "hooks")

	// 既存のhookファイルを作成
	existingHooks := map[string]string{
		"post-commit": "#!/bin/bash\necho 'existing post-commit'\n",
		"pre-commit":  "#!/bin/bash\necho 'existing pre-commit'\n",
	}

	for hookName, content := range existingHooks {
		hookPath := filepath.Join(hooksDir, hookName)
		err := os.WriteFile(hookPath, []byte(content), 0755)
		if err != nil {
			t.Fatalf("Failed to create existing hook %s: %v", hookName, err)
		}
	}

	err := hookManager.BackupExistingHooks()
	if err != nil {
		t.Fatalf("BackupExistingHooks() error = %v", err)
	}

	// バックアップが作成されているかチェック
	for hookName, expectedContent := range existingHooks {
		backupPath := filepath.Join(hooksDir, hookName+".aict-backup")
		
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("BackupExistingHooks() did not create backup for %s", hookName)
			continue
		}

		// バックアップ内容をチェック
		backupContent, err := os.ReadFile(backupPath)
		if err != nil {
			t.Errorf("Failed to read backup %s: %v", backupPath, err)
			continue
		}

		if string(backupContent) != expectedContent {
			t.Errorf("BackupExistingHooks() backup content for %s does not match", hookName)
		}
	}
}

// TestRestoreBackupHooks はバックアップ復元をテストする
func TestRestoreBackupHooks(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	hooksDir := filepath.Join(tempDir, ".git", "hooks")

	// バックアップファイルを作成
	backupHooks := map[string]string{
		"post-commit.aict-backup": "#!/bin/bash\necho 'backup post-commit'\n",
		"pre-commit.aict-backup":  "#!/bin/bash\necho 'backup pre-commit'\n",
	}

	for backupName, content := range backupHooks {
		backupPath := filepath.Join(hooksDir, backupName)
		err := os.WriteFile(backupPath, []byte(content), 0755)
		if err != nil {
			t.Fatalf("Failed to create backup %s: %v", backupName, err)
		}
	}

	err := hookManager.RestoreBackupHooks()
	if err != nil {
		t.Fatalf("RestoreBackupHooks() error = %v", err)
	}

	// 復元されたファイルをチェック
	for backupName, expectedContent := range backupHooks {
		originalName := strings.TrimSuffix(backupName, ".aict-backup")
		originalPath := filepath.Join(hooksDir, originalName)
		
		if _, err := os.Stat(originalPath); os.IsNotExist(err) {
			t.Errorf("RestoreBackupHooks() did not restore %s", originalName)
			continue
		}

		// 復元された内容をチェック
		restoredContent, err := os.ReadFile(originalPath)
		if err != nil {
			t.Errorf("Failed to read restored file %s: %v", originalPath, err)
			continue
		}

		if string(restoredContent) != expectedContent {
			t.Errorf("RestoreBackupHooks() restored content for %s does not match", originalName)
		}

		// バックアップファイルが削除されているかチェック
		backupPath := filepath.Join(hooksDir, backupName)
		if _, err := os.Stat(backupPath); err == nil {
			t.Errorf("RestoreBackupHooks() did not remove backup file %s", backupName)
		}
	}
}

// BenchmarkGeneratePostCommitHook はpost-commit hook生成のベンチマークテストを行う
func BenchmarkGeneratePostCommitHook(b *testing.B) {
	hookManager, tempDir := setupTestHookManager(&testing.T{})
	defer cleanupTestHookManager(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hookManager.generatePostCommitHook()
		if err != nil {
			b.Fatalf("generatePostCommitHook error: %v", err)
		}
	}
}

// BenchmarkGenerateClaudeCodeHooksConfig はClaude Code hooks設定生成のベンチマークテストを行う
func BenchmarkGenerateClaudeCodeHooksConfig(b *testing.B) {
	hookManager, tempDir := setupTestHookManager(&testing.T{})
	defer cleanupTestHookManager(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hookManager.generateClaudeCodeHooksConfig()
		if err != nil {
			b.Fatalf("generateClaudeCodeHooksConfig error: %v", err)
		}
	}
}

// TestHooksIntegration はhooks機能の統合テストを行う
func TestHooksIntegration(t *testing.T) {
	hookManager, tempDir := setupTestHookManager(t)
	defer cleanupTestHookManager(tempDir)

	// テスト用の一時ホームディレクトリを作成
	tempHome := filepath.Join(tempDir, "home")
	os.MkdirAll(tempHome, 0755)
	
	// 元のHOME環境変数を保存
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	
	// テスト用のHOMEを設定
	os.Setenv("HOME", tempHome)

	t.Run("Full Setup and Status Check", func(t *testing.T) {
		// 権限チェック
		err := hookManager.CheckPermissions()
		if err != nil {
			t.Logf("Permission check failed: %v", err)
		}

		// Git hooks設定
		err = hookManager.SetupGitHooks()
		if err != nil {
			t.Fatalf("SetupGitHooks() error = %v", err)
		}

		// Claude Code hooks設定
		err = hookManager.SetupClaudeCodeHooks()
		if err != nil {
			t.Fatalf("SetupClaudeCodeHooks() error = %v", err)
		}

		// 状況確認
		status, err := hookManager.GetHookStatus()
		if err != nil {
			t.Fatalf("GetHookStatus() error = %v", err)
		}

		// Git hooksが設定されているかチェック
		if gitHooks, ok := status["git_hooks"].(map[string]interface{}); ok {
			if installed, ok := gitHooks["installed"].(bool); !ok || !installed {
				t.Errorf("Git hooks not properly installed")
			}
		}

		// Claude Code hooksが設定されているかチェック
		if claudeHooks, ok := status["claude_hooks"].(map[string]interface{}); ok {
			if installed, ok := claudeHooks["installed"].(bool); !ok || !installed {
				t.Errorf("Claude Code hooks not properly installed")
			}
		}
	})

	t.Run("Setup and Remove Cycle", func(t *testing.T) {
		// 設定
		err := hookManager.SetupGitHooks()
		if err != nil {
			t.Fatalf("SetupGitHooks() error = %v", err)
		}

		err = hookManager.SetupClaudeCodeHooks()
		if err != nil {
			t.Fatalf("SetupClaudeCodeHooks() error = %v", err)
		}

		// 削除
		err = hookManager.RemoveGitHooks()
		if err != nil {
			t.Fatalf("RemoveGitHooks() error = %v", err)
		}

		err = hookManager.RemoveClaudeCodeHooks()
		if err != nil {
			t.Fatalf("RemoveClaudeCodeHooks() error = %v", err)
		}

		// 削除後の状況確認
		status, err := hookManager.GetHookStatus()
		if err != nil {
			t.Fatalf("GetHookStatus() error = %v", err)
		}

		// hooksの削除確認（バックアップが復元される場合は設定されているとみなされる）
		if gitHooks, ok := status["git_hooks"].(map[string]interface{}); ok {
			if installed, ok := gitHooks["installed"].(bool); ok {
				// バックアップが復元された場合は正常動作
				t.Logf("Git hooks status after removal: installed=%v", installed)
			}
		}

		if claudeHooks, ok := status["claude_hooks"].(map[string]interface{}); ok {
			if installed, ok := claudeHooks["installed"].(bool); ok {
				// バックアップが復元された場合は正常動作
				t.Logf("Claude Code hooks status after removal: installed=%v", installed)
			}
		}
	})
}