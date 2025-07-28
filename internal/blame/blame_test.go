package blame

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// setupTestBlamer はテスト用のBlamerを作成する
func setupTestBlamer(t *testing.T) (*Blamer, *storage.Storage, string) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "aict-blame-test-*")
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

	// Blamerを作成
	blamer := NewBlamer(storageInstance, tempDir)

	return blamer, storageInstance, tempDir
}

// cleanupTestBlamer はテスト用のBlamerを削除する
func cleanupTestBlamer(storage *storage.Storage, tempDir string) {
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

// TestNewBlamer はBlamerの初期化をテストする
func TestNewBlamer(t *testing.T) {
	_, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	blamer := NewBlamer(storage, tempDir)

	if blamer.storage != storage {
		t.Errorf("NewBlamer().storage が設定されていません")
	}

	if blamer.gitRepo != tempDir {
		t.Errorf("NewBlamer().gitRepo = %s, want %s", blamer.gitRepo, tempDir)
	}
}

// TestParseGitBlameOutput はGit blame出力のパースをテストする
func TestParseGitBlameOutput(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	tests := []struct {
		name     string
		output   string
		expected int // 期待される行数
	}{
		{
			name: "Valid Git Blame Output",
			output: `a1b2c3d4e5f6g7h890abcdef1234567890abcdef 1 1 1
author John Doe
author-time 1640995200
	package main
e5f6g7h890abcdef1234567890abcdefabcdef12 2 2 1
author Claude Code
author-time 1640995260
	import "fmt"`,
			expected: 2,
		},
		{
			name:     "Empty Output",
			output:   "",
			expected: 0,
		},
		{
			name: "Single Line Output",
			output: `a1b2c3d4e5f6g7h890abcdef1234567890abcdef 1 1 1
author John Doe
author-time 1640995200
	single line`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := blamer.parseGitBlameOutput(tt.output)
			if err != nil {
				t.Fatalf("parseGitBlameOutput() error = %v", err)
			}

			if len(result) != tt.expected {
				t.Errorf("parseGitBlameOutput() returned %d lines, want %d", len(result), tt.expected)
			}

			// 最初の行の詳細チェック（有効な出力の場合）
			if tt.expected > 0 && len(result) > 0 {
				firstLine := result[0]
				if firstLine.CommitHash == "" {
					t.Errorf("parseGitBlameOutput() first line has empty commit hash")
				}
				if firstLine.LineNumber != 1 {
					t.Errorf("parseGitBlameOutput() first line number = %d, want 1", firstLine.LineNumber)
				}
			}
		})
	}
}

// TestIsClaudeCodeAuthor はClaude Code作成者判定をテストする
func TestIsClaudeCodeAuthor(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	tests := []struct {
		name     string
		author   string
		expected bool
	}{
		{"Claude Code", "Claude Code", true},
		{"Claude", "Claude", true},
		{"claude", "claude", true},
		{"AI Assistant", "AI Assistant", true},
		{"noreply@anthropic.com", "noreply@anthropic.com", true},
		{"John Doe", "John Doe", false},
		{"user@example.com", "user@example.com", false},
		{"Empty", "", false},
		{"Claude in name", "Claude Smith", true},
		{"Mixed case", "CLAUDE CODE", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := blamer.isClaudeCodeAuthor(tt.author)
			if result != tt.expected {
				t.Errorf("isClaudeCodeAuthor(%q) = %v, want %v", tt.author, result, tt.expected)
			}
		})
	}
}

// TestGuessModelFromDate は日付からのモデル推測をテストする
func TestGuessModelFromDate(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "Recent Date (Claude 4)",
			date:     time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			expected: "claude-sonnet-4",
		},
		{
			name:     "Old Date (Claude 3)",
			date:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "claude-3-sonnet",
		},
		{
			name:     "Boundary Date",
			date:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			expected: "claude-3-sonnet",
		},
		{
			name:     "Very Recent Date",
			date:     time.Now(),
			expected: "claude-sonnet-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := blamer.guessModelFromDate(tt.date)
			if result != tt.expected {
				t.Errorf("guessModelFromDate(%v) = %s, want %s", tt.date, result, tt.expected)
			}
		})
	}
}

