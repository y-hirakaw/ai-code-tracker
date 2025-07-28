package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// setupTestEnvironment はテスト環境を準備する
func setupTestEnvironment(t *testing.T) (string, string) {
	t.Helper()
	
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict_cli_test_*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	
	// .gitディレクトリを作成（Gitリポジトリをシミュレート）
	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf(".gitディレクトリの作成に失敗: %v", err)
	}
	
	// ai-trackerディレクトリを作成
	dataDir := filepath.Join(tempDir, storage.DefaultDataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("データディレクトリの作成に失敗: %v", err)
	}
	
	return tempDir, dataDir
}

// teardownTestEnvironment はテスト環境をクリーンアップする
func teardownTestEnvironment(tempDir string) {
	os.RemoveAll(tempDir)
}

// createTestData はテスト用のデータを作成する
func createTestData(t *testing.T, dataDir string) {
	t.Helper()
	
	// DuckDBストレージを作成
	config := storage.StorageConfig{
		Type:    storage.StorageTypeDuckDB,
		DataDir: dataDir,
		Debug:   true,
	}
	
	store, err := storage.NewAdvancedStorageByType(config)
	if err != nil {
		t.Fatalf("ストレージの作成に失敗: %v", err)
	}
	defer store.Close()
	
	// テストイベントを作成
	now := time.Now()
	events := []*types.TrackEvent{
		{
			ID:        "test-1",
			Timestamp: now.Add(-2 * time.Hour),
			EventType: types.EventTypeAI,
			Author:    "Claude",
			Model:     "claude-sonnet-4",
			Files: []types.FileInfo{
				{
					Path:         "main.go",
					LinesAdded:   50,
					LinesModified: 10,
					LinesDeleted:  5,
				},
			},
			SessionID: "session-1",
		},
		{
			ID:        "test-2",
			Timestamp: now.Add(-time.Hour),
			EventType: types.EventTypeHuman,
			Author:    "Developer",
			Files: []types.FileInfo{
				{
					Path:         "utils.go",
					LinesAdded:   30,
					LinesModified: 5,
					LinesDeleted:  2,
				},
				{
					Path:         "helper.py",
					LinesAdded:   20,
					LinesModified: 0,
					LinesDeleted:  0,
				},
			},
			SessionID: "session-2",
		},
		{
			ID:        "test-3",
			Timestamp: now,
			EventType: types.EventTypeAI,
			Author:    "Claude",
			Model:     "claude-sonnet-4",
			Files: []types.FileInfo{
				{
					Path:         "api.js",
					LinesAdded:   40,
					LinesModified: 8,
					LinesDeleted:  3,
				},
			},
			SessionID: "session-3",
		},
	}
	
	// イベントを保存
	for _, event := range events {
		if err := store.StoreTrackEvent(event); err != nil {
			t.Fatalf("イベントの保存に失敗: %v", err)
		}
	}
}

func TestPeriodHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		setup   bool // テストデータを作成するかどうか
	}{
		{
			name:    "引数なしでエラー",
			args:    []string{},
			wantErr: true,
			setup:   false,
		},
		{
			name:    "Q1 2024の期間分析",
			args:    []string{"Q1", "2024"},
			wantErr: false,
			setup:   true,
		},
		{
			name:    "this yearの期間分析",
			args:    []string{"this", "year"},
			wantErr: false,
			setup:   true,
		},
		{
			name:    "無効な期間表現",
			args:    []string{"invalid", "period"},
			wantErr: true,
			setup:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト環境を準備
			tempDir, dataDir := setupTestEnvironment(t)
			defer teardownTestEnvironment(tempDir)
			
			// 現在の作業ディレクトリを保存
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("現在のディレクトリの取得に失敗: %v", err)
			}
			
			// テスト用ディレクトリに移動
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("テストディレクトリへの移動に失敗: %v", err)
			}
			defer os.Chdir(originalDir)
			
			// テストデータを作成
			if tt.setup {
				createTestData(t, dataDir)
			}
			
			// ハンドラーを作成して実行
			handler := NewPeriodHandler()
			err = handler.Handle(tt.args)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("PeriodHandler.Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPeriodHandler_Handle_NoGitRepository(t *testing.T) {
	// Gitリポジトリではないディレクトリでテスト
	tempDir, err := os.MkdirTemp("", "aict_no_git_test_*")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// 現在の作業ディレクトリを保存
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗: %v", err)
	}
	
	// テスト用ディレクトリに移動
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("テストディレクトリへの移動に失敗: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// ハンドラーを作成して実行
	handler := NewPeriodHandler()
	err = handler.Handle([]string{"Q1", "2024"})
	
	if err == nil {
		t.Error("Gitリポジトリではない場合、エラーが発生すべきです")
	}
}

