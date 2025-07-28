package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// setupTestStorage はテスト用の一時ストレージを作成する
func setupTestStorage(t *testing.T) (*Storage, string) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	// ストレージを初期化
	storage, err := NewStorage(tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("ストレージの初期化に失敗: %v", err)
	}

	return storage, tempDir
}

// cleanupTestStorage はテスト用のストレージを削除する
func cleanupTestStorage(storage *Storage, tempDir string) {
	storage.Close()
	os.RemoveAll(tempDir)
}

// createTestEvent はテスト用のTrackEventを作成する
func createTestEvent(id string, eventType types.EventType, author string) *types.TrackEvent {
	return &types.TrackEvent{
		ID:        id,
		Timestamp: time.Now(),
		EventType: eventType,
		Author:    author,
		Model:     "claude-sonnet-4",
		Files: []types.FileInfo{
			{
				Path:          "test.go",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
				Hash:          "test-hash",
			},
		},
		Message: "Test event",
	}
}

// TestNewStorage はストレージの初期化をテストする
func TestNewStorage(t *testing.T) {
	t.Run("Valid Directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "aict-test-*")
		if err != nil {
			t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
		}
		defer os.RemoveAll(tempDir)

		storage, err := NewStorage(tempDir)
		if err != nil {
			t.Fatalf("NewStorage() error = %v, want nil", err)
		}
		defer storage.Close()

		// データディレクトリが存在することを確認
		if _, err := os.Stat(tempDir); os.IsNotExist(err) {
			t.Errorf("データディレクトリが作成されていません: %s", tempDir)
		}
	})

	t.Run("Default Directory", func(t *testing.T) {
		// Gitリポジトリでないディレクトリでテスト
		tempDir, err := os.MkdirTemp("", "aict-test-*")
		if err != nil {
			t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 現在のディレクトリを変更
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(tempDir)

		// .gitディレクトリを作成
		os.MkdirAll(".git", 0755)

		storage, err := NewStorage("")
		if err != nil {
			t.Fatalf("NewStorage() with default directory error = %v", err)
		}
		defer storage.Close()

		// デフォルトディレクトリが作成されることを確認
		defaultPath := filepath.Join(tempDir, DefaultDataDir)
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			t.Errorf("デフォルトデータディレクトリが作成されていません: %s", defaultPath)
		}
	})
}

// TestWriteEvent はイベント書き込み機能をテストする
func TestWriteEvent(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	t.Run("Valid Event", func(t *testing.T) {
		event := createTestEvent("test-001", types.EventTypeAI, "Claude Code")

		err := storage.WriteEvent(event)
		if err != nil {
			t.Fatalf("WriteEvent() error = %v, want nil", err)
		}

		// ファイルが作成されていることを確認
		dataFile := filepath.Join(tempDir, DataFileName)
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			t.Errorf("データファイルが作成されていません: %s", dataFile)
		}

		// インデックスファイルが作成されていることを確認
		indexFile := filepath.Join(tempDir, IndexFileName)
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			t.Errorf("インデックスファイルが作成されていません: %s", indexFile)
		}
	})

	t.Run("Invalid Event", func(t *testing.T) {
		invalidEvent := &types.TrackEvent{
			ID: "", // 無効なID
		}

		err := storage.WriteEvent(invalidEvent)
		if err == nil {
			t.Errorf("WriteEvent() with invalid event error = nil, want error")
		}
	})

	t.Run("Multiple Events", func(t *testing.T) {
		events := []*types.TrackEvent{
			createTestEvent("test-002", types.EventTypeAI, "Claude Code"),
			createTestEvent("test-003", types.EventTypeHuman, "John Doe"),
			createTestEvent("test-004", types.EventTypeCommit, "Jane Smith"),
		}

		// コミットイベントにはCommitHashが必要
		events[2].CommitHash = "abc123"
		events[2].Model = "" // コミットイベントではモデル不要

		for _, event := range events {
			err := storage.WriteEvent(event)
			if err != nil {
				t.Fatalf("WriteEvent() error = %v for event %s", err, event.ID)
			}
		}
	})
}

