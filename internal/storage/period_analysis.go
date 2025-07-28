package storage

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PeriodAnalysis は期間別分析の結果
type PeriodAnalysis struct {
	StartDate       time.Time           `json:"start_date"`
	EndDate         time.Time           `json:"end_date"`
	TotalLines      int                 `json:"total_lines"`
	AILines         int                 `json:"ai_lines"`
	HumanLines      int                 `json:"human_lines"`
	AIPercentage    float64             `json:"ai_percentage"`
	FileCount       int                 `json:"file_count"`
	SessionCount    int                 `json:"session_count"`
	ActiveDays      int                 `json:"active_days"`
	FileBreakdown   []FileAnalysis      `json:"file_breakdown"`
	LanguageStats   []LanguageAnalysis  `json:"language_stats"`
	DailyTimeline   []DailyPoint        `json:"daily_timeline"`
	ContributorStats []ContributorStat  `json:"contributor_stats"`
}

// FileAnalysis はファイル別の分析結果
type FileAnalysis struct {
	FilePath            string    `json:"file_path"`
	Language            string    `json:"language"`
	TotalLines          int       `json:"total_lines"`
	AILines             int       `json:"ai_lines"`
	HumanLines          int       `json:"human_lines"`
	AIPercentage        float64   `json:"ai_percentage"`
	FirstEdit           time.Time `json:"first_edit"`
	LastEdit            time.Time `json:"last_edit"`
	ActiveDays          int       `json:"active_days"`
	AISessions          int       `json:"ai_sessions"`
	HumanSessions       int       `json:"human_sessions"`
	AIEfficiency        float64   `json:"ai_efficiency"`        // lines per session
	HumanEfficiency     float64   `json:"human_efficiency"`     // lines per session
	NetContribution     int       `json:"net_contribution"`     // lines_added - lines_deleted
}

// LanguageAnalysis は言語別の分析結果
type LanguageAnalysis struct {
	Language        string  `json:"language"`
	TotalLines      int     `json:"total_lines"`
	AILines         int     `json:"ai_lines"`
	HumanLines      int     `json:"human_lines"`
	AIPercentage    float64 `json:"ai_percentage"`
	FileCount       int     `json:"file_count"`
	SessionCount    int     `json:"session_count"`
}

// DailyPoint は日別統計のポイント
type DailyPoint struct {
	Date            time.Time `json:"date"`
	AILines         int       `json:"ai_lines"`
	HumanLines      int       `json:"human_lines"`
	TotalLines      int       `json:"total_lines"`
	FilesEdited     int       `json:"files_edited"`
	Sessions        int       `json:"sessions"`
	ActiveHours     int       `json:"active_hours"`
}

// ContributorStat は貢献者統計
type ContributorStat struct {
	Author          string  `json:"author"`
	AuthorType      string  `json:"author_type"`
	Lines           int     `json:"lines"`
	Percentage      float64 `json:"percentage"`
	Sessions        int     `json:"sessions"`
	FilesEdited     int     `json:"files_edited"`
	AvgLinesPerDay  float64 `json:"avg_lines_per_day"`
}

