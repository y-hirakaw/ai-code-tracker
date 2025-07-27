package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/pkg/types"
)

// setupTestStatsManager はテスト用のStatsManagerを作成する
func setupTestStatsManager(t *testing.T) (*StatsManager, *storage.Storage, string) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict-stats-test-*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}

	// ストレージを初期化
	storageInstance, err := storage.NewStorage(filepath.Join(tempDir, "ai-tracker"))
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("ストレージの初期化に失敗: %v", err)
	}

	// StatsManagerを作成
	statsManager := NewStatsManager(storageInstance)

	return statsManager, storageInstance, tempDir
}

// cleanupTestStatsManager はテスト用のStatsManagerを削除する
func cleanupTestStatsManager(storage *storage.Storage, tempDir string) {
	storage.Close()
	os.RemoveAll(tempDir)
}

// createTestEvents はテスト用のイベントを作成する
func createTestEvents(t *testing.T, storage *storage.Storage) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []*types.TrackEvent{
		// 1日目: AI イベント
		{
			ID:        "event-1",
			Timestamp: baseTime,
			EventType: types.EventTypeAI,
			Author:    "Claude Code",
			Model:     "claude-sonnet-4",
			Files: []types.FileInfo{
				{Path: "main.go", LinesAdded: 10, LinesModified: 5, LinesDeleted: 2},
				{Path: "utils.go", LinesAdded: 8, LinesModified: 0, LinesDeleted: 0},
			},
			Message: "AI generated code",
		},
		// 1日目: 人間 イベント
		{
			ID:        "event-2",
			Timestamp: baseTime.Add(2 * time.Hour),
			EventType: types.EventTypeHuman,
			Author:    "John Doe",
			Files: []types.FileInfo{
				{Path: "main.go", LinesAdded: 5, LinesModified: 3, LinesDeleted: 1},
			},
			Message: "Human bug fix",
		},
		// 2日目: AI イベント
		{
			ID:        "event-3",
			Timestamp: baseTime.Add(24 * time.Hour),
			EventType: types.EventTypeAI,
			Author:    "Claude Code",
			Model:     "claude-opus-4",
			Files: []types.FileInfo{
				{Path: "test.go", LinesAdded: 15, LinesModified: 0, LinesDeleted: 0},
			},
			Message: "AI test generation",
		},
		// 3日目: 人間 イベント
		{
			ID:        "event-4",
			Timestamp: baseTime.Add(48 * time.Hour),
			EventType: types.EventTypeHuman,
			Author:    "Jane Smith",
			Files: []types.FileInfo{
				{Path: "main.go", LinesAdded: 2, LinesModified: 8, LinesDeleted: 3},
				{Path: "config.go", LinesAdded: 20, LinesModified: 0, LinesDeleted: 0},
			},
			Message: "Configuration updates",
		},
		// 4日目: コミット イベント
		{
			ID:         "event-5",
			Timestamp:  baseTime.Add(72 * time.Hour),
			EventType:  types.EventTypeCommit,
			Author:     "John Doe",
			CommitHash: "abc123",
			Files: []types.FileInfo{
				{Path: "main.go", LinesAdded: 0, LinesModified: 1, LinesDeleted: 0},
			},
			Message: "Fix typo",
		},
	}

	// イベントを保存
	for _, event := range events {
		err := storage.WriteEvent(event)
		if err != nil {
			t.Fatalf("テストイベントの保存に失敗: %v", err)
		}
	}
}

// TestNewStatsManager はStatsManagerの初期化をテストする
func TestNewStatsManager(t *testing.T) {
	_, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	statsManager := NewStatsManager(storage)

	if statsManager.storage != storage {
		t.Errorf("NewStatsManager().storage が設定されていません")
	}
}

