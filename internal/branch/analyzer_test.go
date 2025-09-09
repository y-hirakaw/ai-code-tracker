package branch

import (
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func createTestRecords() []tracker.CheckpointRecord {
	baseTime := time.Date(2025, 1, 20, 10, 0, 0, 0, time.UTC)

	return []tracker.CheckpointRecord{
		{
			Timestamp: baseTime,
			Author:    "human",
			Branch:    "main",
			Added:     100,
			Deleted:   10,
		},
		{
			Timestamp: baseTime.Add(1 * time.Hour),
			Author:    "ai-assistant",
			Branch:    "main",
			Added:     200,
			Deleted:   20,
		},
		{
			Timestamp: baseTime.Add(2 * time.Hour),
			Author:    "human",
			Branch:    "feature/ui-improve",
			Added:     150,
			Deleted:   5,
		},
		{
			Timestamp: baseTime.Add(3 * time.Hour),
			Author:    "ai-assistant",
			Branch:    "feature/ui-improve",
			Added:     300,
			Deleted:   15,
		},
		{
			Timestamp: baseTime.Add(4 * time.Hour),
			Author:    "human",
			Branch:    "feature/api-v2",
			Added:     80,
			Deleted:   8,
		},
		{
			Timestamp: baseTime.Add(5 * time.Hour),
			Author:    "ai-assistant",
			Branch:    "hotfix/critical-bug",
			Added:     50,
			Deleted:   2,
		},
		// Test record without explicit branch info (should use backward compatibility)
		{
			Timestamp: baseTime.Add(6 * time.Hour),
			Author:    "claude",
			Branch:    "", // Empty branch - should default to "main"
			Added:     75,
			Deleted:   3,
		},
	}
}

func TestBranchAnalyzer_AnalyzeByBranch(t *testing.T) {
	records := createTestRecords()
	analyzer := NewBranchAnalyzer(records)

	tests := []struct {
		name               string
		branchName         string
		expectedRecords    int
		expectedTotalAdded int
		expectedAuthors    []string
	}{
		{
			name:               "main branch",
			branchName:         "main",
			expectedRecords:    3, // 2 explicit + 1 inferred from empty branch
			expectedTotalAdded: 375, // 100 + 200 + 75
			expectedAuthors:    []string{"ai-assistant", "claude", "human"},
		},
		{
			name:               "feature/ui-improve branch",
			branchName:         "feature/ui-improve",
			expectedRecords:    2,
			expectedTotalAdded: 450, // 150 + 300
			expectedAuthors:    []string{"ai-assistant", "human"},
		},
		{
			name:               "non-existent branch",
			branchName:         "non-existent",
			expectedRecords:    0,
			expectedTotalAdded: 0,
			expectedAuthors:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := analyzer.AnalyzeByBranch(tt.branchName)
			if err != nil {
				t.Errorf("BranchAnalyzer.AnalyzeByBranch() error = %v", err)
				return
			}

			if report.BranchName != tt.branchName {
				t.Errorf("BranchReport.BranchName = %v, want %v", report.BranchName, tt.branchName)
			}

			if report.RecordCount != tt.expectedRecords {
				t.Errorf("BranchReport.RecordCount = %v, want %v", report.RecordCount, tt.expectedRecords)
			}

			if report.TotalAdded != tt.expectedTotalAdded {
				t.Errorf("BranchReport.TotalAdded = %v, want %v", report.TotalAdded, tt.expectedTotalAdded)
			}

			if len(report.Authors) != len(tt.expectedAuthors) {
				t.Errorf("BranchReport.Authors length = %v, want %v", len(report.Authors), len(tt.expectedAuthors))
			}

			// Check if all expected authors are present (order may vary due to sorting)
			authorMap := make(map[string]bool)
			for _, author := range report.Authors {
				authorMap[author] = true
			}

			for _, expectedAuthor := range tt.expectedAuthors {
				if !authorMap[expectedAuthor] {
					t.Errorf("BranchReport.Authors missing expected author: %v", expectedAuthor)
				}
			}

			// Check AI ratio calculation (should be > 0 for records with data)
			if tt.expectedRecords > 0 && report.AIRatio <= 0 {
				t.Errorf("BranchReport.AIRatio = %v, want > 0 for branch with records", report.AIRatio)
			}
		})
	}
}

