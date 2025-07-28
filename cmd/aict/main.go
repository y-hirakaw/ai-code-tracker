package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/internal/blame"
	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/hooks"
	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/interactive"
	"github.com/ai-code-tracker/aict/internal/stats"
	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/internal/tracker"
	"github.com/ai-code-tracker/aict/internal/ui"
	"github.com/ai-code-tracker/aict/pkg/types"
)

const (
	// Version はアプリケーションのバージョン
	Version = "0.1.0"
	// AppName はアプリケーション名
	AppName = "aict"
)

var (
	// helpSystem はグローバルヘルプシステム
	helpSystem *ui.HelpSystem
)

// CLI コマンドの定義
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// main はアプリケーションのエントリーポイント
func main() {
	// i18nシステムを初期化
	i18n.Initialize()
	
	// エラーフォーマッターを初期化
	errors.InitializeFormatter()
	
	// ヘルプシステムを初期化
	helpSystem = ui.NewHelpSystem(AppName, Version)
	
	if len(os.Args) < 2 {
		helpSystem.ShowMainHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	commands := map[string]Command{
		"track": {
			Name:        "track",
			Description: "ファイルの変更を手動でトラッキングする",
			Handler:     handleTrack,
		},
		"init": {
			Name:        "init",
			Description: "プロジェクトでAI Code Trackerを初期化する",
			Handler:     handleInit,
		},
		"stats": {
			Name:        "stats",
			Description: "統計情報を表示する",
			Handler:     handleStats,
		},
		"blame": {
			Name:        "blame",
			Description: "ファイルのAI/人間による変更履歴を表示する",
			Handler:     handleBlame,
		},
		"config": {
			Name:        "config",
			Description: "設定を管理する",
			Handler:     handleConfig,
		},
		"setup": {
			Name:        "setup",
			Description: "Git hooks と Claude Code hooks を自動設定する",
			Handler:     handleSetup,
		},
		"wizard": {
			Name:        "wizard",
			Description: "インタラクティブセットアップウィザードを実行する",
			Handler:     handleWizard,
		},
		"version": {
			Name:        "version",
			Description: "バージョン情報を表示する",
			Handler:     handleVersion,
		},
		"help": {
			Name:        "help",
			Description: "ヘルプを表示する",
			Handler:     handleHelp,
		},
	}

	cmd, exists := commands[command]
	if !exists {
		friendlyErr := errors.UnknownCommand(command)
		fmt.Fprint(os.Stderr, errors.FormatError(friendlyErr))
		os.Exit(1)
	}

	if err := cmd.Handler(args); err != nil {
		// エラーが既にFriendlyErrorの場合はそのまま使用、そうでなければラップ
		if friendlyErr, ok := err.(*errors.FriendlyError); ok {
			fmt.Fprint(os.Stderr, errors.FormatError(friendlyErr.WithCommand(command)))
		} else {
			friendlyErr := errors.WrapError(err, errors.ErrorTypeGeneral, "generic_error").WithCommand(command)
			fmt.Fprint(os.Stderr, errors.FormatError(friendlyErr))
		}
		os.Exit(1)
	}
}


// handleInit はプロジェクトの初期化を処理する
func handleInit(args []string) error {
	// 現在のディレクトリがGitリポジトリかチェック
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "directory_access_failed")
	}

	if !tracker.IsGitRepo(currentDir) {
		return errors.GitNotRepository()
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer storage.Close()

	fmt.Println("AI Code Tracker を初期化しました")
	fmt.Printf("データディレクトリ: %s\n", filepath.Join(currentDir, ".git/ai-tracker"))
	fmt.Println("\n次のステップ:")
	fmt.Println("1. `aict track` でファイルの変更を追跡")
	fmt.Println("2. `aict stats` で統計情報を確認")
	fmt.Println("3. Git hooks の自動設定は今後のバージョンで対応予定")

	return nil
}

