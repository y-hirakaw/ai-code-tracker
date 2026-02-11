package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestGetCommitsInRange(t *testing.T) {
	// Setup test git repository
	tmpDir := testutil.TempGitRepo(t)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create multiple commits
	testutil.CreateTestFile(t, tmpDir, "file1.go", "package main\n")
	commit1 := testutil.GitCommit(t, tmpDir, "First commit")

	testutil.CreateTestFile(t, tmpDir, "file2.go", "package main\n")
	commit2 := testutil.GitCommit(t, tmpDir, "Second commit")

	testutil.CreateTestFile(t, tmpDir, "file3.go", "package main\n")
	testutil.GitCommit(t, tmpDir, "Third commit")

	// Test getCommitsInRange with commit range
	commits, err := getCommitsInRange(commit1[:7] + ".." + commit2[:7])

	if err != nil {
		t.Fatalf("getCommitsInRange() error = %v", err)
	}

	if len(commits) == 0 {
		t.Error("getCommitsInRange() returned no commits")
	}

	// Verify commits contain expected hash
	foundCommit2 := false
	for _, commit := range commits {
		if strings.HasPrefix(commit, commit2[:7]) {
			foundCommit2 = true
			break
		}
	}

	if !foundCommit2 {
		t.Errorf("getCommitsInRange() did not include expected commit %s", commit2[:7])
	}
}

func TestHandleRangeReport_EnvironmentSetup(t *testing.T) {
	// handleRangeReport()の統合テスト前提条件（環境セットアップ）を検証する
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n")
	testutil.GitCommit(t, tmpDir, "Test commit")

	commits, err := getCommitsInRange("HEAD")
	if err != nil {
		t.Fatalf("Failed to get commits: %v", err)
	}

	if len(commits) == 0 {
		t.Error("Expected at least one commit")
	}
}

func TestFormatRangeReport_EnvironmentSetup(t *testing.T) {
	// AICT設定ファイルが正しく生成されることを検証する
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	testutil.AssertFileExists(t, tmpDir+"/.git/aict/config.json")
}

func TestExpandShorthandDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"7d", "7 days ago"},
		{"1d", "1 days ago"},
		{"2w", "2 weeks ago"},
		{"1m", "1 months ago"},
		{"3m", "3 months ago"},
		{"1y", "1 years ago"},
		{"2025-01-01", "2025-01-01"},
		{"yesterday", "yesterday"},
		{"d", "d"},
		{"", ""},
		{"abc", "abc"},
		{"12x", "12x"},
		{"0d", "0 days ago"},
	}

	for _, tt := range tests {
		result := expandShorthandDate(tt.input)
		if result != tt.expected {
			t.Errorf("expandShorthandDate(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"0", true},
		{"", false},
		{"12a", false},
		{"abc", false},
		{"-1", false},
		{"1.5", false},
	}

	for _, tt := range tests {
		result := isNumeric(tt.input)
		if result != tt.expected {
			t.Errorf("isNumeric(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestBuildReport(t *testing.T) {
	tests := []struct {
		name         string
		opts         *ReportOptions
		commitCount  int
		result       *authorStatsResult
		wantRange    string
		wantCommits  int
		wantTotal    int
		wantAIPct    float64
	}{
		{
			name:        "AI and human stats",
			opts:        &ReportOptions{Range: "abc..def", Since: ""},
			commitCount: 5,
			result: &authorStatsResult{
				byAuthor: map[string]*tracker.AuthorStats{
					"claude": {Name: "claude", Type: tracker.AuthorTypeAI, Lines: 80, Commits: 3},
					"human":  {Name: "human", Type: tracker.AuthorTypeHuman, Lines: 20, Commits: 2},
				},
				totalAI:    80,
				totalHuman: 20,
			},
			wantRange:   "abc..def",
			wantCommits: 5,
			wantTotal:   100,
			wantAIPct:   80.0,
		},
		{
			name:        "since option in range display",
			opts:        &ReportOptions{Range: "converted..HEAD", Since: "7d"},
			commitCount: 3,
			result: &authorStatsResult{
				byAuthor:   map[string]*tracker.AuthorStats{},
				totalAI:    0,
				totalHuman: 0,
			},
			wantRange:   "since 7d",
			wantCommits: 3,
			wantTotal:   0,
			wantAIPct:   0,
		},
		{
			name:        "single author",
			opts:        &ReportOptions{Range: "HEAD~1..HEAD"},
			commitCount: 1,
			result: &authorStatsResult{
				byAuthor: map[string]*tracker.AuthorStats{
					"dev": {Name: "dev", Type: tracker.AuthorTypeHuman, Lines: 50, Commits: 1},
				},
				totalAI:    0,
				totalHuman: 50,
			},
			wantRange:   "HEAD~1..HEAD",
			wantCommits: 1,
			wantTotal:   50,
			wantAIPct:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := buildReport(tt.opts, tt.commitCount, tt.result)

			if report.Range != tt.wantRange {
				t.Errorf("Range = %q, want %q", report.Range, tt.wantRange)
			}
			if report.Commits != tt.wantCommits {
				t.Errorf("Commits = %d, want %d", report.Commits, tt.wantCommits)
			}
			if report.Summary.TotalLines != tt.wantTotal {
				t.Errorf("TotalLines = %d, want %d", report.Summary.TotalLines, tt.wantTotal)
			}
			if report.Summary.AIPercentage != tt.wantAIPct {
				t.Errorf("AIPercentage = %.1f, want %.1f", report.Summary.AIPercentage, tt.wantAIPct)
			}
			if len(report.ByAuthor) != len(tt.result.byAuthor) {
				t.Errorf("ByAuthor count = %d, want %d", len(report.ByAuthor), len(tt.result.byAuthor))
			}
		})
	}
}

func TestFormatRangeReport_JSON(t *testing.T) {
	report := &tracker.Report{
		Range:   "test..HEAD",
		Commits: 2,
		Summary: tracker.SummaryStats{
			TotalLines:   100,
			AILines:      60,
			HumanLines:   40,
			AIPercentage: 60.0,
		},
	}

	err := formatRangeReport(report, "json", nil)
	if err != nil {
		t.Errorf("formatRangeReport(json) error = %v", err)
	}
}

func TestFormatRangeReport_Table(t *testing.T) {
	report := &tracker.Report{
		Range:   "test..HEAD",
		Commits: 1,
		Summary: tracker.SummaryStats{TotalLines: 50},
		ByAuthor: []tracker.AuthorStats{
			{Name: "claude", Type: tracker.AuthorTypeAI, Lines: 30, Percentage: 60.0, Commits: 1},
		},
	}
	metrics := &tracker.DetailedMetrics{
		Contributions: tracker.ContributionMetrics{AIAdditions: 30, HumanAdditions: 20},
		WorkVolume:    tracker.WorkVolumeMetrics{AIChanges: 35, HumanChanges: 25},
	}

	err := formatRangeReport(report, "table", metrics)
	if err != nil {
		t.Errorf("formatRangeReport(table) error = %v", err)
	}
}

func TestFormatRangeReport_UnknownFormat(t *testing.T) {
	report := &tracker.Report{Range: "test", Commits: 1}

	err := formatRangeReport(report, "xml", nil)
	if err == nil {
		t.Error("formatRangeReport(xml) should return error")
	}
	if err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "xml") {
			t.Errorf("error should contain format name, got: %v", err)
		}
		if !strings.Contains(msg, "table, json") {
			t.Errorf("error should list available formats, got: %v", err)
		}
	}
}

func TestValidateSinceInput(t *testing.T) {
	tests := []struct {
		input   string
		wantOK  bool // true = no warning
	}{
		// 正常な短縮形式
		{"7d", true},
		{"1d", true},
		{"2w", true},
		{"1m", true},
		{"1y", true},

		// 正常な日付形式
		{"2025-01-01", true},
		{"2024-12-31", true},

		// 正常なgit日付表現
		{"yesterday", true},
		{"today", true},
		{"7 days ago", true},
		{"2 weeks ago", true},
		{"1 months ago", true},

		// 不正な形式（警告が出る）
		{"invalid", false},
		{"abc", false},
		{"7x", false},
		{"", false},
		{"last week", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			warning := validateSinceInput(tt.input)
			if tt.wantOK && warning != "" {
				t.Errorf("validateSinceInput(%q) returned warning %q, want empty", tt.input, warning)
			}
			if !tt.wantOK && warning == "" {
				t.Errorf("validateSinceInput(%q) returned no warning, want warning", tt.input)
			}
		})
	}
}

