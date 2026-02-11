package tracker

import (
	"fmt"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/git"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

type Analyzer struct {
	config   *Config
	executor gitexec.Executor
}

func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config:   config,
		executor: gitexec.NewExecutor(),
	}
}

// NewAnalyzerWithExecutor creates an Analyzer with a custom executor (for testing)
func NewAnalyzerWithExecutor(config *Config, executor gitexec.Executor) *Analyzer {
	return &Analyzer{
		config:   config,
		executor: executor,
	}
}

func (a *Analyzer) AnalyzeCheckpoints(before, after *Checkpoint) (*AnalysisResult, error) {
	isAIAuthor := a.IsAIAuthor(after.Author)

	// Try to use numstat data from checkpoints for accurate counting
	if before.NumstatData != nil && after.NumstatData != nil {
		return a.analyzeFromNumstat(before, after, isAIAuthor)
	}

	// Fall back to commit-based comparison
	if before.CommitHash != "" && after.CommitHash != "" {
		result, err := a.analyzeFromCommits(before, after, isAIAuthor)
		if err == nil {
			return result, nil
		}
		// Fall back to line-by-line comparison if git numstat fails
	}

	// Final fallback: line-by-line file comparison
	return a.analyzeFromFiles(before, after, isAIAuthor), nil
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
	return git.GetNumstatBetweenCommits(a.executor, fromCommit, toCommit)
}

// shouldTrackFile checks if a file should be tracked based on config.
// Delegates to the shared IsTrackedFile function.
func (a *Analyzer) shouldTrackFile(filepath string) bool {
	return IsTrackedFile(filepath, a.config)
}

// calculatePercentage calculates AI percentage from AI and human lines
func calculatePercentage(aiLines, humanLines int) float64 {
	total := aiLines + humanLines
	if total == 0 {
		return 0.0
	}
	return float64(aiLines) / float64(total) * 100
}

// aggregateLinesByAuthor adds lines to appropriate counter based on author type
func (a *Analyzer) aggregateLinesByAuthor(lines int, isAI bool, aiLines, humanLines *int) {
	if isAI {
		*aiLines += lines
	} else {
		*humanLines += lines
	}
}

// analyzeFromNumstat analyzes checkpoints using numstat data from checkpoints
func (a *Analyzer) analyzeFromNumstat(before, after *Checkpoint, isAI bool) (*AnalysisResult, error) {
	result := &AnalysisResult{
		LastUpdated: after.Timestamp,
	}

	for filePath, afterStats := range after.NumstatData {
		if !a.shouldTrackFile(filePath) {
			continue
		}

		beforeStats, existed := before.NumstatData[filePath]
		if !existed {
			beforeStats = [2]int{0, 0}
		}

		addedLinesDiff := afterStats[0] - beforeStats[0]
		if addedLinesDiff > 0 {
			a.aggregateLinesByAuthor(addedLinesDiff, isAI, &result.AILines, &result.HumanLines)
		}

		// 詳細メトリクス: 作業量貢献（追加+削除）
		added := afterStats[0]
		deleted := afterStats[1]
		if isAI {
			result.Metrics.WorkVolume.AIAdded += added
			result.Metrics.WorkVolume.AIDeleted += deleted
			result.Metrics.WorkVolume.AIChanges += added + deleted
		} else {
			result.Metrics.WorkVolume.HumanAdded += added
			result.Metrics.WorkVolume.HumanDeleted += deleted
			result.Metrics.WorkVolume.HumanChanges += added + deleted
		}

		// 詳細メトリクス: コードベース貢献（純粋な追加）
		if addedLinesDiff > 0 {
			if isAI {
				result.Metrics.Contributions.AIAdditions += addedLinesDiff
			} else {
				result.Metrics.Contributions.HumanAdditions += addedLinesDiff
			}
		}

		// 詳細メトリクス: 新規ファイル
		if !existed {
			if isAI {
				result.Metrics.NewFiles.AINewLines += afterStats[0]
			} else {
				result.Metrics.NewFiles.HumanNewLines += afterStats[0]
			}
		}
	}

	// Calculate total lines
	for _, file := range after.Files {
		result.TotalLines += len(file.Lines)
	}

	result.Percentage = calculatePercentage(result.AILines, result.HumanLines)
	return result, nil
}

// analyzeFromCommits analyzes checkpoints using git diff between commit hashes
func (a *Analyzer) analyzeFromCommits(before, after *Checkpoint, isAI bool) (*AnalysisResult, error) {
	result := &AnalysisResult{
		LastUpdated: after.Timestamp,
	}

	// Get git numstat between commits
	numstatData, err := a.getGitNumstat(before.CommitHash, after.CommitHash)
	if err != nil {
		return nil, err
	}

	// Process files from git diff
	for filepath, stats := range numstatData {
		if !a.shouldTrackFile(filepath) {
			continue
		}

		addedLines := stats[0]
		deletedLines := stats[1]
		a.aggregateLinesByAuthor(addedLines, isAI, &result.AILines, &result.HumanLines)

		// 詳細メトリクス: 作業量貢献（追加+削除）
		if isAI {
			result.Metrics.WorkVolume.AIAdded += addedLines
			result.Metrics.WorkVolume.AIDeleted += deletedLines
			result.Metrics.WorkVolume.AIChanges += addedLines + deletedLines
			result.Metrics.Contributions.AIAdditions += addedLines
		} else {
			result.Metrics.WorkVolume.HumanAdded += addedLines
			result.Metrics.WorkVolume.HumanDeleted += deletedLines
			result.Metrics.WorkVolume.HumanChanges += addedLines + deletedLines
			result.Metrics.Contributions.HumanAdditions += addedLines
		}
	}

	// Calculate total lines
	for _, file := range after.Files {
		result.TotalLines += len(file.Lines)
	}

	result.Percentage = calculatePercentage(result.AILines, result.HumanLines)
	return result, nil
}

func (a *Analyzer) IsAIAuthor(author string) bool {
	return IsAIAgent(author, a.config.AIAgents, a.config.AuthorMappings)
}

// analyzeFromFiles analyzes checkpoints using line-by-line file comparison (fallback)
func (a *Analyzer) analyzeFromFiles(before, after *Checkpoint, isAI bool) *AnalysisResult {
	result := &AnalysisResult{
		LastUpdated: after.Timestamp,
	}

	// Process files: new and modified
	for path, afterFile := range after.Files {
		beforeFile, exists := before.Files[path]
		if !exists {
			// New file
			lineCount := len(afterFile.Lines)
			a.aggregateLinesByAuthor(lineCount, isAI, &result.AILines, &result.HumanLines)
			result.TotalLines += lineCount
			continue
		}

		// Modified file - count added lines
		stats := a.compareFiles(beforeFile, afterFile, isAI)
		result.AILines += stats.AILines
		result.HumanLines += stats.HumanLines
		// Don't add total lines here - we only want the diff
	}

	// Process deleted files
	for path, beforeFile := range before.Files {
		if _, exists := after.Files[path]; !exists {
			result.TotalLines -= len(beforeFile.Lines)
		}
	}

	result.Percentage = calculatePercentage(result.AILines, result.HumanLines)
	return result
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