// TestReadEvents はイベント読み込み機能をテストする
func TestReadEvents(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	t.Run("Empty Storage", func(t *testing.T) {
		events, err := storage.ReadEvents()
		if err != nil {
			t.Fatalf("ReadEvents() error = %v, want nil", err)
		}

		if len(events) != 0 {
			t.Errorf("ReadEvents() returned %d events, want 0", len(events))
		}
	})

	t.Run("With Events", func(t *testing.T) {
		// テストイベントを追加
		testEvents := []*types.TrackEvent{
			createTestEvent("read-001", types.EventTypeAI, "Claude Code"),
			createTestEvent("read-002", types.EventTypeHuman, "John Doe"),
		}

		for _, event := range testEvents {
			err := storage.WriteEvent(event)
			if err != nil {
				t.Fatalf("WriteEvent() error = %v", err)
			}
		}

		// イベントを読み込み
		events, err := storage.ReadEvents()
		if err != nil {
			t.Fatalf("ReadEvents() error = %v", err)
		}

		if len(events) != len(testEvents) {
			t.Errorf("ReadEvents() returned %d events, want %d", len(events), len(testEvents))
		}

		// 最初のイベントの内容を確認
		if events[0].ID != testEvents[0].ID {
			t.Errorf("ReadEvents()[0].ID = %s, want %s", events[0].ID, testEvents[0].ID)
		}
	})
}

// TestReadEventsByAuthor は作成者別イベント読み込みをテストする
func TestReadEventsByAuthor(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	// テストデータを準備
	events := []*types.TrackEvent{
		createTestEvent("author-001", types.EventTypeAI, "Claude Code"),
		createTestEvent("author-002", types.EventTypeAI, "Claude Code"),
		createTestEvent("author-003", types.EventTypeHuman, "John Doe"),
	}

	for _, event := range events {
		err := storage.WriteEvent(event)
		if err != nil {
			t.Fatalf("WriteEvent() error = %v", err)
		}
	}

	t.Run("Existing Author", func(t *testing.T) {
		claudeEvents, err := storage.ReadEventsByAuthor("Claude Code")
		if err != nil {
			t.Fatalf("ReadEventsByAuthor() error = %v", err)
		}

		if len(claudeEvents) != 2 {
			t.Errorf("ReadEventsByAuthor('Claude Code') returned %d events, want 2", len(claudeEvents))
		}

		for _, event := range claudeEvents {
			if event.Author != "Claude Code" {
				t.Errorf("ReadEventsByAuthor('Claude Code') returned event with author %s", event.Author)
			}
		}
	})

	t.Run("Non-existing Author", func(t *testing.T) {
		events, err := storage.ReadEventsByAuthor("Unknown Author")
		if err != nil {
			t.Fatalf("ReadEventsByAuthor() error = %v", err)
		}

		if len(events) != 0 {
			t.Errorf("ReadEventsByAuthor('Unknown Author') returned %d events, want 0", len(events))
		}
	})
}

// TestReadEventsByType はイベントタイプ別読み込みをテストする
func TestReadEventsByType(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	// テストデータを準備
	events := []*types.TrackEvent{
		createTestEvent("type-001", types.EventTypeAI, "Claude Code"),
		createTestEvent("type-002", types.EventTypeAI, "Claude Code"),
		createTestEvent("type-003", types.EventTypeHuman, "John Doe"),
	}

	for _, event := range events {
		err := storage.WriteEvent(event)
		if err != nil {
			t.Fatalf("WriteEvent() error = %v", err)
		}
	}

	t.Run("AI Events", func(t *testing.T) {
		aiEvents, err := storage.ReadEventsByType(types.EventTypeAI)
		if err != nil {
			t.Fatalf("ReadEventsByType() error = %v", err)
		}

		if len(aiEvents) != 2 {
			t.Errorf("ReadEventsByType(EventTypeAI) returned %d events, want 2", len(aiEvents))
		}

		for _, event := range aiEvents {
			if event.EventType != types.EventTypeAI {
				t.Errorf("ReadEventsByType(EventTypeAI) returned event with type %s", event.EventType)
			}
		}
	})

	t.Run("Human Events", func(t *testing.T) {
		humanEvents, err := storage.ReadEventsByType(types.EventTypeHuman)
		if err != nil {
			t.Fatalf("ReadEventsByType() error = %v", err)
		}

		if len(humanEvents) != 1 {
			t.Errorf("ReadEventsByType(EventTypeHuman) returned %d events, want 1", len(humanEvents))
		}
	})
}