// TestAuthorCommitCountAccuracy tests that commit counts are accurate
// when a single commit has multiple files (regression test for v1.1.3)
func TestAuthorCommitCountAccuracy(t *testing.T) {
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Scenario: Single commit with 3 files
	// Before fix: Would count as 3 commits (1 per file)
	// After fix: Should count as 1 commit

	testutil.CreateTestFile(t, tmpDir, "file1.go", "package main\nfunc main() {}\n")
	testutil.CreateTestFile(t, tmpDir, "file2.go", "package utils\nfunc Helper() {}\n")
	testutil.CreateTestFile(t, tmpDir, "file3.go", "package models\ntype User struct {}\n")
	commitHash := testutil.GitCommit(t, tmpDir, "Add multiple files")

	// Verify that getCommitsInRange returns exactly 1 commit
	commits, err := getCommitsInRange(commitHash)
	if err != nil {
		t.Fatalf("getCommitsInRange() error = %v", err)
	}

	if len(commits) != 1 {
		t.Errorf("getCommitsInRange() = %d commits, want 1", len(commits))
	}

	// Verify the commit hash matches
	if len(commits) > 0 && !strings.HasPrefix(commits[0], commitHash[:7]) {
		t.Errorf("getCommitsInRange() returned unexpected commit: got %s, want %s", commits[0][:7], commitHash[:7])
	}
}


// TestCalculateAuthorContribution は按分計算ロジックをテーブル駆動で検証する
func TestCalculateAuthorContribution(t *testing.T) {
	tests := []struct {
		name             string
		authorLines      int
		totalAuthorLines int
		totalAdded       int
		totalDeleted     int
		authorCount      int
		wantAdded        int
		wantDeleted      int
	}{
		{
			name:             "正常系: 30/100の按分でtotalAdded=50",
			authorLines:      30,
			totalAuthorLines: 100,
			totalAdded:       50,
			totalDeleted:     10,
			authorCount:      2,
			wantAdded:        15, // int(50 * 30/100) = 15
			wantDeleted:      3,  // int(10 * 30/100) = 3
		},
		{
			name:             "正常系: 100%の按分（単独作成者）",
			authorLines:      100,
			totalAuthorLines: 100,
			totalAdded:       50,
			totalDeleted:     20,
			authorCount:      1,
			wantAdded:        50,
			wantDeleted:      20,
		},
		{
			name:             "正常系: 50/50の按分",
			authorLines:      50,
			totalAuthorLines: 100,
			totalAdded:       80,
			totalDeleted:     40,
			authorCount:      2,
			wantAdded:        40, // int(80 * 0.5) = 40
			wantDeleted:      20, // int(40 * 0.5) = 20
		},
		{
			name:             "エッジケース: totalAuthorLines=0, authorCount=1（削除のみ返す）",
			authorLines:      0,
			totalAuthorLines: 0,
			totalAdded:       0,
			totalDeleted:     15,
			authorCount:      1,
			wantAdded:        0,
			wantDeleted:      15,
		},
		{
			name:             "エッジケース: totalAuthorLines=0, authorCount>1（ゼロ返却）",
			authorLines:      0,
			totalAuthorLines: 0,
			totalAdded:       10,
			totalDeleted:     5,
			authorCount:      3,
			wantAdded:        0,
			wantDeleted:      0,
		},
		{
			name:             "ゼロ値テスト: すべて0",
			authorLines:      0,
			totalAuthorLines: 0,
			totalAdded:       0,
			totalDeleted:     0,
			authorCount:      1,
			wantAdded:        0,
			wantDeleted:      0,
		},
		{
			name:             "按分の端数切り捨て",
			authorLines:      1,
			totalAuthorLines: 3,
			totalAdded:       10,
			totalDeleted:     7,
			authorCount:      3,
			wantAdded:        3, // int(10 * 1/3) = int(3.33) = 3
			wantDeleted:      2, // int(7 * 1/3) = int(2.33) = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			added, deleted := calculateAuthorContribution(
				tt.authorLines, tt.totalAuthorLines,
				tt.totalAdded, tt.totalDeleted, tt.authorCount,
			)
			if added != tt.wantAdded {
				t.Errorf("added = %d, want %d", added, tt.wantAdded)
			}
			if deleted != tt.wantDeleted {
				t.Errorf("deleted = %d, want %d", deleted, tt.wantDeleted)
			}
		})
	}
}