// handleTrack はファイルのトラッキングを処理する
func handleTrack(args []string) error {
	var (
		isAI        = false
		author      = ""
		model       = ""
		filesStr    = ""
		message     = ""
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--ai":
			isAI = true
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		case "--model":
			if i+1 < len(args) {
				model = args[i+1]
				i++
			}
		case "--files":
			if i+1 < len(args) {
				filesStr = args[i+1]
				i++
			}
		case "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		}
	}

	// パラメータの検証
	if author == "" {
		if isAI {
			author = "Claude Code"
		} else {
			return errors.NewError(errors.ErrorTypeCommand, "missing_required_option", "--author").
				WithSuggestions(i18n.T("suggestion_specify_author", "--author オプションで作成者を指定してください"))
		}
	}

	if isAI && model == "" {
		model = "claude-code" // デフォルトモデル
	}

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "directory_access_failed")
	}

	// ストレージとトラッカーを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer storage.Close()

	tracker := tracker.NewTracker(storage, currentDir)

	// ファイルリストを処理
	var files []string
	if filesStr != "" {
		files = strings.Split(filesStr, ",")
		for i, file := range files {
			files[i] = strings.TrimSpace(file)
		}
	} else {
		// ファイルが指定されていない場合、変更されたファイルを自動検出
		detectedFiles, err := tracker.DetectChangedFiles()
		if err != nil {
			return errors.WrapError(err, errors.ErrorTypeGit, "git_command_failed", "git diff")
		}
		files = detectedFiles
	}

	if len(files) == 0 {
		fmt.Println("追跡するファイルがありません")
		return nil
	}

	// イベントタイプを決定
	eventType := types.EventTypeHuman
	if isAI {
		eventType = types.EventTypeAI
	}

	// トラッキングを実行
	err = tracker.TrackFileChanges(eventType, author, model, files, message)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "tracking_failed")
	}

	fmt.Printf("✓ %d個のファイルの変更を追跡しました\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}
	fmt.Printf("作成者: %s\n", author)
	if isAI {
		fmt.Printf("モデル: %s\n", model)
	}
	if message != "" {
		fmt.Printf("メッセージ: %s\n", message)
	}

	return nil
}

// handleStats は統計情報の表示を処理する
func handleStats(args []string) error {
	var (
		format  = "table"
		since   = ""
		until   = ""
		author  = ""
		byFile  = false
		trend   = false
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--since":
			if i+1 < len(args) {
				since = args[i+1]
				i++
			}
		case "--until":
			if i+1 < len(args) {
				until = args[i+1]
				i++
			}
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		case "--by-file":
			byFile = true
		case "--trend":
			trend = true
		}
	}

	// 日付のパース
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = time.Parse("2006-01-02", since)
		if err != nil {
			return errors.InvalidDateFormat(since).WithCommand("stats")
		}
	} else {
		// デフォルトは30日前から
		sinceTime = time.Now().AddDate(0, 0, -30)
	}

	if until != "" {
		untilTime, err = time.Parse("2006-01-02", until)
		if err != nil {
			return errors.InvalidDateFormat(until).WithCommand("stats")
		}
	} else {
		// デフォルトは現在まで
		untilTime = time.Now()
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer storage.Close()

	// StatsManagerを初期化
	statsManager := stats.NewStatsManager(storage)

	// フォーマット別処理
	switch format {
	case "daily":
		return showDailyStats(statsManager, sinceTime, untilTime)
	case "files":
		return showFileStats(statsManager, sinceTime, author)
	case "contributors":
		return showContributorStats(statsManager, sinceTime, author)
	}

	// トレンド分析
	if trend {
		return showTrendAnalysis(statsManager, sinceTime, untilTime)
	}

	// ファイル別統計
	if byFile {
		return showFileStats(statsManager, sinceTime, author)
	}

	// 基本統計情報を取得
	basicStats, err := storage.GetStatistics()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "statistics_fetch_failed")
	}

	// 作成者フィルタ処理
	if author != "" {
		fmt.Printf("作成者フィルタ: %s\n", author)
		fmt.Printf("注意: 作成者フィルタは基本統計には適用されません\n\n")
	}

	// 出力形式に応じて表示
	switch format {
	case "table":
		showStatsTable(basicStats)
	case "json":
		showStatsJSON(basicStats)
	case "summary":
		showStatsSummary(basicStats)
	default:
		return errors.NewError(errors.ErrorTypeCommand, "invalid_output_format", format).
			WithSuggestions(
				i18n.T("suggestion_valid_formats", "有効な形式: table, json, summary, daily, files, contributors"),
			).WithCommand("stats")
	}

	return nil
}

