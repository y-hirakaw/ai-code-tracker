package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/stats"
	"github.com/y-hirakaw/ai-code-tracker/internal/utils"
	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// StatsPresenter は統計情報の表示を担当する
type StatsPresenter struct{}

// NewStatsPresenter は新しいStatsPresenterを作成する
func NewStatsPresenter() *StatsPresenter {
	return &StatsPresenter{}
}

// ShowAuthorFilter は作成者フィルタの情報を表示する
func (p *StatsPresenter) ShowAuthorFilter(author string) {
	fmt.Printf("作成者フィルタ: %s\n", author)
	fmt.Printf("注意: 作成者フィルタは基本統計には適用されません\n\n")
}

// ShowStatsTable はテーブル形式で統計を表示する
func (p *StatsPresenter) ShowStatsTable(stats *types.Statistics) {
	fmt.Println("=== AI Code Tracker 統計情報 ===")
	
	fmt.Printf("%-20s: %d\n", "総イベント数", stats.TotalEvents)
	fmt.Printf("%-20s: %d (%s)\n", "AI イベント", stats.AIEvents, utils.FormatPercentage(stats.AIPercentage()))
	fmt.Printf("%-20s: %d (%s)\n", "人間 イベント", stats.HumanEvents, utils.FormatPercentage(stats.HumanPercentage()))
	fmt.Printf("%-20s: %d\n", "コミット イベント", stats.CommitEvents)
	fmt.Println()
	
	fmt.Printf("%-20s: %d\n", "追加行数", stats.TotalLinesAdded)
	fmt.Printf("%-20s: %d\n", "変更行数", stats.TotalLinesModified)
	fmt.Printf("%-20s: %d\n", "削除行数", stats.TotalLinesDeleted)
	fmt.Printf("%-20s: %d\n", "総変更行数", stats.TotalChanges())
	fmt.Println()
	
	if stats.FirstEvent != nil {
		fmt.Printf("%-20s: %s\n", "最初のイベント", utils.FormatTimestamp(*stats.FirstEvent))
	}
	if stats.LastEvent != nil {
		fmt.Printf("%-20s: %s\n", "最後のイベント", utils.FormatTimestamp(*stats.LastEvent))
	}
}

// ShowStatsJSON はJSON形式で統計を表示する
func (p *StatsPresenter) ShowStatsJSON(stats *types.Statistics) {
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

// ShowStatsSummary はサマリー形式で統計を表示する
func (p *StatsPresenter) ShowStatsSummary(stats *types.Statistics) {
	fmt.Println("📊 AI Code Tracker サマリー")
	fmt.Println(utils.CreateSeparatorLine("=", 30))
	
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

// ShowDailyStats は日次統計を表示する
func (p *StatsPresenter) ShowDailyStats(statsManager *stats.StatsManager, since, until time.Time) error {
	dailyStats, err := statsManager.GetDailyStats(since, until)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "statistics_fetch_failed")
	}

	fmt.Printf("=== 日次統計 (%s - %s) ===\n\n", 
		utils.FormatDate(since), utils.FormatDate(until))

	if len(dailyStats) == 0 {
		fmt.Println("指定期間内にデータがありません")
		return nil
	}

	fmt.Printf("%-12s %-8s %-8s %-8s %-8s %-8s\n", 
		"日付", "AI", "人間", "コミット", "変更行", "AI率")
	fmt.Println(utils.CreateSeparatorLine("-", 60))

	for _, daily := range dailyStats {
		fmt.Printf("%-12s %-8d %-8d %-8d %-8d %6.1f%%\n",
			utils.FormatDate(daily.Date),
			daily.AIEvents,
			daily.HumanEvents,
			daily.CommitEvents,
			daily.TotalChanges,
			daily.AIPercentage)
	}

	return nil
}

// ShowFileStats はファイル別統計を表示する
func (p *StatsPresenter) ShowFileStats(statsManager *stats.StatsManager, since time.Time, authorFilter string) error {
	fileStats, err := statsManager.GetFileStats(since)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "statistics_fetch_failed")
	}

	fmt.Printf("=== ファイル別統計 (%s以降) ===\n\n", utils.FormatDate(since))

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
	fmt.Println(utils.CreateSeparatorLine("-", 90))

	limit := 20 // 上位20ファイルを表示
	for i, file := range fileStats {
		if i >= limit {
			break
		}

		// ファイル名を短縮
		fileName := utils.TruncateStringPrefix(file.FilePath, 28)

		fmt.Printf("%-30s %-6d %-6d %-8d %-12s %-20s\n",
			fileName,
			file.AIEvents,
			file.HumanEvents,
			file.TotalChanges,
			utils.FormatDate(file.LastModified),
			file.MainContributor)
	}

	if len(fileStats) > limit {
		fmt.Printf("\n... 他 %d ファイル\n", len(fileStats)-limit)
	}

	return nil
}

// ShowContributorStats は貢献者別統計を表示する
func (p *StatsPresenter) ShowContributorStats(statsManager *stats.StatsManager, since time.Time, authorFilter string) error {
	contributorStats, err := statsManager.GetContributorStats(since)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "statistics_fetch_failed")
	}

	fmt.Printf("=== 貢献者別統計 (%s以降) ===\n\n", utils.FormatDate(since))

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
	fmt.Println(utils.CreateSeparatorLine("-", 85))

	for _, contributor := range contributorStats {
		typeIndicator := "👤"
		model := "-"
		if contributor.IsAI {
			typeIndicator = "🤖"
			model = utils.TruncateString(contributor.Model, 13)
		}

		name := utils.TruncateString(contributor.Name, 18)

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

// ShowTrendAnalysis はトレンド分析を表示する
func (p *StatsPresenter) ShowTrendAnalysis(statsManager *stats.StatsManager, since, until time.Time) error {
	analysis, err := statsManager.GetTrendAnalysis(since, until)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "statistics_fetch_failed")
	}

	fmt.Printf("=== トレンド分析 (%s - %s) ===\n\n", 
		utils.FormatDate(since), utils.FormatDate(until))

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