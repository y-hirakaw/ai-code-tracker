package tracker

import (
	"strings"
	"testing"
	"time"
)

func TestNewAnalyzer(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".js"},
		ExcludePatterns:    []string{"test", "vendor"},
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	if analyzer.config != config {
		t.Error("Expected analyzer config to match provided config")
	}
}

func TestIsAIAuthor(t *testing.T) {
	config := &Config{
		AuthorMappings: map[string]string{
			"GPT Assistant": "ai",
			"Human Dev":     "human",
		},
	}

	analyzer := NewAnalyzer(config)

	tests := []struct {
		author   string
		expected bool
	}{
		{"claude", true},
		{"Claude AI", true},
		{"AI Assistant", true},
		{"Bot User", true},
		{"human-developer", false},
		{"John Doe", false},
		{"GPT Assistant", true}, // Mapped to "ai"
		{"Human Dev", false},    // Mapped to "human"
	}

	for _, test := range tests {
		result := analyzer.IsAIAuthor(test.author)
		if result != test.expected {
			t.Errorf("IsAIAuthor(%s) = %v, expected %v", test.author, result, test.expected)
		}
	}
}

func TestShouldTrackFile(t *testing.T) {
	config := &Config{
		TrackedExtensions: []string{".go", ".js", ".py"},
		ExcludePatterns:   []string{"test", "vendor", "_generated"},
	}

	analyzer := NewAnalyzer(config)

	tests := []struct {
		filepath string
		expected bool
	}{
		{"main.go", true},
		{"src/app.js", true},
		{"lib/helper.py", true},
		{"test/main_test.go", false},  // Contains "test"
		{"vendor/lib/code.go", false}, // Contains "vendor"
		{"code_generated.go", false},  // Contains "_generated"
		{"README.md", false},          // Wrong extension
		{"config.json", false},        // Wrong extension
		{"src/valid.go", true},
		{"src/test_file.go", false}, // Contains "test"
	}

	for _, test := range tests {
		result := analyzer.shouldTrackFile(test.filepath)
		if result != test.expected {
			t.Errorf("shouldTrackFile(%s) = %v, expected %v", test.filepath, result, test.expected)
		}
	}
}

func TestCompareFiles(t *testing.T) {
	config := &Config{}
	analyzer := NewAnalyzer(config)

	before := FileContent{
		Path:  "test.go",
		Lines: []string{"line1", "line2", "line3"},
	}

	after := FileContent{
		Path:  "test.go",
		Lines: []string{"line1", "line2", "line3", "line4", "line5"},
	}

	// Test AI author
	stats := analyzer.compareFiles(before, after, true)
	if stats.AILines != 2 {
		t.Errorf("Expected 2 AI lines added, got %d", stats.AILines)
	}
	if stats.HumanLines != 0 {
		t.Errorf("Expected 0 human lines added, got %d", stats.HumanLines)
	}

	// Test human author
	stats = analyzer.compareFiles(before, after, false)
	if stats.AILines != 0 {
		t.Errorf("Expected 0 AI lines added, got %d", stats.AILines)
	}
	if stats.HumanLines != 2 {
		t.Errorf("Expected 2 human lines added, got %d", stats.HumanLines)
	}

	// Test file with fewer lines (no additions counted)
	smaller := FileContent{
		Path:  "test.go",
		Lines: []string{"line1"},
	}
	stats = analyzer.compareFiles(before, smaller, true)
	if stats.AILines != 0 {
		t.Errorf("Expected 0 AI lines when file shrinks, got %d", stats.AILines)
	}
}

func TestAnalyzeCheckpoints(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go"},
		ExcludePatterns:    []string{},
		AuthorMappings:     make(map[string]string),
	}

	analyzer := NewAnalyzer(config)

	// Create before checkpoint
	before := &Checkpoint{
		ID:        "before",
		Timestamp: time.Now().Add(-1 * time.Hour),
		Author:    "human",
		Files: map[string]FileContent{
			"main.go": {
				Path:  "main.go",
				Lines: []string{"package main", "func main() {}"},
			},
		},
		NumstatData: map[string][2]int{
			"main.go": {2, 0},
		},
	}

	// Create after checkpoint with AI additions
	after := &Checkpoint{
		ID:        "after",
		Timestamp: time.Now(),
		Author:    "claude",
		Files: map[string]FileContent{
			"main.go": {
				Path:  "main.go",
				Lines: []string{"package main", "import \"fmt\"", "func main() {", "fmt.Println(\"Hello\")", "}"},
			},
			"helper.go": {
				Path:  "helper.go",
				Lines: []string{"package main", "func helper() {}"},
			},
		},
		NumstatData: map[string][2]int{
			"main.go":   {5, 2}, // 3 lines added (5-2)
			"helper.go": {2, 0}, // 2 lines added (new file)
		},
	}

	result, err := analyzer.AnalyzeCheckpoints(before, after)
	if err != nil {
		t.Fatalf("Failed to analyze checkpoints: %v", err)
	}

	// With numstat data: 5 total AI lines from main.go + 2 AI lines from helper.go = 7 AI lines
	// Note: The analyzer counts the total added lines from numstat, not the difference
	if result.AILines != 7 {
		t.Errorf("Expected 7 AI lines, got %d", result.AILines)
	}

	if result.HumanLines != 0 {
		t.Errorf("Expected 0 human lines, got %d", result.HumanLines)
	}

	// Total lines should be sum of all lines in after checkpoint
	expectedTotal := 5 + 2 // main.go + helper.go
	if result.TotalLines != expectedTotal {
		t.Errorf("Expected %d total lines, got %d", expectedTotal, result.TotalLines)
	}
}