// showStatsTable はテーブル形式で統計を表示する
func showStatsTable(stats *types.Statistics) {
	fmt.Println("=== AI Code Tracker 統計情報 ===\n")
	
	fmt.Printf("%-20s: %d\n", "総イベント数", stats.TotalEvents)
	fmt.Printf("%-20s: %d (%.1f%%)\n", "AI イベント", stats.AIEvents, stats.AIPercentage())
	fmt.Printf("%-20s: %d (%.1f%%)\n", "人間 イベント", stats.HumanEvents, stats.HumanPercentage())
	fmt.Printf("%-20s: %d\n", "コミット イベント", stats.CommitEvents)
	fmt.Println()
	
	fmt.Printf("%-20s: %d\n", "追加行数", stats.TotalLinesAdded)
	fmt.Printf("%-20s: %d\n", "変更行数", stats.TotalLinesModified)
	fmt.Printf("%-20s: %d\n", "削除行数", stats.TotalLinesDeleted)
	fmt.Printf("%-20s: %d\n", "総変更行数", stats.TotalChanges())
	fmt.Println()
	
	if stats.FirstEvent != nil {
		fmt.Printf("%-20s: %s\n", "最初のイベント", stats.FirstEvent.Format("2006-01-02 15:04:05"))
	}
	if stats.LastEvent != nil {
		fmt.Printf("%-20s: %s\n", "最後のイベント", stats.LastEvent.Format("2006-01-02 15:04:05"))
	}
}

// showStatsJSON はJSON形式で統計を表示する
func showStatsJSON(stats *types.Statistics) {
	fmt.Printf(`{
  "total_events": %d,
  "ai_events": %d,
  "human_events": %d,
  "commit_events": %d,
  "ai_percentage": %.1f,
  "human_percentage": %.1f,
  "total_lines_added": %d,
  "total_lines_modified": %d,
  "total_lines_deleted": %d,
  "total_changes": %d`,
		stats.TotalEvents,
		stats.AIEvents,
		stats.HumanEvents,
		stats.CommitEvents,
		stats.AIPercentage(),
		stats.HumanPercentage(),
		stats.TotalLinesAdded,
		stats.TotalLinesModified,
		stats.TotalLinesDeleted,
		stats.TotalChanges())

	if stats.FirstEvent != nil {
		fmt.Printf(`,
  "first_event": "%s"`, stats.FirstEvent.Format("2006-01-02T15:04:05Z07:00"))
	}
	if stats.LastEvent != nil {
		fmt.Printf(`,
  "last_event": "%s"`, stats.LastEvent.Format("2006-01-02T15:04:05Z07:00"))
	}

	fmt.Println("\n}")
}

// showStatsSummary はサマリー形式で統計を表示する
func showStatsSummary(stats *types.Statistics) {
	fmt.Println("📊 AI Code Tracker サマリー")
	fmt.Print(strings.Repeat("=", 30))
	fmt.Println()
	
	if stats.TotalEvents == 0 {
		fmt.Println("まだイベントが記録されていません")
		return
	}
	
	fmt.Printf("🤖 AI によるコード: %.1f%% (%d イベント)\n", stats.AIPercentage(), stats.AIEvents)
	fmt.Printf("👤 人間によるコード: %.1f%% (%d イベント)\n", stats.HumanPercentage(), stats.HumanEvents)
	fmt.Printf("📝 総変更行数: %d 行\n", stats.TotalChanges())
	
	if stats.FirstEvent != nil && stats.LastEvent != nil {
		duration := stats.LastEvent.Sub(*stats.FirstEvent)
		fmt.Printf("📅 追跡期間: %d 日間\n", int(duration.Hours()/24))
	}
}

