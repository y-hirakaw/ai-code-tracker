package period

import (
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// Analyzer provides period-based analysis functionality
type Analyzer struct {
	config *tracker.Config
}

// NewAnalyzer creates a new period analyzer
func NewAnalyzer(config *tracker.Config) *Analyzer {
	return &Analyzer{
		config: config,
	}
}

// AnalyzePeriod analyzes records within the specified time range
func (a *Analyzer) AnalyzePeriod(records []tracker.CheckpointRecord, timeRange *TimeRange) (*PeriodReport, error) {
	filtered := FilterRecordsInclusive(records, timeRange)

	report := &PeriodReport{
		Range: *timeRange,
	}

	// Calculate totals using the same logic as the main analyzer
	analyzer := tracker.NewAnalyzer(a.config)

	for _, record := range filtered {
		if analyzer.IsAIAuthor(record.Author) {
			report.AILines += record.Added
		} else {
			report.HumanLines += record.Added
		}
	}

	report.TotalLines = report.AILines + report.HumanLines

	if report.TotalLines > 0 {
		report.Percentage = float64(report.AILines) / float64(report.TotalLines) * 100
	}

	// Generate daily statistics if requested
	report.DailyStats = a.generateDailyStats(filtered)

	return report, nil
}

// generateDailyStats creates daily aggregated statistics
func (a *Analyzer) generateDailyStats(records []tracker.CheckpointRecord) []DailyStat {
	dailyMap := make(map[string]*DailyStat)
	analyzer := tracker.NewAnalyzer(a.config)

	for _, record := range records {
		dateKey := record.Timestamp.Format("2006-01-02")

		if _, exists := dailyMap[dateKey]; !exists {
			date, _ := time.Parse("2006-01-02", dateKey)
			dailyMap[dateKey] = &DailyStat{
				Date: date,
			}
		}

		if analyzer.IsAIAuthor(record.Author) {
			dailyMap[dateKey].AILines += record.Added
		} else {
			dailyMap[dateKey].HumanLines += record.Added
		}
	}

	// Convert map to sorted slice
	dailyStats := make([]DailyStat, 0, len(dailyMap))
	for _, stat := range dailyMap {
		dailyStats = append(dailyStats, *stat)
	}

	// Sort by date (simple bubble sort for small datasets)
	for i := 0; i < len(dailyStats)-1; i++ {
		for j := 0; j < len(dailyStats)-i-1; j++ {
			if dailyStats[j].Date.After(dailyStats[j+1].Date) {
				dailyStats[j], dailyStats[j+1] = dailyStats[j+1], dailyStats[j]
			}
		}
	}

	return dailyStats
}