// TestAccumulateMetrics はAI/Humanの作成者タイプに基づくメトリクス累積を検証する
func TestAccumulateMetrics(t *testing.T) {
	tests := []struct {
		name       string
		authorType tracker.AuthorType
		added      int
		deleted    int
		// 期待値: WorkVolume と Contributions
		wantAIWorkChanges    int
		wantAIWorkAdded      int
		wantAIWorkDeleted    int
		wantHumanWorkChanges int
		wantHumanWorkAdded   int
		wantHumanWorkDeleted int
		wantAIContrib        int
		wantHumanContrib     int
		wantTotalAI          int
		wantTotalHuman       int
	}{
		{
			name:                 "AI作成者のメトリクス累積",
			authorType:           tracker.AuthorTypeAI,
			added:                30,
			deleted:              10,
			wantAIWorkChanges:    40,
			wantAIWorkAdded:      30,
			wantAIWorkDeleted:    10,
			wantHumanWorkChanges: 0,
			wantHumanWorkAdded:   0,
			wantHumanWorkDeleted: 0,
			wantAIContrib:        30,
			wantHumanContrib:     0,
			wantTotalAI:          30,
			wantTotalHuman:       0,
		},
		{
			name:                 "Human作成者のメトリクス累積",
			authorType:           tracker.AuthorTypeHuman,
			added:                20,
			deleted:              5,
			wantAIWorkChanges:    0,
			wantAIWorkAdded:      0,
			wantAIWorkDeleted:    0,
			wantHumanWorkChanges: 25,
			wantHumanWorkAdded:   20,
			wantHumanWorkDeleted: 5,
			wantAIContrib:        0,
			wantHumanContrib:     20,
			wantTotalAI:          0,
			wantTotalHuman:       20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 各テストケースで新しいresultを初期化
			result := &authorStatsResult{
				byAuthor: make(map[string]*tracker.AuthorStats),
			}

			accumulateMetrics(result, tt.authorType, tt.added, tt.deleted)

			// WorkVolume検証
			wv := result.detailedMetrics.WorkVolume
			if wv.AIChanges != tt.wantAIWorkChanges {
				t.Errorf("WorkVolume.AIChanges = %d, want %d", wv.AIChanges, tt.wantAIWorkChanges)
			}
			if wv.AIAdded != tt.wantAIWorkAdded {
				t.Errorf("WorkVolume.AIAdded = %d, want %d", wv.AIAdded, tt.wantAIWorkAdded)
			}
			if wv.AIDeleted != tt.wantAIWorkDeleted {
				t.Errorf("WorkVolume.AIDeleted = %d, want %d", wv.AIDeleted, tt.wantAIWorkDeleted)
			}
			if wv.HumanChanges != tt.wantHumanWorkChanges {
				t.Errorf("WorkVolume.HumanChanges = %d, want %d", wv.HumanChanges, tt.wantHumanWorkChanges)
			}
			if wv.HumanAdded != tt.wantHumanWorkAdded {
				t.Errorf("WorkVolume.HumanAdded = %d, want %d", wv.HumanAdded, tt.wantHumanWorkAdded)
			}
			if wv.HumanDeleted != tt.wantHumanWorkDeleted {
				t.Errorf("WorkVolume.HumanDeleted = %d, want %d", wv.HumanDeleted, tt.wantHumanWorkDeleted)
			}

			// Contributions検証
			c := result.detailedMetrics.Contributions
			if c.AIAdditions != tt.wantAIContrib {
				t.Errorf("Contributions.AIAdditions = %d, want %d", c.AIAdditions, tt.wantAIContrib)
			}
			if c.HumanAdditions != tt.wantHumanContrib {
				t.Errorf("Contributions.HumanAdditions = %d, want %d", c.HumanAdditions, tt.wantHumanContrib)
			}

			// totalAI/totalHuman検証
			if result.totalAI != tt.wantTotalAI {
				t.Errorf("totalAI = %d, want %d", result.totalAI, tt.wantTotalAI)
			}
			if result.totalHuman != tt.wantTotalHuman {
				t.Errorf("totalHuman = %d, want %d", result.totalHuman, tt.wantTotalHuman)
			}
		})
	}
}