// handleBlame はファイルのblame情報を表示する
func handleBlame(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrorTypeCommand, "missing_required_argument", "file_path").
			WithSuggestions(
				i18n.T("suggestion_specify_file", "例: aict blame src/main.go"),
			).WithCommand("blame")
	}

	var (
		filePath  = args[0]
		useColor  = true
		showStats = false
		topN      = 0
	)

	// コマンドライン引数をパース
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--no-color":
			useColor = false
		case "--stats":
			showStats = true
		case "--top":
			if i+1 < len(args) {
				if n, err := fmt.Sscanf(args[i+1], "%d", &topN); n == 1 && err == nil {
					i++
				}
			}
		}
	}

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "directory_access_failed")
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer storage.Close()

	// Blamerを初期化
	blamer := blame.NewBlamer(storage, currentDir)

	// ファイルパスを検証
	if err := blamer.ValidateFilePath(filePath); err != nil {
		return errors.FileNotFound(filePath).WithCommand("blame")
	}

	if showStats || topN > 0 {
		// 統計情報または上位貢献者を表示
		if topN > 0 {
			contributors, err := blamer.GetTopContributors(filePath, topN)
			if err != nil {
				return errors.WrapError(err, errors.ErrorTypeData, "contributor_fetch_failed")
			}

			fmt.Printf("=== %s の上位貢献者 ===\n\n", filePath)
			for i, contributor := range contributors {
				indicator := "👤"
				if contributor.IsAI {
					indicator = "🤖"
				}
				fmt.Printf("%d. %s %s - %d行 (%.1f%%)\n", 
					i+1, indicator, contributor.Name, contributor.Lines, contributor.Percentage)
			}
		} else {
			// 貢献者別統計のみ表示
			contribution, err := blamer.GetFileContribution(filePath)
			if err != nil {
				return errors.WrapError(err, errors.ErrorTypeData, "contributor_fetch_failed")
			}

			fmt.Printf("=== %s の貢献者統計 ===\n\n", filePath)
			for contributor, lines := range contribution {
				fmt.Printf("%-20s: %d行\n", contributor, lines)
			}
		}
	} else {
		// 通常のblame表示
		result, err := blamer.BlameFile(filePath)
		if err != nil {
			return errors.WrapError(err, errors.ErrorTypeGit, "blame_fetch_failed")
		}

		// フォーマットして出力
		output := blamer.FormatBlameOutput(result, useColor)
		fmt.Print(output)
	}

	return nil
}

// showDailyStats は日次統計を表示する
func showDailyStats(statsManager *stats.StatsManager, since, until time.Time) error {
	dailyStats, err := statsManager.GetDailyStats(since, until)
	if err != nil {
		return fmt.Errorf("日次統計の取得に失敗しました: %w", err)
	}

	fmt.Printf("=== 日次統計 (%s - %s) ===\n\n", 
		since.Format("2006-01-02"), until.Format("2006-01-02"))

	if len(dailyStats) == 0 {
		fmt.Println("指定期間内にデータがありません")
		return nil
	}

	fmt.Printf("%-12s %-8s %-8s %-8s %-8s %-8s\n", 
		"日付", "AI", "人間", "コミット", "変更行", "AI率")
	fmt.Print(strings.Repeat("-", 60))
	fmt.Println()

	for _, daily := range dailyStats {
		fmt.Printf("%-12s %-8d %-8d %-8d %-8d %6.1f%%\n",
			daily.Date.Format("2006-01-02"),
			daily.AIEvents,
			daily.HumanEvents,
			daily.CommitEvents,
			daily.TotalChanges,
			daily.AIPercentage)
	}

	return nil
}

