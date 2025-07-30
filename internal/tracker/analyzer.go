package tracker

import (
	"fmt"
	"os/exec"
	"strconv"
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

	// Try to use numstat data from checkpoints for accurate counting
	if before.NumstatData != nil && after.NumstatData != nil {
		for filepath, afterStats := range after.NumstatData {
			// Check if file should be tracked
			if !a.shouldTrackFile(filepath) {
				continue
			}
			
			beforeStats, existed := before.NumstatData[filepath]
			if !existed {
				beforeStats = [2]int{0, 0} // File didn't exist in previous checkpoint
			}
			
			// Calculate the difference in added lines between checkpoints
			addedLinesDiff := afterStats[0] - beforeStats[0]
			
			if addedLinesDiff > 0 {
				if isAIAuthor {
					result.AILines += addedLinesDiff
				} else {
					result.HumanLines += addedLinesDiff
				}
			}
		}
		
		// Check for new files in after that weren't in before
		for filepath, afterStats := range after.NumstatData {
			if !a.shouldTrackFile(filepath) {
				continue
			}
			
			if _, existed := before.NumstatData[filepath]; !existed {
				// New file
				if isAIAuthor {
					result.AILines += afterStats[0]
				} else {
					result.HumanLines += afterStats[0]
				}
			}
		}
		
		// Calculate total lines from current checkpoint
		for _, file := range after.Files {
			result.TotalLines += len(file.Lines)
		}
		
		if result.AILines+result.HumanLines > 0 {
			result.Percentage = float64(result.AILines) / float64(result.AILines + result.HumanLines) * 100
		}
		
		return result, nil
	}

	// Fall back to commit-based or file-based comparison
	if before.CommitHash != "" && after.CommitHash != "" {
		numstatData, err := a.getGitNumstat(before.CommitHash, after.CommitHash)
		if err == nil {
			// Use git numstat data for accurate counting
			for filepath, stats := range numstatData {
				// Check if file should be tracked
				if !a.shouldTrackFile(filepath) {
					continue
				}
				
				addedLines := stats[0]
				if isAIAuthor {
					result.AILines += addedLines
				} else {
					result.HumanLines += addedLines
				}
			}
			
			// Calculate total lines from current checkpoint
			for _, file := range after.Files {
				result.TotalLines += len(file.Lines)
			}
			
			if result.AILines+result.HumanLines > 0 {
				result.Percentage = float64(result.AILines) / float64(result.AILines + result.HumanLines) * 100
			}
			
			return result, nil
		}
		// Fall back to line-by-line comparison if git numstat fails
	}

	// Original implementation (fallback)
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

	// Simple fallback when git numstat is not available
	// Only count net additions
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

// getGitNumstat runs git diff --numstat between two commits
func (a *Analyzer) getGitNumstat(fromCommit, toCommit string) (map[string][2]int, error) {
	// Result: map[filepath] -> [added_lines, deleted_lines]
	result := make(map[string][2]int)
	
	cmd := exec.Command("git", "diff", "--numstat", fromCommit, toCommit)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		// Format: "added\tdeleted\tfilepath"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		
		added, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // Skip binary files which show "-"
		}
		
		deleted, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		
		// Handle renames: "path1 => path2" becomes just "path2"
		filepath := strings.Join(parts[2:], " ")
		if idx := strings.Index(filepath, " => "); idx != -1 {
			filepath = filepath[idx+4:]
		}
		
		result[filepath] = [2]int{added, deleted}
	}
	
	return result, nil
}

// shouldTrackFile checks if a file should be tracked based on config
func (a *Analyzer) shouldTrackFile(filepath string) bool {
	// Check extension
	hasValidExt := false
	for _, ext := range a.config.TrackedExtensions {
		if strings.HasSuffix(filepath, ext) {
			hasValidExt = true
			break
		}
	}
	
	if !hasValidExt {
		return false
	}
	
	// Check exclusion patterns
	for _, pattern := range a.config.ExcludePatterns {
		// Simple pattern matching (can be improved with glob)
		if strings.Contains(filepath, pattern) {
			return false
		}
	}
	
	return true
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

	return report
}