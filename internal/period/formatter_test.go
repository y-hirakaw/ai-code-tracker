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

func TestFormatterCSV(t *testing.T) {
	formatter := NewFormatter(80.0)

	// Test with daily stats
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 200,
		AILines:    150,
		HumanLines: 50,
		Percentage: 75.0,
		DailyStats: []DailyStat{
			{
				Date:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				AILines:    70,
				HumanLines: 30,
			},
			{
				Date:       time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				AILines:    80,
				HumanLines: 20,
			},
		},
	}

	output, err := formatter.Format(report, FormatCSV)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header
	expectedHeader := "Date,AI_Lines,Human_Lines,Total_Lines,AI_Percentage,Human_Percentage,Target_Percentage,Progress"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header: %s, got: %s", expectedHeader, lines[0])
	}

	// Check first data row
	expectedFirstRow := "2025-01-01,70,30,100,70.0,30.0,80.0,87.5"
	if lines[1] != expectedFirstRow {
		t.Errorf("Expected first row: %s, got: %s", expectedFirstRow, lines[1])
	}

	// Check second data row
	expectedSecondRow := "2025-01-02,80,20,100,80.0,20.0,80.0,100.0"
	if lines[2] != expectedSecondRow {
		t.Errorf("Expected second row: %s, got: %s", expectedSecondRow, lines[2])
	}

	// Verify total number of lines (header + 2 data rows)
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines total, got %d", len(lines))
	}
}

func TestFormatterCSVSummary(t *testing.T) {
	formatter := NewFormatter(80.0)

	// Test without daily stats (summary mode)
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 200,
		AILines:    160,
		HumanLines: 40,
		Percentage: 80.0,
		DailyStats: []DailyStat{}, // Empty daily stats
	}

	output, err := formatter.Format(report, FormatCSV)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header for summary format
	expectedHeader := "Period_From,Period_To,AI_Lines,Human_Lines,Total_Lines,AI_Percentage,Human_Percentage,Target_Percentage,Progress"
	if lines[0] != expectedHeader {
		t.Errorf("Expected header: %s, got: %s", expectedHeader, lines[0])
	}

	// Check data row
	expectedDataRow := "2025-01-01,2025-01-05,160,40,200,80.0,20.0,80.0,100.0"
	if lines[1] != expectedDataRow {
		t.Errorf("Expected data row: %s, got: %s", expectedDataRow, lines[1])
	}

	// Verify total number of lines (header + 1 data row)
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines total, got %d", len(lines))
	}
}

func TestFormatterCSVEmptyData(t *testing.T) {
	formatter := NewFormatter(80.0)

	// Test with empty data
	report := &PeriodReport{
		Range: TimeRange{
			From: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		TotalLines: 0,
		AILines:    0,
		HumanLines: 0,
		Percentage: 0.0,
		DailyStats: []DailyStat{}, // Empty
	}

	output, err := formatter.Format(report, FormatCSV)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check that it outputs summary format with zeros
	expectedDataRow := "2025-01-01,2025-01-02,0,0,0,0.0,0.0,80.0,0.0"
	if lines[1] != expectedDataRow {
		t.Errorf("Expected data row with zeros: %s, got: %s", expectedDataRow, lines[1])
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