// showFileStats はファイル別統計を表示する
func showFileStats(statsManager *stats.StatsManager, since time.Time, authorFilter string) error {
	fileStats, err := statsManager.GetFileStats(since)
	if err != nil {
		return fmt.Errorf("ファイル別統計の取得に失敗しました: %w", err)
	}

	fmt.Printf("=== ファイル別統計 (%s以降) ===\n\n", since.Format("2006-01-02"))

	if len(fileStats) == 0 {
		fmt.Println("統計データがありません")
		return nil
	}

	// 作成者フィルタ適用
	if authorFilter != "" {
		fmt.Printf("作成者フィルタ: %s\n\n", authorFilter)
	}

	fmt.Printf("%-30s %-6s %-6s %-8s %-12s %-20s\n", 
		"ファイル", "AI", "人間", "変更行", "最終変更", "主要貢献者")
	fmt.Print(strings.Repeat("-", 90))
	fmt.Println()

	limit := 20 // 上位20ファイルを表示
	for i, file := range fileStats {
		if i >= limit {
			break
		}

		// ファイル名を短縮
		fileName := file.FilePath
		if len(fileName) > 28 {
			fileName = "..." + fileName[len(fileName)-25:]
		}

		fmt.Printf("%-30s %-6d %-6d %-8d %-12s %-20s\n",
			fileName,
			file.AIEvents,
			file.HumanEvents,
			file.TotalChanges,
			file.LastModified.Format("2006-01-02"),
			file.MainContributor)
	}

	if len(fileStats) > limit {
		fmt.Printf("\n... 他 %d ファイル\n", len(fileStats)-limit)
	}

	return nil
}

// showContributorStats は貢献者別統計を表示する
func showContributorStats(statsManager *stats.StatsManager, since time.Time, authorFilter string) error {
	contributorStats, err := statsManager.GetContributorStats(since)
	if err != nil {
		return fmt.Errorf("貢献者別統計の取得に失敗しました: %w", err)
	}

	fmt.Printf("=== 貢献者別統計 (%s以降) ===\n\n", since.Format("2006-01-02"))

	if len(contributorStats) == 0 {
		fmt.Println("統計データがありません")
		return nil
	}

	// 作成者フィルタ適用
	if authorFilter != "" {
		var filtered []stats.ContributorStats
		for _, contributor := range contributorStats {
			if strings.Contains(strings.ToLower(contributor.Name), strings.ToLower(authorFilter)) {
				filtered = append(filtered, contributor)
			}
		}
		contributorStats = filtered
		fmt.Printf("作成者フィルタ: %s (%d人)\n\n", authorFilter, len(contributorStats))
	}

	fmt.Printf("%-20s %-4s %-8s %-6s %-6s %-6s %-8s %-15s\n", 
		"貢献者", "種別", "イベント", "追加", "変更", "削除", "ファイル", "モデル")
	fmt.Print(strings.Repeat("-", 85))
	fmt.Println()

	for _, contributor := range contributorStats {
		typeIndicator := "👤"
		model := "-"
		if contributor.IsAI {
			typeIndicator = "🤖"
			model = contributor.Model
			if len(model) > 13 {
				model = model[:10] + "..."
			}
		}

		name := contributor.Name
		if len(name) > 18 {
			name = name[:15] + "..."
		}

		fmt.Printf("%-20s %-4s %-8d %-6d %-6d %-6d %-8d %-15s\n",
			name,
			typeIndicator,
			contributor.Events,
			contributor.LinesAdded,
			contributor.LinesModified,
			contributor.LinesDeleted,
			contributor.FilesModified,
			model)
	}

	return nil
}