// TestCombineBlameWithTracking はblame情報とトラッキング情報の結合をテストする
func TestCombineBlameWithTracking(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	// テスト用のGit blame行を作成
	now := time.Now()
	gitBlameLines := []GitBlameLine{
		{
			CommitHash: "abc123",
			Author:     "John Doe",
			Date:       now,
			LineNumber: 1,
			Content:    "package main",
		},
		{
			CommitHash: "def456",
			Author:     "Claude Code",
			Date:       now.Add(time.Hour),
			LineNumber: 2,
			Content:    "import \"fmt\"",
		},
	}

	// テスト用のトラッキングイベントを作成
	events := []*types.TrackEvent{
		{
			ID:         "event-1",
			Timestamp:  now,
			EventType:  types.EventTypeCommit,
			Author:     "John Doe",
			CommitHash: "abc123",
			Files: []types.FileInfo{
				{Path: "main.go", LinesAdded: 1},
			},
		},
		{
			ID:         "event-2",
			Timestamp:  now.Add(time.Hour),
			EventType:  types.EventTypeAI,
			Author:     "Claude Code",
			Model:      "claude-sonnet-4",
			CommitHash: "def456",
			Files: []types.FileInfo{
				{Path: "main.go", LinesAdded: 1},
			},
		},
	}

	result, err := blamer.combineBlameWithTracking(gitBlameLines, events, "main.go")
	if err != nil {
		t.Fatalf("combineBlameWithTracking() error = %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("combineBlameWithTracking() returned %d lines, want 2", len(result))
	}

	// 最初の行（人間）
	if result[0].IsAI {
		t.Errorf("First line should be human, got AI")
	}
	if result[0].Author != "John Doe" {
		t.Errorf("First line author = %s, want 'John Doe'", result[0].Author)
	}

	// 二番目の行（AI）
	if !result[1].IsAI {
		t.Errorf("Second line should be AI, got human")
	}
	if result[1].Model != "claude-sonnet-4" {
		t.Errorf("Second line model = %s, want 'claude-sonnet-4'", result[1].Model)
	}
}

// TestCalculateStatistics は統計計算をテストする
func TestCalculateStatistics(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	// テスト用のblame行を作成
	lines := []BlameInfo{
		{
			LineNumber: 1,
			Author:     "John Doe",
			IsAI:       false,
		},
		{
			LineNumber: 2,
			Author:     "Claude Code",
			Model:      "claude-sonnet-4",
			IsAI:       true,
		},
		{
			LineNumber: 3,
			Author:     "Claude Code",
			Model:      "claude-sonnet-4",
			IsAI:       true,
		},
		{
			LineNumber: 4,
			Author:     "Jane Smith",
			IsAI:       false,
		},
	}

	stats := blamer.calculateStatistics(lines)

	// 基本統計のテスト
	if stats.TotalLines != 4 {
		t.Errorf("TotalLines = %d, want 4", stats.TotalLines)
	}
	if stats.AILines != 2 {
		t.Errorf("AILines = %d, want 2", stats.AILines)
	}
	if stats.HumanLines != 2 {
		t.Errorf("HumanLines = %d, want 2", stats.HumanLines)
	}

	// パーセンテージのテスト
	if stats.AIPercentage != 50.0 {
		t.Errorf("AIPercentage = %f, want 50.0", stats.AIPercentage)
	}
	if stats.HumanPercentage != 50.0 {
		t.Errorf("HumanPercentage = %f, want 50.0", stats.HumanPercentage)
	}

	// 最多モデルのテスト
	if stats.TopAIModel != "claude-sonnet-4" {
		t.Errorf("TopAIModel = %s, want 'claude-sonnet-4'", stats.TopAIModel)
	}
}

// TestFormatBlameLine は行フォーマットをテストする
func TestFormatBlameLine(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	now := time.Now()
	
	tests := []struct {
		name     string
		line     BlameInfo
		useColor bool
		contains []string // 出力に含まれるべき文字列
	}{
		{
			name: "Human Line with Color",
			line: BlameInfo{
				LineNumber: 1,
				Author:     "John Doe",
				Date:       now,
				Content:    "package main",
				IsAI:       false,
			},
			useColor: true,
			contains: []string{"1", "John Doe", "👤", "package main"},
		},
		{
			name: "AI Line with Color",
			line: BlameInfo{
				LineNumber: 2,
				Author:     "Claude Code",
				Date:       now,
				Model:      "claude-sonnet-4",
				Content:    "import \"fmt\"",
				IsAI:       true,
			},
			useColor: true,
			contains: []string{"2", "Claude Code", "🤖", "S4", "import \"fmt\""},
		},
		{
			name: "AI Line without Color",
			line: BlameInfo{
				LineNumber: 3,
				Author:     "Claude Code",
				Date:       now,
				Model:      "claude-opus-4",
				Content:    "func main() {",
				IsAI:       true,
			},
			useColor: false,
			contains: []string{"3", "Claude Code", "🤖", "O4", "func main() {"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := blamer.formatBlameLine(tt.line, tt.useColor)
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatBlameLine() result does not contain %q\nResult: %s", expected, result)
				}
			}
		})
	}
}

