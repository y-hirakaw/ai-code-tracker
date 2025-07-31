package period

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Formatter handles different output formats for period reports
type Formatter struct {
	targetAIPercentage float64
}

// NewFormatter creates a new formatter
func NewFormatter(targetAIPercentage float64) *Formatter {
	return &Formatter{
		targetAIPercentage: targetAIPercentage,
	}
}

// Format formats the period report according to the specified format
func (f *Formatter) Format(report *PeriodReport, format ReportFormat) (string, error) {
	switch format {
	case FormatTable:
		return f.formatTable(report), nil
	case FormatGraph:
		return f.formatGraph(report), nil
	case FormatJSON:
		return f.formatJSON(report)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatTable formats as a text table
func (f *Formatter) formatTable(report *PeriodReport) string {
	humanPercentage := 0.0
	if report.TotalLines > 0 {
		humanPercentage = float64(report.HumanLines) / float64(report.TotalLines) * 100
	}
	
	progress := report.Percentage / f.targetAIPercentage * 100
	if progress > 100 {
		progress = 100
	}
	
	var builder strings.Builder
	
	builder.WriteString("AI Code Tracking Report (Period)\n")
	builder.WriteString("=================================\n")
	builder.WriteString(fmt.Sprintf("Period: %s to %s\n", 
		report.Range.From.Format("2006-01-02 15:04:05"),
		report.Range.To.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Total Lines: %d\n", report.TotalLines))
	builder.WriteString(fmt.Sprintf("  AI Lines: %d (%.1f%%)\n", report.AILines, report.Percentage))
	builder.WriteString(fmt.Sprintf("  Human Lines: %d (%.1f%%)\n", report.HumanLines, humanPercentage))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Target: %.1f%% AI code\n", f.targetAIPercentage))
	builder.WriteString(fmt.Sprintf("Progress: %.1f%%\n", progress))
	
	if len(report.DailyStats) > 0 {
		builder.WriteString("\nDaily Breakdown:\n")
		builder.WriteString("Date       | AI Lines | Human Lines | AI %\n")
		builder.WriteString("-----------+----------+-------------+------\n")
		
		for _, stat := range report.DailyStats {
			dailyTotal := stat.AILines + stat.HumanLines
			dailyPercentage := 0.0
			if dailyTotal > 0 {
				dailyPercentage = float64(stat.AILines) / float64(dailyTotal) * 100
			}
			
			builder.WriteString(fmt.Sprintf("%s | %8d | %11d | %4.1f\n",
				stat.Date.Format("2006-01-02"),
				stat.AILines,
				stat.HumanLines,
				dailyPercentage))
		}
	}
	
	return builder.String()
}

// formatGraph formats as ASCII graph
func (f *Formatter) formatGraph(report *PeriodReport) string {
	var builder strings.Builder
	
	builder.WriteString("AI vs Human Code Contributions (Period)\n")
	builder.WriteString("========================================\n")
	builder.WriteString(fmt.Sprintf("Period: %s to %s\n\n", 
		report.Range.From.Format("2006-01-02"),
		report.Range.To.Format("2006-01-02")))
	
	if len(report.DailyStats) == 0 {
		// Simple bar chart for totals
		builder.WriteString("Total Contributions:\n")
		
		maxLines := max(report.AILines, report.HumanLines)
		if maxLines == 0 {
			builder.WriteString("No contributions in this period.\n")
			return builder.String()
		}
		
		aiBarLength := (report.AILines * 50) / maxLines
		humanBarLength := (report.HumanLines * 50) / maxLines
		
		builder.WriteString(fmt.Sprintf("AI     [%s%s] %d lines (%.1f%%)\n",
			strings.Repeat("█", aiBarLength),
			strings.Repeat("░", 50-aiBarLength),
			report.AILines,
			report.Percentage))
		
		humanPercentage := 0.0
		if report.TotalLines > 0 {
			humanPercentage = float64(report.HumanLines) / float64(report.TotalLines) * 100
		}
		
		builder.WriteString(fmt.Sprintf("Human  [%s%s] %d lines (%.1f%%)\n",
			strings.Repeat("█", humanBarLength),
			strings.Repeat("░", 50-humanBarLength),
			report.HumanLines,
			humanPercentage))
		
		return builder.String()
	}
	
	// Daily trend graph
	builder.WriteString("Daily AI Percentage Trend:\n")
	
	maxPercentage := 100.0
	
	for _, stat := range report.DailyStats {
		dailyTotal := stat.AILines + stat.HumanLines
		dailyPercentage := 0.0
		if dailyTotal > 0 {
			dailyPercentage = float64(stat.AILines) / float64(dailyTotal) * 100
		}
		
		barLength := int((dailyPercentage / maxPercentage) * 50)
		
		builder.WriteString(fmt.Sprintf("%s [%s%s] %.1f%% (%d/%d)\n",
			stat.Date.Format("01-02"),
			strings.Repeat("█", barLength),
			strings.Repeat("░", 50-barLength),
			dailyPercentage,
			stat.AILines,
			dailyTotal))
	}
	
	// Add target line
	targetBarLength := int((f.targetAIPercentage / maxPercentage) * 50)
	builder.WriteString(fmt.Sprintf("\nTarget [%s%s] %.1f%%\n",
		strings.Repeat("─", targetBarLength),
		strings.Repeat(" ", 50-targetBarLength),
		f.targetAIPercentage))
	
	return builder.String()
}

// formatJSON formats as JSON
func (f *Formatter) formatJSON(report *PeriodReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}