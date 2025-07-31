package period

import (
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestFilterRecords(t *testing.T) {
	now := time.Now()
	
	records := []tracker.CheckpointRecord{
		{Timestamp: now.Add(-2 * time.Hour), Author: "human", Added: 10},
		{Timestamp: now.Add(-1 * time.Hour), Author: "claude", Added: 15},
		{Timestamp: now.Add(-30 * time.Minute), Author: "human", Added: 5},
	}
	
	timeRange := &TimeRange{
		From: now.Add(-90 * time.Minute),
		To:   now,
	}
	
	filtered := FilterRecords(records, timeRange)
	
	// Should include records from last 90 minutes (2 records)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered records, got %d", len(filtered))
	}
	
	// First record should be the one from 1 hour ago
	if filtered[0].Author != "claude" {
		t.Errorf("Expected first filtered record to be from claude, got %s", filtered[0].Author)
	}
}

func TestFilterRecordsInclusive(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	
	records := []tracker.CheckpointRecord{
		{Timestamp: baseTime, Author: "human", Added: 10},
		{Timestamp: baseTime.Add(1 * time.Hour), Author: "claude", Added: 15},
		{Timestamp: baseTime.Add(2 * time.Hour), Author: "human", Added: 5},
	}
	
	timeRange := &TimeRange{
		From: baseTime,
		To:   baseTime.Add(2 * time.Hour),
	}
	
	filtered := FilterRecordsInclusive(records, timeRange)
	
	// Should include all 3 records (inclusive boundaries)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 filtered records, got %d", len(filtered))
	}
}

func TestFilterRecordsNilRange(t *testing.T) {
	records := []tracker.CheckpointRecord{
		{Timestamp: time.Now(), Author: "human", Added: 10},
	}
	
	filtered := FilterRecords(records, nil)
	
	// Should return all records when range is nil
	if len(filtered) != len(records) {
		t.Errorf("Expected %d records when range is nil, got %d", len(records), len(filtered))
	}
}