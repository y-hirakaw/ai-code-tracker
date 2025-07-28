package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ai-code-tracker/aict/pkg/types"
)

func TestParsePeriodExpression(t *testing.T) {
	// 2025年のテスト用基準時刻
	baseTime := time.Date(2025, time.June, 15, 12, 0, 0, 0, time.Local)
	
	tests := []struct {
		name      string
		expr      string
		wantStart time.Time
		wantEnd   time.Time
		wantErr   bool
	}{
		{
			name:      "Q1 2024",
			expr:      "Q1 2024",
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 3, 31, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "q2 2024",
			expr:      "q2 2024",
			wantStart: time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 6, 30, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "Q3 2024",
			expr:      "Q3 2024",
			wantStart: time.Date(2024, 7, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 9, 30, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "Q4 2024",
			expr:      "Q4 2024",
			wantStart: time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "Jan-Mar 2024",
			expr:      "Jan-Mar 2024",
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 3, 31, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "Apr-Jun 2024",
			expr:      "Apr-Jun 2024",
			wantStart: time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 6, 30, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "this year",
			expr:      "this year",
			wantStart: time.Date(baseTime.Year(), 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(baseTime.Year(), 12, 31, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:      "last year",
			expr:      "last year",
			wantStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local),
			wantEnd:   time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.Local),
			wantErr:   false,
		},
		{
			name:    "無効な表現",
			expr:    "invalid expression",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ParsePeriodExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePeriodExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if !start.Equal(tt.wantStart) {
					t.Errorf("期待する開始日: %v, 実際: %v", tt.wantStart, start)
				}
				if !end.Equal(tt.wantEnd) {
					t.Errorf("期待する終了日: %v, 実際: %v", tt.wantEnd, end)
				}
			}
		})
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		defaultYear int
		want        int
	}{
		{
			name:        "2024年を抽出",
			expr:        "Q1 2024",
			defaultYear: 2023,
			want:        2024,
		},
		{
			name:        "年が含まれていない場合はデフォルト",
			expr:        "Q1",
			defaultYear: 2023,
			want:        2023,
		},
		{
			name:        "複数の数字がある場合は年を優先",
			expr:        "Q1 2024 test 123",
			defaultYear: 2023,
			want:        2024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractYear(tt.expr, tt.defaultYear)
			if got != tt.want {
				t.Errorf("extractYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		defaultNum int
		want       int
	}{
		{
			name:       "数字を抽出",
			expr:       "last 3 months",
			defaultNum: 1,
			want:       3,
		},
		{
			name:       "英語の数詞を抽出",
			expr:       "last three months",
			defaultNum: 1,
			want:       3,
		},
		{
			name:       "数字がない場合はデフォルト",
			expr:       "last month",
			defaultNum: 1,
			want:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNumber(tt.expr, tt.defaultNum)
			if got != tt.want {
				t.Errorf("extractNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

// createTestEventForPeriod はperiod_analysis_test.go専用のテストイベント作成関数
func createTestEventForPeriod(id, author string, eventType types.EventType, timestamp time.Time, files []types.FileInfo) *types.TrackEvent {
	return &types.TrackEvent{
		ID:        id,
		Timestamp: timestamp,
		EventType: eventType,
		Author:    author,
		Model:     "claude-sonnet-4",
		Files:     files,
		SessionID: "session-" + id,
	}
}

func TestDuckDBStorage_GetPeriodAnalysis(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	// テストデータの準備
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	events := []*types.TrackEvent{
		createTestEventForPeriod("test-1", "Claude", types.EventTypeAI, startDate.Add(time.Hour), []types.FileInfo{
			{
				Path:         "main.go",
				LinesAdded:   20,
				LinesModified: 5,
				LinesDeleted:  2,
			},
		}),
		createTestEventForPeriod("test-2", "Developer", types.EventTypeHuman, startDate.Add(2*time.Hour), []types.FileInfo{
			{
				Path:         "main.go",
				LinesAdded:   10,
				LinesModified: 3,
				LinesDeleted:  1,
			},
			{
				Path:         "utils.js",
				LinesAdded:   15,
				LinesModified: 0,
				LinesDeleted:  0,
			},
		}),
		createTestEventForPeriod("test-3", "Claude", types.EventTypeAI, startDate.Add(3*time.Hour), []types.FileInfo{
			{
				Path:         "helper.py",
				LinesAdded:   30,
				LinesModified: 0,
				LinesDeleted:  5,
			},
		}),
	}

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// 期間分析を実行
	ctx := context.Background()
	analysis, err := storage.GetPeriodAnalysis(ctx, startDate, endDate)
	if err != nil {
		t.Fatalf("期間分析に失敗: %v", err)
	}

	// 基本統計の検証
	if analysis.SessionCount != 3 {
		t.Errorf("期待するセッション数: 3, 実際: %d", analysis.SessionCount)
	}

	expectedTotalLines := (20 - 2) + (10 - 1) + (15 - 0) + (30 - 5) // lines_added - lines_deleted
	if analysis.TotalLines != expectedTotalLines {
		t.Errorf("期待する総行数: %d, 実際: %d", expectedTotalLines, analysis.TotalLines)
	}

	expectedAILines := (20 - 2) + (30 - 5) // AI events
	if analysis.AILines != expectedAILines {
		t.Errorf("期待するAI行数: %d, 実際: %d", expectedAILines, analysis.AILines)
	}

	expectedHumanLines := (10 - 1) + (15 - 0) // Human events
	if analysis.HumanLines != expectedHumanLines {
		t.Errorf("期待する人間行数: %d, 実際: %d", expectedHumanLines, analysis.HumanLines)
	}

	if analysis.FileCount != 3 {
		t.Errorf("期待するファイル数: 3, 実際: %d", analysis.FileCount)
	}

	// AI比率の検証
	expectedAIPercentage := float64(expectedAILines) / float64(expectedTotalLines) * 100
	if abs(analysis.AIPercentage-expectedAIPercentage) > 0.1 {
		t.Errorf("期待するAI比率: %.1f%%, 実際: %.1f%%", expectedAIPercentage, analysis.AIPercentage)
	}

	// ファイル別分析の検証
	if len(analysis.FileBreakdown) == 0 {
		t.Error("ファイル別分析が空です")
	}

	// ファイル別分析の内容を確認
	fileMap := make(map[string]*FileAnalysis)
	for i := range analysis.FileBreakdown {
		file := &analysis.FileBreakdown[i]
		fileMap[file.FilePath] = file
	}

	// main.goの検証（AIとHuman両方が編集）
	if mainGo, exists := fileMap["main.go"]; exists {
		expectedMainGoTotal := (20 - 2) + (10 - 1)
		if mainGo.TotalLines != expectedMainGoTotal {
			t.Errorf("main.goの期待する総行数: %d, 実際: %d", expectedMainGoTotal, mainGo.TotalLines)
		}
		if mainGo.Language != "Go" {
			t.Errorf("main.goの期待する言語: Go, 実際: %s", mainGo.Language)
		}
	} else {
		t.Error("main.goのファイル分析が見つかりません")
	}

	// 言語別統計の検証
	if len(analysis.LanguageStats) == 0 {
		t.Error("言語別統計が空です")
	}

	// 貢献者統計の検証
	if len(analysis.ContributorStats) == 0 {
		t.Error("貢献者統計が空です")
	}

	// 日別タイムラインの検証
	if len(analysis.DailyTimeline) == 0 {
		t.Error("日別タイムラインが空です")
	}
}

func TestDuckDBStorage_GetPeriodAnalysis_EmptyData(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	// データがない状態で期間分析を実行
	ctx := context.Background()
	analysis, err := storage.GetPeriodAnalysis(ctx, startDate, endDate)
	if err != nil {
		t.Fatalf("期間分析に失敗: %v", err)
	}

	// 空のデータでも正常に動作することを確認
	if analysis.TotalLines != 0 {
		t.Errorf("期待する総行数: 0, 実際: %d", analysis.TotalLines)
	}

	if analysis.SessionCount != 0 {
		t.Errorf("期待するセッション数: 0, 実際: %d", analysis.SessionCount)
	}

	if analysis.FileCount != 0 {
		t.Errorf("期待するファイル数: 0, 実際: %d", analysis.FileCount)
	}

	if len(analysis.FileBreakdown) != 0 {
		t.Errorf("期待するファイル分析数: 0, 実際: %d", len(analysis.FileBreakdown))
	}
}

func TestDuckDBStorage_GetPeriodAnalysis_MultipleLanguages(t *testing.T) {
	storage, tempDir := setupTestDuckDB(t)
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	// 複数言語のファイルを含むイベント
	events := []*types.TrackEvent{
		createTestEventForPeriod("test-go", "Claude", types.EventTypeAI, startDate.Add(time.Hour), []types.FileInfo{
			{Path: "main.go", LinesAdded: 20},
		}),
		createTestEventForPeriod("test-js", "Developer", types.EventTypeHuman, startDate.Add(2*time.Hour), []types.FileInfo{
			{Path: "app.js", LinesAdded: 15},
		}),
		createTestEventForPeriod("test-py", "Claude", types.EventTypeAI, startDate.Add(3*time.Hour), []types.FileInfo{
			{Path: "utils.py", LinesAdded: 25},
		}),
		createTestEventForPeriod("test-ts", "Developer", types.EventTypeHuman, startDate.Add(4*time.Hour), []types.FileInfo{
			{Path: "types.ts", LinesAdded: 10},
		}),
	}

	// イベントを保存
	for _, event := range events {
		if err := storage.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// 期間分析を実行
	ctx := context.Background()
	analysis, err := storage.GetPeriodAnalysis(ctx, startDate, endDate)
	if err != nil {
		t.Fatalf("期間分析に失敗: %v", err)
	}

	// 言語別統計の検証
	expectedLanguages := map[string]int{
		"Go":         20,
		"JavaScript": 15,
		"Python":     25,
		"TypeScript": 10,
	}

	if len(analysis.LanguageStats) != len(expectedLanguages) {
		t.Errorf("期待する言語数: %d, 実際: %d", len(expectedLanguages), len(analysis.LanguageStats))
	}

	langMap := make(map[string]*LanguageAnalysis)
	for i := range analysis.LanguageStats {
		lang := &analysis.LanguageStats[i]
		langMap[lang.Language] = lang
	}

	for expectedLang, expectedLines := range expectedLanguages {
		if lang, exists := langMap[expectedLang]; exists {
			if lang.TotalLines != expectedLines {
				t.Errorf("%sの期待する行数: %d, 実際: %d", expectedLang, expectedLines, lang.TotalLines)
			}
		} else {
			t.Errorf("言語 %s の統計が見つかりません", expectedLang)
		}
	}
}

// ベンチマークテスト
func BenchmarkParsePeriodExpression(b *testing.B) {
	expressions := []string{
		"Q1 2024",
		"q2 2024",
		"Jan-Mar 2024",
		"this year",
		"last year",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := expressions[i%len(expressions)]
		_, _, err := ParsePeriodExpression(expr)
		if err != nil {
			b.Errorf("ParsePeriodExpression failed: %v", err)
		}
	}
}

func BenchmarkDuckDBStorage_GetPeriodAnalysis(b *testing.B) {
	storage, tempDir := setupTestDuckDB(&testing.T{})
	defer teardownTestDuckDB(storage, tempDir)

	now := time.Now()
	startDate := now.Add(-30 * 24 * time.Hour) // 30日前
	endDate := now

	// テストデータを大量生成
	for i := 0; i < 1000; i++ {
		event := createTestEventForPeriod(
			fmt.Sprintf("bench-test-%d", i),
			"BenchAuthor",
			types.EventTypeAI,
			startDate.Add(time.Duration(i)*time.Minute),
			[]types.FileInfo{
				{
					Path:       fmt.Sprintf("file%d.go", i%10),
					LinesAdded: 10 + (i % 20),
				},
			},
		)
		if err := storage.StoreTrackEvent(event); err != nil {
			b.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := storage.GetPeriodAnalysis(ctx, startDate, endDate)
		if err != nil {
			b.Fatalf("期間分析に失敗: %v", err)
		}
	}
}