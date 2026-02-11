package main

import (
	"os"
	"strings"
	"testing"

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
