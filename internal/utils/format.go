package utils

import (
	"fmt"
	"time"
)

// FormatDuration は期間を人間が読みやすい形式でフォーマットする
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%d日間", days)
	} else if hours > 0 {
		return fmt.Sprintf("%d時間", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分", minutes)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}

// FormatFileSize はファイルサイズを人間が読みやすい形式でフォーマットする
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatMemoryUsage はメモリ使用量をフォーマットする
func FormatMemoryUsage(bytes uint64) string {
	return FormatFileSize(int64(bytes))
}

// FormatNumber は数値を3桁区切りでフォーマットする
func FormatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	
	str := fmt.Sprintf("%d", n)
	length := len(str)
	result := ""
	
	for i, char := range str {
		if i > 0 && (length-i)%3 == 0 {
			result += ","
		}
		result += string(char)
	}
	
	return result
}

// FormatEventCount はイベント数を適切にフォーマットする
func FormatEventCount(count int) string {
	return fmt.Sprintf("%s events", FormatNumber(count))
}

// FormatTableRow はテーブル行を指定された幅でフォーマットする
func FormatTableRow(columns []string, widths []int) string {
	if len(columns) != len(widths) {
		return ""
	}
	
	result := ""
	for i, col := range columns {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%-*s", widths[i], col)
	}
	
	return result
}

// FormatTimestamp はタイムスタンプを標準形式でフォーマットする
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate は日付を標準形式でフォーマットする
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTime は時刻を標準形式でフォーマットする
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// FormatISO8601 はISO8601形式でタイムスタンプをフォーマットする
func FormatISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}