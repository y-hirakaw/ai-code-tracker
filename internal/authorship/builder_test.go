package authorship

import (
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestBuildAuthorshipLog(t *testing.T) {
	checkpoints := []*tracker.CheckpointV2{
		{
			Timestamp: time.Now(),
			Author:    "Alice",
			Type:      tracker.AuthorTypeHuman,
			Changes: map[string]tracker.Change{
				"main.go": {
					Added:   10,
					Deleted: 2,
					Lines:   [][]int{{1, 10}},
				},
			},
		},
		{
			Timestamp: time.Now(),
			Author:    "Claude Code",
			Type:      tracker.AuthorTypeAI,
			Metadata:  map[string]string{"model": "claude-sonnet-4"},
			Changes: map[string]tracker.Change{
				"main.go": {
					Added:   50,
					Deleted: 5,
					Lines:   [][]int{{21, 70}},
				},
			},
		},
	}

	log, err := BuildAuthorshipLog(checkpoints, "abc123")
	if err != nil {
		t.Fatalf("BuildAuthorshipLog failed: %v", err)
	}

	if log.Version != AuthorshipLogVersion {
		t.Errorf("Expected version %s, got %s", AuthorshipLogVersion, log.Version)
	}

	if log.Commit != "abc123" {
		t.Errorf("Expected commit abc123, got %s", log.Commit)
	}

	fileInfo, exists := log.Files["main.go"]
	if !exists {
		t.Fatal("main.go not found in files")
	}

	if len(fileInfo.Authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(fileInfo.Authors))
	}

	// Check Alice's contribution
	found := false
	for _, author := range fileInfo.Authors {
		if author.Name == "Alice" && author.Type == tracker.AuthorTypeHuman {
			found = true
			if len(author.Lines) != 1 || len(author.Lines[0]) != 2 || author.Lines[0][0] != 1 || author.Lines[0][1] != 10 {
				t.Errorf("Alice's line ranges incorrect: %v", author.Lines)
			}
		}
	}
	if !found {
		t.Error("Alice not found in authors")
	}

	// Check Claude Code's contribution
	found = false
	for _, author := range fileInfo.Authors {
		if author.Name == "Claude Code" && author.Type == tracker.AuthorTypeAI {
			found = true
			if len(author.Lines) != 1 || len(author.Lines[0]) != 2 || author.Lines[0][0] != 21 || author.Lines[0][1] != 70 {
				t.Errorf("Claude Code's line ranges incorrect: %v", author.Lines)
			}
			if author.Metadata["model"] != "claude-sonnet-4" {
				t.Errorf("Expected model claude-sonnet-4, got %s", author.Metadata["model"])
			}
		}
	}
	if !found {
		t.Error("Claude Code not found in authors")
	}
}

func TestBuildAuthorshipLogMultipleFiles(t *testing.T) {
	checkpoints := []*tracker.CheckpointV2{
		{
			Timestamp: time.Now(),
			Author:    "Bob",
			Type:      tracker.AuthorTypeHuman,
			Changes: map[string]tracker.Change{
				"main.go":  {Added: 10, Lines: [][]int{{1, 10}}},
				"utils.go": {Added: 5, Lines: [][]int{{1, 5}}},
			},
		},
	}

	log, err := BuildAuthorshipLog(checkpoints, "def456")
	if err != nil {
		t.Fatalf("BuildAuthorshipLog failed: %v", err)
	}

	if len(log.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(log.Files))
	}

	if _, exists := log.Files["main.go"]; !exists {
		t.Error("main.go not found")
	}

	if _, exists := log.Files["utils.go"]; !exists {
		t.Error("utils.go not found")
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name     string
		ranges   [][]int
		expected int
	}{
		{
			name:     "Single line",
			ranges:   [][]int{{10}},
			expected: 1,
		},
		{
			name:     "Range",
			ranges:   [][]int{{1, 10}},
			expected: 10,
		},
		{
			name:     "Multiple ranges",
			ranges:   [][]int{{1, 10}, {20, 30}},
			expected: 21,
		},
		{
			name:     "Mixed single and range",
			ranges:   [][]int{{1, 10}, {15}, {20, 25}},
			expected: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountLines(tt.ranges)
			if result != tt.expected {
				t.Errorf("Expected %d lines, got %d", tt.expected, result)
			}
		})
	}
}