// TestGetFileContribution はファイル貢献度取得をテストする
func TestGetFileContribution(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	// モック実装（実際のGitコマンドを使わない）
	t.Run("Mock File Contribution", func(t *testing.T) {
		// このテストは実際のGitリポジトリが必要なため、
		// モック版として基本的な動作のみテスト
		
		// ファイルが存在しない場合のエラーハンドリングをテスト
		_, err := blamer.GetFileContribution("nonexistent.go")
		if err == nil {
			t.Log("GetFileContribution() succeeded - running in git repository")
		} else {
			t.Logf("GetFileContribution() failed as expected: %v", err)
		}
	})
}

// TestGetTopContributors は上位貢献者取得をテストする
func TestGetTopContributors(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	t.Run("Mock Top Contributors", func(t *testing.T) {
		// このテストは実際のGitリポジトリが必要なため、
		// モック版として基本的な動作のみテスト
		
		_, err := blamer.GetTopContributors("nonexistent.go", 5)
		if err == nil {
			t.Log("GetTopContributors() succeeded - running in git repository")
		} else {
			t.Logf("GetTopContributors() failed as expected: %v", err)
		}
	})
}

// TestValidateFilePath はファイルパス検証をテストする
func TestValidateFilePath(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	t.Run("Nonexistent File", func(t *testing.T) {
		err := blamer.ValidateFilePath("nonexistent.go")
		if err == nil {
			t.Errorf("ValidateFilePath() should return error for nonexistent file")
		}
		if !strings.Contains(err.Error(), "ファイルが存在しません") {
			t.Errorf("ValidateFilePath() error = %v, should contain 'ファイルが存在しません'", err)
		}
	})

	t.Run("Existing File", func(t *testing.T) {
		// テストファイルを作成
		testFile := createTestFile(t, tempDir, "test.go", "package main\n")
		
		err := blamer.ValidateFilePath(testFile)
		// Gitで追跡されていないため、エラーが発生する
		if err == nil {
			t.Log("ValidateFilePath() succeeded - file is tracked by git")
		} else {
			t.Logf("ValidateFilePath() failed as expected (not tracked by git): %v", err)
		}
	})
}

// TestBlameStatistics はBlameStatistics構造体をテストする
func TestBlameStatistics(t *testing.T) {
	stats := BlameStatistics{
		TotalLines:     100,
		AILines:        60,
		HumanLines:     40,
		AIPercentage:   60.0,
		HumanPercentage: 40.0,
		TopAIModel:     "claude-sonnet-4",
		TopHumanAuthor: "John Doe",
	}

	// 基本的な値の確認
	if stats.TotalLines != 100 {
		t.Errorf("BlameStatistics.TotalLines = %d, want 100", stats.TotalLines)
	}
	if stats.AILines != 60 {
		t.Errorf("BlameStatistics.AILines = %d, want 60", stats.AILines)
	}
	if stats.AIPercentage != 60.0 {
		t.Errorf("BlameStatistics.AIPercentage = %f, want 60.0", stats.AIPercentage)
	}
}