func TestGenerateReport(t *testing.T) {
	config := &Config{
		TargetAIPercentage: 80.0,
	}

	analyzer := NewAnalyzer(config)

	result := &AnalysisResult{
		TotalLines:  1000,
		AILines:     600,
		HumanLines:  400,
		Percentage:  60.0,
		LastUpdated: time.Now(),
	}

	report := analyzer.GenerateReport(result)

	// Check that report contains expected information
	if !strings.Contains(report, "AI Code Tracking Report") {
		t.Error("Report should contain title")
	}

	if !strings.Contains(report, "Added Lines: 1000") {
		t.Error("Report should contain total added lines")
	}

	if !strings.Contains(report, "AI Lines: 600") {
		t.Error("Report should contain AI lines count")
	}

	if !strings.Contains(report, "Human Lines: 400") {
		t.Error("Report should contain human lines count")
	}

	if !strings.Contains(report, "Target: 80.0% AI code") {
		t.Error("Report should contain target percentage")
	}

	if !strings.Contains(report, "Progress: 75.0%") {
		t.Error("Report should contain progress percentage")
	}
}

func TestAnalyzeFromGitDiff(t *testing.T) {
	config := &Config{}
	analyzer := NewAnalyzer(config)

	currentMetrics := &AnalysisResult{
		TotalLines:  100,
		AILines:     50,
		HumanLines:  50,
		LastUpdated: time.Now(),
	}

	// Simulate git diff output
	gitDiff := `commit abc123
Author: Claude AI
Date: Mon Jan 1 12:00:00 2024

    Add new feature

diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,6 @@
 package main
 
 func main() {
+    // New line 1
+    // New line 2
+    // New line 3
 }`

	result, err := analyzer.AnalyzeFromGitDiff(gitDiff, currentMetrics)
	if err != nil {
		t.Fatalf("Failed to analyze git diff: %v", err)
	}

	// Should have added 3 AI lines
	if result.AILines != 53 {
		t.Errorf("Expected 53 AI lines (50 + 3), got %d", result.AILines)
	}

	if result.HumanLines != 50 {
		t.Errorf("Expected 50 human lines (unchanged), got %d", result.HumanLines)
	}

	if result.TotalLines != 103 {
		t.Errorf("Expected 103 total lines, got %d", result.TotalLines)
	}
}

func TestGetFileStats(t *testing.T) {
	config := &Config{}
	analyzer := NewAnalyzer(config)

	checkpoint := &Checkpoint{
		Author: "claude",
		Files: map[string]FileContent{
			"file1.go": {
				Path:  "file1.go",
				Lines: []string{"line1", "line2", "line3"},
			},
			"file2.js": {
				Path:  "file2.js",
				Lines: []string{"line1", "line2"},
			},
		},
	}

	stats := analyzer.GetFileStats(checkpoint)

	if len(stats) != 2 {
		t.Errorf("Expected 2 file stats, got %d", len(stats))
	}

	// Find file1.go stats
	var file1Stats *FileStats
	for i := range stats {
		if stats[i].Path == "file1.go" {
			file1Stats = &stats[i]
			break
		}
	}

	if file1Stats == nil {
		t.Fatal("Could not find stats for file1.go")
	}

	if file1Stats.TotalLines != 3 {
		t.Errorf("Expected 3 total lines for file1.go, got %d", file1Stats.TotalLines)
	}

	if file1Stats.AILines != 3 {
		t.Errorf("Expected 3 AI lines for file1.go (AI author), got %d", file1Stats.AILines)
	}

	if file1Stats.HumanLines != 0 {
		t.Errorf("Expected 0 human lines for file1.go (AI author), got %d", file1Stats.HumanLines)
	}
}
