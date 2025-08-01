package period

import "time"

// TimeRange represents a time range for filtering
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// PeriodReport contains statistics for a specific time period
type PeriodReport struct {
	Range      TimeRange   `json:"range"`
	TotalLines int         `json:"total_lines"`
	AILines    int         `json:"ai_lines"`
	HumanLines int         `json:"human_lines"`
	Percentage float64     `json:"percentage"`
	DailyStats []DailyStat `json:"daily_stats,omitempty"`
}

// DailyStat represents daily aggregated statistics
type DailyStat struct {
	Date       time.Time `json:"date"`
	AILines    int       `json:"ai_lines"`
	HumanLines int       `json:"human_lines"`
}

// ReportFormat defines output format options
type ReportFormat string

// String returns the string representation of ReportFormat
func (r ReportFormat) String() string {
	return string(r)
}

const (
	FormatTable ReportFormat = "table"
	FormatGraph ReportFormat = "graph"
	FormatJSON  ReportFormat = "json"
	FormatCSV   ReportFormat = "csv"
)