// TestGetDailyStats は日次統計取得をテストする
func TestGetDailyStats(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	// 日次統計を取得
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	dailyStats, err := statsManager.GetDailyStats(since, until)
	if err != nil {
		t.Fatalf("GetDailyStats() error = %v", err)
	}

	if len(dailyStats) == 0 {
		t.Errorf("GetDailyStats() returned empty results")
		return
	}

	// 1日目の統計を確認
	day1Found := false
	for _, daily := range dailyStats {
		if daily.Date.Day() == 1 {
			day1Found = true
			if daily.AIEvents != 1 {
				t.Errorf("Day 1 AIEvents = %d, want 1", daily.AIEvents)
			}
			if daily.HumanEvents != 1 {
				t.Errorf("Day 1 HumanEvents = %d, want 1", daily.HumanEvents)
			}
			if daily.AIPercentage != 50.0 {
				t.Errorf("Day 1 AIPercentage = %.1f, want 50.0", daily.AIPercentage)
			}
			break
		}
	}

	if !day1Found {
		t.Errorf("Day 1 stats not found in results")
	}
}

// TestGetFileStats はファイル別統計取得をテストする
func TestGetFileStats(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	// ファイル別統計を取得
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	fileStats, err := statsManager.GetFileStats(since)
	if err != nil {
		t.Fatalf("GetFileStats() error = %v", err)
	}

	if len(fileStats) == 0 {
		t.Errorf("GetFileStats() returned empty results")
		return
	}

	// main.goの統計を確認
	mainGoFound := false
	for _, file := range fileStats {
		if file.FilePath == "main.go" {
			mainGoFound = true
			if file.AIEvents != 1 {
				t.Errorf("main.go AIEvents = %d, want 1", file.AIEvents)
			}
			if file.HumanEvents != 2 { // John Doe + Jane Smith
				t.Errorf("main.go HumanEvents = %d, want 2", file.HumanEvents)
			}
			if file.TotalChanges == 0 {
				t.Errorf("main.go TotalChanges = %d, want > 0", file.TotalChanges)
			}
			break
		}
	}

	if !mainGoFound {
		t.Errorf("main.go stats not found in results")
	}
}

// TestGetContributorStats は貢献者別統計取得をテストする
func TestGetContributorStats(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	// 貢献者別統計を取得
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	contributorStats, err := statsManager.GetContributorStats(since)
	if err != nil {
		t.Fatalf("GetContributorStats() error = %v", err)
	}

	if len(contributorStats) == 0 {
		t.Errorf("GetContributorStats() returned empty results")
		return
	}

	// Claude Codeの統計を確認
	claudeFound := false
	johnFound := false
	
	for _, contributor := range contributorStats {
		if contributor.Name == "Claude Code" {
			claudeFound = true
			if !contributor.IsAI {
				t.Errorf("Claude Code IsAI = false, want true")
			}
			if contributor.Events != 2 {
				t.Errorf("Claude Code Events = %d, want 2", contributor.Events)
			}
			if contributor.LinesAdded == 0 {
				t.Errorf("Claude Code LinesAdded = %d, want > 0", contributor.LinesAdded)
			}
		}
		if contributor.Name == "John Doe" {
			johnFound = true
			if contributor.IsAI {
				t.Errorf("John Doe IsAI = true, want false")
			}
			if contributor.Events != 2 { // Human + Commit
				t.Errorf("John Doe Events = %d, want 2", contributor.Events)
			}
		}
	}

	if !claudeFound {
		t.Errorf("Claude Code stats not found in results")
	}
	if !johnFound {
		t.Errorf("John Doe stats not found in results")
	}
}

// TestGetPeriodStats は期間統計取得をテストする
func TestGetPeriodStats(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	// 期間統計を取得
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	periodStats, err := statsManager.GetPeriodStats(since, until)
	if err != nil {
		t.Fatalf("GetPeriodStats() error = %v", err)
	}

	if periodStats.TotalEvents != 5 {
		t.Errorf("PeriodStats.TotalEvents = %d, want 5", periodStats.TotalEvents)
	}

	if periodStats.AIEvents != 2 {
		t.Errorf("PeriodStats.AIEvents = %d, want 2", periodStats.AIEvents)
	}

	if periodStats.HumanEvents != 2 {
		t.Errorf("PeriodStats.HumanEvents = %d, want 2", periodStats.HumanEvents)
	}

	if len(periodStats.DailyStats) == 0 {
		t.Errorf("PeriodStats.DailyStats is empty")
	}

	if len(periodStats.TopContributors) == 0 {
		t.Errorf("PeriodStats.TopContributors is empty")
	}

	if len(periodStats.TopFiles) == 0 {
		t.Errorf("PeriodStats.TopFiles is empty")
	}
}

