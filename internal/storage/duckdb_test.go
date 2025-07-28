package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// setupTestDuckDB はテスト用のDuckDBストレージを作成
func setupTestDuckDB(t *testing.T) (*DuckDBStorage, string) {
	t.Helper()
	
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict_duckdb_test_*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	
	// DuckDBストレージを作成
	storage, err := NewDuckDBStorage(tempDir, true)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("DuckDBストレージの作成に失敗: %v", err)
	}
	
	return storage, tempDir
}

// teardownTestDuckDB はテスト用のリソースをクリーンアップ
func teardownTestDuckDB(storage *DuckDBStorage, tempDir string) {
	if storage != nil {
		storage.Close()
	}
	os.RemoveAll(tempDir)
}

// createTestEventForDuckDB はDuckDBテスト用のTrackEventを作成
func createTestEventForDuckDB(id, author string, eventType types.EventType, timestamp time.Time) *types.TrackEvent {
	return &types.TrackEvent{
		ID:        id,
		Timestamp: timestamp,
		EventType: eventType,
		Author:    author,
		Model:     "claude-sonnet-4",
		Files: []types.FileInfo{
			{
				Path:         "test.go",
				LinesAdded:   10,
				LinesModified: 5,
				LinesDeleted: 2,
				Hash:         "abc123",
			},
		},
		Message:   "Test event",
		SessionID: "session-1",
	}
}

func TestDuckDBStorage_NewDuckDBStorage(t *testing.T) {
	tests := []struct {
		name    string
		dataDir string
		debug   bool
		wantErr bool
	}{
		{
			name:    "正常な作成",
			dataDir: "",
			debug:   false,
			wantErr: false,
		},
		{
			name:    "デバッグモード有効",
			dataDir: "",
			debug:   true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "aict_test_*")
			if err != nil {
				t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
			}
			defer os.RemoveAll(tempDir)

			dataDir := tempDir
			if tt.dataDir != "" {
				dataDir = tt.dataDir
			}

			storage, err := NewDuckDBStorage(dataDir, tt.debug)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDuckDBStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Error("NewDuckDBStorage() returned nil storage")
					return
				}

				// データベースファイルが作成されているかチェック
				dbPath := filepath.Join(dataDir, "aict.duckdb")
				if _, err := os.Stat(dbPath); os.IsNotExist(err) {
					t.Error("データベースファイルが作成されていません")
				}

				// 接続テスト
				if err := storage.TestConnection(); err != nil {
					t.Errorf("接続テストに失敗: %v", err)
				}

				storage.Close()
			}
		})
	}
}

func TestDuckDBStorage_StoreTrackEvent(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	event := createTestEventForDuckDB("test-1", "test-author", types.EventTypeAI, now)

	err := storage.StoreTrackEvent(event)
	if err != nil {
		t.Fatalf("イベントの保存に失敗: %v", err)
	}

	// データベース情報を確認
	info, err := storage.GetDatabaseInfo()
	if err != nil {
		t.Fatalf("データベース情報の取得に失敗: %v", err)
	}

	if info.TrackCount != 1 {
		t.Errorf("期待するトラック数: 1, 実際: %d", info.TrackCount)
	}

	if info.FileChangeCount != 1 {
		t.Errorf("期待するファイル変更数: 1, 実際: %d", info.FileChangeCount)
	}
}

func TestDuckDBStorage_ReadEvents(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	events := []*types.TrackEvent{
		createTestEventForDuckDB("test-1", "author1", types.EventTypeAI, now),
		createTestEventForDuckDB("test-2", "author2", types.EventTypeHuman, now.Add(time.Hour)),
		createTestEventForDuckDB("test-3", "author1", types.EventTypeCommit, now.Add(2*time.Hour)),
	}

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// イベントを読み取り
	readEvents, err := storage.ReadEvents()
	if err != nil {
		t.Fatalf("イベントの読み取りに失敗: %v", err)
	}

	if len(readEvents) != len(events) {
		t.Errorf("期待するイベント数: %d, 実際: %d", len(events), len(readEvents))
	}

	// イベントの内容を検証
	for i, readEvent := range readEvents {
		original := events[i]
		if readEvent.ID != original.ID {
			t.Errorf("イベントID不一致: 期待 %s, 実際 %s", original.ID, readEvent.ID)
		}
		if readEvent.Author != original.Author {
			t.Errorf("作者不一致: 期待 %s, 実際 %s", original.Author, readEvent.Author)
		}
		if readEvent.EventType != original.EventType {
			t.Errorf("イベントタイプ不一致: 期待 %s, 実際 %s", original.EventType, readEvent.EventType)
		}
		if len(readEvent.Files) != len(original.Files) {
			t.Errorf("ファイル数不一致: 期待 %d, 実際 %d", len(original.Files), len(readEvent.Files))
		}
	}
}