// TestContributorInfo はContributorInfo構造体をテストする
func TestContributorInfo(t *testing.T) {
	contributor := ContributorInfo{
		Name:       "Claude Code",
		Lines:      50,
		Percentage: 45.5,
		IsAI:       true,
	}

	if contributor.Name != "Claude Code" {
		t.Errorf("ContributorInfo.Name = %s, want 'Claude Code'", contributor.Name)
	}
	if contributor.Lines != 50 {
		t.Errorf("ContributorInfo.Lines = %d, want 50", contributor.Lines)
	}
	if !contributor.IsAI {
		t.Errorf("ContributorInfo.IsAI = false, want true")
	}
}

// BenchmarkParseGitBlameOutput はGit blame出力解析のベンチマークテストを行う
func BenchmarkParseGitBlameOutput(b *testing.B) {
	blamer, storage, tempDir := setupTestBlamer(&testing.T{})
	defer cleanupTestBlamer(storage, tempDir)

	// 大きなGit blame出力を模擬
	output := strings.Repeat(`a1b2c3d4 1 1 1
author John Doe
author-time 1640995200
	package main
`, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := blamer.parseGitBlameOutput(output)
		if err != nil {
			b.Fatalf("parseGitBlameOutput error: %v", err)
		}
	}
}

// BenchmarkCalculateStatistics は統計計算のベンチマークテストを行う
func BenchmarkCalculateStatistics(b *testing.B) {
	blamer, storage, tempDir := setupTestBlamer(&testing.T{})
	defer cleanupTestBlamer(storage, tempDir)

	// 大量のblame行を作成
	lines := make([]BlameInfo, 1000)
	for i := 0; i < 1000; i++ {
		lines[i] = BlameInfo{
			LineNumber: i + 1,
			Author:     "Test Author",
			IsAI:       i%2 == 0, // 半分をAI、半分を人間
			Model:      "claude-sonnet-4",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = blamer.calculateStatistics(lines)
	}
}

// TestFormatBlameOutput は出力フォーマットの統合テストを行う
func TestFormatBlameOutput(t *testing.T) {
	blamer, storage, tempDir := setupTestBlamer(t)
	defer cleanupTestBlamer(storage, tempDir)

	// テスト用のFileBlameResultを作成
	now := time.Now()
	result := &FileBlameResult{
		FilePath: "test.go",
		Lines: []BlameInfo{
			{
				LineNumber: 1,
				Author:     "John Doe",
				Date:       now,
				Content:    "package main",
				IsAI:       false,
			},
			{
				LineNumber: 2,
				Author:     "Claude Code",
				Date:       now,
				Model:      "claude-sonnet-4",
				Content:    "import \"fmt\"",
				IsAI:       true,
			},
		},
		Statistics: BlameStatistics{
			TotalLines:      2,
			AILines:         1,
			HumanLines:      1,
			AIPercentage:    50.0,
			HumanPercentage: 50.0,
			TopAIModel:      "claude-sonnet-4",
			TopHumanAuthor:  "John Doe",
		},
	}

	t.Run("With Color", func(t *testing.T) {
		output := blamer.FormatBlameOutput(result, true)
		
		// 基本的な要素が含まれているかチェック
		expectedElements := []string{
			"test.go",
			"統計情報",
			"総行数: 2",
			"AI による行: 1 (50.0%)",
			"人間による行: 1 (50.0%)",
			"主要AIモデル: claude-sonnet-4",
			"主要貢献者: John Doe",
			"行別情報",
			"package main",
			"import \"fmt\"",
			"👤",
			"🤖",
		}
		
		for _, expected := range expectedElements {
			if !strings.Contains(output, expected) {
				t.Errorf("FormatBlameOutput() output does not contain %q", expected)
			}
		}
	})

	t.Run("Without Color", func(t *testing.T) {
		output := blamer.FormatBlameOutput(result, false)
		
		// カラーコード（ANSI エスケープシーケンス）が含まれていないことを確認
		if strings.Contains(output, "\033[") {
			t.Errorf("FormatBlameOutput() with useColor=false should not contain ANSI color codes")
		}
		
		// 基本要素は含まれている
		if !strings.Contains(output, "test.go") {
			t.Errorf("FormatBlameOutput() should contain file path")
		}
	})
}