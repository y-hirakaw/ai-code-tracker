package tracker

import (
	"testing"
	"time"
)

func TestCheckpointStructure(t *testing.T) {
	checkpoint := &Checkpoint{
		ID:          "test123",
		Timestamp:   time.Now(),
		Author:      "test-author",
		CommitHash:  "abc123",
		Files:       make(map[string]FileContent),
		NumstatData: make(map[string][2]int),
	}

	if checkpoint.ID != "test123" {
		t.Errorf("Expected ID to be 'test123', got '%s'", checkpoint.ID)
	}

	if checkpoint.Author != "test-author" {
		t.Errorf("Expected Author to be 'test-author', got '%s'", checkpoint.Author)
	}

	if checkpoint.CommitHash != "abc123" {
		t.Errorf("Expected CommitHash to be 'abc123', got '%s'", checkpoint.CommitHash)
	}
}

func TestCheckpointRecordStructure(t *testing.T) {
	record := &CheckpointRecord{
		Timestamp: time.Now(),
		Author:    "ai-assistant",
		Branch:    "feature/test",
		Commit:    "def456",
		Added:     100,
		Deleted:   20,
	}

	if record.Author != "ai-assistant" {
		t.Errorf("Expected Author to be 'ai-assistant', got '%s'", record.Author)
	}

	if record.Branch != "feature/test" {
		t.Errorf("Expected Branch to be 'feature/test', got '%s'", record.Branch)
	}

	if record.Added != 100 {
		t.Errorf("Expected Added to be 100, got %d", record.Added)
	}

	if record.Deleted != 20 {
		t.Errorf("Expected Deleted to be 20, got %d", record.Deleted)
	}
}

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
			// Test GetBranch()
			if got := tt.record.GetBranch(); got != tt.expectedBranch {
				t.Errorf("GetBranch() = %v, expected %v", got, tt.expectedBranch)
			}

			// Test HasBranchInfo()
			if got := tt.record.HasBranchInfo(); got != tt.hasBranchInfo {
				t.Errorf("HasBranchInfo() = %v, expected %v", got, tt.hasBranchInfo)
			}
		})
	}
}

func TestFileContentStructure(t *testing.T) {
	content := FileContent{
		Path:  "test/file.go",
		Lines: []string{"line1", "line2", "line3"},
	}

	if content.Path != "test/file.go" {
		t.Errorf("Expected Path to be 'test/file.go', got '%s'", content.Path)
	}

	if len(content.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(content.Lines))
	}
}

func TestAnalysisResultStructure(t *testing.T) {
	result := &AnalysisResult{
		TotalLines:  1000,
		AILines:     800,
		HumanLines:  200,
		Percentage:  80.0,
		LastUpdated: time.Now(),
	}

	if result.TotalLines != 1000 {
		t.Errorf("Expected TotalLines to be 1000, got %d", result.TotalLines)
	}

	if result.AILines != 800 {
		t.Errorf("Expected AILines to be 800, got %d", result.AILines)
	}

	if result.HumanLines != 200 {
		t.Errorf("Expected HumanLines to be 200, got %d", result.HumanLines)
	}

	if result.Percentage != 80.0 {
		t.Errorf("Expected Percentage to be 80.0, got %.1f", result.Percentage)
	}
}

func TestConfigStructure(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".js", ".py"},
		ExcludePatterns:    []string{"test", "vendor"},
		AuthorMappings: map[string]string{
			"Claude AI": "ai",
			"GPT-4":     "ai",
		},
	}

	if config.TargetAIPercentage != 80.0 {
		t.Errorf("Expected TargetAIPercentage to be 80.0, got %.1f", config.TargetAIPercentage)
	}

	if len(config.TrackedExtensions) != 3 {
		t.Errorf("Expected 3 tracked extensions, got %d", len(config.TrackedExtensions))
	}

	if len(config.ExcludePatterns) != 2 {
		t.Errorf("Expected 2 exclude patterns, got %d", len(config.ExcludePatterns))
	}

	if len(config.AuthorMappings) != 2 {
		t.Errorf("Expected 2 author mappings, got %d", len(config.AuthorMappings))
	}
}