func TestPeriodHandler_Handle_NoTrackingData(t *testing.T) {
	// トラッキングデータがない場合のテスト
	tempDir, _ := setupTestEnvironment(t)
	defer teardownTestEnvironment(tempDir)
	
	// 現在の作業ディレクトリを保存
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("現在のディレクトリの取得に失敗: %v", err)
	}
	
	// テスト用ディレクトリに移動
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("テストディレクトリへの移動に失敗: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// ai-trackerディレクトリを削除（トラッキングデータなしをシミュレート）
	dataDir := filepath.Join(tempDir, storage.DefaultDataDir)
	os.RemoveAll(dataDir)
	
	// ハンドラーを作成して実行
	handler := NewPeriodHandler()
	err = handler.Handle([]string{"Q1", "2024"})
	
	if err == nil {
		t.Error("トラッキングデータがない場合、エラーが発生すべきです")
	}
}

func TestHasExportFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "エクスポートフラグあり（--export）",
			args: []string{"Q1", "2024", "--export", "markdown"},
			want: true,
		},
		{
			name: "エクスポートフラグあり（-e）",
			args: []string{"Q1", "2024", "-e", "csv"},
			want: true,
		},
		{
			name: "エクスポートフラグなし",
			args: []string{"Q1", "2024"},
			want: false,
		},
		{
			name: "部分一致でエクスポートフラグあり",
			args: []string{"Q1", "2024", "--export-all"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasExportFlag(tt.args)
			if got != tt.want {
				t.Errorf("hasExportFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{
			name: "3桁以下",
			n:    123,
			want: "123",
		},
		{
			name: "4桁",
			n:    1234,
			want: "1,234",
		},
		{
			name: "7桁",
			n:    1234567,
			want: "1,234,567",
		},
		{
			name: "ゼロ",
			n:    0,
			want: "0",
		},
		{
			name: "大きな数値",
			n:    1000000000,
			want: "1,000,000,000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatNumber(tt.n)
			if got != tt.want {
				t.Errorf("formatNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "通常の文字列",
			s:    "hello",
			want: "olleh",
		},
		{
			name: "数字",
			s:    "12345",
			want: "54321",
		},
		{
			name: "空文字列",
			s:    "",
			want: "",
		},
		{
			name: "単文字",
			s:    "a",
			want: "a",
		},
		{
			name: "日本語",
			s:    "こんにちは",
			want: "はちにんこ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reverse(tt.s)
			if got != tt.want {
				t.Errorf("reverse() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ベンチマークテスト
func BenchmarkFormatNumber(b *testing.B) {
	numbers := []int{123, 1234, 12345, 123456, 1234567, 12345678, 123456789}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		num := numbers[i%len(numbers)]
		formatNumber(num)
	}
}

func BenchmarkPeriodHandler_Handle(b *testing.B) {
	// テスト環境を準備
	tempDir, dataDir := setupTestEnvironment(&testing.T{})
	defer teardownTestEnvironment(tempDir)
	
	// 現在の作業ディレクトリを保存
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)
	
	// テストデータを作成
	createTestData(&testing.T{}, dataDir)
	
	handler := NewPeriodHandler()
	args := []string{"Q1", "2024"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(args)
	}
}