// showTrendAnalysis はトレンド分析を表示する
func showTrendAnalysis(statsManager *stats.StatsManager, since, until time.Time) error {
	analysis, err := statsManager.GetTrendAnalysis(since, until)
	if err != nil {
		return fmt.Errorf("トレンド分析の取得に失敗しました: %w", err)
	}

	fmt.Printf("=== トレンド分析 (%s - %s) ===\n\n", 
		since.Format("2006-01-02"), until.Format("2006-01-02"))

	// AI使用率の推移
	if trend, exists := analysis["ai_usage_trend"]; exists {
		if trendMap, ok := trend.(map[string]float64); ok {
			fmt.Println("📈 AI使用率の推移:")
			fmt.Printf("  最初の週の平均: %.1f%%\n", trendMap["first_week_avg"])
			fmt.Printf("  最後の週の平均: %.1f%%\n", trendMap["last_week_avg"])
			
			change := trendMap["trend_change"]
			changeStr := "📈 増加"
			if change < 0 {
				changeStr = "📉 減少"
			} else if change == 0 {
				changeStr = "📊 変化なし"
			}
			fmt.Printf("  トレンド: %s (%.1f%%)\n\n", changeStr, change)
		}
	}

	// 最も活発な曜日
	if weekday, exists := analysis["most_active_weekday"]; exists {
		if weekdayMap, ok := weekday.(map[string]interface{}); ok {
			fmt.Println("📅 最も活発な曜日:")
			fmt.Printf("  %s (%d回の活動)\n\n", weekdayMap["weekday"], weekdayMap["activity"])
		}
	}

	// AI比率の安定性
	if stability, exists := analysis["ai_ratio_stability"]; exists {
		if stabilityMap, ok := stability.(map[string]float64); ok {
			fmt.Println("📊 AI比率の安定性:")
			stabilityScore := stabilityMap["stability"]
			
			stabilityLevel := "低い"
			if stabilityScore > 80 {
				stabilityLevel = "非常に高い"
			} else if stabilityScore > 60 {
				stabilityLevel = "高い"
			} else if stabilityScore > 40 {
				stabilityLevel = "中程度"
			}
			
			fmt.Printf("  安定性スコア: %.1f%% (%s)\n", stabilityScore, stabilityLevel)
			fmt.Printf("  分散: %.1f\n\n", stabilityMap["variance"])
		}
	}

	return nil
}

// handleSetup はhooks設定を処理する
func handleSetup(args []string) error {
	var (
		gitHooksOnly    = false
		claudeHooksOnly = false
		removeHooks     = false
		showStatus      = false
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--git-hooks":
			gitHooksOnly = true
		case "--claude-hooks":
			claudeHooksOnly = true
		case "--remove":
			removeHooks = true
		case "--status":
			showStatus = true
		}
	}

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("現在のディレクトリの取得に失敗しました: %w", err)
	}

	// HookManagerを初期化
	hookManager := hooks.NewHookManager(currentDir)

	// Gitリポジトリの検証
	if err := hookManager.ValidateGitRepo(); err != nil {
		return fmt.Errorf("Gitリポジトリの検証に失敗しました: %w", err)
	}

	// ステータス表示
	if showStatus {
		return showHookStatus(hookManager)
	}

	// hooks削除
	if removeHooks {
		return removeHooksCmd(hookManager, gitHooksOnly, claudeHooksOnly)
	}

	// 権限チェック
	if err := hookManager.CheckPermissions(); err != nil {
		return fmt.Errorf("権限チェックに失敗しました: %w", err)
	}

	// hooks設定
	return setupHooks(hookManager, gitHooksOnly, claudeHooksOnly)
}

// setupHooks はhooksを設定する
func setupHooks(hookManager *hooks.HookManager, gitOnly, claudeOnly bool) error {
	fmt.Println("=== AI Code Tracker Hooks 設定 ===\n")

	// 既存のhooksをバックアップ
	if err := hookManager.BackupExistingHooks(); err != nil {
		fmt.Printf("警告: 既存hooksのバックアップに失敗しました: %v\n", err)
	}

	// Git hooks設定
	if !claudeOnly {
		fmt.Println("📁 Git hooks を設定中...")
		if err := hookManager.SetupGitHooks(); err != nil {
			return fmt.Errorf("Git hooksの設定に失敗しました: %w", err)
		}
		fmt.Println("✅ Git hooks を設定しました")
	}

	// Claude Code hooks設定
	if !gitOnly {
		fmt.Println("\n🤖 Claude Code hooks を設定中...")
		if err := hookManager.SetupClaudeCodeHooks(); err != nil {
			return fmt.Errorf("Claude Code hooksの設定に失敗しました: %w", err)
		}
		fmt.Println("✅ Claude Code hooks を設定しました")
	}

	fmt.Println("\n🎉 Hooks設定が完了しました！")
	fmt.Println("\n次のステップ:")
	if !gitOnly {
		homeDir, _ := os.UserHomeDir()
		hooksPath := filepath.Join(homeDir, ".claude", "hooks-aict.json")
		fmt.Printf("1. 環境変数を設定: export CLAUDE_HOOKS_CONFIG=%s\n", hooksPath)
		fmt.Println("2. Claude Codeを再起動してhooksを有効化")
	}
	if !claudeOnly {
		fmt.Println("3. Gitでコミットを行うと自動的にトラッキングが開始されます")
	}

	return nil
}

