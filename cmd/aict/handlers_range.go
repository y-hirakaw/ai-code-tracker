package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
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
	executor := gitexec.NewExecutor()

	totalAI := 0
	totalHuman := 0
	byAuthor := make(map[string]*tracker.AuthorStats)
	byFile := make(map[string]*FileStatsRange)

	// 詳細メトリクス用
	var detailedMetrics tracker.DetailedMetrics

	// 作成者ごとのコミット参加記録（重複カウント防止）
	authorCommits := make(map[string]map[string]bool) // author -> commitHash -> bool

	for _, commitHash := range commits {
		log, err := nm.GetAuthorshipLog(commitHash)
		if err != nil {
			// Authorship Logがないコミットはスキップ
			continue
		}
		if log == nil {
			continue
		}

		// git show --numstat でコミットの追加/削除行数を取得
		numstatOutput, err := executor.Run("show", "--numstat", "--format=", commitHash)
		if err != nil {
			// numstatが取得できない場合はスキップ
			continue
		}

		// numstatデータをパース (filepath -> [added, deleted])
		numstatMap := parseNumstatOutput(numstatOutput)

		// このコミットに参加した作成者を追跡
		authorsInCommit := make(map[string]bool)

		// 3. 集計（numstatベースのみ - 差分追跡方式）
		for filepath, fileInfo := range log.Files {
			// numstatデータから追加/削除を取得
			numstat, found := numstatMap[filepath]
			if !found {
				continue // numstatがないファイルはスキップ
			}

			totalAdded := numstat[0]
			totalDeleted := numstat[1]

			// Authorship Logから各作成者の行数を計算
			// 複数作成者がいる場合は、行範囲から実際の行数を計算して按分
			authorLineCount := make(map[string]int)
			totalAuthorLines := 0

			for _, author := range fileInfo.Authors {
				lines := authorship.CountLines(author.Lines)
				authorLineCount[author.Name] = lines
				totalAuthorLines += lines
			}

			// 作成者ごとに集計
			for _, author := range fileInfo.Authors {
				stats, exists := byAuthor[author.Name]
				if !exists {
					stats = &tracker.AuthorStats{
						Name: author.Name,
						Type: author.Type,
					}
					byAuthor[author.Name] = stats
				}

				// この作成者の行数を取得
				authorLines := authorLineCount[author.Name]

				// numstatの追加行数・削除行数を作成者の比率で按分
				var added, deleted int
				if totalAuthorLines > 0 {
					// 通常のケース: 追加行がある場合、比率で按分
					ratio := float64(authorLines) / float64(totalAuthorLines)
					added = int(float64(totalAdded) * ratio)
					deleted = int(float64(totalDeleted) * ratio)
				} else if len(fileInfo.Authors) == 1 {
					// 削除のみのファイル: 作成者が1人の場合、全削除行をその作成者に割り当て
					added = 0
					deleted = int(totalDeleted)
				}

				stats.Lines += added

				// このコミットに参加したことを記録
				authorsInCommit[author.Name] = true

				// 詳細メトリクス
				if author.Type == tracker.AuthorTypeAI {
					detailedMetrics.WorkVolume.AIAdded += added
					detailedMetrics.WorkVolume.AIDeleted += deleted
					detailedMetrics.WorkVolume.AIChanges += added + deleted
					detailedMetrics.Contributions.AIAdditions += added
					totalAI += added
				} else {
					detailedMetrics.WorkVolume.HumanAdded += added
					detailedMetrics.WorkVolume.HumanDeleted += deleted
					detailedMetrics.WorkVolume.HumanChanges += added + deleted
					detailedMetrics.Contributions.HumanAdditions += added
					totalHuman += added
				}
			}
		}

		// このコミットに参加した作成者のコミット数を更新
		for authorName := range authorsInCommit {
			if authorCommits[authorName] == nil {
				authorCommits[authorName] = make(map[string]bool)
			}
			authorCommits[authorName][commitHash] = true
		}
	}

	// コミット数を集計（重複なし）
	for authorName, commits := range authorCommits {
		if stats, exists := byAuthor[authorName]; exists {
			stats.Commits = len(commits)
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
	formatRangeReport(report, opts.Format, &detailedMetrics)
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
	executor := gitexec.NewExecutor()
	output, err := executor.Run("log", "--since="+expandedSince, "--format=%H", "--reverse")
	if err != nil {
		return "", fmt.Errorf("failed to get commits since %s: %w", since, err)
	}

	commits := strings.Split(output, "\n")
	if len(commits) == 0 || commits[0] == "" {
		return "", fmt.Errorf("no commits found since %s", since)
	}

	// 最初のコミットの1つ前からHEADまでの範囲を作成
	firstCommit := commits[0]

	// 最初のコミットの親が存在するか確認
	executor2 := gitexec.NewExecutor()
	_, err = executor2.Run("rev-parse", firstCommit+"^")
	if err != nil {
		// 親がない（初回コミット、またはリポジトリ初期化直後）場合
		// 最初のコミット自体から開始: firstCommit..HEAD
		// ただし、firstCommitのみが対象の場合もあるので、firstCommit^..HEAD を使う
		// git では ^ が無効な場合でも --not を使える
		return firstCommit + "..HEAD", nil
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
	executor := gitexec.NewExecutor()
	output, err := executor.Run("log", "--format=%H", rangeSpec)
	if err != nil {
		return nil, err
	}

	var commits []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			commits = append(commits, line)
		}
	}

	return commits, nil
}

