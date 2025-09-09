package branch

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// BranchAnalyzer provides branch-based analysis of checkpoint records
type BranchAnalyzer struct {
	records []tracker.CheckpointRecord
}

// NewBranchAnalyzer creates a new branch analyzer with the given records
func NewBranchAnalyzer(records []tracker.CheckpointRecord) *BranchAnalyzer {
	return &BranchAnalyzer{
		records: records,
	}
}

// BranchReport contains analysis results for a single branch
type BranchReport struct {
	BranchName    string    `json:"branch_name"`
	RecordCount   int       `json:"record_count"`
	TotalAdded    int       `json:"total_added"`
	TotalDeleted  int       `json:"total_deleted"`
	FirstRecord   time.Time `json:"first_record"`
	LastRecord    time.Time `json:"last_record"`
	Authors       []string  `json:"authors"`
	AIRatio       float64   `json:"ai_ratio"`
}

// GroupReport contains analysis results for multiple branches matching a pattern
type GroupReport struct {
	PatternDescription string                   `json:"pattern_description"`
	MatchingBranches   []string                 `json:"matching_branches"`
	TotalRecords       int                      `json:"total_records"`
	TotalAdded         int                      `json:"total_added"`
	TotalDeleted       int                      `json:"total_deleted"`
	GroupAIRatio       float64                  `json:"group_ai_ratio"`
	BranchReports      map[string]*BranchReport `json:"branch_reports"`
}

// AnalyzeByBranch analyzes records for a specific branch (exact match)
func (a *BranchAnalyzer) AnalyzeByBranch(branchName string) (*BranchReport, error) {
	filter := NewExactFilter(branchName)
	filteredRecords, err := a.filterRecords(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to filter records for branch '%s': %w", branchName, err)
	}

	if len(filteredRecords) == 0 {
		return &BranchReport{
			BranchName:  branchName,
			RecordCount: 0,
			Authors:     []string{},
		}, nil
	}

	return a.analyzeRecords(branchName, filteredRecords), nil
}

// AnalyzeByPattern analyzes records for branches matching a pattern
func (a *BranchAnalyzer) AnalyzeByPattern(pattern string, isRegex bool) (*GroupReport, error) {
	var filter *BranchFilter
	if isRegex {
		filter = NewRegexFilter(pattern)
	} else {
		// Check if pattern looks like a glob pattern
		if strings.ContainsAny(pattern, "*?[]") {
			filter = NewGlobFilter(pattern)
		} else {
			filter = NewExactFilter(pattern)
		}
	}
	
	// Validate filter
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter pattern: %w", err)
	}

	// Group records by branch
	branchGroups := make(map[string][]tracker.CheckpointRecord)
	matchingBranches := make(map[string]bool)

	for _, record := range a.records {
		branchName := record.GetBranch()
		
		matches, err := filter.Matches(branchName)
		if err != nil {
			return nil, fmt.Errorf("pattern matching failed for branch '%s': %w", branchName, err)
		}

		if matches {
			branchGroups[branchName] = append(branchGroups[branchName], record)
			matchingBranches[branchName] = true
		}
	}

	// Create group report
	groupReport := &GroupReport{
		PatternDescription: filter.String(),
		MatchingBranches:   make([]string, 0, len(matchingBranches)),
		BranchReports:      make(map[string]*BranchReport),
	}

	// Sort branch names for consistent output
	for branch := range matchingBranches {
		groupReport.MatchingBranches = append(groupReport.MatchingBranches, branch)
	}
	sort.Strings(groupReport.MatchingBranches)

	// Analyze each matching branch
	for _, branchName := range groupReport.MatchingBranches {
		records := branchGroups[branchName]
		branchReport := a.analyzeRecords(branchName, records)
		
		groupReport.BranchReports[branchName] = branchReport
		groupReport.TotalRecords += branchReport.RecordCount
		groupReport.TotalAdded += branchReport.TotalAdded
		groupReport.TotalDeleted += branchReport.TotalDeleted
	}

	// Calculate group AI ratio
	if groupReport.TotalAdded > 0 {
		aiLines := a.calculateAILines(groupReport.TotalAdded)
		groupReport.GroupAIRatio = float64(aiLines) / float64(groupReport.TotalAdded) * 100
	}

	return groupReport, nil
}

// AnalyzeAllBranches analyzes records for all branches
func (a *BranchAnalyzer) AnalyzeAllBranches() (*GroupReport, error) {
	// Use empty pattern to match all branches
	return a.AnalyzeByPattern("", false)
}