// TestAccumulateMetrics_Cumulative は複数回の累積呼び出しが正しく加算されることを検証する
func TestAccumulateMetrics_Cumulative(t *testing.T) {
	result := &authorStatsResult{
		byAuthor: make(map[string]*tracker.AuthorStats),
	}

	// AI 30+10, Human 20+5 を順次累積
	accumulateMetrics(result, tracker.AuthorTypeAI, 30, 10)
	accumulateMetrics(result, tracker.AuthorTypeHuman, 20, 5)
	accumulateMetrics(result, tracker.AuthorTypeAI, 15, 3)

	// AI: added=30+15=45, deleted=10+3=13, changes=40+18=58
	if result.detailedMetrics.WorkVolume.AIChanges != 58 {
		t.Errorf("cumulative AIChanges = %d, want 58", result.detailedMetrics.WorkVolume.AIChanges)
	}
	if result.detailedMetrics.WorkVolume.AIAdded != 45 {
		t.Errorf("cumulative AIAdded = %d, want 45", result.detailedMetrics.WorkVolume.AIAdded)
	}
	if result.detailedMetrics.Contributions.AIAdditions != 45 {
		t.Errorf("cumulative AIAdditions = %d, want 45", result.detailedMetrics.Contributions.AIAdditions)
	}
	if result.totalAI != 45 {
		t.Errorf("cumulative totalAI = %d, want 45", result.totalAI)
	}

	// Human: added=20, deleted=5, changes=25
	if result.detailedMetrics.WorkVolume.HumanChanges != 25 {
		t.Errorf("cumulative HumanChanges = %d, want 25", result.detailedMetrics.WorkVolume.HumanChanges)
	}
	if result.totalHuman != 20 {
		t.Errorf("cumulative totalHuman = %d, want 20", result.totalHuman)
	}
}

