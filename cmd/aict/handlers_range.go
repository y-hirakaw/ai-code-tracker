package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// handleRangeReport handles report for commit range (SPEC.md準拠)
func handleRangeReport(opts *ReportOptions) {
	// 1. git log <range> でコミット一覧を取得
	commits, err := getCommitsInRange(opts.Range)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		fmt.Println("No commits found in range:", opts.Range)
		return
	}

	// 2. 各コミットのAuthorship Logを読み込み
	nm := gitnotes.NewNotesManager()

	totalAI := 0
	totalHuman := 0
	byAuthor := make(map[string]*tracker.AuthorStats)
	byFile := make(map[string]*FileStatsRange)

	for _, commitHash := range commits {
		log, err := nm.GetAuthorshipLog(commitHash)
		if err != nil {
			// Authorship Logがないコミットはスキップ
			continue
		}
		if log == nil {
			continue
		}

		// 3. 集計
		for filepath, fileInfo := range log.Files {
			for _, author := range fileInfo.Authors {
				lineCount := authorship.CountLines(author.Lines)

				// 作成者別集計
				stats, exists := byAuthor[author.Name]
				if !exists {
					stats = &tracker.AuthorStats{
						Name: author.Name,
						Type: author.Type,
					}
					byAuthor[author.Name] = stats
				}
				stats.Lines += lineCount
				stats.Commits++

				// ファイル別集計
				fileStats, exists := byFile[filepath]
				if !exists {
					fileStats = &FileStatsRange{Path: filepath}
					byFile[filepath] = fileStats
				}
				fileStats.TotalLines += lineCount

				if author.Type == tracker.AuthorTypeAI {
					totalAI += lineCount
					fileStats.AILines += lineCount
				} else {
					totalHuman += lineCount
					fileStats.HumanLines += lineCount
				}
			}
		}
	}

	// 4. レポート生成
	report := &tracker.Report{
		Range:   opts.Range,
		Commits: len(commits),
		Summary: tracker.SummaryStats{
			TotalLines:   totalAI + totalHuman,
			AILines:      totalAI,
			HumanLines:   totalHuman,
			AIPercentage: 0,
		},
	}

	if report.Summary.TotalLines > 0 {
		report.Summary.AIPercentage = float64(totalAI) / float64(totalAI+totalHuman) * 100
	}

	// ByAuthor を追加
	for _, stats := range byAuthor {
		stats.Percentage = float64(stats.Lines) / float64(report.Summary.TotalLines) * 100
		report.ByAuthor = append(report.ByAuthor, *stats)
	}

	// ByFile を追加
	for _, stats := range byFile {
		if stats.TotalLines > 0 {
			stats.AIPercentage = float64(stats.AILines) / float64(stats.TotalLines) * 100
		}
		// FileStatsにはAIPercentageフィールドがないので、別途計算
		report.ByFile = append(report.ByFile, tracker.FileStats{
			Path:       stats.Path,
			TotalLines: stats.TotalLines,
			AILines:    stats.AILines,
			HumanLines: stats.HumanLines,
		})
	}

	// 5. フォーマットに応じて出力
	formatRangeReport(report, opts.Format)
}

// FileStatsRange is a temporary struct for file statistics during range report
type FileStatsRange struct {
	Path         string
	TotalLines   int
	AILines      int
	HumanLines   int
	AIPercentage float64
}

// getCommitsInRange retrieves commit hashes in the given range
func getCommitsInRange(rangeSpec string) ([]string, error) {
	cmd := exec.Command("git", "log", "--format=%H", rangeSpec)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var commits []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			commits = append(commits, line)
		}
	}

	return commits, nil
}

// formatRangeReport formats and displays the range report
func formatRangeReport(report *tracker.Report, format string) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))

	case "table", "graph":
		// Table format
		fmt.Println("📊 AI Code Generation Report")
		fmt.Println()
		fmt.Printf("Range: %s (%d commits)\n", report.Range, report.Commits)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf("  Total Lines:        %d\n", report.Summary.TotalLines)
		fmt.Printf("  🤖 AI Generated:    %d (%.1f%%)\n", report.Summary.AILines, report.Summary.AIPercentage)
		fmt.Printf("  👤 Human Written:   %d (%.1f%%)\n", report.Summary.HumanLines, 100-report.Summary.AIPercentage)
		fmt.Println()

		if len(report.ByAuthor) > 0 {
			fmt.Println("By Author:")
			for _, author := range report.ByAuthor {
				icon := "👤"
				if author.Type == tracker.AuthorTypeAI {
					icon = "🤖"
				}
				fmt.Printf("  %s %-20s %6d lines (%.1f%%) - %d commits\n",
					icon, author.Name, author.Lines, author.Percentage, author.Commits)
			}
			fmt.Println()
		}

		if len(report.ByFile) > 0 && len(report.ByFile) <= 10 {
			fmt.Println("Top Files:")
			for i, file := range report.ByFile {
				if i >= 10 {
					break
				}
				aiPct := 0.0
				if file.TotalLines > 0 {
					aiPct = float64(file.AILines) / float64(file.TotalLines) * 100
				}
				fmt.Printf("  %-40s %5d lines (%.0f%% AI)\n",
					file.Path, file.TotalLines, aiPct)
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		os.Exit(1)
	}
}
