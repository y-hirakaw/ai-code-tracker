package period

import (
	"testing"
	"time"
)

func TestParseLastDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"7d", 7 * 24 * time.Hour, false},
		{"2w", 14 * 24 * time.Hour, false},
		{"1m", 30 * 24 * time.Hour, false},
		{"invalid", 0, true},
		{"7x", 0, true},
		{"", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			timeRange, err := ParseLastDuration(test.input)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", test.input)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
				return
			}
			
			actualDuration := timeRange.To.Sub(timeRange.From)
			if actualDuration != test.expected {
				t.Errorf("Expected duration %v, got %v", test.expected, actualDuration)
			}
		})
	}
}

func TestParseTimeRange(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
	}{
		{"2025-01-01", false},
		{"2025-01-01 12:00:00", false},
		{"2 days ago", false},
		{"1 week ago", false},
		{"invalid date", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			_, err := ParseTimeRange(test.input)
			
			if test.hasError && err == nil {
				t.Errorf("Expected error for input %s, but got none", test.input)
			}
			
			if !test.hasError && err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
			}
		})
	}
}

func TestParseFromTo(t *testing.T) {
	tests := []struct {
		from     string
		to       string
		hasError bool
	}{
		{"2025-01-01", "2025-01-02", false},
		{"2025-01-02", "2025-01-01", true}, // from after to
		{"invalid", "2025-01-01", true},
		{"2025-01-01", "invalid", true},
	}

	for _, test := range tests {
		t.Run(test.from+"_to_"+test.to, func(t *testing.T) {
			_, err := ParseFromTo(test.from, test.to)
			
			if test.hasError && err == nil {
				t.Errorf("Expected error for range %s to %s, but got none", test.from, test.to)
			}
			
			if !test.hasError && err != nil {
				t.Errorf("Unexpected error for range %s to %s: %v", test.from, test.to, err)
			}
		})
	}
}