// GetPeriodAnalysis は指定期間の詳細分析を実行する
func (s *DuckDBStorage) GetPeriodAnalysis(ctx context.Context, startDate, endDate time.Time) (*PeriodAnalysis, error) {
	analysis := &PeriodAnalysis{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// 1. 全体サマリーを取得
	if err := s.getPeriodSummary(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to get period summary: %w", err)
	}

	// 2. ファイル別詳細分析
	if err := s.getFileBreakdown(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to get file breakdown: %w", err)
	}

	// 3. 言語別統計
	if err := s.getLanguageStats(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to get language stats: %w", err)
	}

	// 4. 日別タイムライン
	if err := s.getDailyTimeline(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to get daily timeline: %w", err)
	}

	// 5. 貢献者统计
	if err := s.getContributorStats(ctx, analysis); err != nil {
		return nil, fmt.Errorf("failed to get contributor stats: %w", err)
	}

	return analysis, nil
}

// getPeriodSummary は期間の全体サマリーを取得
func (s *DuckDBStorage) getPeriodSummary(ctx context.Context, analysis *PeriodAnalysis) error {
	query := `
	WITH period_summary AS (
		SELECT 
			COUNT(DISTINCT t.id) as total_events,
			COUNT(DISTINCT t.session_id) as session_count,
			COUNT(DISTINCT fc.file_path) as file_count,
			COUNT(DISTINCT CAST(t.timestamp AS DATE)) as active_days,
			COALESCE(SUM(CASE WHEN t.author_type = 'ai' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as ai_lines,
			COALESCE(SUM(CASE WHEN t.author_type = 'human' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as human_lines
		FROM tracks t
		LEFT JOIN file_changes fc ON t.id = fc.track_id
		WHERE t.timestamp >= ? AND t.timestamp <= ?
	)
	SELECT 
		COALESCE(session_count, 0),
		COALESCE(file_count, 0),
		COALESCE(active_days, 0),
		COALESCE(ai_lines, 0),
		COALESCE(human_lines, 0),
		COALESCE(ai_lines + human_lines, 0) as total_lines,
		CASE 
			WHEN (ai_lines + human_lines) > 0 
			THEN ai_lines::FLOAT / (ai_lines + human_lines) * 100
			ELSE 0 
		END as ai_percentage
	FROM period_summary
	`

	row := s.db.QueryRowContext(ctx, query, analysis.StartDate, analysis.EndDate)
	
	return row.Scan(
		&analysis.SessionCount,
		&analysis.FileCount,
		&analysis.ActiveDays,
		&analysis.AILines,
		&analysis.HumanLines,
		&analysis.TotalLines,
		&analysis.AIPercentage,
	)
}

// getFileBreakdown はファイル別の詳細分析を取得
func (s *DuckDBStorage) getFileBreakdown(ctx context.Context, analysis *PeriodAnalysis) error {
	query := `
	WITH file_stats AS (
		SELECT 
			fc.file_path,
			fc.language,
			MIN(t.timestamp) as first_edit,
			MAX(t.timestamp) as last_edit,
			COUNT(DISTINCT CAST(t.timestamp AS DATE)) as active_days,
			
			-- AI統計
			COALESCE(SUM(CASE WHEN t.author_type = 'ai' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as ai_lines,
			COUNT(CASE WHEN t.author_type = 'ai' THEN 1 END) as ai_sessions,
			
			-- 人間統計
			COALESCE(SUM(CASE WHEN t.author_type = 'human' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as human_lines,
			COUNT(CASE WHEN t.author_type = 'human' THEN 1 END) as human_sessions,
			
			-- 全体統計
			COALESCE(SUM(fc.lines_added - fc.lines_deleted), 0) as net_contribution
			
		FROM tracks t
		JOIN file_changes fc ON t.id = fc.track_id
		WHERE t.timestamp >= ? AND t.timestamp <= ?
		GROUP BY fc.file_path, fc.language
		HAVING net_contribution != 0  -- 実質的な変更があったファイルのみ
	)
	SELECT 
		file_path,
		language,
		first_edit,
		last_edit,
		active_days,
		ai_lines,
		ai_sessions,
		human_lines,
		human_sessions,
		ai_lines + human_lines as total_lines,
		CASE 
			WHEN (ai_lines + human_lines) > 0 
			THEN ai_lines::FLOAT / (ai_lines + human_lines) * 100
			ELSE 0 
		END as ai_percentage,
		CASE WHEN ai_sessions > 0 THEN ai_lines::FLOAT / ai_sessions ELSE 0 END as ai_efficiency,
		CASE WHEN human_sessions > 0 THEN human_lines::FLOAT / human_sessions ELSE 0 END as human_efficiency,
		net_contribution
	FROM file_stats
	ORDER BY (ai_lines + human_lines) DESC
	LIMIT 100  -- 上位100ファイル
	`

	rows, err := s.db.QueryContext(ctx, query, analysis.StartDate, analysis.EndDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var file FileAnalysis
		err := rows.Scan(
			&file.FilePath,
			&file.Language,
			&file.FirstEdit,
			&file.LastEdit,
			&file.ActiveDays,
			&file.AILines,
			&file.AISessions,
			&file.HumanLines,
			&file.HumanSessions,
			&file.TotalLines,
			&file.AIPercentage,
			&file.AIEfficiency,
			&file.HumanEfficiency,
			&file.NetContribution,
		)
		if err != nil {
			return err
		}
		
		analysis.FileBreakdown = append(analysis.FileBreakdown, file)
	}

	return rows.Err()
}

// getLanguageStats は言語別統計を取得
func (s *DuckDBStorage) getLanguageStats(ctx context.Context, analysis *PeriodAnalysis) error {
	query := `
	SELECT 
		fc.language,
		COALESCE(SUM(CASE WHEN t.author_type = 'ai' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as ai_lines,
		COALESCE(SUM(CASE WHEN t.author_type = 'human' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as human_lines,
		COALESCE(SUM(fc.lines_added - fc.lines_deleted), 0) as total_lines,
		COUNT(DISTINCT fc.file_path) as file_count,
		COUNT(DISTINCT t.session_id) as session_count
	FROM tracks t
	JOIN file_changes fc ON t.id = fc.track_id
	WHERE t.timestamp >= ? AND t.timestamp <= ?
	GROUP BY fc.language
	HAVING total_lines > 0
	ORDER BY total_lines DESC
	`

	rows, err := s.db.QueryContext(ctx, query, analysis.StartDate, analysis.EndDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var lang LanguageAnalysis
		err := rows.Scan(
			&lang.Language,
			&lang.AILines,
			&lang.HumanLines,
			&lang.TotalLines,
			&lang.FileCount,
			&lang.SessionCount,
		)
		if err != nil {
			return err
		}

		if lang.TotalLines > 0 {
			lang.AIPercentage = float64(lang.AILines) / float64(lang.TotalLines) * 100
		}

		analysis.LanguageStats = append(analysis.LanguageStats, lang)
	}

	return rows.Err()
}

// getDailyTimeline は日別タイムラインを取得
func (s *DuckDBStorage) getDailyTimeline(ctx context.Context, analysis *PeriodAnalysis) error {
	query := `
	SELECT 
		CAST(t.timestamp AS DATE) as date,
		COALESCE(SUM(CASE WHEN t.author_type = 'ai' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as ai_lines,
		COALESCE(SUM(CASE WHEN t.author_type = 'human' THEN fc.lines_added - fc.lines_deleted ELSE 0 END), 0) as human_lines,
		COALESCE(SUM(fc.lines_added - fc.lines_deleted), 0) as total_lines,
		COUNT(DISTINCT fc.file_path) as files_edited,
		COUNT(DISTINCT t.session_id) as sessions,
		COUNT(DISTINCT EXTRACT(hour FROM t.timestamp)) as active_hours
	FROM tracks t
	JOIN file_changes fc ON t.id = fc.track_id
	WHERE t.timestamp >= ? AND t.timestamp <= ?
	GROUP BY CAST(t.timestamp AS DATE)
	ORDER BY date
	`

	rows, err := s.db.QueryContext(ctx, query, analysis.StartDate, analysis.EndDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var point DailyPoint
		err := rows.Scan(
			&point.Date,
			&point.AILines,
			&point.HumanLines,
			&point.TotalLines,
			&point.FilesEdited,
			&point.Sessions,
			&point.ActiveHours,
		)
		if err != nil {
			return err
		}

		analysis.DailyTimeline = append(analysis.DailyTimeline, point)
	}

	return rows.Err()
}

// getContributorStats は貢献者統計を取得
func (s *DuckDBStorage) getContributorStats(ctx context.Context, analysis *PeriodAnalysis) error {
	query := `
	WITH contributor_data AS (
		SELECT 
			t.author,
			t.author_type,
			COALESCE(SUM(fc.lines_added - fc.lines_deleted), 0) as lines,
			COUNT(DISTINCT t.session_id) as sessions,
			COUNT(DISTINCT fc.file_path) as files_edited,
			COUNT(DISTINCT CAST(t.timestamp AS DATE)) as active_days
		FROM tracks t
		JOIN file_changes fc ON t.id = fc.track_id
		WHERE t.timestamp >= ? AND t.timestamp <= ?
		GROUP BY t.author, t.author_type
		HAVING lines > 0
	),
	total_lines AS (
		SELECT SUM(lines) as total FROM contributor_data
	)
	SELECT 
		cd.author,
		cd.author_type,
		cd.lines,
		cd.lines::FLOAT / tl.total * 100 as percentage,
		cd.sessions,
		cd.files_edited,
		CASE WHEN cd.active_days > 0 THEN cd.lines::FLOAT / cd.active_days ELSE 0 END as avg_lines_per_day
	FROM contributor_data cd
	CROSS JOIN total_lines tl
	ORDER BY cd.lines DESC
	`

	rows, err := s.db.QueryContext(ctx, query, analysis.StartDate, analysis.EndDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var stat ContributorStat
		err := rows.Scan(
			&stat.Author,
			&stat.AuthorType,
			&stat.Lines,
			&stat.Percentage,
			&stat.Sessions,
			&stat.FilesEdited,
			&stat.AvgLinesPerDay,
		)
		if err != nil {
			return err
		}

		analysis.ContributorStats = append(analysis.ContributorStats, stat)
	}

	return rows.Err()
}

// ParsePeriodExpression は自然言語風の期間表現を解析
func ParsePeriodExpression(expr string) (start, end time.Time, err error) {
	expr = strings.ToLower(strings.TrimSpace(expr))
	now := time.Now()
	
	switch {
	// 四半期指定
	case strings.Contains(expr, "q1"):
		year := extractYear(expr, now.Year())
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 3, 31, 23, 59, 59, 999999999, time.Local), nil
			   
	case strings.Contains(expr, "q2"):
		year := extractYear(expr, now.Year())
		return time.Date(year, 4, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 6, 30, 23, 59, 59, 999999999, time.Local), nil
			   
	case strings.Contains(expr, "q3"):
		year := extractYear(expr, now.Year())
		return time.Date(year, 7, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 9, 30, 23, 59, 59, 999999999, time.Local), nil
			   
	case strings.Contains(expr, "q4"):
		year := extractYear(expr, now.Year())
		return time.Date(year, 10, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 12, 31, 23, 59, 59, 999999999, time.Local), nil
	
	// 月名での指定 (Jan-Mar形式)
	case strings.Contains(expr, "jan") && strings.Contains(expr, "mar"):
		year := extractYear(expr, now.Year())
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 3, 31, 23, 59, 59, 999999999, time.Local), nil
			   
	case strings.Contains(expr, "apr") && strings.Contains(expr, "jun"):
		year := extractYear(expr, now.Year())
		return time.Date(year, 4, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 6, 30, 23, 59, 59, 999999999, time.Local), nil
	
	// 相対的な期間指定
	case strings.Contains(expr, "last") && strings.Contains(expr, "month"):
		months := extractNumber(expr, 1)
		start = now.AddDate(0, -months, 0)
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.Local)
		end = now
		return start, end, nil
		
	case strings.Contains(expr, "this year"):
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local),
			   time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, time.Local), nil
			   
	case strings.Contains(expr, "last year"):
		year := now.Year() - 1
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.Local),
			   time.Date(year, 12, 31, 23, 59, 59, 999999999, time.Local), nil
	
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unsupported period expression: %s", expr)
	}
}

// extractYear は文字列から年を抽出
func extractYear(expr string, defaultYear int) int {
	// "2024", "24" などを探す
	for _, part := range strings.Fields(expr) {
		if len(part) == 4 && part >= "2000" && part <= "2099" {
			var year int
			if _, err := fmt.Sscanf(part, "%d", &year); err == nil {
				return year
			}
		}
	}
	return defaultYear
}

// extractNumber は文字列から数値を抽出
func extractNumber(expr string, defaultNum int) int {
	words := strings.Fields(expr)
	for _, word := range words {
		var num int
		if _, err := fmt.Sscanf(word, "%d", &num); err == nil {
			return num
		}
		// 英語の数詞も対応
		switch word {
		case "one", "1": return 1
		case "two", "2": return 2
		case "three", "3": return 3
		case "six", "6": return 6
		case "twelve", "12": return 12
		}
	}
	return defaultNum
}