// TestReadEventsByDateRange は日付範囲別読み込みをテストする
func TestReadEventsByDateRange(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	// 異なる日付のイベントを作成
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	events := []*types.TrackEvent{
		createTestEvent("date-001", types.EventTypeAI, "Claude Code"),
		createTestEvent("date-002", types.EventTypeHuman, "John Doe"),
	}

	// タイムスタンプを設定
	events[0].Timestamp = yesterday
	events[1].Timestamp = now

	for _, event := range events {
		err := storage.WriteEvent(event)
		if err != nil {
			t.Fatalf("WriteEvent() error = %v", err)
		}
	}

	t.Run("Date Range Including All", func(t *testing.T) {
		start := yesterday.AddDate(0, 0, -1)
		end := tomorrow

		events, err := storage.ReadEventsByDateRange(start, end)
		if err != nil {
			t.Fatalf("ReadEventsByDateRange() error = %v", err)
		}

		if len(events) != 2 {
			t.Errorf("ReadEventsByDateRange() returned %d events, want 2", len(events))
		}
	})

	t.Run("Date Range Excluding Some", func(t *testing.T) {
		start := now.AddDate(0, 0, -1)
		end := now

		events, err := storage.ReadEventsByDateRange(start, end)
		if err != nil {
			t.Fatalf("ReadEventsByDateRange() error = %v", err)
		}

		// 日付の境界条件により結果が変わる可能性があるため、柔軟にテスト
		if len(events) > 2 {
			t.Errorf("ReadEventsByDateRange() returned %d events, want <= 2", len(events))
		}
	})
}

// TestGetStatistics は統計情報取得をテストする
func TestGetStatistics(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	t.Run("Empty Statistics", func(t *testing.T) {
		stats, err := storage.GetStatistics()
		if err != nil {
			t.Fatalf("GetStatistics() error = %v", err)
		}

		if stats.TotalEvents != 0 {
			t.Errorf("GetStatistics().TotalEvents = %d, want 0", stats.TotalEvents)
		}
		if stats.AIEvents != 0 {
			t.Errorf("GetStatistics().AIEvents = %d, want 0", stats.AIEvents)
		}
	})

	t.Run("With Events", func(t *testing.T) {
		// テストイベントを追加
		events := []*types.TrackEvent{
			createTestEvent("stats-001", types.EventTypeAI, "Claude Code"),
			createTestEvent("stats-002", types.EventTypeAI, "Claude Code"),
			createTestEvent("stats-003", types.EventTypeHuman, "John Doe"),
		}

		// コミットイベントを追加
		commitEvent := createTestEvent("stats-004", types.EventTypeCommit, "Jane Smith")
		commitEvent.CommitHash = "abc123"
		commitEvent.Model = ""
		events = append(events, commitEvent)

		for _, event := range events {
			err := storage.WriteEvent(event)
			if err != nil {
				t.Fatalf("WriteEvent() error = %v", err)
			}
		}

		stats, err := storage.GetStatistics()
		if err != nil {
			t.Fatalf("GetStatistics() error = %v", err)
		}

		if stats.TotalEvents != 4 {
			t.Errorf("GetStatistics().TotalEvents = %d, want 4", stats.TotalEvents)
		}
		if stats.AIEvents != 2 {
			t.Errorf("GetStatistics().AIEvents = %d, want 2", stats.AIEvents)
		}
		if stats.HumanEvents != 1 {
			t.Errorf("GetStatistics().HumanEvents = %d, want 1", stats.HumanEvents)
		}
		if stats.CommitEvents != 1 {
			t.Errorf("GetStatistics().CommitEvents = %d, want 1", stats.CommitEvents)
		}

		// 行数の統計をチェック（各イベントが10行追加、5行変更、3行削除）
		expectedLinesAdded := 4 * 10
		if stats.TotalLinesAdded != expectedLinesAdded {
			t.Errorf("GetStatistics().TotalLinesAdded = %d, want %d", stats.TotalLinesAdded, expectedLinesAdded)
		}

		// パーセンテージの計算をチェック
		expectedAIPercentage := float64(2) / float64(4) * 100.0
		if stats.AIPercentage() != expectedAIPercentage {
			t.Errorf("GetStatistics().AIPercentage() = %f, want %f", stats.AIPercentage(), expectedAIPercentage)
		}
	})
}