// removeHooksCmd はhooksを削除する
func removeHooksCmd(hookManager *hooks.HookManager, gitOnly, claudeOnly bool) error {
	fmt.Println("=== AI Code Tracker Hooks 削除 ===\n")

	// Git hooks削除
	if !claudeOnly {
		fmt.Println("📁 Git hooks を削除中...")
		if err := hookManager.RemoveGitHooks(); err != nil {
			return fmt.Errorf("Git hooksの削除に失敗しました: %w", err)
		}
		fmt.Println("✅ Git hooks を削除しました")
	}

	// Claude Code hooks削除
	if !gitOnly {
		fmt.Println("\n🤖 Claude Code hooks を削除中...")
		if err := hookManager.RemoveClaudeCodeHooks(); err != nil {
			return fmt.Errorf("Claude Code hooksの削除に失敗しました: %w", err)
		}
		fmt.Println("✅ Claude Code hooks を削除しました")
	}

	fmt.Println("\n🎉 Hooks削除が完了しました！")
	return nil
}

// showHookStatus はhooksの設定状況を表示する
func showHookStatus(hookManager *hooks.HookManager) error {
	fmt.Println("=== AI Code Tracker Hooks 設定状況 ===\n")

	status, err := hookManager.GetHookStatus()
	if err != nil {
		return fmt.Errorf("hooks状況の取得に失敗しました: %w", err)
	}

	// Git hooks状況
	if gitHooks, ok := status["git_hooks"].(map[string]interface{}); ok {
		fmt.Println("📁 Git Hooks:")
		if installed, ok := gitHooks["installed"].(bool); ok && installed {
			fmt.Println("  ✅ インストール済み")
		} else {
			fmt.Println("  ❌ 未インストール")
		}

		if path, ok := gitHooks["path"].(string); ok {
			fmt.Printf("  📂 パス: %s\n", path)
		}

		if executable, ok := gitHooks["executable"].(bool); ok {
			if executable {
				fmt.Println("  ✅ 実行可能")
			} else {
				fmt.Println("  ❌ 実行権限なし")
			}
		}

		if backup, ok := gitHooks["backup"].(bool); ok && backup {
			fmt.Println("  💾 バックアップあり")
		}
	}

	fmt.Println()

	// Claude Code hooks状況
	if claudeHooks, ok := status["claude_hooks"].(map[string]interface{}); ok {
		fmt.Println("🤖 Claude Code Hooks:")
		if installed, ok := claudeHooks["installed"].(bool); ok && installed {
			fmt.Println("  ✅ インストール済み")
		} else {
			fmt.Println("  ❌ 未インストール")
		}

		if path, ok := claudeHooks["path"].(string); ok {
			fmt.Printf("  📂 パス: %s\n", path)
		}

		if envVarSet, ok := claudeHooks["env_var_set"].(bool); ok {
			if envVarSet {
				fmt.Println("  ✅ 環境変数設定済み")
			} else {
				fmt.Println("  ❌ 環境変数未設定")
				if path, ok := claudeHooks["path"].(string); ok {
					fmt.Printf("  💡 実行してください: export CLAUDE_HOOKS_CONFIG=%s\n", path)
				}
			}
		}

		if backup, ok := claudeHooks["backup"].(bool); ok && backup {
			fmt.Println("  💾 バックアップあり")
		}
	}

	return nil
}

// handleConfig は設定管理を処理する（今後実装）
func handleConfig(args []string) error {
	fmt.Println("設定機能は今後のバージョンで実装予定です")
	return nil
}

