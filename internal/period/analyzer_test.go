package period

import (
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestAnalyzePeriod(t *testing.T) {
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".py"},
		AuthorMappings:     map[string]string{"user": "human"},
	}
	
	analyzer := NewAnalyzer(config)
	
	now := time.Now()
	records := []tracker.CheckpointRecord{
		{Timestamp: now.Add(-2 * time.Hour), Author: "human", Added: 10},
		{Timestamp: now.Add(-1 * time.Hour), Author: "claude", Added: 20},
		{Timestamp: now.Add(-30 * time.Minute), Author: "human", Added: 5},
	}
	
	timeRange := &TimeRange{
		From: now.Add(-3 * time.Hour),
		To:   now,
	}
	
	report, err := analyzer.AnalyzePeriod(records, timeRange)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Expected: AI=20, Human=15, Total=35
	expectedAI := 20
	expectedHuman := 15
	expectedTotal := 35
	expectedPercentage := float64(20) / float64(35) * 100
	
	if report.AILines != expectedAI {
		t.Errorf("Expected AI lines %d, got %d", expectedAI, report.AILines)
	}
	
	if report.HumanLines != expectedHuman {
		t.Errorf("Expected Human lines %d, got %d", expectedHuman, report.HumanLines)
	}
	
	if report.TotalLines != expectedTotal {
		t.Errorf("Expected Total lines %d, got %d", expectedTotal, report.TotalLines)
	}
	
	if report.Percentage != expectedPercentage {
		t.Errorf("Expected Percentage %.2f, got %.2f", expectedPercentage, report.Percentage)
	}
	
	if report.Range.From != timeRange.From {
		t.Errorf("Expected From time to match")
	}
	
	if report.Range.To != timeRange.To {
		t.Errorf("Expected To time to match")
	}
}

func TestGenerateDailyStats(t *testing.T) {
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go"},
		AuthorMappings:     map[string]string{},
	}
	
	analyzer := NewAnalyzer(config)
	
	day1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 1, 2, 15, 0, 0, 0, time.UTC)
	
	records := []tracker.CheckpointRecord{
		{Timestamp: day1, Author: "claude", Added: 10},
		{Timestamp: day1, Author: "human", Added: 5},
		{Timestamp: day2, Author: "claude", Added: 15},
	}
	
	dailyStats := analyzer.generateDailyStats(records)
	
	if len(dailyStats) != 2 {
		t.Errorf("Expected 2 daily stats, got %d", len(dailyStats))
	}
	
	// Check first day stats
	if dailyStats[0].AILines != 10 {
		t.Errorf("Expected day 1 AI lines 10, got %d", dailyStats[0].AILines)
	}
	
	if dailyStats[0].HumanLines != 5 {
		t.Errorf("Expected day 1 Human lines 5, got %d", dailyStats[0].HumanLines)
	}
	
	// Check second day stats
	if dailyStats[1].AILines != 15 {
		t.Errorf("Expected day 2 AI lines 15, got %d", dailyStats[1].AILines)
	}
	
	if dailyStats[1].HumanLines != 0 {
		t.Errorf("Expected day 2 Human lines 0, got %d", dailyStats[1].HumanLines)
	}
	
	// Check dates are sorted
	if dailyStats[0].Date.After(dailyStats[1].Date) {
		t.Errorf("Daily stats should be sorted by date")
	}
}

func TestAnalyzePeriodEmptyRecords(t *testing.T) {
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
	}
	
	analyzer := NewAnalyzer(config)
	
	timeRange := &TimeRange{
		From: time.Now().Add(-1 * time.Hour),
		To:   time.Now(),
	}
	
	report, err := analyzer.AnalyzePeriod([]tracker.CheckpointRecord{}, timeRange)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if report.TotalLines != 0 {
		t.Errorf("Expected 0 total lines for empty records, got %d", report.TotalLines)
	}
	
	if report.Percentage != 0 {
		t.Errorf("Expected 0 percentage for empty records, got %.2f", report.Percentage)
	}
}