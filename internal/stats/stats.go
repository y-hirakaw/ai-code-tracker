package stats

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/pkg/types"
)

// StatsManager は統計情報管理を提供する
type StatsManager struct {
	storage *storage.Storage
}

// NewStatsManager は新しいStatsManagerインスタンスを作成する
func NewStatsManager(storage *storage.Storage) *StatsManager {
	return &StatsManager{
		storage: storage,
	}
}

// DailyStats は日次統計情報を表す
type DailyStats struct {
	// Date は日付
	Date time.Time
	// AIEvents はAIイベント数
	AIEvents int
	// HumanEvents は人間イベント数
	HumanEvents int
	// CommitEvents はコミットイベント数
	CommitEvents int
	// TotalChanges は総変更行数
	TotalChanges int
	// AIPercentage はAIの貢献率
	AIPercentage float64
}

// FileStats はファイル別統計情報を表す
type FileStats struct {
	// FilePath はファイルパス
	FilePath string
	// AIEvents はAIイベント数
	AIEvents int
	// HumanEvents は人間イベント数
	HumanEvents int
	// TotalChanges は総変更行数
	TotalChanges int
	// LastModified は最終変更日時
	LastModified time.Time
	// MainContributor は主要貢献者
	MainContributor string
}

// ContributorStats は貢献者別統計情報を表す
type ContributorStats struct {
	// Name は貢献者名
	Name string
	// IsAI はAIかどうか
	IsAI bool
	// Events はイベント数
	Events int
	// LinesAdded は追加行数
	LinesAdded int
	// LinesModified は変更行数
	LinesModified int
	// LinesDeleted は削除行数
	LinesDeleted int
	// FilesModified は変更ファイル数
	FilesModified int
	// FirstActivity は初回活動日時
	FirstActivity time.Time
	// LastActivity は最終活動日時
	LastActivity time.Time
	// Model はAIモデル（AIの場合のみ）
	Model string
}

// PeriodStats は期間統計情報を表す
type PeriodStats struct {
	// StartDate は開始日
	StartDate time.Time
	// EndDate は終了日
	EndDate time.Time
	// TotalEvents は総イベント数
	TotalEvents int
	// AIEvents はAIイベント数
	AIEvents int
	// HumanEvents は人間イベント数
	HumanEvents int
	// DailyStats は日次統計のスライス
	DailyStats []DailyStats
	// TopContributors は上位貢献者
	TopContributors []ContributorStats
	// TopFiles は最も変更されたファイル
	TopFiles []FileStats
}

// GetDailyStats は日次統計を取得する
func (sm *StatsManager) GetDailyStats(since time.Time, until time.Time) ([]DailyStats, error) {
	events, err := sm.storage.ReadEventsByDateRange(since, until)
	if err != nil {
		return nil, fmt.Errorf("イベントの取得に失敗しました: %w", err)
	}

	// 日付ごとにイベントをグループ化
	dailyEventMap := make(map[string][]*types.TrackEvent)
	for _, event := range events {
		dateKey := event.Timestamp.Format("2006-01-02")
		dailyEventMap[dateKey] = append(dailyEventMap[dateKey], event)
	}

	var dailyStats []DailyStats
	for dateStr, dayEvents := range dailyEventMap {
		date, _ := time.Parse("2006-01-02", dateStr)
		
		stats := DailyStats{
			Date: date,
		}

		for _, event := range dayEvents {
			switch event.EventType {
			case types.EventTypeAI:
				stats.AIEvents++
			case types.EventTypeHuman:
				stats.HumanEvents++
			case types.EventTypeCommit:
				stats.CommitEvents++
			}

			// 変更行数を集計
			for _, file := range event.Files {
				stats.TotalChanges += file.LinesAdded + file.LinesModified + file.LinesDeleted
			}
		}

		// AIの貢献率を計算
		totalActivityEvents := stats.AIEvents + stats.HumanEvents
		if totalActivityEvents > 0 {
			stats.AIPercentage = float64(stats.AIEvents) / float64(totalActivityEvents) * 100.0
		}

		dailyStats = append(dailyStats, stats)
	}

	// 日付でソート
	sort.Slice(dailyStats, func(i, j int) bool {
		return dailyStats[i].Date.Before(dailyStats[j].Date)
	})

	return dailyStats, nil
}