// TestProcessFileAuthors は1ファイル内の作成者ごとの行数按分を検証する
func TestProcessFileAuthors(t *testing.T) {
	t.Run("単独AI作成者_numstat_10_2", func(t *testing.T) {
		result := &authorStatsResult{
			byAuthor: make(map[string]*tracker.AuthorStats),
		}
		authorsInCommit := make(map[string]bool)

		fileInfo := tracker.FileInfo{
			Authors: []tracker.AuthorInfo{
				{
					Name:  "claude",
					Type:  tracker.AuthorTypeAI,
					Lines: [][]int{{1, 10}}, // CountLines = 10
				},
			},
		}
		numstat := [2]int{10, 2} // added=10, deleted=2

		processFileAuthors(result, fileInfo, numstat, authorsInCommit)

		// 単独作成者なので100%按分: added=10, deleted=2
		stats := result.byAuthor["claude"]
		if stats == nil {
			t.Fatal("claude のAuthorStatsが作成されていない")
		}
		if stats.Lines != 10 {
			t.Errorf("claude.Lines = %d, want 10", stats.Lines)
		}
		if stats.Type != tracker.AuthorTypeAI {
			t.Errorf("claude.Type = %q, want %q", stats.Type, tracker.AuthorTypeAI)
		}

		// authorsInCommitに登録されている
		if !authorsInCommit["claude"] {
			t.Error("claude がauthorsInCommitに登録されていない")
		}

		// メトリクス検証
		if result.totalAI != 10 {
			t.Errorf("totalAI = %d, want 10", result.totalAI)
		}
		if result.detailedMetrics.WorkVolume.AIAdded != 10 {
			t.Errorf("WorkVolume.AIAdded = %d, want 10", result.detailedMetrics.WorkVolume.AIAdded)
		}
		if result.detailedMetrics.WorkVolume.AIDeleted != 2 {
			t.Errorf("WorkVolume.AIDeleted = %d, want 2", result.detailedMetrics.WorkVolume.AIDeleted)
		}
	})

	t.Run("複数作成者_AI_Human_按分", func(t *testing.T) {
		result := &authorStatsResult{
			byAuthor: make(map[string]*tracker.AuthorStats),
		}
		authorsInCommit := make(map[string]bool)

		fileInfo := tracker.FileInfo{
			Authors: []tracker.AuthorInfo{
				{
					Name:  "claude",
					Type:  tracker.AuthorTypeAI,
					Lines: [][]int{{1, 30}}, // CountLines = 30
				},
				{
					Name:  "developer",
					Type:  tracker.AuthorTypeHuman,
					Lines: [][]int{{31, 40}, {45}}, // CountLines = 10 + 1 = 11
				},
			},
		}
		// totalAuthorLines = 30 + 11 = 41
		numstat := [2]int{41, 10} // added=41, deleted=10

		processFileAuthors(result, fileInfo, numstat, authorsInCommit)

		// claude: ratio = 30/41
		// added = int(41 * 30/41) = int(30.0) = 30
		// deleted = int(10 * 30/41) = int(7.31) = 7
		claudeStats := result.byAuthor["claude"]
		if claudeStats == nil {
			t.Fatal("claude のAuthorStatsが作成されていない")
		}
		if claudeStats.Lines != 30 {
			t.Errorf("claude.Lines = %d, want 30", claudeStats.Lines)
		}

		// developer: ratio = 11/41
		// added = int(41 * 11/41) = int(11.0) = 11
		devStats := result.byAuthor["developer"]
		if devStats == nil {
			t.Fatal("developer のAuthorStatsが作成されていない")
		}
		if devStats.Lines != 11 {
			t.Errorf("developer.Lines = %d, want 11", devStats.Lines)
		}

		// 両方がauthorsInCommitに登録
		if !authorsInCommit["claude"] {
			t.Error("claude がauthorsInCommitに登録されていない")
		}
		if !authorsInCommit["developer"] {
			t.Error("developer がauthorsInCommitに登録されていない")
		}

		// メトリクス: AI=30 added, Human=11 added
		if result.totalAI != 30 {
			t.Errorf("totalAI = %d, want 30", result.totalAI)
		}
		if result.totalHuman != 11 {
			t.Errorf("totalHuman = %d, want 11", result.totalHuman)
		}
	})
}