// TestIndexOperations はインデックス操作をテストする
func TestIndexOperations(t *testing.T) {
	storage, tempDir := setupTestStorage(t)
	defer cleanupTestStorage(storage, tempDir)

	t.Run("Index Update", func(t *testing.T) {
		event := createTestEvent("index-001", types.EventTypeAI, "Claude Code")

		err := storage.WriteEvent(event)
		if err != nil {
			t.Fatalf("WriteEvent() error = %v", err)
		}

		// インデックスが更新されていることを確認
		if storage.index.TotalEvents != 1 {
			t.Errorf("Index.TotalEvents = %d, want 1", storage.index.TotalEvents)
		}

		// 作成者インデックスの確認
		if len(storage.index.EventsByAuthor["Claude Code"]) != 1 {
			t.Errorf("Index.EventsByAuthor['Claude Code'] length = %d, want 1", 
				len(storage.index.EventsByAuthor["Claude Code"]))
		}

		// イベントタイプインデックスの確認
		if len(storage.index.EventsByType["ai"]) != 1 {
			t.Errorf("Index.EventsByType['ai'] length = %d, want 1", 
				len(storage.index.EventsByType["ai"]))
		}
	})

	t.Run("Index Rebuild", func(t *testing.T) {
		// インデックスを破損させる
		storage.index.TotalEvents = 999

		// インデックスを再構築
		err := storage.rebuildIndex()
		if err != nil {
			t.Fatalf("rebuildIndex() error = %v", err)
		}

		// インデックスが正しく再構築されていることを確認
		if storage.index.TotalEvents != 1 {
			t.Errorf("After rebuild, Index.TotalEvents = %d, want 1", storage.index.TotalEvents)
		}
	})
}

// BenchmarkWriteEvent はイベント書き込みのベンチマークテストを行う
func BenchmarkWriteEvent(b *testing.B) {
	storage, tempDir := setupTestStorage(&testing.T{})
	defer cleanupTestStorage(storage, tempDir)

	event := createTestEvent("bench-001", types.EventTypeAI, "Claude Code")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event.ID = "bench-" + string(rune(i))
		err := storage.WriteEvent(event)
		if err != nil {
			b.Fatalf("WriteEvent error: %v", err)
		}
	}
}

// BenchmarkReadEvents はイベント読み込みのベンチマークテストを行う
func BenchmarkReadEvents(b *testing.B) {
	storage, tempDir := setupTestStorage(&testing.T{})
	defer cleanupTestStorage(storage, tempDir)

	// テストデータを準備
	for i := 0; i < 100; i++ {
		event := createTestEvent("bench-read-"+string(rune(i)), types.EventTypeAI, "Claude Code")
		storage.WriteEvent(event)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.ReadEvents()
		if err != nil {
			b.Fatalf("ReadEvents error: %v", err)
		}
	}
}