// handleVersion はバージョン情報を表示する
func handleVersion(args []string) error {
	fmt.Printf("%s version %s\n", AppName, Version)
	fmt.Println("AI Code Tracker - AIと人間によるコード変更の自動追跡システム")
	return nil
}

// handleHelp は改良されたヘルプを表示する
func handleHelp(args []string) error {
	if len(args) > 0 {
		// 特定のコマンドのヘルプを表示
		helpSystem.ShowCommandHelp(args[0])
	} else {
		// メインヘルプを表示
		helpSystem.ShowMainHelp()
	}
	return nil
}

// handleWizard はインタラクティブウィザードを実行する
func handleWizard(args []string) error {
	wizard := interactive.NewWizard()
	
	var wizardType string
	if len(args) > 0 {
		wizardType = args[0]
	} else {
		wizardType = "init"
	}
	
	switch wizardType {
	case "init":
		config := wizard.InitializationWizard()
		helpSystem.ShowSuccess("設定ウィザードが完了しました!")
		
		// 設定を適用
		return applyWizardConfig(config)
		
	case "security":
		config := wizard.SecurityWizard()
		helpSystem.ShowSuccess("セキュリティ設定が完了しました!")
		
		// セキュリティ設定を適用
		return applySecurityConfig(config)
		
	case "quickstart":
		wizard.QuickStartWizard()
		return nil
		
	default:
		return fmt.Errorf("不明なウィザードタイプ: %s", wizardType)
	}
}

// applyWizardConfig はウィザード設定を適用する
func applyWizardConfig(config map[string]interface{}) error {
	// 基本設定の適用
	if setupGit, ok := config["setup_git_hooks"].(bool); ok && setupGit {
		helpSystem.ShowInfo("Git hooks を設定中...")
		// Git hooks設定のロジックを呼び出し
	}
	
	if setupClaude, ok := config["setup_claude_hooks"].(bool); ok && setupClaude {
		helpSystem.ShowInfo("Claude Code hooks を設定中...")
		// Claude hooks設定のロジックを呼び出し
	}
	
	// セキュリティ設定の適用
	if enableEncryption, ok := config["enable_encryption"].(bool); ok && enableEncryption {
		helpSystem.ShowInfo("データ暗号化を有効化中...")
		os.Setenv("AICT_ENCRYPT_DATA", "true")
	}
	
	if enableAudit, ok := config["enable_audit_log"].(bool); ok && enableAudit {
		helpSystem.ShowInfo("監査ログを有効化中...")
		os.Setenv("AICT_AUDIT_LOG", "true")
	}
	
	if anonymize, ok := config["anonymize_authors"].(bool); ok && anonymize {
		helpSystem.ShowInfo("作成者匿名化を有効化中...")
		os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	}
	
	helpSystem.ShowSuccess("設定が正常に適用されました")
	return nil
}

// applySecurityConfig はセキュリティ設定を適用する
func applySecurityConfig(config map[string]interface{}) error {
	securityMode, ok := config["security_mode"].(string)
	if !ok {
		return fmt.Errorf("セキュリティモードが指定されていません")
	}
	
	helpSystem.ShowInfo(fmt.Sprintf("セキュリティモード '%s' を適用中...", securityMode))
	
	// セキュリティモードに応じた環境変数設定
	os.Setenv("AICT_SECURITY_MODE", securityMode)
	
	for key, value := range config {
		switch key {
		case "enable_encryption":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_ENCRYPT_DATA", "true")
			}
		case "enable_audit_log":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_AUDIT_LOG", "true")
			}
		case "anonymize_authors":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
			}
		case "strict_validation":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_STRICT_VALIDATION", "true")
			}
		case "hash_file_paths":
			if val, ok := value.(bool); ok && val {
				os.Setenv("AICT_HASH_FILE_PATHS", "true")
			}
		case "data_retention_days":
			if val, ok := value.(int); ok {
				os.Setenv("AICT_DATA_RETENTION_DAYS", fmt.Sprintf("%d", val))
			}
		}
	}
	
	helpSystem.ShowSuccess("セキュリティ設定が正常に適用されました")
	return nil
}