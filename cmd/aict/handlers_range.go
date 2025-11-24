package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// ReportOptions holds options for the report command
type ReportOptions struct {
	Range  string
	Since  string
	Format string
}

// handleRangeReport is the entry point called from main
func handleRangeReport() {
	fs := flag.NewFlagSet("report", flag.ExitOnError)

	opts := &ReportOptions{}
	fs.StringVar(&opts.Range, "range", "", "Commit range (e.g., 'origin/main..HEAD')")
	fs.StringVar(&opts.Since, "since", "", "Show commits since date (e.g., '7 days ago', '2025-01-01')")
	fs.StringVar(&opts.Format, "format", "table", "Output format: table or json")

	fs.Parse(os.Args[2:])

	// --range と --since の排他チェック
	if opts.Range != "" && opts.Since != "" {
		fmt.Println("Error: --range and --since are mutually exclusive")
		fmt.Println("Please use either --range or --since, not both")
		os.Exit(1)
	}

	// どちらも指定されていない場合
	if opts.Range == "" && opts.Since == "" {
		fmt.Println("Error: either --range or --since is required")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  aict report --range <base>..<head>")
		fmt.Println("  aict report --since <date>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  aict report --range origin/main..HEAD")
		fmt.Println("  aict report --since 7d        # 7 days ago")
		fmt.Println("  aict report --since 2w        # 2 weeks ago")
		fmt.Println("  aict report --since 1m        # 1 month ago")
		fmt.Println("  aict report --since '7 days ago'")
		fmt.Println("  aict report --since '2025-01-01'")
		fmt.Println("  aict report --since yesterday")
		os.Exit(1)
	}

	// --since を --range に変換
	if opts.Since != "" {
		convertedRange, err := convertSinceToRange(opts.Since)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		opts.Range = convertedRange
	}

	handleRangeReportWithOptions(opts)
}

// handleRangeReportWithOptions handles report for commit range (SPEC.md準拠)
func handleRangeReportWithOptions(opts *ReportOptions) {
	// 1. git log <range> でコミット一覧を取得
	commits, err := getCommitsInRange(opts.Range)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(commits) == 0 {
		rangeDisplay := opts.Range
		if opts.Since != "" {
			rangeDisplay = "since " + opts.Since
		}
		fmt.Println("No commits found in range:", rangeDisplay)
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
	rangeDisplay := opts.Range
	if opts.Since != "" {
		rangeDisplay = "since " + opts.Since
	}

	report := &tracker.Report{
		Range:   rangeDisplay,
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

// convertSinceToRange converts --since date to --range format
func convertSinceToRange(since string) (string, error) {
	// 簡潔な表記を展開（3d → 3 days ago, 2w → 2 weeks ago, 1m → 1 month ago）
	expandedSince := expandShorthandDate(since)

	// git log --since でコミットハッシュリストを取得（古い順）
	cmd := exec.Command("git", "log", "--since="+expandedSince, "--format=%H", "--reverse")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git log failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to get commits since %s: %w", since, err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commits) == 0 || commits[0] == "" {
		return "", fmt.Errorf("no commits found since %s", since)
	}

	// 最初のコミットの1つ前からHEADまでの範囲を作成
	firstCommit := commits[0]

	// 最初のコミットの親が存在するか確認
	parentCmd := exec.Command("git", "rev-parse", firstCommit+"^")
	_, err = parentCmd.Output()
	if err != nil {
		// 親がない（最初のコミット）場合は、そのコミット自体から開始
		return firstCommit + "^.." + firstCommit, nil
	}

	return firstCommit + "^..HEAD", nil
}

// expandShorthandDate expands shorthand date notation to git-compatible format
// Examples: 3d → 3 days ago, 2w → 2 weeks ago, 1m → 1 month ago
func expandShorthandDate(since string) string {
	if len(since) < 2 {
		return since
	}

	// 末尾の単位文字を確認
	lastChar := since[len(since)-1]
	numPart := since[:len(since)-1]

	// 数値部分が有効か確認
	if !isNumeric(numPart) {
		return since
	}

	switch lastChar {
	case 'd':
		return numPart + " days ago"
	case 'w':
		return numPart + " weeks ago"
	case 'm':
		return numPart + " months ago"
	case 'y':
		return numPart + " years ago"
	default:
		return since
	}
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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
