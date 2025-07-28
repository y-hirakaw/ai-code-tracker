package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

func TestNewStorageByType(t *testing.T) {
	tests := []struct {
		name    string
		config  StorageConfig
		wantErr bool
	}{
		{
			name: "JSONLストレージの作成",
			config: StorageConfig{
				Type:    StorageTypeJSONL,
				DataDir: "",
				Debug:   false,
			},
			wantErr: false,
		},
		{
			name: "DuckDBストレージの作成",
			config: StorageConfig{
				Type:    StorageTypeDuckDB,
				DataDir: "",
				Debug:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "aict_interface_test_*")
			if err != nil {
				t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// データディレクトリを設定
			tt.config.DataDir = tempDir

			storage, err := NewStorageByType(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStorageByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Error("NewStorageByType() returned nil storage")
					return
				}

				// 基本的な操作をテスト
				testEvent := &types.TrackEvent{
					ID:        "interface-test",
					Timestamp: time.Now(),
					EventType: types.EventTypeAI,
					Author:    "test-author",
					Model:     "test-model",
					Files: []types.FileInfo{
						{
							Path:       "test.go",
							LinesAdded: 10,
						},
					},
				}

				// イベントの保存
				if err := storage.StoreTrackEvent(testEvent); err != nil {
					t.Errorf("StoreTrackEvent() failed: %v", err)
				}

				// イベントの読み取り
				events, err := storage.ReadEvents()
				if err != nil {
					t.Errorf("ReadEvents() failed: %v", err)
				}

				if len(events) != 1 {
					t.Errorf("期待するイベント数: 1, 実際: %d", len(events))
				}

				// 統計の取得
				stats, err := storage.GetStatistics()
				if err != nil {
					t.Errorf("GetStatistics() failed: %v", err)
				}

				if stats.TotalEvents != 1 {
					t.Errorf("期待する総イベント数: 1, 実際: %d", stats.TotalEvents)
				}

				storage.Close()
			}
		})
	}
}

func TestNewAdvancedStorageByType(t *testing.T) {
	tests := []struct {
		name    string
		config  StorageConfig
		wantErr bool
	}{
		{
			name: "DuckDBストレージの作成（高度機能）",
			config: StorageConfig{
				Type:    StorageTypeDuckDB,
				DataDir: "",
				Debug:   true,
			},
			wantErr: false,
		},
		{
			name: "JSONLストレージのラッパー作成",
			config: StorageConfig{
				Type:    StorageTypeJSONL,
				DataDir: "",
				Debug:   false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tempDir, err := os.MkdirTemp("", "aict_advanced_test_*")
			if err != nil {
				t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
			}
			defer os.RemoveAll(tempDir)

			// データディレクトリを設定
			tt.config.DataDir = tempDir

			storage, err := NewAdvancedStorageByType(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAdvancedStorageByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Error("NewAdvancedStorageByType() returned nil storage")
					return
				}

				// 基本機能のテスト
				testEvent := &types.TrackEvent{
					ID:        "advanced-test",
					Timestamp: time.Now(),
					EventType: types.EventTypeAI,
					Author:    "test-author",
					Model:     "test-model",
					Files: []types.FileInfo{
						{
							Path:       "test.go",
							LinesAdded: 15,
						},
					},
				}

				if err := storage.StoreTrackEvent(testEvent); err != nil {
					t.Errorf("StoreTrackEvent() failed: %v", err)
				}

				// 基本統計の取得
				ctx := context.Background()
				basicStats, err := storage.GetBasicStats(ctx)
				if err != nil {
					t.Errorf("GetBasicStats() failed: %v", err)
				}

				if basicStats == nil {
					t.Error("GetBasicStats() returned nil")
				}

				// 期間分析の取得
				now := time.Now()
				startDate := now.Add(-24 * time.Hour)
				endDate := now.Add(24 * time.Hour)

				analysis, err := storage.GetPeriodAnalysis(ctx, startDate, endDate)
				if err != nil {
					t.Errorf("GetPeriodAnalysis() failed: %v", err)
				}

				if analysis == nil {
					t.Error("GetPeriodAnalysis() returned nil")
				}

				// データベース情報の取得
				info, err := storage.GetDatabaseInfo()
				if err != nil {
					t.Errorf("GetDatabaseInfo() failed: %v", err)
				}

				if info == nil {
					t.Error("GetDatabaseInfo() returned nil")
				}

				// 接続テスト
				if err := storage.TestConnection(); err != nil {
					t.Errorf("TestConnection() failed: %v", err)
				}

				storage.Close()
			}
		})
	}
}