func TestBranchAnalyzer_AnalyzeByPattern(t *testing.T) {
	records := createTestRecords()
	analyzer := NewBranchAnalyzer(records)

	tests := []struct {
		name                    string
		pattern                 string
		isRegex                 bool
		expectedBranchCount     int
		expectedTotalRecords    int
		shouldContainBranches   []string
		shouldNotContainBranches []string
	}{
		{
			name:                  "feature branches regex",
			pattern:               "^feature/",
			isRegex:               true,
			expectedBranchCount:   2,
			expectedTotalRecords:  3, // 2 from feature/ui-improve + 1 from feature/api-v2
			shouldContainBranches: []string{"feature/ui-improve", "feature/api-v2"},
			shouldNotContainBranches: []string{"main", "hotfix/critical-bug"},
		},
		{
			name:                  "hotfix branches regex",
			pattern:               "^hotfix/",
			isRegex:               true,
			expectedBranchCount:   1,
			expectedTotalRecords:  1,
			shouldContainBranches: []string{"hotfix/critical-bug"},
			shouldNotContainBranches: []string{"main", "feature/ui-improve"},
		},
		{
			name:                  "all branches (empty pattern)",
			pattern:               "",
			isRegex:               false,
			expectedBranchCount:   4, // main, feature/ui-improve, feature/api-v2, hotfix/critical-bug
			expectedTotalRecords:  7,
			shouldContainBranches: []string{"main", "feature/ui-improve", "feature/api-v2", "hotfix/critical-bug"},
			shouldNotContainBranches: []string{},
		},
		{
			name:                  "non-matching pattern",
			pattern:               "^release/",
			isRegex:               true,
			expectedBranchCount:   0,
			expectedTotalRecords:  0,
			shouldContainBranches: []string{},
			shouldNotContainBranches: []string{"main", "feature/ui-improve"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := analyzer.AnalyzeByPattern(tt.pattern, tt.isRegex)
			if err != nil {
				t.Errorf("BranchAnalyzer.AnalyzeByPattern() error = %v", err)
				return
			}

			if len(report.MatchingBranches) != tt.expectedBranchCount {
				t.Errorf("GroupReport.MatchingBranches count = %v, want %v", len(report.MatchingBranches), tt.expectedBranchCount)
			}

			if report.TotalRecords != tt.expectedTotalRecords {
				t.Errorf("GroupReport.TotalRecords = %v, want %v", report.TotalRecords, tt.expectedTotalRecords)
			}

			// Check that expected branches are present
			branchMap := make(map[string]bool)
			for _, branch := range report.MatchingBranches {
				branchMap[branch] = true
			}

			for _, expectedBranch := range tt.shouldContainBranches {
				if !branchMap[expectedBranch] {
					t.Errorf("GroupReport.MatchingBranches missing expected branch: %v", expectedBranch)
				}
			}

			for _, unexpectedBranch := range tt.shouldNotContainBranches {
				if branchMap[unexpectedBranch] {
					t.Errorf("GroupReport.MatchingBranches contains unexpected branch: %v", unexpectedBranch)
				}
			}

			// Check that branch reports exist for all matching branches
			for _, branch := range report.MatchingBranches {
				if _, exists := report.BranchReports[branch]; !exists {
					t.Errorf("GroupReport.BranchReports missing report for branch: %v", branch)
				}
			}

			// Check group AI ratio calculation
			if tt.expectedTotalRecords > 0 && report.GroupAIRatio <= 0 {
				t.Errorf("GroupReport.GroupAIRatio = %v, want > 0 for group with records", report.GroupAIRatio)
			}
		})
	}
}

func TestBranchAnalyzer_AnalyzeByPattern_InvalidRegex(t *testing.T) {
	records := createTestRecords()
	analyzer := NewBranchAnalyzer(records)

	_, err := analyzer.AnalyzeByPattern("[invalid", true)
	if err == nil {
		t.Errorf("BranchAnalyzer.AnalyzeByPattern() expected error for invalid regex, got nil")
	}
}