func TestDuckDBStorage_ReadEventsByDateRange(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	events := []*types.TrackEvent{
		createTestEventForDuckDB("test-1", "author1", types.EventTypeAI, now.Add(-2*time.Hour)),
		createTestEventForDuckDB("test-2", "author2", types.EventTypeHuman, now),
		createTestEventForDuckDB("test-3", "author1", types.EventTypeCommit, now.Add(2*time.Hour)),
	}

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// 期間を指定してイベントを読み取り
	startTime := now.Add(-time.Hour)
	endTime := now.Add(time.Hour)
	
	readEvents, err := storage.ReadEventsByDateRange(startTime, endTime)
	if err != nil {
		t.Fatalf("期間指定イベント読み取りに失敗: %v", err)
	}

	// 期間内のイベントは1つのみ（events[1]）
	if len(readEvents) != 1 {
		t.Errorf("期待するイベント数: 1, 実際: %d", len(readEvents))
	}

	if len(readEvents) > 0 && readEvents[0].ID != "test-2" {
		t.Errorf("期待するイベントID: test-2, 実際: %s", readEvents[0].ID)
	}
}

func TestDuckDBStorage_ReadEventsByAuthor(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	events := []*types.TrackEvent{
		createTestEventForDuckDB("test-1", "author1", types.EventTypeAI, now),
		createTestEventForDuckDB("test-2", "author2", types.EventTypeHuman, now),
		createTestEventForDuckDB("test-3", "author1", types.EventTypeCommit, now),
	}

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// author1のイベントを読み取り
	readEvents, err := storage.ReadEventsByAuthor("author1")
	if err != nil {
		t.Fatalf("作者指定イベント読み取りに失敗: %v", err)
	}

	// author1のイベントは2つ
	if len(readEvents) != 2 {
		t.Errorf("期待するイベント数: 2, 実際: %d", len(readEvents))
	}

	for _, event := range readEvents {
		if event.Author != "author1" {
			t.Errorf("期待する作者: author1, 実際: %s", event.Author)
		}
	}
}

func TestDuckDBStorage_ReadEventsByType(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	events := []*types.TrackEvent{
		createTestEventForDuckDB("test-1", "author1", types.EventTypeAI, now),
		createTestEventForDuckDB("test-2", "author2", types.EventTypeHuman, now),
		createTestEventForDuckDB("test-3", "author1", types.EventTypeAI, now),
	}

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// AIイベントを読み取り
	readEvents, err := storage.ReadEventsByType(types.EventTypeAI)
	if err != nil {
		t.Fatalf("タイプ指定イベント読み取りに失敗: %v", err)
	}

	// AIイベントは2つ
	if len(readEvents) != 2 {
		t.Errorf("期待するイベント数: 2, 実際: %d", len(readEvents))
	}

	for _, event := range readEvents {
		if event.EventType != types.EventTypeAI {
			t.Errorf("期待するイベントタイプ: %s, 実際: %s", types.EventTypeAI, event.EventType)
		}
	}
}

func TestDuckDBStorage_GetBasicStats(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	events := []*types.TrackEvent{
		createTestEventForDuckDB("test-1", "author1", types.EventTypeAI, now),
		createTestEventForDuckDB("test-2", "author2", types.EventTypeHuman, now),
	}

	// 異なる行数を設定
	events[0].Files[0].LinesAdded = 20
	events[1].Files[0].LinesAdded = 30

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// 基本統計を取得
	ctx := context.Background()
	stats, err := storage.GetBasicStats(ctx)
	if err != nil {
		t.Fatalf("基本統計の取得に失敗: %v", err)
	}

	if stats.TotalEvents != 2 {
		t.Errorf("期待する総イベント数: 2, 実際: %d", stats.TotalEvents)
	}

	if stats.TotalLines != 50 {
		t.Errorf("期待する総行数: 50, 実際: %d", stats.TotalLines)
	}

	if stats.AILines != 20 {
		t.Errorf("期待するAI行数: 20, 実際: %d", stats.AILines)
	}

	if stats.HumanLines != 30 {
		t.Errorf("期待する人間行数: 30, 実際: %d", stats.HumanLines)
	}

	expectedAIPercentage := float64(20) / float64(50) * 100
	if abs(stats.AIPercentage-expectedAIPercentage) > 0.1 {
		t.Errorf("期待するAI比率: %.1f%%, 実際: %.1f%%", expectedAIPercentage, stats.AIPercentage)
	}
}

