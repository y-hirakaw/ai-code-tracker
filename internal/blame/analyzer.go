package blame

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// BlameAnalyzer analyzes code using git blame
type BlameAnalyzer struct {
	config *tracker.Config
}

// NewBlameAnalyzer creates a new BlameAnalyzer
func NewBlameAnalyzer(config *tracker.Config) *BlameAnalyzer {
	return &BlameAnalyzer{
		config: config,
	}
}

// AnalyzeCodebase analyzes the entire codebase using git blame
func (ba *BlameAnalyzer) AnalyzeCodebase() (*tracker.AnalysisResult, error) {
	// Get all tracked files
	files, err := ba.getTrackedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked files: %w", err)
	}

	aiLines := 0
	humanLines := 0

	// Analyze each file
	for _, file := range files {
		ai, human, err := ba.analyzeFile(file)
		if err != nil {
			// Skip files that can't be blamed (binary, etc.)
			continue
		}

		aiLines += ai
		humanLines += human
	}

	totalLines := aiLines + humanLines
	percentage := 0.0
	if totalLines > 0 {
		percentage = float64(aiLines) / float64(totalLines) * 100
	}

	return &tracker.AnalysisResult{
		TotalLines: totalLines,
		AILines:    aiLines,
		HumanLines: humanLines,
		Percentage: percentage,
	}, nil
}

// analyzeFile analyzes a single file using git blame
func (ba *BlameAnalyzer) analyzeFile(filePath string) (aiLines, humanLines int, err error) {
	// Run git blame --line-porcelain
	cmd := exec.Command("git", "blame", "--line-porcelain", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("git blame failed: %w", err)
	}

	// Parse git blame output
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var currentAuthor string

	for scanner.Scan() {
		line := scanner.Text()

		// Parse author line
		if strings.HasPrefix(line, "author ") {
			currentAuthor = strings.TrimPrefix(line, "author ")
			continue
		}

		// When we hit the actual code line (starts with tab)
		if strings.HasPrefix(line, "\t") && currentAuthor != "" {
			if ba.isAIAuthor(currentAuthor) {
				aiLines++
			} else {
				humanLines++
			}
			currentAuthor = "" // Reset for next line
		}
	}

	return aiLines, humanLines, scanner.Err()
}

// getTrackedFiles returns all files that should be tracked
func (ba *BlameAnalyzer) getTrackedFiles() ([]string, error) {
	// Get all files tracked by git
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed: %w", err)
	}

	var trackedFiles []string
	scanner := bufio.Scanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		filePath := scanner.Text()

		// Check if file should be tracked
		if ba.shouldTrackFile(filePath) {
			trackedFiles = append(trackedFiles, filePath)
		}
	}

	return trackedFiles, scanner.Err()
}

// shouldTrackFile checks if a file should be tracked based on config
func (ba *BlameAnalyzer) shouldTrackFile(filePath string) bool {
	// Check extension
	hasValidExt := false
	for _, ext := range ba.config.TrackedExtensions {
		if strings.HasSuffix(filePath, ext) {
			hasValidExt = true
			break
		}
	}

	if !hasValidExt {
		return false
	}

	// Check exclusion patterns
	for _, pattern := range ba.config.ExcludePatterns {
		// Simple pattern matching
		if matched, _ := filepath.Match(pattern, filepath.Base(filePath)); matched {
			return false
		}

		// Check if path contains pattern
		if strings.Contains(filePath, strings.TrimSuffix(pattern, "/*")) {
			return false
		}
	}

	return true
}

// isAIAuthor checks if an author should be considered AI
func (ba *BlameAnalyzer) isAIAuthor(author string) bool {
	// Check author mappings first
	if mapping, exists := ba.config.AuthorMappings[author]; exists {
		return strings.ToLower(mapping) == "ai"
	}

	// Check common AI author names
	authorLower := strings.ToLower(author)
	aiAuthors := []string{"claude", "ai", "assistant", "bot", "copilot", "codewhisperer"}

	for _, aiAuthor := range aiAuthors {
		if strings.Contains(authorLower, aiAuthor) {
			return true
		}
	}

	return false
}
