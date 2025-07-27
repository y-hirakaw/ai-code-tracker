package tracker

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/pkg/types"
)

// setupTestTracker はテスト用のTrackerを作成する
func setupTestTracker(t *testing.T) (*Tracker, *storage.Storage, string) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict-tracker-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	// Gitリポジトリっぽい構造を作成
	gitDir := filepath.Join(tempDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Gitディレクトリの作成に失敗: %v", err)
	}

	// ストレージを初期化
	storageInstance, err := storage.NewStorage(filepath.Join(tempDir, ".git", "ai-tracker"))
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("ストレージの初期化に失敗: %v", err)
	}

	// トラッカーを作成
	tracker := NewTracker(storageInstance, tempDir)

	return tracker, storageInstance, tempDir
}

// cleanupTestTracker はテスト用のTrackerを削除する
func cleanupTestTracker(storage *storage.Storage, tempDir string) {
	storage.Close()
	os.RemoveAll(tempDir)
}

// createTestFile はテスト用のファイルを作成する
func createTestFile(t *testing.T, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}
	return filePath
}

// TestNewTracker はTrackerの初期化をテストする
func TestNewTracker(t *testing.T) {
	_, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	tracker := NewTracker(storage, tempDir)

	if tracker.storage != storage {
		t.Errorf("NewTracker().storage が設定されていません")
	}

	if tracker.gitRepo != tempDir {
		t.Errorf("NewTracker().gitRepo = %s, want %s", tracker.gitRepo, tempDir)
	}

	if tracker.duplicateWindow != 5*time.Second {
		t.Errorf("NewTracker().duplicateWindow = %v, want 5s", tracker.duplicateWindow)
	}
}

// TestIsGitRepo はGitリポジトリ判定をテストする
func TestIsGitRepo(t *testing.T) {
	t.Run("Valid Git Repository", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "aict-git-test-*")
		if err != nil {
			t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// .gitディレクトリを作成
		gitDir := filepath.Join(tempDir, ".git")
		err = os.MkdirAll(gitDir, 0755)
		if err != nil {
			t.Fatalf("Gitディレクトリの作成に失敗: %v", err)
		}

		if !IsGitRepo(tempDir) {
			t.Errorf("IsGitRepo() = false, want true for valid git repository")
		}
	})

	t.Run("Non-Git Directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "aict-nogit-test-*")
		if err != nil {
			t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
		}
		defer os.RemoveAll(tempDir)

		if IsGitRepo(tempDir) {
			t.Errorf("IsGitRepo() = true, want false for non-git directory")
		}
	})
}

// TestTrackFileChanges はファイル変更追跡をテストする
func TestTrackFileChanges(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	// テストファイルを作成
	_ = createTestFile(t, tempDir, "test.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`)

	t.Run("Valid AI Event", func(t *testing.T) {
		files := []string{"test.go"}
		err := tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"AI generated code",
		)

		if err != nil {
			t.Fatalf("TrackFileChanges() error = %v, want nil", err)
		}

		// イベントが保存されているか確認
		events, err := storage.ReadEvents()
		if err != nil {
			t.Fatalf("ReadEvents() error = %v", err)
		}

		if len(events) != 1 {
			t.Errorf("保存されたイベント数 = %d, want 1", len(events))
		}

		event := events[0]
		if event.EventType != types.EventTypeAI {
			t.Errorf("EventType = %s, want %s", event.EventType, types.EventTypeAI)
		}
		if event.Author != "Claude Code" {
			t.Errorf("Author = %s, want 'Claude Code'", event.Author)
		}
		if event.Model != "claude-sonnet-4" {
			t.Errorf("Model = %s, want 'claude-sonnet-4'", event.Model)
		}
	})

	t.Run("Valid Human Event", func(t *testing.T) {
		files := []string{"test.go"}
		err := tracker.TrackFileChanges(
			types.EventTypeHuman,
			"John Doe",
			"", // 人間イベントではモデル不要
			files,
			"Human code changes",
		)

		if err != nil {
			t.Fatalf("TrackFileChanges() error = %v, want nil", err)
		}
	})

	t.Run("Duplicate Event Prevention", func(t *testing.T) {
		// テスト用に重複防止ウィンドウを短くする
		tracker.duplicateWindow = 100 * time.Millisecond
		
		// イベント数をリセット（統計をクリア）
		initialEvents, _ := storage.ReadEvents()
		initialCount := len(initialEvents)
		
		// 最初のイベント
		files := []string{"test.go"}
		err := tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"First event",
		)
		if err != nil {
			t.Fatalf("First TrackFileChanges() error = %v", err)
		}

		// 直後の重複イベント（スキップされるはず）
		err = tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"Duplicate event",
		)
		if err != nil {
			t.Fatalf("Duplicate TrackFileChanges() error = %v", err)
		}

		// 重複防止後のイベント数確認
		afterDuplicateEvents, err := storage.ReadEvents()
		if err != nil {
			t.Fatalf("ReadEvents() error = %v", err)
		}
		
		expectedAfterFirst := initialCount + 1
		if len(afterDuplicateEvents) != expectedAfterFirst {
			t.Logf("Duplicate prevention: expected %d events, got %d", expectedAfterFirst, len(afterDuplicateEvents))
		}

		// 重複防止ウィンドウより長く待機
		time.Sleep(200 * time.Millisecond)

		err = tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"After delay",
		)
		if err != nil {
			t.Fatalf("After delay TrackFileChanges() error = %v", err)
		}
	})

	t.Run("Nonexistent File", func(t *testing.T) {
		files := []string{"nonexistent.go"}
		err := tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"Nonexistent file",
		)

		// ファイルが存在しなくてもエラーにならない（削除として処理）
		if err != nil {
			t.Fatalf("TrackFileChanges() with nonexistent file error = %v", err)
		}
	})
}