func TestJSONLStorageWrapper_GetBasicStats(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict_wrapper_test_*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// JSONLストレージを作成
	jsonlStorage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("JSONLストレージの作成に失敗: %v", err)
	}

	// ラッパーを作成
	wrapper := &JSONLStorageWrapper{jsonlStorage}

	// テストデータを追加
	testEvent := &types.TrackEvent{
		ID:        "wrapper-test",
		Timestamp: time.Now(),
		EventType: types.EventTypeAI,
		Author:    "test-author",
		Model:     "test-model",
		Files: []types.FileInfo{
			{
				Path:       "test.go",
				LinesAdded: 20,
			},
		},
	}

	if err := wrapper.StoreTrackEvent(testEvent); err != nil {
		t.Fatalf("イベントの保存に失敗: %v", err)
	}

	// 基本統計を取得
	ctx := context.Background()
	stats, err := wrapper.GetBasicStats(ctx)
	if err != nil {
		t.Fatalf("基本統計の取得に失敗: %v", err)
	}

	if stats.TotalEvents != 1 {
		t.Errorf("期待する総イベント数: 1, 実際: %d", stats.TotalEvents)
	}

	if stats.TotalLines != 20 {
		t.Errorf("期待する総行数: 20, 実際: %d", stats.TotalLines)
	}
}

