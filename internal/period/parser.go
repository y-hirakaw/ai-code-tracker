package period

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseTimeRange parses various time range formats into TimeRange
func ParseTimeRange(input string) (*TimeRange, error) {
	now := time.Now()
	
	// Handle "since" format (relative)
	if matched, duration := parseRelativeTime(input); matched {
		return &TimeRange{
			From: now.Add(-duration),
			To:   now,
		}, nil
	}
	
	// Handle absolute date format (YYYY-MM-DD)
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return &TimeRange{
			From: t,
			To:   now,
		}, nil
	}
	
	// Handle datetime format (YYYY-MM-DD HH:MM:SS)
	if t, err := time.Parse("2006-01-02 15:04:05", input); err == nil {
		return &TimeRange{
			From: t,
			To:   now,
		}, nil
	}
	
	return nil, fmt.Errorf("unsupported time format: %s", input)
}

// ParseLastDuration parses "last Nx" format (e.g., "7d", "2w", "1m")
func ParseLastDuration(input string) (*TimeRange, error) {
	re := regexp.MustCompile(`^(\d+)([dwm])$`)
	matches := re.FindStringSubmatch(input)
	
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid duration format: %s (expected format: Nd/Nw/Nm)", input)
	}
	
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid number in duration: %s", matches[1])
	}
	
	unit := matches[2]
	now := time.Now()
	var duration time.Duration
	
	switch unit {
	case "d":
		duration = time.Duration(num) * 24 * time.Hour
	case "w":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "m":
		// Approximate month as 30 days
		duration = time.Duration(num) * 30 * 24 * time.Hour
	default:
		return nil, fmt.Errorf("invalid time unit: %s", unit)
	}
	
	return &TimeRange{
		From: now.Add(-duration),
		To:   now,
	}, nil
}

// ParseFromTo parses from and to dates for range queries
func ParseFromTo(from, to string) (*TimeRange, error) {
	fromTime, err := parseDate(from)
	if err != nil {
		return nil, fmt.Errorf("invalid from date: %w", err)
	}
	
	toTime, err := parseDate(to)
	if err != nil {
		return nil, fmt.Errorf("invalid to date: %w", err)
	}
	
	if fromTime.After(toTime) {
		return nil, fmt.Errorf("from date must be before to date")
	}
	
	return &TimeRange{
		From: fromTime,
		To:   toTime,
	}, nil
}

// parseRelativeTime parses relative time expressions like "2 weeks ago", "1 day ago"
func parseRelativeTime(input string) (bool, time.Duration) {
	input = strings.ToLower(strings.TrimSpace(input))
	
	// Pattern: "N (days|weeks|months) ago"
	re := regexp.MustCompile(`^(\d+)\s+(days?|weeks?|months?)\s+ago$`)
	matches := re.FindStringSubmatch(input)
	
	if len(matches) != 3 {
		return false, 0
	}
	
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return false, 0
	}
	
	unit := matches[2]
	var duration time.Duration
	
	switch {
	case strings.HasPrefix(unit, "day"):
		duration = time.Duration(num) * 24 * time.Hour
	case strings.HasPrefix(unit, "week"):
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case strings.HasPrefix(unit, "month"):
		duration = time.Duration(num) * 30 * 24 * time.Hour
	default:
		return false, 0
	}
	
	return true, duration
}

// parseDate parses various date formats
func parseDate(input string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"01/02/2006",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unsupported date format: %s", input)
}