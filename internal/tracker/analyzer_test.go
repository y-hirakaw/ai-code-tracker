package tracker

import (
	"strings"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
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
		name     string
		author   string
		expected bool
	}{
		{"claude detected", "claude", true},
		{"Claude AI detected", "Claude AI", true},
		{"AI Assistant detected", "AI Assistant", true},
		{"Bot User detected", "Bot User", true},
		{"human developer", "human-developer", false},
		{"human name", "John Doe", false},
		{"mapped to ai", "GPT Assistant", true},
		{"mapped to human", "Human Dev", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.IsAIAuthor(tt.author)
			if result != tt.expected {
				t.Errorf("IsAIAuthor(%s) = %v, expected %v", tt.author, result, tt.expected)
			}
		})
	}
}

func TestIsAIAgent(t *testing.T) {
	configuredAgents := []string{"My Custom Bot", "Internal AI"}
	authorMappings := map[string]string{
		"GPT Helper": "ai-assistant",
		"Dev Lead":   "human-manager",
	}

	tests := []struct {
		name     string
		author   string
		expected bool
	}{
		// DefaultAINames patterns
		{"claude lowercase", "claude", true},
		{"Claude Code", "Claude Code", true},
		{"copilot", "GitHub Copilot", true},
		{"chatgpt", "chatgpt-4o", true},
		{"bot in name", "my-bot", true},
		{"ai in name", "AI Helper", true},
		{"assistant", "Code Assistant", true},

		// Human authors
		{"human name", "John Doe", false},
		{"empty string", "", false},

		// configuredAgents exact match
		{"configured agent exact", "My Custom Bot", true},
		{"configured agent case mismatch", "my custom bot", true}, // "bot" matches DefaultAINames

		// authorMappings resolution
		{"mapping to ai-like name", "GPT Helper", true}, // resolved to "ai-assistant", contains "ai"
		{"mapping to human-like name", "Dev Lead", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAIAgent(tt.author, configuredAgents, authorMappings)
			if result != tt.expected {
				t.Errorf("IsAIAgent(%q) = %v, expected %v", tt.author, result, tt.expected)
			}
		})
	}
}

func TestIsAIAgentNilMappings(t *testing.T) {
	// authorMappings が nil でもパニックしないことを確認
	result := IsAIAgent("claude", nil, nil)
	if !result {
		t.Error("IsAIAgent('claude', nil, nil) should return true")
	}

	result = IsAIAgent("John Doe", nil, nil)
	if result {
		t.Error("IsAIAgent('John Doe', nil, nil) should return false")
	}
}