func TestJSONLStorageWrapper_GetPeriodAnalysis(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict_wrapper_period_test_*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// JSONLストレージを作成
	jsonlStorage, err := NewStorage(tempDir)
	if err != nil {
		t.Fatalf("JSONLストレージの作成に失敗: %v", err)
	}

	// ラッパーを作成
	wrapper := &JSONLStorageWrapper{jsonlStorage}

	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	// テストデータを追加
	events := []*types.TrackEvent{
		{
			ID:        "period-test-1",
			Timestamp: startDate.Add(time.Hour),
			EventType: types.EventTypeAI,
			Author:    "Claude",
			Model:     "claude-sonnet-4",
			Files: []types.FileInfo{
				{
					Path:       "main.go",
					LinesAdded: 30,
				},
			},
		},
		{
			ID:        "period-test-2",
			Timestamp: startDate.Add(2 * time.Hour),
			EventType: types.EventTypeHuman,
			Author:    "Developer",
			Files: []types.FileInfo{
				{
					Path:       "utils.py",
					LinesAdded: 25,
				},
			},
		},
	}

	for _, event := range events {
		if err := wrapper.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}

	// 期間分析を取得
	ctx := context.Background()
	analysis, err := wrapper.GetPeriodAnalysis(ctx, startDate, endDate)
	if err != nil {
		t.Fatalf("期間分析の取得に失敗: %v", err)
	}

	if analysis.TotalLines != 55 {
		t.Errorf("期待する総行数: 55, 実際: %d", analysis.TotalLines)
	}

	if analysis.AILines != 30 {
		t.Errorf("期待するAI行数: 30, 実際: %d", analysis.AILines)
	}

	if analysis.HumanLines != 25 {
		t.Errorf("期待する人間行数: 25, 実際: %d", analysis.HumanLines)
	}

	if len(analysis.FileBreakdown) != 2 {
		t.Errorf("期待するファイル数: 2, 実際: %d", len(analysis.FileBreakdown))
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Goファイル",
			filePath: "main.go",
			want:     "Go",
		},
		{
			name:     "JavaScriptファイル",
			filePath: "app.js",
			want:     "JavaScript",
		},
		{
			name:     "TypeScriptファイル",
			filePath: "types.ts",
			want:     "TypeScript",
		},
		{
			name:     "TSXファイル",
			filePath: "component.tsx",
			want:     "TypeScript",
		},
		{
			name:     "Pythonファイル",
			filePath: "script.py",
			want:     "Python",
		},
		{
			name:     "Javaファイル",
			filePath: "Main.java",
			want:     "Java",
		},
		{
			name:     "不明なファイル",
			filePath: "README.txt",
			want:     "Other",
		},
		{
			name:     "拡張子なし",
			filePath: "Makefile",
			want:     "Other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.filePath)
			if got != tt.want {
				t.Errorf("detectLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildBasicPeriodAnalysis(t *testing.T) {
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)
	endDate := now

	events := []*types.TrackEvent{
		{
			ID:        "basic-1",
			Timestamp: startDate.Add(time.Hour),
			EventType: types.EventTypeAI,
			Author:    "Claude",
			Model:     "claude-sonnet-4",
			Files: []types.FileInfo{
				{
					Path:       "main.go",
					LinesAdded: 40,
				},
			},
		},
		{
			ID:        "basic-2",
			Timestamp: startDate.Add(2 * time.Hour),
			EventType: types.EventTypeHuman,
			Author:    "Developer",
			Files: []types.FileInfo{
				{
					Path:       "utils.js",
					LinesAdded: 30,
				},
			},
		},
	}

	analysis := buildBasicPeriodAnalysis(events, startDate, endDate)

	if analysis.TotalLines != 70 {
		t.Errorf("期待する総行数: 70, 実際: %d", analysis.TotalLines)
	}

	if analysis.AILines != 40 {
		t.Errorf("期待するAI行数: 40, 実際: %d", analysis.AILines)
	}

	if analysis.HumanLines != 30 {
		t.Errorf("期待する人間行数: 30, 実際: %d", analysis.HumanLines)
	}

	expectedAIPercentage := float64(40) / float64(70) * 100
	if abs(analysis.AIPercentage-expectedAIPercentage) > 0.1 {
		t.Errorf("期待するAI比率: %.1f%%, 実際: %.1f%%", expectedAIPercentage, analysis.AIPercentage)
	}

	if analysis.FileCount != 2 {
		t.Errorf("期待するファイル数: 2, 実際: %d", analysis.FileCount)
	}

	if analysis.SessionCount != 2 {
		t.Errorf("期待するセッション数: 2, 実際: %d", analysis.SessionCount)
	}

	// ファイル別分析の確認
	if len(analysis.FileBreakdown) != 2 {
		t.Errorf("期待するファイル分析数: 2, 実際: %d", len(analysis.FileBreakdown))
	}

	// ファイルマップを作成
	fileMap := make(map[string]*FileAnalysis)
	for i := range analysis.FileBreakdown {
		file := &analysis.FileBreakdown[i]
		fileMap[file.FilePath] = file
	}

	// main.goの確認
	if mainGo, exists := fileMap["main.go"]; exists {
		if mainGo.Language != "Go" {
			t.Errorf("main.goの期待する言語: Go, 実際: %s", mainGo.Language)
		}
		if mainGo.TotalLines != 40 {
			t.Errorf("main.goの期待する行数: 40, 実際: %d", mainGo.TotalLines)
		}
		if mainGo.AIPercentage != 100.0 {
			t.Errorf("main.goの期待するAI比率: 100.0%%, 実際: %.1f%%", mainGo.AIPercentage)
		}
	} else {
		t.Error("main.goのファイル分析が見つかりません")
	}

	// utils.jsの確認
	if utilsJs, exists := fileMap["utils.js"]; exists {
		if utilsJs.Language != "JavaScript" {
			t.Errorf("utils.jsの期待する言語: JavaScript, 実際: %s", utilsJs.Language)
		}
		if utilsJs.TotalLines != 30 {
			t.Errorf("utils.jsの期待する行数: 30, 実際: %d", utilsJs.TotalLines)
		}
		if utilsJs.AIPercentage != 0.0 {
			t.Errorf("utils.jsの期待するAI比率: 0.0%%, 実際: %.1f%%", utilsJs.AIPercentage)
		}
	} else {
		t.Error("utils.jsのファイル分析が見つかりません")
	}
}