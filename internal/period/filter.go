package period

import (
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// FilterRecords filters checkpoint records by time range
func FilterRecords(records []tracker.CheckpointRecord, timeRange *TimeRange) []tracker.CheckpointRecord {
	if timeRange == nil {
		return records
	}
	
	filtered := make([]tracker.CheckpointRecord, 0, len(records))
	
	for _, record := range records {
		if record.Timestamp.After(timeRange.From) && record.Timestamp.Before(timeRange.To) {
			filtered = append(filtered, record)
		}
	}
	
	return filtered
}

// FilterRecordsInclusive filters records including boundary timestamps
func FilterRecordsInclusive(records []tracker.CheckpointRecord, timeRange *TimeRange) []tracker.CheckpointRecord {
	if timeRange == nil {
		return records
	}
	
	filtered := make([]tracker.CheckpointRecord, 0, len(records))
	
	for _, record := range records {
		if (record.Timestamp.Equal(timeRange.From) || record.Timestamp.After(timeRange.From)) &&
		   (record.Timestamp.Equal(timeRange.To) || record.Timestamp.Before(timeRange.To)) {
			filtered = append(filtered, record)
		}
	}
	
	return filtered
}