// TestProcessCommitFiles はAuthorshipLogとnumstatMapから正しく集計されることを検証する
func TestProcessCommitFiles(t *testing.T) {
	t.Run("正常系: 2ファイルの集計", func(t *testing.T) {
		result := &authorStatsResult{
			byAuthor: make(map[string]*tracker.AuthorStats),
		}

		alog := &tracker.AuthorshipLog{
			Version: "1",
			Commit:  "abc123",
			Files: map[string]tracker.FileInfo{
				"main.go": {
					Authors: []tracker.AuthorInfo{
						{Name: "claude", Type: tracker.AuthorTypeAI, Lines: [][]int{{1, 20}}},
					},
				},
				"utils.go": {
					Authors: []tracker.AuthorInfo{
						{Name: "developer", Type: tracker.AuthorTypeHuman, Lines: [][]int{{1, 10}}},
					},
				},
			},
		}

		numstatMap := map[string][2]int{
			"main.go":  {20, 5},
			"utils.go": {10, 3},
		}

		authorsInCommit := processCommitFiles(result, alog, numstatMap)

		// claude: main.go から added=20
		if result.byAuthor["claude"] == nil {
			t.Fatal("claude のAuthorStatsが作成されていない")
		}
		if result.byAuthor["claude"].Lines != 20 {
			t.Errorf("claude.Lines = %d, want 20", result.byAuthor["claude"].Lines)
		}

		// developer: utils.go から added=10
		if result.byAuthor["developer"] == nil {
			t.Fatal("developer のAuthorStatsが作成されていない")
		}
		if result.byAuthor["developer"].Lines != 10 {
			t.Errorf("developer.Lines = %d, want 10", result.byAuthor["developer"].Lines)
		}

		// authorsInCommitに両方登録
		if !authorsInCommit["claude"] {
			t.Error("claude がauthorsInCommitに登録されていない")
		}
		if !authorsInCommit["developer"] {
			t.Error("developer がauthorsInCommitに登録されていない")
		}
	})

	t.Run("ファイルがnumstatにない場合スキップ", func(t *testing.T) {
		result := &authorStatsResult{
			byAuthor: make(map[string]*tracker.AuthorStats),
		}

		alog := &tracker.AuthorshipLog{
			Version: "1",
			Commit:  "abc123",
			Files: map[string]tracker.FileInfo{
				"main.go": {
					Authors: []tracker.AuthorInfo{
						{Name: "claude", Type: tracker.AuthorTypeAI, Lines: [][]int{{1, 20}}},
					},
				},
				"missing.go": {
					Authors: []tracker.AuthorInfo{
						{Name: "developer", Type: tracker.AuthorTypeHuman, Lines: [][]int{{1, 10}}},
					},
				},
			},
		}

		// missing.go はnumstatMapにない
		numstatMap := map[string][2]int{
			"main.go": {20, 5},
		}

		authorsInCommit := processCommitFiles(result, alog, numstatMap)

		// claude は存在する（main.goがnumstatにある）
		if result.byAuthor["claude"] == nil {
			t.Fatal("claude のAuthorStatsが作成されていない")
		}
		if result.byAuthor["claude"].Lines != 20 {
			t.Errorf("claude.Lines = %d, want 20", result.byAuthor["claude"].Lines)
		}

		// developer は存在しない（missing.goがnumstatにない）
		if result.byAuthor["developer"] != nil {
			t.Errorf("developer は作成されるべきではないが、Lines=%d で存在する", result.byAuthor["developer"].Lines)
		}

		// authorsInCommitにはclaudeのみ
		if !authorsInCommit["claude"] {
			t.Error("claude がauthorsInCommitに登録されていない")
		}
		if authorsInCommit["developer"] {
			t.Error("developer がauthorsInCommitに登録されるべきではない")
		}
	})
}