func TestDuckDBStorage_GetStatistics(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	events := []*types.TrackEvent{
		createTestEventForDuckDB("test-1", "author1", types.EventTypeAI, now),
		createTestEventForDuckDB("test-2", "author2", types.EventTypeHuman, now),
		createTestEventForDuckDB("test-3", "author1", types.EventTypeCommit, now),
	}
	
	// Commitイベントにはコミットハッシュが必要
	events[2].CommitHash = "abc123def456"

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// 統計情報を取得
	stats, err := storage.GetStatistics()
	if err != nil {
		t.Fatalf("統計情報の取得に失敗: %v", err)
	}

	if stats.TotalEvents != 3 {
		t.Errorf("期待する総イベント数: 3, 実際: %d", stats.TotalEvents)
	}

	if stats.AIEvents != 1 {
		t.Errorf("期待するAIイベント数: 1, 実際: %d", stats.AIEvents)
	}

	if stats.HumanEvents != 1 {
		t.Errorf("期待する人間イベント数: 1, 実際: %d", stats.HumanEvents)
	}

	if stats.CommitEvents != 1 {
		t.Errorf("期待するコミットイベント数: 1, 実際: %d", stats.CommitEvents)
	}
}

func TestDuckDBStorage_TestConnection(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	err := storage.TestConnection()
	if err != nil {
		t.Errorf("接続テストに失敗: %v", err)
	}
}

func TestDuckDBStorage_GetDatabaseInfo(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	// テストイベントを追加
	event := createTestEventForDuckDB("test-1", "test-author", types.EventTypeAI, time.Now())
	if err := storage.StoreTrackEvent(event); err != nil {
		t.Fatalf("イベントの保存に失敗: %v", err)
	}

	info, err := storage.GetDatabaseInfo()
	if err != nil {
		t.Fatalf("データベース情報の取得に失敗: %v", err)
	}

	if info.TrackCount != 1 {
		t.Errorf("期待するトラック数: 1, 実際: %d", info.TrackCount)
	}

	if info.FileChangeCount != 1 {
		t.Errorf("期待するファイル変更数: 1, 実際: %d", info.FileChangeCount)
	}

	if info.Size <= 0 {
		t.Error("データベースサイズが0以下です")
	}

	expectedPath := filepath.Join(tempDir, "aict.duckdb")
	if info.Path != expectedPath {
		t.Errorf("期待するパス: %s, 実際: %s", expectedPath, info.Path)
	}
}

// abs は浮動小数点数の絶対値を返す
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ベンチマークテスト
func BenchmarkDuckDBStorage_StoreTrackEvent(b *testing.B) {
	storage, tempDir := setupTestDuckDB(&testing.T{})
	defer teardownTestDuckDB(storage, tempDir)

	event := createTestEventForDuckDB("bench-test", "bench-author", types.EventTypeAI, time.Now())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event.ID = fmt.Sprintf("bench-test-%d", i)
		if err := storage.StoreTrackEvent(event); err != nil {
			b.Fatalf("イベントの保存に失敗: %v", err)
		}
	}
}

func BenchmarkDuckDBStorage_ReadEvents(b *testing.B) {
	storage, tempDir := setupTestDuckDB(&testing.T{})
	defer teardownTestDuckDB(storage, tempDir)

	// テストデータを準備
	for i := 0; i < 1000; i++ {
		event := createTestEventForDuckDB(fmt.Sprintf("test-%d", i), "author", types.EventTypeAI, time.Now())
		if err := storage.StoreTrackEvent(event); err != nil {
			b.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.ReadEvents()
		if err != nil {
			b.Fatalf("イベントの読み取りに失敗: %v", err)
		}
	}
}