// GetFileStats はファイル別統計を取得する
func (sm *StatsManager) GetFileStats(since time.Time) ([]FileStats, error) {
	events, err := sm.storage.ReadEventsByDateRange(since, time.Now())
	if err != nil {
		return nil, fmt.Errorf("イベントの取得に失敗しました: %w", err)
	}

	fileStatsMap := make(map[string]*FileStats)

	for _, event := range events {
		for _, file := range event.Files {
			filePath := file.Path
			
			if _, exists := fileStatsMap[filePath]; !exists {
				fileStatsMap[filePath] = &FileStats{
					FilePath:     filePath,
					LastModified: event.Timestamp,
				}
			}

			stat := fileStatsMap[filePath]
			
			// イベントタイプ別カウント
			switch event.EventType {
			case types.EventTypeAI:
				stat.AIEvents++
			case types.EventTypeHuman:
				stat.HumanEvents++
			}

			// 変更行数を集計
			stat.TotalChanges += file.LinesAdded + file.LinesModified + file.LinesDeleted

			// 最終変更日時を更新
			if event.Timestamp.After(stat.LastModified) {
				stat.LastModified = event.Timestamp
				stat.MainContributor = event.Author
			}
		}
	}

	// スライスに変換
	var fileStats []FileStats
	for _, stat := range fileStatsMap {
		fileStats = append(fileStats, *stat)
	}

	// 変更回数でソート
	sort.Slice(fileStats, func(i, j int) bool {
		totalI := fileStats[i].AIEvents + fileStats[i].HumanEvents
		totalJ := fileStats[j].AIEvents + fileStats[j].HumanEvents
		return totalI > totalJ
	})

	return fileStats, nil
}

// GetContributorStats は貢献者別統計を取得する
func (sm *StatsManager) GetContributorStats(since time.Time) ([]ContributorStats, error) {
	events, err := sm.storage.ReadEventsByDateRange(since, time.Now())
	if err != nil {
		return nil, fmt.Errorf("イベントの取得に失敗しました: %w", err)
	}

	contributorStatsMap := make(map[string]*ContributorStats)
	filesModifiedMap := make(map[string]map[string]bool) // contributor -> files

	for _, event := range events {
		author := event.Author
		
		if _, exists := contributorStatsMap[author]; !exists {
			contributorStatsMap[author] = &ContributorStats{
				Name:          author,
				IsAI:          event.EventType == types.EventTypeAI,
				FirstActivity: event.Timestamp,
				LastActivity:  event.Timestamp,
				Model:         event.Model,
			}
			filesModifiedMap[author] = make(map[string]bool)
		}

		stat := contributorStatsMap[author]
		stat.Events++

		// 活動期間を更新
		if event.Timestamp.Before(stat.FirstActivity) {
			stat.FirstActivity = event.Timestamp
		}
		if event.Timestamp.After(stat.LastActivity) {
			stat.LastActivity = event.Timestamp
		}

		// ファイル変更情報を集計
		for _, file := range event.Files {
			stat.LinesAdded += file.LinesAdded
			stat.LinesModified += file.LinesModified
			stat.LinesDeleted += file.LinesDeleted
			
			// ファイル変更カウント
			filesModifiedMap[author][file.Path] = true
		}
	}

	// ファイル変更数を計算
	for author, stat := range contributorStatsMap {
		stat.FilesModified = len(filesModifiedMap[author])
	}

	// スライスに変換
	var contributorStats []ContributorStats
	for _, stat := range contributorStatsMap {
		contributorStats = append(contributorStats, *stat)
	}

	// イベント数でソート
	sort.Slice(contributorStats, func(i, j int) bool {
		return contributorStats[i].Events > contributorStats[j].Events
	})

	return contributorStats, nil
}

