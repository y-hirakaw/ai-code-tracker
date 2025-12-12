package authorship

import (
	"encoding/json"
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

	log, err := BuildAuthorshipLog(checkpoints, "abc123", nil)
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

	log, err := BuildAuthorshipLog(checkpoints, "def456", nil)
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

func TestToJSON(t *testing.T) {
	log := &tracker.AuthorshipLog{
		Version:   AuthorshipLogVersion,
		Commit:    "abc123",
		Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Files: map[string]tracker.FileInfo{
			"test.go": {
				Authors: []tracker.AuthorInfo{
					{
						Name:  "Test User",
						Type:  tracker.AuthorTypeHuman,
						Lines: [][]int{{1, 10}},
					},
				},
			},
		},
	}

	data, err := ToJSON(log)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON data")
	}

	// Verify it's valid JSON by parsing it back
	var parsed tracker.AuthorshipLog
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Generated JSON is not valid: %v", err)
	}

	if parsed.Version != log.Version {
		t.Errorf("Version = %s, want %s", parsed.Version, log.Version)
	}

	if parsed.Commit != log.Commit {
		t.Errorf("Commit = %s, want %s", parsed.Commit, log.Commit)
	}
}

func TestFromJSON(t *testing.T) {
	jsonData := []byte(`{
		"version": "1.0",
		"commit": "abc123",
		"timestamp": "2024-01-01T00:00:00Z",
		"files": {
			"test.go": {
				"authors": [
					{
						"name": "Test User",
						"type": "human",
						"lines": [[1, 10]],
						"metadata": {}
					}
				]
			}
		}
	}`)

	log, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if log.Version != "1.0" {
		t.Errorf("Version = %s, want 1.0", log.Version)
	}

	if log.Commit != "abc123" {
		t.Errorf("Commit = %s, want abc123", log.Commit)
	}

	if len(log.Files) != 1 {
		t.Errorf("Files count = %d, want 1", len(log.Files))
	}

	fileInfo, exists := log.Files["test.go"]
	if !exists {
		t.Fatal("test.go not found in files")
	}

	if len(fileInfo.Authors) != 1 {
		t.Errorf("Authors count = %d, want 1", len(fileInfo.Authors))
	}

	if fileInfo.Authors[0].Name != "Test User" {
		t.Errorf("Author name = %s, want Test User", fileInfo.Authors[0].Name)
	}
}

func TestFromJSONInvalidJSON(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)

	_, err := FromJSON(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestValidateAuthorshipLog(t *testing.T) {
	tests := []struct {
		name    string
		log     *tracker.AuthorshipLog
		wantErr bool
	}{
		{
			name: "Valid log",
			log: &tracker.AuthorshipLog{
				Version: AuthorshipLogVersion,
				Commit:  "abc123",
				Files: map[string]tracker.FileInfo{
					"test.go": {
						Authors: []tracker.AuthorInfo{
							{
								Name: "Test User",
								Type: tracker.AuthorTypeHuman,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing version",
			log: &tracker.AuthorshipLog{
				Commit: "abc123",
				Files:  map[string]tracker.FileInfo{},
			},
			wantErr: true,
		},
		{
			name: "Missing commit",
			log: &tracker.AuthorshipLog{
				Version: AuthorshipLogVersion,
				Files:   map[string]tracker.FileInfo{},
			},
			wantErr: true,
		},
		{
			name: "Wrong version",
			log: &tracker.AuthorshipLog{
				Version: "999.0",
				Commit:  "abc123",
				Files:   map[string]tracker.FileInfo{},
			},
			wantErr: true,
		},
		{
			name: "File with no authors",
			log: &tracker.AuthorshipLog{
				Version: AuthorshipLogVersion,
				Commit:  "abc123",
				Files: map[string]tracker.FileInfo{
					"test.go": {
						Authors: []tracker.AuthorInfo{},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Author with empty name",
			log: &tracker.AuthorshipLog{
				Version: AuthorshipLogVersion,
				Commit:  "abc123",
				Files: map[string]tracker.FileInfo{
					"test.go": {
						Authors: []tracker.AuthorInfo{
							{
								Name: "",
								Type: tracker.AuthorTypeHuman,
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid author type",
			log: &tracker.AuthorshipLog{
				Version: AuthorshipLogVersion,
				Commit:  "abc123",
				Files: map[string]tracker.FileInfo{
					"test.go": {
						Authors: []tracker.AuthorInfo{
							{
								Name: "Test User",
								Type: "invalid",
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthorshipLog(tt.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthorshipLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