// TestConvertSinceToRange はモックExecutorを使ってconvertSinceToRangeの変換ロジックを検証する
func TestConvertSinceToRange(t *testing.T) {
	t.Run("正常系: 親コミットが存在する場合", func(t *testing.T) {
		// DI差し替え
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		mock.RunFunc = func(args ...string) (string, error) {
			// 1回目: git log --since=... --format=%H --reverse
			if len(args) >= 3 && args[0] == "log" {
				return "aaa111\nbbb222\nccc333", nil
			}
			// 2回目: git rev-parse aaa111^（親の存在確認）
			if len(args) >= 1 && args[0] == "rev-parse" {
				return "parent-hash", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}
		newExecutor = func() gitexec.Executor { return mock }

		result, err := convertSinceToRange("7d")
		if err != nil {
			t.Fatalf("convertSinceToRange(\"7d\") error = %v", err)
		}

		// 親が存在するので firstCommit^..HEAD
		expected := "aaa111^..HEAD"
		if result != expected {
			t.Errorf("convertSinceToRange(\"7d\") = %q, want %q", result, expected)
		}
	})

	t.Run("初回コミット: 親がない場合", func(t *testing.T) {
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		mock.RunFunc = func(args ...string) (string, error) {
			if len(args) >= 3 && args[0] == "log" {
				return "initial-commit-hash", nil
			}
			if len(args) >= 1 && args[0] == "rev-parse" {
				// 親がない場合はエラーを返す
				return "", fmt.Errorf("rev-parse failed: no parent")
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}
		newExecutor = func() gitexec.Executor { return mock }

		result, err := convertSinceToRange("1y")
		if err != nil {
			t.Fatalf("convertSinceToRange(\"1y\") error = %v", err)
		}

		// 親がないので firstCommit..HEAD
		expected := "initial-commit-hash..HEAD"
		if result != expected {
			t.Errorf("convertSinceToRange(\"1y\") = %q, want %q", result, expected)
		}
	})

	t.Run("コミットなしの場合のエラー", func(t *testing.T) {
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		mock.RunFunc = func(args ...string) (string, error) {
			if len(args) >= 3 && args[0] == "log" {
				// 空の出力 = コミットなし
				return "", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}
		newExecutor = func() gitexec.Executor { return mock }

		_, err := convertSinceToRange("7d")
		if err == nil {
			t.Fatal("convertSinceToRange should return error when no commits found")
		}
		if !strings.Contains(err.Error(), "no commits found") {
			t.Errorf("error message should contain 'no commits found', got: %v", err)
		}
	})

	t.Run("git logコマンドのエラー", func(t *testing.T) {
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		mock.RunFunc = func(args ...string) (string, error) {
			if len(args) >= 3 && args[0] == "log" {
				return "", fmt.Errorf("git log failed: not a git repository")
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}
		newExecutor = func() gitexec.Executor { return mock }

		_, err := convertSinceToRange("7d")
		if err == nil {
			t.Fatal("convertSinceToRange should return error when git log fails")
		}
		if !strings.Contains(err.Error(), "failed to get commits") {
			t.Errorf("error message should contain 'failed to get commits', got: %v", err)
		}
	})

	t.Run("shorthand展開の確認: 2wが正しくgit logに渡される", func(t *testing.T) {
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		var capturedSinceArg string
		mock.RunFunc = func(args ...string) (string, error) {
			if len(args) >= 3 && args[0] == "log" {
				// --since=引数をキャプチャ
				for _, arg := range args {
					if strings.HasPrefix(arg, "--since=") {
						capturedSinceArg = arg
					}
				}
				return "commit-abc", nil
			}
			if args[0] == "rev-parse" {
				return "parent-hash", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}
		newExecutor = func() gitexec.Executor { return mock }

		_, err := convertSinceToRange("2w")
		if err != nil {
			t.Fatalf("convertSinceToRange(\"2w\") error = %v", err)
		}

		// 2w は "2 weeks ago" に展開されるはず
		expectedArg := "--since=2 weeks ago"
		if capturedSinceArg != expectedArg {
			t.Errorf("git log received %q, want %q", capturedSinceArg, expectedArg)
		}
	})
}