// GetPeriodStats は期間統計を取得する
func (sm *StatsManager) GetPeriodStats(since time.Time, until time.Time) (*PeriodStats, error) {
	events, err := sm.storage.ReadEventsByDateRange(since, until)
	if err != nil {
		return nil, fmt.Errorf("イベントの取得に失敗しました: %w", err)
	}

	stats := &PeriodStats{
		StartDate:   since,
		EndDate:     until,
		TotalEvents: len(events),
	}

	// イベントタイプ別カウント
	for _, event := range events {
		switch event.EventType {
		case types.EventTypeAI:
			stats.AIEvents++
		case types.EventTypeHuman:
			stats.HumanEvents++
		}
	}

	// 日次統計を取得
	dailyStats, err := sm.GetDailyStats(since, until)
	if err != nil {
		return nil, fmt.Errorf("日次統計の取得に失敗しました: %w", err)
	}
	stats.DailyStats = dailyStats

	// 上位貢献者を取得（上位10名）
	contributorStats, err := sm.GetContributorStats(since)
	if err != nil {
		return nil, fmt.Errorf("貢献者統計の取得に失敗しました: %w", err)
	}
	if len(contributorStats) > 10 {
		stats.TopContributors = contributorStats[:10]
	} else {
		stats.TopContributors = contributorStats
	}

	// 上位ファイルを取得（上位10ファイル）
	fileStats, err := sm.GetFileStats(since)
	if err != nil {
		return nil, fmt.Errorf("ファイル統計の取得に失敗しました: %w", err)
	}
	if len(fileStats) > 10 {
		stats.TopFiles = fileStats[:10]
	} else {
		stats.TopFiles = fileStats
	}

	return stats, nil
}

// FilterByAuthor は指定した作成者でイベントをフィルタリングする
func (sm *StatsManager) FilterByAuthor(events []*types.TrackEvent, author string) []*types.TrackEvent {
	var filtered []*types.TrackEvent
	for _, event := range events {
		if strings.Contains(strings.ToLower(event.Author), strings.ToLower(author)) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetTrendAnalysis はトレンド分析を行う
func (sm *StatsManager) GetTrendAnalysis(since time.Time, until time.Time) (map[string]interface{}, error) {
	periodStats, err := sm.GetPeriodStats(since, until)
	if err != nil {
		return nil, fmt.Errorf("期間統計の取得に失敗しました: %w", err)
	}

	analysis := make(map[string]interface{})

	// AI使用率の推移
	if len(periodStats.DailyStats) > 1 {
		firstWeek := periodStats.DailyStats[:min(7, len(periodStats.DailyStats))]
		lastWeek := periodStats.DailyStats[max(0, len(periodStats.DailyStats)-7):]

		firstWeekAvg := calculateAverageAIPercentage(firstWeek)
		lastWeekAvg := calculateAverageAIPercentage(lastWeek)

		analysis["ai_usage_trend"] = map[string]float64{
			"first_week_avg": firstWeekAvg,
			"last_week_avg":  lastWeekAvg,
			"trend_change":   lastWeekAvg - firstWeekAvg,
		}
	}

	// 最も活発な曜日
	weekdayActivity := make(map[time.Weekday]int)
	for _, daily := range periodStats.DailyStats {
		weekdayActivity[daily.Date.Weekday()] += daily.AIEvents + daily.HumanEvents
	}

	var mostActiveWeekday time.Weekday
	maxActivity := 0
	for weekday, activity := range weekdayActivity {
		if activity > maxActivity {
			maxActivity = activity
			mostActiveWeekday = weekday
		}
	}

	analysis["most_active_weekday"] = map[string]interface{}{
		"weekday":  mostActiveWeekday.String(),
		"activity": maxActivity,
	}

	// AI/人間比率の安定性
	var aiPercentages []float64
	for _, daily := range periodStats.DailyStats {
		aiPercentages = append(aiPercentages, daily.AIPercentage)
	}

	if len(aiPercentages) > 0 {
		variance := calculateVariance(aiPercentages)
		analysis["ai_ratio_stability"] = map[string]float64{
			"variance":   variance,
			"stability": 100.0 - variance, // 低い分散 = 高い安定性
		}
	}

	return analysis, nil
}

// ヘルパー関数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func calculateAverageAIPercentage(dailyStats []DailyStats) float64 {
	if len(dailyStats) == 0 {
		return 0.0
	}

	total := 0.0
	for _, daily := range dailyStats {
		total += daily.AIPercentage
	}
	return total / float64(len(dailyStats))
}

func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// 平均を計算
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 分散を計算
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	return variance / float64(len(values))
}