func TestShouldTrackFile(t *testing.T) {
	config := &Config{
		TrackedExtensions: []string{".go", ".js", ".py"},
		ExcludePatterns:   []string{"*_test.go", "vendor/*", "*_generated.go"},
	}

	analyzer := NewAnalyzer(config)

	tests := []struct {
		name     string
		filepath string
		expected bool
	}{
		{"go file tracked", "main.go", true},
		{"js file tracked", "src/app.js", true},
		{"py file tracked", "lib/helper.py", true},
		{"test file excluded", "main_test.go", false},
		{"nested test excluded", "pkg/handler_test.go", false},
		{"vendor excluded", "vendor/lib/code.go", false},
		{"generated excluded", "code_generated.go", false},
		{"wrong extension md", "README.md", false},
		{"wrong extension json", "config.json", false},
		{"nested go tracked", "src/valid.go", true},
		{"non-test file with test in path", "src/test_helper.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.shouldTrackFile(tt.filepath)
			if result != tt.expected {
				t.Errorf("shouldTrackFile(%s) = %v, expected %v", tt.filepath, result, tt.expected)
			}
		})
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

	// With numstat data: main.go diff = 5-2 = 3 AI lines, helper.go (new) = 2 AI lines → 5 total
	if result.AILines != 5 {
		t.Errorf("Expected 5 AI lines, got %d", result.AILines)
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

func TestAnalyzeFromNumstat_DetailedMetrics(t *testing.T) {
	config := &Config{
		TrackedExtensions: []string{".go"},
		ExcludePatterns:   []string{},
		AuthorMappings:    make(map[string]string),
	}
	analyzer := NewAnalyzer(config)

	before := &Checkpoint{
		ID:        "before",
		Timestamp: time.Now().Add(-1 * time.Hour),
		Author:    "human",
		NumstatData: map[string][2]int{
			"existing.go": {10, 0},
		},
		Files: map[string]FileContent{},
	}

	after := &Checkpoint{
		ID:        "after",
		Timestamp: time.Now(),
		Author:    "claude",
		NumstatData: map[string][2]int{
			"existing.go": {15, 3}, // 5 lines added diff, 15 added total, 3 deleted
			"newfile.go":  {8, 0},  // new file, 8 lines added
		},
		Files: map[string]FileContent{
			"existing.go": {Path: "existing.go", Lines: []string{"a", "b", "c"}},
			"newfile.go":  {Path: "newfile.go", Lines: []string{"x", "y"}},
		},
	}

	result, err := analyzer.AnalyzeCheckpoints(before, after)
	if err != nil {
		t.Fatalf("AnalyzeCheckpoints failed: %v", err)
	}

	// AILines = diff of added: (15-10) + 8 = 13
	if result.AILines != 13 {
		t.Errorf("AILines: got %d, want 13", result.AILines)
	}

	// WorkVolume: AI added = 15 + 8 = 23, AI deleted = 3 + 0 = 3
	if result.Metrics.WorkVolume.AIAdded != 23 {
		t.Errorf("WorkVolume.AIAdded: got %d, want 23", result.Metrics.WorkVolume.AIAdded)
	}
	if result.Metrics.WorkVolume.AIDeleted != 3 {
		t.Errorf("WorkVolume.AIDeleted: got %d, want 3", result.Metrics.WorkVolume.AIDeleted)
	}
	if result.Metrics.WorkVolume.AIChanges != 26 {
		t.Errorf("WorkVolume.AIChanges: got %d, want 26", result.Metrics.WorkVolume.AIChanges)
	}

	// Contributions: AI additions = 5 + 8 = 13
	if result.Metrics.Contributions.AIAdditions != 13 {
		t.Errorf("Contributions.AIAdditions: got %d, want 13", result.Metrics.Contributions.AIAdditions)
	}

	// NewFiles: only newfile.go is new (8 lines)
	if result.Metrics.NewFiles.AINewLines != 8 {
		t.Errorf("NewFiles.AINewLines: got %d, want 8", result.Metrics.NewFiles.AINewLines)
	}

	// Human metrics should all be zero
	if result.Metrics.WorkVolume.HumanAdded != 0 {
		t.Errorf("WorkVolume.HumanAdded: got %d, want 0", result.Metrics.WorkVolume.HumanAdded)
	}
	if result.Metrics.NewFiles.HumanNewLines != 0 {
		t.Errorf("NewFiles.HumanNewLines: got %d, want 0", result.Metrics.NewFiles.HumanNewLines)
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


func TestNewAnalyzerWithExecutor(t *testing.T) {
	config := &Config{TargetAIPercentage: 80}
	mock := &gitexec.MockExecutor{}
	analyzer := NewAnalyzerWithExecutor(config, mock)
	if analyzer.config != config {
		t.Error("config mismatch: analyzer.config does not match provided config")
	}
	if analyzer.executor != mock {
		t.Error("executor mismatch: analyzer.executor does not match provided mock")
	}
}

func TestAnalyzeFromCommits(t *testing.T) {
	config := &Config{
		TrackedExtensions: []string{".go"},
		ExcludePatterns:   []string{},
		AuthorMappings:    make(map[string]string),
	}
	mock := gitexec.NewMockExecutor()
	mock.RunFunc = func(args ...string) (string, error) {
		// Expected call: "diff", "--numstat", "abc123", "def456"
		if len(args) >= 1 && args[0] == "diff" {
			return "10\t2\tmain.go\n5\t1\thelper.go", nil
		}
		return "", nil
	}
	analyzer := NewAnalyzerWithExecutor(config, mock)

	before := &Checkpoint{
		CommitHash: "abc123",
		Timestamp:  time.Now().Add(-1 * time.Hour),
		Author:     "human",
		Files:      map[string]FileContent{},
		// NumstatData is nil -> triggers analyzeFromCommits
	}
	after := &Checkpoint{
		CommitHash: "def456",
		Timestamp:  time.Now(),
		Author:     "claude",
		Files: map[string]FileContent{
			"main.go":   {Path: "main.go", Lines: []string{"a", "b", "c", "d", "e"}},
			"helper.go": {Path: "helper.go", Lines: []string{"x", "y", "z"}},
		},
		// NumstatData is nil -> triggers analyzeFromCommits
	}

	result, err := analyzer.AnalyzeCheckpoints(before, after)
	if err != nil {
		t.Fatalf("AnalyzeCheckpoints failed: %v", err)
	}

	// AI lines = added lines from numstat: 10 (main.go) + 5 (helper.go) = 15
	if result.AILines != 15 {
		t.Errorf("AILines: got %d, want 15", result.AILines)
	}

	if result.HumanLines != 0 {
		t.Errorf("HumanLines: got %d, want 0", result.HumanLines)
	}

	// TotalLines = sum of lines in after.Files: 5 (main.go) + 3 (helper.go) = 8
	if result.TotalLines != 8 {
		t.Errorf("TotalLines: got %d, want 8", result.TotalLines)
	}

	// Verify percentage: 15/(15+0) * 100 = 100.0
	if result.Percentage != 100.0 {
		t.Errorf("Percentage: got %.1f, want 100.0", result.Percentage)
	}

	// Verify detailed metrics: WorkVolume
	if result.Metrics.WorkVolume.AIAdded != 15 {
		t.Errorf("WorkVolume.AIAdded: got %d, want 15", result.Metrics.WorkVolume.AIAdded)
	}
	if result.Metrics.WorkVolume.AIDeleted != 3 {
		t.Errorf("WorkVolume.AIDeleted: got %d, want 3 (2+1)", result.Metrics.WorkVolume.AIDeleted)
	}
	if result.Metrics.Contributions.AIAdditions != 15 {
		t.Errorf("Contributions.AIAdditions: got %d, want 15", result.Metrics.Contributions.AIAdditions)
	}

	// Verify mock was called with correct args
	calls := mock.GetCalls("Run")
	if len(calls) == 0 {
		t.Fatal("Expected at least one Run call to mock executor")
	}
	diffCall := calls[0]
	if len(diffCall.Args) < 4 || diffCall.Args[0] != "diff" || diffCall.Args[1] != "--numstat" ||
		diffCall.Args[2] != "abc123" || diffCall.Args[3] != "def456" {
		t.Errorf("Expected diff --numstat abc123 def456, got %v", diffCall.Args)
	}
}

func TestAnalyzeFromFiles(t *testing.T) {
	config := &Config{
		TrackedExtensions: []string{".go"},
		ExcludePatterns:   []string{},
		AuthorMappings:    make(map[string]string),
	}
	analyzer := NewAnalyzer(config)

	t.Run("new file counted as AI lines", func(t *testing.T) {
		before := &Checkpoint{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "human",
			Files:     map[string]FileContent{},
		}
		after := &Checkpoint{
			Timestamp: time.Now(),
			Author:    "claude",
			Files: map[string]FileContent{
				"newfile.go": {Path: "newfile.go", Lines: []string{"package main", "func hello() {}", "// comment"}},
			},
		}

		result, err := analyzer.AnalyzeCheckpoints(before, after)
		if err != nil {
			t.Fatalf("AnalyzeCheckpoints failed: %v", err)
		}

		if result.AILines != 3 {
			t.Errorf("AILines: got %d, want 3", result.AILines)
		}
		if result.HumanLines != 0 {
			t.Errorf("HumanLines: got %d, want 0", result.HumanLines)
		}
		if result.TotalLines != 3 {
			t.Errorf("TotalLines: got %d, want 3", result.TotalLines)
		}
	})

	t.Run("modified file counts added lines", func(t *testing.T) {
		before := &Checkpoint{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "human",
			Files: map[string]FileContent{
				"main.go": {Path: "main.go", Lines: []string{"line1", "line2"}},
			},
		}
		after := &Checkpoint{
			Timestamp: time.Now(),
			Author:    "claude",
			Files: map[string]FileContent{
				"main.go": {Path: "main.go", Lines: []string{"line1", "line2", "line3", "line4", "line5"}},
			},
		}

		result, err := analyzer.AnalyzeCheckpoints(before, after)
		if err != nil {
			t.Fatalf("AnalyzeCheckpoints failed: %v", err)
		}

		if result.AILines != 3 {
			t.Errorf("AILines: got %d, want 3", result.AILines)
		}
		if result.HumanLines != 0 {
			t.Errorf("HumanLines: got %d, want 0", result.HumanLines)
		}
	})

	t.Run("deleted file reduces TotalLines", func(t *testing.T) {
		before := &Checkpoint{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "human",
			Files: map[string]FileContent{
				"old.go": {Path: "old.go", Lines: []string{"line1", "line2", "line3", "line4"}},
			},
		}
		after := &Checkpoint{
			Timestamp: time.Now(),
			Author:    "claude",
			Files:     map[string]FileContent{},
		}

		result, err := analyzer.AnalyzeCheckpoints(before, after)
		if err != nil {
			t.Fatalf("AnalyzeCheckpoints failed: %v", err)
		}

		if result.TotalLines != -4 {
			t.Errorf("TotalLines: got %d, want -4", result.TotalLines)
		}
		if result.AILines != 0 {
			t.Errorf("AILines: got %d, want 0", result.AILines)
		}
	})

	t.Run("human author new file", func(t *testing.T) {
		before := &Checkpoint{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "someone",
			Files:     map[string]FileContent{},
		}
		after := &Checkpoint{
			Timestamp: time.Now(),
			Author:    "John Doe",
			Files: map[string]FileContent{
				"util.go": {Path: "util.go", Lines: []string{"package util", "func Add(a, b int) int { return a + b }"}},
			},
		}

		result, err := analyzer.AnalyzeCheckpoints(before, after)
		if err != nil {
			t.Fatalf("AnalyzeCheckpoints failed: %v", err)
		}

		if result.AILines != 0 {
			t.Errorf("AILines: got %d, want 0", result.AILines)
		}
		if result.HumanLines != 2 {
			t.Errorf("HumanLines: got %d, want 2", result.HumanLines)
		}
	})

	t.Run("mixed new and modified files", func(t *testing.T) {
		before := &Checkpoint{
			Timestamp: time.Now().Add(-1 * time.Hour),
			Author:    "human",
			Files: map[string]FileContent{
				"existing.go": {Path: "existing.go", Lines: []string{"line1", "line2"}},
			},
		}
		after := &Checkpoint{
			Timestamp: time.Now(),
			Author:    "claude",
			Files: map[string]FileContent{
				"existing.go":  {Path: "existing.go", Lines: []string{"line1", "line2", "line3"}},
				"brand_new.go": {Path: "brand_new.go", Lines: []string{"a", "b"}},
			},
		}

		result, err := analyzer.AnalyzeCheckpoints(before, after)
		if err != nil {
			t.Fatalf("AnalyzeCheckpoints failed: %v", err)
		}

		// existing.go: 1 line added, brand_new.go: 2 lines (new file) = 3 AI lines total
		if result.AILines != 3 {
			t.Errorf("AILines: got %d, want 3", result.AILines)
		}
		// TotalLines: 2 (new file lines) + 0 (modified diff doesn't add to total in this path)
		if result.TotalLines != 2 {
			t.Errorf("TotalLines: got %d, want 2", result.TotalLines)
		}
	})
}