func TestBranchAnalyzer_GetUniqueBranches(t *testing.T) {
	records := createTestRecords()
	analyzer := NewBranchAnalyzer(records)

	branches := analyzer.GetUniqueBranches()

	// Should have 4 unique branches: main, feature/ui-improve, feature/api-v2, hotfix/critical-bug
	// Note: empty branch gets converted to "main" via GetBranch()
	expected := []string{"feature/api-v2", "feature/ui-improve", "hotfix/critical-bug", "main"}

	if len(branches) != len(expected) {
		t.Errorf("GetUniqueBranches() returned %d branches, want %d", len(branches), len(expected))
	}

	for i, expectedBranch := range expected {
		if i >= len(branches) || branches[i] != expectedBranch {
			t.Errorf("GetUniqueBranches()[%d] = %v, want %v", i, branches[i], expectedBranch)
		}
	}
}

func TestBranchAnalyzer_GetRecordStats(t *testing.T) {
	records := createTestRecords()
	analyzer := NewBranchAnalyzer(records)

	stats := analyzer.GetRecordStats()

	expectedTotalRecords := 7
	expectedUniqueBranches := 4
	expectedRecordsWithBranch := 6    // All except the one with empty Branch
	expectedRecordsWithoutBranch := 1 // The one with empty Branch

	if stats.TotalRecords != expectedTotalRecords {
		t.Errorf("RecordStats.TotalRecords = %v, want %v", stats.TotalRecords, expectedTotalRecords)
	}

	if stats.UniqueBranches != expectedUniqueBranches {
		t.Errorf("RecordStats.UniqueBranches = %v, want %v", stats.UniqueBranches, expectedUniqueBranches)
	}

	if stats.RecordsWithBranch != expectedRecordsWithBranch {
		t.Errorf("RecordStats.RecordsWithBranch = %v, want %v", stats.RecordsWithBranch, expectedRecordsWithBranch)
	}

	if stats.RecordsWithoutBranch != expectedRecordsWithoutBranch {
		t.Errorf("RecordStats.RecordsWithoutBranch = %v, want %v", stats.RecordsWithoutBranch, expectedRecordsWithoutBranch)
	}

	// Check totals
	expectedTotalAdded := 955 // Sum of all Added fields
	if stats.TotalAdded != expectedTotalAdded {
		t.Errorf("RecordStats.TotalAdded = %v, want %v", stats.TotalAdded, expectedTotalAdded)
	}
}

func TestBranchAnalyzer_EmptyRecords(t *testing.T) {
	analyzer := NewBranchAnalyzer([]tracker.CheckpointRecord{})

	// Test AnalyzeByBranch with empty records
	report, err := analyzer.AnalyzeByBranch("main")
	if err != nil {
		t.Errorf("BranchAnalyzer.AnalyzeByBranch() with empty records error = %v", err)
	}

	if report.RecordCount != 0 {
		t.Errorf("BranchReport.RecordCount = %v, want 0 for empty records", report.RecordCount)
	}

	// Test GetUniqueBranches with empty records
	branches := analyzer.GetUniqueBranches()
	if len(branches) != 0 {
		t.Errorf("GetUniqueBranches() with empty records = %v, want empty slice", branches)
	}

	// Test GetRecordStats with empty records
	stats := analyzer.GetRecordStats()
	if stats.TotalRecords != 0 {
		t.Errorf("RecordStats.TotalRecords = %v, want 0 for empty records", stats.TotalRecords)
	}
}

func TestBranchReport_AIRatioCalculation(t *testing.T) {
	// Test the AI ratio calculation logic
	records := []tracker.CheckpointRecord{
		{
			Timestamp: time.Now(),
			Author:    "human",
			Branch:    "test",
			Added:     100,
			Deleted:   0,
		},
	}

	analyzer := NewBranchAnalyzer(records)
	report, err := analyzer.AnalyzeByBranch("test")
	if err != nil {
		t.Errorf("BranchAnalyzer.AnalyzeByBranch() error = %v", err)
		return
	}

	// With simplified 80% AI assumption, 100 lines * 0.8 = 80 AI lines
	// AI ratio should be 80/100 * 100 = 80%
	expectedRatio := 80.0
	if report.AIRatio != expectedRatio {
		t.Errorf("BranchReport.AIRatio = %v, want %v", report.AIRatio, expectedRatio)
	}
}