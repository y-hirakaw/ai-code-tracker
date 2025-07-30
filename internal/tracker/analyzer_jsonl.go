package tracker

import (
	"fmt"
)

// AnalyzeRecords analyzes checkpoint records to generate metrics
func (a *Analyzer) AnalyzeRecords(records []CheckpointRecord) (*AnalysisResult, error) {
	if len(records) == 0 {
		return &AnalysisResult{}, nil
	}

	result := &AnalysisResult{
		LastUpdated: records[len(records)-1].Timestamp,
	}

	// Process records in chronological order
	for i := 1; i < len(records); i++ {
		beforeRecord := records[i-1]
		afterRecord := records[i]

		// Calculate differences between consecutive records
		diff := a.calculateRecordDiff(beforeRecord, afterRecord)
		
		// Accumulate changes based on author
		isAIAuthor := a.IsAIAuthor(afterRecord.Author)
		if isAIAuthor {
			result.AILines += diff
		} else {
			result.HumanLines += diff
		}
	}

	// Calculate percentage
	totalAddedLines := result.AILines + result.HumanLines
	if totalAddedLines > 0 {
		result.Percentage = float64(result.AILines) / float64(totalAddedLines) * 100
	}

	return result, nil
}

// calculateRecordDiff calculates the net added lines between two records
func (a *Analyzer) calculateRecordDiff(before, after CheckpointRecord) int {
	// Simple difference between total added lines
	addedDiff := after.Added - before.Added
	if addedDiff > 0 {
		return addedDiff
	}
	return 0
}

// GenerateReportFromRecords generates a report from checkpoint records
func (a *Analyzer) GenerateReportFromRecords(records []CheckpointRecord, baselineLines int) (string, error) {
	result, err := a.AnalyzeRecords(records)
	if err != nil {
		return "", err
	}

	// Calculate progress toward target
	progress := result.Percentage / a.config.TargetAIPercentage * 100
	if progress > 100 {
		progress = 100
	}

	addedLines := result.AILines + result.HumanLines
	humanPercentage := 0.0
	if addedLines > 0 {
		humanPercentage = float64(result.HumanLines) / float64(addedLines) * 100
	}

	report := fmt.Sprintf(`AI Code Tracking Report
======================
Added Lines: %d
  AI Lines: %d (%.1f%%)
  Human Lines: %d (%.1f%%)

Target: %.1f%% AI code
Progress: %.1f%%

Last Updated: %s
`,
		addedLines,
		result.AILines, result.Percentage,
		result.HumanLines, humanPercentage,
		a.config.TargetAIPercentage,
		progress,
		result.LastUpdated.Format("2006-01-02 15:04:05"))

	return report, nil
}

// GetFileStatsFromRecords extracts file statistics from records
// Note: With simplified format, per-file statistics are not available
func (a *Analyzer) GetFileStatsFromRecords(records []CheckpointRecord) []FileStats {
	// Return empty slice since we no longer track per-file data
	return []FileStats{}
}