// parseNumstatOutput parses git show --numstat output
// Returns map[filepath][added, deleted]
func parseNumstatOutput(output string) map[string][2]int {
	result := make(map[string][2]int)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// numstat format: <added>\t<deleted>\t<filepath>
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}

		added := 0
		deleted := 0

		// "-" means binary file
		if parts[0] != "-" {
			fmt.Sscanf(parts[0], "%d", &added)
		}
		if parts[1] != "-" {
			fmt.Sscanf(parts[1], "%d", &deleted)
		}

		filepath := parts[2]
		result[filepath] = [2]int{added, deleted}
	}

	return result
}

// formatRangeReport formats and displays the range report
func formatRangeReport(report *tracker.Report, format string, metrics *tracker.DetailedMetrics) {
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
		fmt.Printf("📊 AI Code Generation Report (%s)\n", report.Range)
		fmt.Println()
		fmt.Printf("Commits: %d\n", report.Commits)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		// 詳細メトリクスを常時表示
		if metrics != nil {
			printDetailedMetrics(metrics)
		}

		// By Author（追加行数ベース）
		if len(report.ByAuthor) > 0 {
			fmt.Println("By Author:")
			for _, author := range report.ByAuthor {
				icon := "👤"
				if author.Type == tracker.AuthorTypeAI {
					icon = "🤖"
				}
				fmt.Printf("  %s %-20s %6d行追加 (%.1f%%) - %d commits\n",
					icon, author.Name, author.Lines, author.Percentage, author.Commits)
			}
			fmt.Println()
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
		os.Exit(1)
	}
}

// printDetailedMetrics prints detailed metrics
func printDetailedMetrics(metrics *tracker.DetailedMetrics) {
	if metrics == nil {
		return
	}

	// コードベース貢献（純粋な追加）
	totalContributions := metrics.Contributions.AIAdditions + metrics.Contributions.HumanAdditions
	aiContribPct := 0.0
	humanContribPct := 0.0
	if totalContributions > 0 {
		aiContribPct = float64(metrics.Contributions.AIAdditions) / float64(totalContributions) * 100
		humanContribPct = float64(metrics.Contributions.HumanAdditions) / float64(totalContributions) * 100
	}

	fmt.Println("【コードベース貢献】（最終的なコード量への寄与）")
	fmt.Printf("  総変更行数: %d行\n", totalContributions)
	fmt.Printf("    🤖 AI追加:   %6d行 (%.1f%%)\n", metrics.Contributions.AIAdditions, aiContribPct)
	fmt.Printf("    👤 人間追加: %6d行 (%.1f%%)\n", metrics.Contributions.HumanAdditions, humanContribPct)
	fmt.Println()

	// 作業量貢献（追加+削除）
	totalWork := metrics.WorkVolume.AIChanges + metrics.WorkVolume.HumanChanges
	aiWorkPct := 0.0
	humanWorkPct := 0.0
	if totalWork > 0 {
		aiWorkPct = float64(metrics.WorkVolume.AIChanges) / float64(totalWork) * 100
		humanWorkPct = float64(metrics.WorkVolume.HumanChanges) / float64(totalWork) * 100
	}

	fmt.Println("【作業量貢献】（実際の作業量）")
	fmt.Printf("  総作業量: %d行\n", totalWork)
	fmt.Printf("    🤖 AI作業:   %6d行 (%.1f%%)\n", metrics.WorkVolume.AIChanges, aiWorkPct)
	fmt.Printf("       └ 追加: %d行, 削除: %d行\n", metrics.WorkVolume.AIAdded, metrics.WorkVolume.AIDeleted)
	fmt.Printf("    👤 人間作業: %6d行 (%.1f%%)\n", metrics.WorkVolume.HumanChanges, humanWorkPct)
	fmt.Printf("       └ 追加: %d行, 削除: %d行\n", metrics.WorkVolume.HumanAdded, metrics.WorkVolume.HumanDeleted)
	fmt.Println()

	// 新規ファイル（オプション）
	totalNewFiles := metrics.NewFiles.AINewLines + metrics.NewFiles.HumanNewLines
	if totalNewFiles > 0 {
		aiNewPct := float64(metrics.NewFiles.AINewLines) / float64(totalNewFiles) * 100
		humanNewPct := float64(metrics.NewFiles.HumanNewLines) / float64(totalNewFiles) * 100

		fmt.Println("【新規ファイル】（完全新規のコードのみ）")
		fmt.Printf("  新規コード: %d行\n", totalNewFiles)
		fmt.Printf("    🤖 AI新規:   %6d行 (%.1f%%)\n", metrics.NewFiles.AINewLines, aiNewPct)
		fmt.Printf("    👤 人間新規: %6d行 (%.1f%%)\n", metrics.NewFiles.HumanNewLines, humanNewPct)
		fmt.Println()
	}
}
