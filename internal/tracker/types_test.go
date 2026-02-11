package tracker

import (
	"testing"
	"time"
)

func TestCheckpointRecordBranchCompatibility(t *testing.T) {
	tests := []struct {
		name           string
		record         CheckpointRecord
		expectedBranch string
		hasBranchInfo  bool
	}{
		{
			name: "Record with branch info",
			record: CheckpointRecord{
				Timestamp: time.Now(),
				Author:    "ai-assistant",
				Branch:    "feature/new-ui",
				Added:     50,
				Deleted:   10,
			},
			expectedBranch: "feature/new-ui",
			hasBranchInfo:  true,
		},
		{
			name: "Record without branch info (backward compatibility)",
			record: CheckpointRecord{
				Timestamp: time.Now(),
				Author:    "human",
				Branch:    "", // Empty branch
				Added:     30,
				Deleted:   5,
			},
			expectedBranch: "main",
			hasBranchInfo:  false,
		},
		{
			name: "Legacy record (no branch field set)",
			record: CheckpointRecord{
				Timestamp: time.Now(),
				Author:    "claude",
				// Branch field omitted (zero value)
				Added:   25,
				Deleted: 0,
			},
			expectedBranch: "main",
			hasBranchInfo:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.GetBranch(); got != tt.expectedBranch {
				t.Errorf("GetBranch() = %v, expected %v", got, tt.expectedBranch)
			}

			if got := tt.record.HasBranchInfo(); got != tt.hasBranchInfo {
				t.Errorf("HasBranchInfo() = %v, expected %v", got, tt.hasBranchInfo)
			}
		})
	}
}

func TestGetDisplayBranch(t *testing.T) {
	tests := []struct {
		name     string
		record   CheckpointRecord
		expected string
	}{
		{
			name:     "explicit branch",
			record:   CheckpointRecord{Branch: "feature/new-ui"},
			expected: "feature/new-ui",
		},
		{
			name:     "inferred main (empty branch)",
			record:   CheckpointRecord{},
			expected: "main (inferred)",
		},
		{
			name:     "explicit main",
			record:   CheckpointRecord{Branch: "main"},
			expected: "main",
		},
		{
			name:     "explicit develop",
			record:   CheckpointRecord{Branch: "develop"},
			expected: "develop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.GetDisplayBranch(); got != tt.expected {
				t.Errorf("GetDisplayBranch() = %v, want %v", got, tt.expected)
			}
		})
	}
}