// TestFilterByAuthor は作成者フィルタをテストする
func TestFilterByAuthor(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	// 全イベントを取得
	events, err := storage.ReadEvents()
	if err != nil {
		t.Fatalf("ReadEvents() error = %v", err)
	}

	// Claude Codeでフィルタ
	claudeEvents := statsManager.FilterByAuthor(events, "Claude")
	if len(claudeEvents) != 2 {
		t.Errorf("FilterByAuthor('Claude') returned %d events, want 2", len(claudeEvents))
	}

	// John Doeでフィルタ
	johnEvents := statsManager.FilterByAuthor(events, "John")
	if len(johnEvents) != 2 {
		t.Errorf("FilterByAuthor('John') returned %d events, want 2", len(johnEvents))
	}

	// 存在しない作成者でフィルタ
	unknownEvents := statsManager.FilterByAuthor(events, "Unknown")
	if len(unknownEvents) != 0 {
		t.Errorf("FilterByAuthor('Unknown') returned %d events, want 0", len(unknownEvents))
	}
}

// TestGetTrendAnalysis はトレンド分析をテストする
func TestGetTrendAnalysis(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	// トレンド分析を取得
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	analysis, err := statsManager.GetTrendAnalysis(since, until)
	if err != nil {
		t.Fatalf("GetTrendAnalysis() error = %v", err)
	}

	// 基本的な分析結果の存在確認
	if _, exists := analysis["most_active_weekday"]; !exists {
		t.Errorf("GetTrendAnalysis() missing 'most_active_weekday'")
	}

	if _, exists := analysis["ai_ratio_stability"]; !exists {
		t.Errorf("GetTrendAnalysis() missing 'ai_ratio_stability'")
	}

	// 最も活発な曜日の確認
	if weekdayData, ok := analysis["most_active_weekday"].(map[string]interface{}); ok {
		if _, hasWeekday := weekdayData["weekday"]; !hasWeekday {
			t.Errorf("most_active_weekday missing 'weekday' field")
		}
		if _, hasActivity := weekdayData["activity"]; !hasActivity {
			t.Errorf("most_active_weekday missing 'activity' field")
		}
	} else {
		t.Errorf("most_active_weekday has incorrect type")
	}
}

// TestCalculateAverageAIPercentage は平均AI貢献率計算をテストする
func TestCalculateAverageAIPercentage(t *testing.T) {
	tests := []struct {
		name       string
		dailyStats []DailyStats
		expected   float64
	}{
		{
			name:       "Empty Stats",
			dailyStats: []DailyStats{},
			expected:   0.0,
		},
		{
			name: "Single Day",
			dailyStats: []DailyStats{
				{AIPercentage: 50.0},
			},
			expected: 50.0,
		},
		{
			name: "Multiple Days",
			dailyStats: []DailyStats{
				{AIPercentage: 40.0},
				{AIPercentage: 60.0},
				{AIPercentage: 50.0},
			},
			expected: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAverageAIPercentage(tt.dailyStats)
			if result != tt.expected {
				t.Errorf("calculateAverageAIPercentage() = %.1f, want %.1f", result, tt.expected)
			}
		})
	}
}