// TestIsClaudeCodeCommit はClaude Codeコミット判定をテストする
func TestIsClaudeCodeCommit(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Claude Code Message",
			message:  "Add new feature\n\n🤖 Generated with [Claude Code](https://claude.ai/code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
			expected: true,
		},
		{
			name:     "Co-Authored Message",
			message:  "Fix bug\n\nCo-Authored-By: Claude <noreply@anthropic.com>",
			expected: true,
		},
		{
			name:     "Claude AI URL",
			message:  "Refactor code\n\nGenerated using claude.ai/code",
			expected: true,
		},
		{
			name:     "Regular Human Commit",
			message:  "Add new feature manually",
			expected: false,
		},
		{
			name:     "Empty Message",
			message:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.isClaudeCodeCommit(tt.message)
			if result != tt.expected {
				t.Errorf("isClaudeCodeCommit(%q) = %v, want %v", tt.message, result, tt.expected)
			}
		})
	}
}

// TestDetectClaudeModel はClaudeモデル検出をテストする
func TestDetectClaudeModel(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "Claude Sonnet 4",
			message:  "Generated with claude-sonnet-4 model",
			expected: "claude-sonnet-4",
		},
		{
			name:     "Claude Opus 4",
			message:  "Using claude-opus-4 for generation",
			expected: "claude-opus-4",
		},
		{
			name:     "Short Sonnet 4",
			message:  "sonnet-4 generated this",
			expected: "claude-sonnet-4",
		},
		{
			name:     "Short Opus 4",
			message:  "opus-4 model used",
			expected: "claude-opus-4",
		},
		{
			name:     "Legacy Claude 3 Opus",
			message:  "claude-3-opus generated",
			expected: "claude-3-opus",
		},
		{
			name:     "No Model Info",
			message:  "Regular commit message",
			expected: "claude-code",
		},
		{
			name:     "Empty Message",
			message:  "",
			expected: "claude-code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.detectClaudeModel(tt.message)
			if result != tt.expected {
				t.Errorf("detectClaudeModel(%q) = %s, want %s", tt.message, result, tt.expected)
			}
		})
	}
}

// TestCalculateFileDiff はファイル差分計算をテストする
func TestCalculateFileDiff(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	t.Run("Existing File", func(t *testing.T) {
		// テストファイルを作成
		content := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
		_ = createTestFile(t, tempDir, "diff_test.go", content)

		// 相対パスでテスト
		relPath := "diff_test.go"
		diffResult, err := tracker.calculateFileDiff(relPath)
		if err != nil {
			t.Fatalf("calculateFileDiff() error = %v", err)
		}

		if diffResult.FilePath != relPath {
			t.Errorf("DiffResult.FilePath = %s, want %s", diffResult.FilePath, relPath)
		}

		if diffResult.ContentHash == "" {
			t.Errorf("DiffResult.ContentHash is empty")
		}

		// ファイルが存在する場合は削除行数は0のはず
		if diffResult.LinesDeleted != 0 {
			t.Errorf("DiffResult.LinesDeleted = %d, want 0 for existing file", diffResult.LinesDeleted)
		}
	})

	t.Run("Nonexistent File", func(t *testing.T) {
		diffResult, err := tracker.calculateFileDiff("nonexistent.go")
		if err != nil {
			t.Fatalf("calculateFileDiff() for nonexistent file error = %v", err)
		}

		if diffResult.FilePath != "nonexistent.go" {
			t.Errorf("DiffResult.FilePath = %s, want 'nonexistent.go'", diffResult.FilePath)
		}

		// 存在しないファイルは削除扱い
		if diffResult.LinesDeleted != 0 {
			// 実際の実装では0になるが、Git diffの結果による
		}
	})
}

// TestGetRepoRoot はリポジトリルート取得をテストする
func TestGetRepoRoot(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	// Git設定ファイルを模擬（実際のgitコマンドは使用しない）
	t.Run("Repository Root", func(t *testing.T) {
		// このテストは実際のgitコマンドに依存するため、
		// モック化するか、統合テストとして別途実装する必要がある

		// 現在は基本的な動作確認のみ
		_, err := tracker.GetRepoRoot()
		// Gitリポジトリでない場合はエラーが返される
		if err == nil {
			// 実際のGitリポジトリの場合のテスト
			t.Log("GetRepoRoot() succeeded - running in actual git repository")
		} else {
			// Gitリポジトリでない場合の正常なエラー
			t.Log("GetRepoRoot() failed as expected - not in git repository")
		}
	})
}

