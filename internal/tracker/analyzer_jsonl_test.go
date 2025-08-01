package tracker

import (
	"strings"
	"testing"
	"time"
)

func TestAnalyzeRecords(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	// Create test records
	records := []CheckpointRecord{
		{
			Timestamp: time.Now().Add(-3 * time.Hour),
			Author:    "human",
			Commit:    "a1",
			Added:     100,
			Deleted:   0,
		},
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Author:    "claude",
			Commit:    "a2",
			Added:     150, // 50 lines added by AI
			Deleted:   10,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "human",
			Commit:    "a3",
			Added:     180, // 30 lines added by human
			Deleted:   20,
		},
		{
			Timestamp: time.Now(),
			Author:    "ai-assistant",
			Commit:    "a4",
			Added:     200, // 20 lines added by AI
			Deleted:   25,
		},
	}

	result, err := analyzer.AnalyzeRecords(records)
	if err != nil {
		t.Fatalf("Failed to analyze records: %v", err)
	}

	// Expected: 50 + 20 = 70 AI lines, 30 human lines
	if result.AILines != 70 {
		t.Errorf("Expected 70 AI lines, got %d", result.AILines)
	}

	if result.HumanLines != 30 {
		t.Errorf("Expected 30 human lines, got %d", result.HumanLines)
	}

	expectedPercentage := 70.0 // 70 / (70 + 30) * 100
	if result.Percentage != expectedPercentage {
		t.Errorf("Expected percentage %.1f%%, got %.1f%%", expectedPercentage, result.Percentage)
	}

	// Check last updated timestamp
	if !result.LastUpdated.Equal(records[len(records)-1].Timestamp) {
		t.Error("Expected last updated to match last record timestamp")
	}
}

func TestAnalyzeRecordsEmpty(t *testing.T) {
	config := &Config{}
	analyzer := NewAnalyzer(config)

	result, err := analyzer.AnalyzeRecords([]CheckpointRecord{})
	if err != nil {
		t.Fatalf("Failed to analyze empty records: %v", err)
	}

	if result.AILines != 0 {
		t.Errorf("Expected 0 AI lines for empty records, got %d", result.AILines)
	}

	if result.HumanLines != 0 {
		t.Errorf("Expected 0 human lines for empty records, got %d", result.HumanLines)
	}

	if result.Percentage != 0 {
		t.Errorf("Expected 0%% for empty records, got %.1f%%", result.Percentage)
	}
}

func TestCalculateRecordDiff(t *testing.T) {
	config := &Config{}
	analyzer := NewAnalyzer(config)

	tests := []struct {
		before   CheckpointRecord
		after    CheckpointRecord
		expected int
	}{
		{
			before:   CheckpointRecord{Added: 100, Deleted: 10},
			after:    CheckpointRecord{Added: 150, Deleted: 20},
			expected: 50, // 150 - 100
		},
		{
			before:   CheckpointRecord{Added: 100, Deleted: 10},
			after:    CheckpointRecord{Added: 100, Deleted: 20},
			expected: 0, // No change in added lines
		},
		{
			before:   CheckpointRecord{Added: 100, Deleted: 10},
			after:    CheckpointRecord{Added: 90, Deleted: 30},
			expected: 0, // Negative diff returns 0
		},
	}

	for i, test := range tests {
		result := analyzer.calculateRecordDiff(test.before, test.after)
		if result != test.expected {
			t.Errorf("Test %d: expected diff %d, got %d", i, test.expected, result)
		}
	}
}

func TestGenerateReportFromRecords(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	records := []CheckpointRecord{
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Author:    "human",
			Added:     0,
			Deleted:   0,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "claude",
			Added:     60,
			Deleted:   0,
		},
		{
			Timestamp: time.Now(),
			Author:    "human",
			Added:     100,
			Deleted:   10,
		},
	}

	report, err := analyzer.GenerateReportFromRecords(records, 1000)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Check report content
	if !strings.Contains(report, "AI Code Tracking Report") {
		t.Error("Report should contain title")
	}

	if !strings.Contains(report, "Added Lines: 100") {
		t.Error("Report should show total added lines")
	}

	if !strings.Contains(report, "AI Lines: 60") {
		t.Error("Report should show AI lines")
	}

	if !strings.Contains(report, "Human Lines: 40") {
		t.Error("Report should show human lines")
	}

	if !strings.Contains(report, "Target: 80.0% AI code") {
		t.Error("Report should show target percentage")
	}

	// Progress should be 60/100 * 100 / 80 = 75%
	if !strings.Contains(report, "Progress: 75.0%") {
		t.Error("Report should show correct progress percentage")
	}
}

func TestGetFileStatsFromRecords(t *testing.T) {
	config := &Config{}
	analyzer := NewAnalyzer(config)

	records := []CheckpointRecord{
		{
			Timestamp: time.Now(),
			Author:    "test",
			Added:     100,
			Deleted:   50,
		},
	}

	stats := analyzer.GetFileStatsFromRecords(records)

	// Should return empty slice since per-file stats are not available
	if len(stats) != 0 {
		t.Errorf("Expected empty file stats, got %d entries", len(stats))
	}
}

func TestAnalyzeRecordsWithAuthorMapping(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		AuthorMappings: map[string]string{
			"GPT Bot":     "ai",
			"Developer 1": "human",
		},
	}

	analyzer := NewAnalyzer(config)

	records := []CheckpointRecord{
		{
			Timestamp: time.Now().Add(-2 * time.Hour),
			Author:    "Developer 1",
			Added:     0,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "GPT Bot", // Mapped to AI
			Added:     50,
		},
		{
			Timestamp: time.Now(),
			Author:    "Developer 1", // Mapped to human
			Added:     80,
		},
	}

	result, err := analyzer.AnalyzeRecords(records)
	if err != nil {
		t.Fatalf("Failed to analyze records: %v", err)
	}

	if result.AILines != 50 {
		t.Errorf("Expected 50 AI lines (from GPT Bot), got %d", result.AILines)
	}

	if result.HumanLines != 30 {
		t.Errorf("Expected 30 human lines (from Developer 1), got %d", result.HumanLines)
	}
}