// GetUniqueBranches returns a sorted list of all unique branch names in the records
func (a *BranchAnalyzer) GetUniqueBranches() []string {
	branchSet := make(map[string]bool)
	
	for _, record := range a.records {
		branchName := record.GetBranch()
		branchSet[branchName] = true
	}

	branches := make([]string, 0, len(branchSet))
	for branch := range branchSet {
		branches = append(branches, branch)
	}
	
	sort.Strings(branches)
	return branches
}

// filterRecords filters records using the given branch filter
func (a *BranchAnalyzer) filterRecords(filter *BranchFilter) ([]tracker.CheckpointRecord, error) {
	var filtered []tracker.CheckpointRecord
	
	for _, record := range a.records {
		branchName := record.GetBranch()
		
		matches, err := filter.Matches(branchName)
		if err != nil {
			return nil, err
		}

		if matches {
			filtered = append(filtered, record)
		}
	}

	return filtered, nil
}

// analyzeRecords performs analysis on a collection of records for a single branch
func (a *BranchAnalyzer) analyzeRecords(branchName string, records []tracker.CheckpointRecord) *BranchReport {
	if len(records) == 0 {
		return &BranchReport{
			BranchName:  branchName,
			RecordCount: 0,
			Authors:     []string{},
		}
	}

	report := &BranchReport{
		BranchName:  branchName,
		RecordCount: len(records),
	}

	// Track authors
	authorSet := make(map[string]bool)
	
	// Find first and last records, calculate totals
	report.FirstRecord = records[0].Timestamp
	report.LastRecord = records[0].Timestamp

	for _, record := range records {
		report.TotalAdded += record.Added
		report.TotalDeleted += record.Deleted

		// Track time range
		if record.Timestamp.Before(report.FirstRecord) {
			report.FirstRecord = record.Timestamp
		}
		if record.Timestamp.After(report.LastRecord) {
			report.LastRecord = record.Timestamp
		}

		// Track unique authors
		if record.Author != "" {
			authorSet[record.Author] = true
		}
	}

	// Convert author set to sorted slice
	for author := range authorSet {
		report.Authors = append(report.Authors, author)
	}
	sort.Strings(report.Authors)

	// Calculate AI ratio (simplified - assumes AI contributes more lines)
	if report.TotalAdded > 0 {
		aiLines := a.calculateAILines(report.TotalAdded)
		report.AIRatio = float64(aiLines) / float64(report.TotalAdded) * 100
	}

	return report
}

// calculateAILines estimates AI-generated lines (placeholder logic)
// TODO: Integrate with actual AI detection logic from main tracker
func (a *BranchAnalyzer) calculateAILines(totalAdded int) int {
	// Simplified assumption: 80% of lines are AI-generated
	// This should be replaced with actual logic from the main tracker
	return int(float64(totalAdded) * 0.8)
}

// RecordStats provides summary statistics across all records
type RecordStats struct {
	TotalRecords       int       `json:"total_records"`
	UniqueBranches     int       `json:"unique_branches"`
	TotalAdded         int       `json:"total_added"`
	TotalDeleted       int       `json:"total_deleted"`
	FirstRecord        time.Time `json:"first_record"`
	LastRecord         time.Time `json:"last_record"`
	RecordsWithBranch  int       `json:"records_with_branch"`
	RecordsWithoutBranch int     `json:"records_without_branch"`
}

// GetRecordStats returns overall statistics about the loaded records
func (a *BranchAnalyzer) GetRecordStats() *RecordStats {
	if len(a.records) == 0 {
		return &RecordStats{}
	}

	stats := &RecordStats{
		TotalRecords:   len(a.records),
		FirstRecord:    a.records[0].Timestamp,
		LastRecord:     a.records[0].Timestamp,
	}

	branchSet := make(map[string]bool)

	for _, record := range a.records {
		stats.TotalAdded += record.Added
		stats.TotalDeleted += record.Deleted

		// Track time range
		if record.Timestamp.Before(stats.FirstRecord) {
			stats.FirstRecord = record.Timestamp
		}
		if record.Timestamp.After(stats.LastRecord) {
			stats.LastRecord = record.Timestamp
		}

		// Track branch info
		branchName := record.GetBranch()
		branchSet[branchName] = true

		if record.HasBranchInfo() {
			stats.RecordsWithBranch++
		} else {
			stats.RecordsWithoutBranch++
		}
	}

	stats.UniqueBranches = len(branchSet)
	return stats
}