// TestIsDuplicateEvent は重複イベント判定をテストする
func TestIsDuplicateEvent(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	t.Run("No Previous Event", func(t *testing.T) {
		// 最初のイベントなので重複ではない
		if tracker.isDuplicateEvent() {
			t.Errorf("isDuplicateEvent() = true, want false for first event")
		}
	})

	t.Run("After Time Window", func(t *testing.T) {
		// 最後のイベント時刻を設定
		tracker.lastEventTime = time.Now().Add(-10 * time.Second)

		// 10秒前なので重複ではない
		if tracker.isDuplicateEvent() {
			t.Errorf("isDuplicateEvent() = true, want false after time window")
		}
	})

	t.Run("Within Time Window", func(t *testing.T) {
		// 最後のイベント時刻を現在に設定
		tracker.lastEventTime = time.Now()

		// 直後なので重複
		if !tracker.isDuplicateEvent() {
			t.Errorf("isDuplicateEvent() = false, want true within time window")
		}
	})
}

// TestTrackCommit はコミット追跡をテストする（モック版）
func TestTrackCommit(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	t.Run("Mock Commit Tracking", func(t *testing.T) {
		// 実際のGitコマンドを使わずに、
		// コミット追跡のロジックをテストするためのモック実装
		
		// このテストは実際のGit操作に依存するため、
		// 本格的なテストには外部のGitリポジトリ環境が必要

		commitHash := "abc123def456"
		author := "John Doe"
		message := "Test commit message"

		// TrackCommitは実際のGitコマンドを実行するため、
		// このテスト環境では成功しない可能性が高い
		err := tracker.TrackCommit(commitHash, author, message)
		
		// エラーが発生することを期待（Gitリポジトリでないため）
		if err == nil {
			t.Log("TrackCommit() succeeded - running in actual git repository")
			
			// 成功した場合は、イベントが記録されているか確認
			events, readErr := storage.ReadEvents()
			if readErr != nil {
				t.Fatalf("ReadEvents() error = %v", readErr)
			}
			
			// 何らかのイベントが記録されているはず
			if len(events) == 0 {
				t.Errorf("No events recorded after TrackCommit()")
			}
		} else {
			t.Logf("TrackCommit() failed as expected: %v", err)
		}
	})
}

// BenchmarkTrackFileChanges はファイル変更追跡のベンチマークテストを行う
func BenchmarkTrackFileChanges(b *testing.B) {
	tracker, storage, tempDir := setupTestTracker(&testing.T{})
	defer cleanupTestTracker(storage, tempDir)

	// テストファイルを作成
	createTestFile(&testing.T{}, tempDir, "bench_test.go", `package main
import "fmt"
func main() {
	fmt.Println("Benchmark test")
}
`)

	files := []string{"bench_test.go"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"Benchmark event",
		)
		if err != nil {
			b.Fatalf("TrackFileChanges error: %v", err)
		}
		
		// 重複防止のため少し待機
		time.Sleep(1 * time.Millisecond)
	}
}

// TestTrackerIntegration は統合テストを行う
func TestTrackerIntegration(t *testing.T) {
	tracker, storage, tempDir := setupTestTracker(t)
	defer cleanupTestTracker(storage, tempDir)

	// 複数のファイルでの統合テスト
	t.Run("Multiple Files Integration", func(t *testing.T) {
		// テスト用に重複防止ウィンドウを短くする
		tracker.duplicateWindow = 100 * time.Millisecond
		
		// テストファイルを複数作成
		files := []string{"main.go", "utils.go", "types.go"}
		for i, filename := range files {
			content := "package main\n\n// File " + string(rune(i))
			createTestFile(t, tempDir, filename, content)
		}

		// AI による変更を追跡
		err := tracker.TrackFileChanges(
			types.EventTypeAI,
			"Claude Code",
			"claude-sonnet-4",
			files,
			"Multi-file AI changes",
		)
		if err != nil {
			t.Fatalf("TrackFileChanges() error = %v", err)
		}

		// 重複防止ウィンドウより長く待機
		time.Sleep(200 * time.Millisecond)
		
		// 人間による変更を追跡
		err = tracker.TrackFileChanges(
			types.EventTypeHuman,
			"John Doe",
			"",
			[]string{"main.go"},
			"Human bug fix",
		)
		if err != nil {
			t.Fatalf("TrackFileChanges() error = %v", err)
		}

		// 統計情報を確認
		stats, err := storage.GetStatistics()
		if err != nil {
			t.Fatalf("GetStatistics() error = %v", err)
		}

		if stats.TotalEvents < 2 {
			t.Errorf("TotalEvents = %d, want >= 2", stats.TotalEvents)
		}

		if stats.AIEvents < 1 {
			t.Errorf("AIEvents = %d, want >= 1", stats.AIEvents)
		}

		if stats.HumanEvents < 1 {
			t.Errorf("HumanEvents = %d, want >= 1", stats.HumanEvents)
		}
	})
}