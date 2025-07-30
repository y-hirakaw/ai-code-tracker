package tracker

import (
	"fmt"
	"strings"
)

type Analyzer struct {
	config *Config
}

func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config: config,
	}
}

func (a *Analyzer) AnalyzeCheckpoints(before, after *Checkpoint) (*AnalysisResult, error) {
	result := &AnalysisResult{
		LastUpdated: after.Timestamp,
	}

	isAIAuthor := a.IsAIAuthor(after.Author)

	for path, afterFile := range after.Files {
		beforeFile, exists := before.Files[path]
		if !exists {
			// New file
			if isAIAuthor {
				result.AILines += len(afterFile.Lines)
			} else {
				result.HumanLines += len(afterFile.Lines)
			}
			result.TotalLines += len(afterFile.Lines)
			continue
		}

		// Modified file - count added lines
		stats := a.compareFiles(beforeFile, afterFile, isAIAuthor)
		result.AILines += stats.AILines
		result.HumanLines += stats.HumanLines
		// Don't add total lines here - we only want the diff
	}

	for path, beforeFile := range before.Files {
		if _, exists := after.Files[path]; !exists {
			result.TotalLines -= len(beforeFile.Lines)
		}
	}

	if result.TotalLines > 0 {
		result.Percentage = float64(result.AILines) / float64(result.TotalLines) * 100
	}

	return result, nil
}

func (a *Analyzer) AnalyzeFromGitDiff(diff string, currentMetrics *AnalysisResult) (*AnalysisResult, error) {
	lines := strings.Split(diff, "\n")
	newAILines := 0
	newHumanLines := 0

	isAICommit := false
	for _, line := range lines {
		if strings.HasPrefix(line, "Author:") {
			author := strings.TrimSpace(strings.TrimPrefix(line, "Author:"))
			isAICommit = a.IsAIAuthor(author)
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			if isAICommit {
				newAILines++
			} else {
				newHumanLines++
			}
		}
	}

	result := &AnalysisResult{
		TotalLines:  currentMetrics.TotalLines + newAILines + newHumanLines,
		AILines:     currentMetrics.AILines + newAILines,
		HumanLines:  currentMetrics.HumanLines + newHumanLines,
		LastUpdated: currentMetrics.LastUpdated,
	}

	if result.TotalLines > 0 {
		result.Percentage = float64(result.AILines) / float64(result.TotalLines) * 100
	}

	return result, nil
}

func (a *Analyzer) GetFileStats(checkpoint *Checkpoint) []FileStats {
	stats := make([]FileStats, 0, len(checkpoint.Files))
	
	for path, file := range checkpoint.Files {
		stat := FileStats{
			Path:       path,
			TotalLines: len(file.Lines),
		}
		
		if a.IsAIAuthor(checkpoint.Author) {
			stat.AILines = len(file.Lines)
		} else {
			stat.HumanLines = len(file.Lines)
		}
		
		stats = append(stats, stat)
	}
	
	return stats
}

func (a *Analyzer) compareFiles(before, after FileContent, isAIAuthor bool) FileStats {
	stats := FileStats{
		Path: after.Path,
	}

	// For now, just count the difference in total lines
	// This is a simplified approach that assumes lines are only added, not modified
	lineDiff := len(after.Lines) - len(before.Lines)
	
	if lineDiff > 0 {
		if isAIAuthor {
			stats.AILines = lineDiff
		} else {
			stats.HumanLines = lineDiff
		}
	}

	return stats
}

func (a *Analyzer) IsAIAuthor(author string) bool {
	aiAuthors := []string{"claude", "ai", "assistant", "bot"}
	authorLower := strings.ToLower(author)
	
	if mapping, exists := a.config.AuthorMappings[author]; exists {
		authorLower = strings.ToLower(mapping)
	}
	
	for _, aiAuthor := range aiAuthors {
		if strings.Contains(authorLower, aiAuthor) {
			return true
		}
	}
	
	return false
}

func (a *Analyzer) GenerateReport(result *AnalysisResult) string {
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
Total Lines: %d (including %d baseline)
Added Lines: %d
  AI Lines: %d (%.1f%%)
  Human Lines: %d (%.1f%%)

Target: %.1f%% AI code
Progress: %.1f%%

Last Updated: %s
`,
		result.TotalLines, result.BaselineLines,
		addedLines,
		result.AILines, result.Percentage,
		result.HumanLines, humanPercentage,
		a.config.TargetAIPercentage,
		progress,
		result.LastUpdated.Format("2006-01-02 15:04:05"))

	return report
}