// TestCalculateVariance は分散計算をテストする
func TestCalculateVariance(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "Empty Values",
			values:   []float64{},
			expected: 0.0,
		},
		{
			name:     "Single Value",
			values:   []float64{50.0},
			expected: 0.0,
		},
		{
			name:     "Identical Values",
			values:   []float64{50.0, 50.0, 50.0},
			expected: 0.0,
		},
		{
			name:     "Variable Values",
			values:   []float64{40.0, 50.0, 60.0},
			expected: 66.66666666666667, // 計算上の分散
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateVariance(tt.values)
			// 浮動小数点の比較のため、誤差を許容
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("calculateVariance() = %.10f, want %.10f", result, tt.expected)
			}
		})
	}
}

// TestMinMax はヘルパー関数をテストする
func TestMinMax(t *testing.T) {
	// min function test
	if min(5, 3) != 3 {
		t.Errorf("min(5, 3) = %d, want 3", min(5, 3))
	}
	if min(2, 8) != 2 {
		t.Errorf("min(2, 8) = %d, want 2", min(2, 8))
	}

	// max function test
	if max(5, 3) != 5 {
		t.Errorf("max(5, 3) = %d, want 5", max(5, 3))
	}
	if max(2, 8) != 8 {
		t.Errorf("max(2, 8) = %d, want 8", max(2, 8))
	}
}

// BenchmarkGetDailyStats は日次統計取得のベンチマークテストを行う
func BenchmarkGetDailyStats(b *testing.B) {
	statsManager, storage, tempDir := setupTestStatsManager(&testing.T{})
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(&testing.T{}, storage)

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := statsManager.GetDailyStats(since, until)
		if err != nil {
			b.Fatalf("GetDailyStats error: %v", err)
		}
	}
}

// BenchmarkGetContributorStats は貢献者統計取得のベンチマークテストを行う
func BenchmarkGetContributorStats(b *testing.B) {
	statsManager, storage, tempDir := setupTestStatsManager(&testing.T{})
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(&testing.T{}, storage)

	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := statsManager.GetContributorStats(since)
		if err != nil {
			b.Fatalf("GetContributorStats error: %v", err)
		}
	}
}

// TestStatsIntegration は統計機能の統合テストを行う
func TestStatsIntegration(t *testing.T) {
	statsManager, storage, tempDir := setupTestStatsManager(t)
	defer cleanupTestStatsManager(storage, tempDir)

	// テストデータを作成
	createTestEvents(t, storage)

	t.Run("Full Period Analysis", func(t *testing.T) {
		since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		until := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

		// 全ての統計機能を呼び出し
		dailyStats, err := statsManager.GetDailyStats(since, until)
		if err != nil {
			t.Fatalf("GetDailyStats() error = %v", err)
		}

		fileStats, err := statsManager.GetFileStats(since)
		if err != nil {
			t.Fatalf("GetFileStats() error = %v", err)
		}

		contributorStats, err := statsManager.GetContributorStats(since)
		if err != nil {
			t.Fatalf("GetContributorStats() error = %v", err)
		}

		periodStats, err := statsManager.GetPeriodStats(since, until)
		if err != nil {
			t.Fatalf("GetPeriodStats() error = %v", err)
		}

		analysis, err := statsManager.GetTrendAnalysis(since, until)
		if err != nil {
			t.Fatalf("GetTrendAnalysis() error = %v", err)
		}

		// 基本的な整合性チェック
		if len(dailyStats) == 0 {
			t.Errorf("No daily stats returned")
		}

		if len(fileStats) == 0 {
			t.Errorf("No file stats returned")
		}

		if len(contributorStats) == 0 {
			t.Errorf("No contributor stats returned")
		}

		if periodStats.TotalEvents == 0 {
			t.Errorf("No events in period stats")
		}

		if len(analysis) == 0 {
			t.Errorf("No trend analysis returned")
		}

		// 合計の整合性チェック
		totalEventsFromDaily := 0
		for _, daily := range dailyStats {
			totalEventsFromDaily += daily.AIEvents + daily.HumanEvents + daily.CommitEvents
		}

		if totalEventsFromDaily != periodStats.TotalEvents {
			t.Errorf("Daily stats total (%d) != period stats total (%d)", 
				totalEventsFromDaily, periodStats.TotalEvents)
		}
	})
}