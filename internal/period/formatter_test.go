package period

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestFormatterTable(t *testing.T) {
	formatter := NewFormatter(80.0)
	
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 100,
		AILines:    70,
		HumanLines: 30,
		Percentage: 70.0,
		DailyStats: []DailyStat{
			{
				Date:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				AILines:    70,
				HumanLines: 30,
			},
		},
	}
	
	output, err := formatter.Format(report, FormatTable)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check that output contains expected elements
	if !strings.Contains(output, "AI Code Tracking Report (Period)") {
		t.Error("Output should contain report title")
	}
	
	if !strings.Contains(output, "Total Lines: 100") {
		t.Error("Output should contain total lines")
	}
	
	if !strings.Contains(output, "AI Lines: 70 (70.0%)") {
		t.Error("Output should contain AI lines and percentage")
	}
	
	if !strings.Contains(output, "Human Lines: 30 (30.0%)") {
		t.Error("Output should contain human lines and percentage")
	}
	
	if !strings.Contains(output, "Target: 80.0% AI code") {
		t.Error("Output should contain target percentage")
	}
	
	if !strings.Contains(output, "Progress: 87.5%") {
		t.Error("Output should contain progress percentage")
	}
	
	if !strings.Contains(output, "Daily Breakdown:") {
		t.Error("Output should contain daily breakdown section")
	}
}

func TestFormatterGraph(t *testing.T) {
	formatter := NewFormatter(80.0)
	
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 100,
		AILines:    60,
		HumanLines: 40,
		Percentage: 60.0,
		DailyStats: []DailyStat{
			{
				Date:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				AILines:    60,
				HumanLines: 40,
			},
		},
	}
	
	output, err := formatter.Format(report, FormatGraph)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check that output contains expected graph elements
	if !strings.Contains(output, "AI vs Human Code Contributions (Period)") {
		t.Error("Output should contain graph title")
	}
	
	if !strings.Contains(output, "Daily AI Percentage Trend:") {
		t.Error("Output should contain trend section")
	}
	
	if !strings.Contains(output, "█") {
		t.Error("Output should contain progress bars (█)")
	}
	
	if !strings.Contains(output, "Target") {
		t.Error("Output should contain target line")
	}
}

func TestFormatterJSON(t *testing.T) {
	formatter := NewFormatter(80.0)
	
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 100,
		AILines:    70,
		HumanLines: 30,
		Percentage: 70.0,
		DailyStats: []DailyStat{
			{
				Date:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				AILines:    70,
				HumanLines: 30,
			},
		},
	}
	
	output, err := formatter.Format(report, FormatJSON)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Verify it's valid JSON by unmarshaling
	var parsed PeriodReport
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("Output should be valid JSON: %v", err)
	}
	
	// Check key fields
	if parsed.TotalLines != 100 {
		t.Errorf("Expected total lines 100, got %d", parsed.TotalLines)
	}
	
	if parsed.AILines != 70 {
		t.Errorf("Expected AI lines 70, got %d", parsed.AILines)
	}
	
	if parsed.Percentage != 70.0 {
		t.Errorf("Expected percentage 70.0, got %.1f", parsed.Percentage)
	}
}

func TestFormatterUnsupportedFormat(t *testing.T) {
	formatter := NewFormatter(80.0)
	
	report := &PeriodReport{}
	
	_, err := formatter.Format(report, ReportFormat("invalid"))
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

func TestFormatterEmptyReport(t *testing.T) {
	formatter := NewFormatter(80.0)
	
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 0,
		AILines:    0,
		HumanLines: 0,
		Percentage: 0.0,
	}
	
	// Test table format with empty data
	output, err := formatter.Format(report, FormatTable)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(output, "Total Lines: 0") {
		t.Error("Output should handle empty report correctly")
	}
	
	// Test graph format with empty data
	graphOutput, err := formatter.Format(report, FormatGraph)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !strings.Contains(graphOutput, "No contributions in this period") {
		t.Error("Graph output should indicate no contributions")
	}
}