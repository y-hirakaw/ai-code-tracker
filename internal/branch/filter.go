package branch

import (
	"fmt"
	"regexp"
	"strings"
)

// BranchFilter provides flexible branch name filtering capabilities
type BranchFilter struct {
	Pattern     string `json:"pattern"`
	IsRegex     bool   `json:"is_regex"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty"`
}

// NewBranchFilter creates a new branch filter with the specified pattern
func NewBranchFilter(pattern string, isRegex bool) *BranchFilter {
	return &BranchFilter{
		Pattern: pattern,
		IsRegex: isRegex,
		CaseInsensitive: false,
	}
}

// NewRegexFilter creates a new regex-based branch filter
func NewRegexFilter(pattern string) *BranchFilter {
	return &BranchFilter{
		Pattern: pattern,
		IsRegex: true,
		CaseInsensitive: false,
	}
}

// NewExactFilter creates a new exact match branch filter
func NewExactFilter(branchName string) *BranchFilter {
	return &BranchFilter{
		Pattern: branchName,
		IsRegex: false,
		CaseInsensitive: false,
	}
}

// WithCaseInsensitive returns a copy of the filter with case insensitive matching enabled
func (f *BranchFilter) WithCaseInsensitive() *BranchFilter {
	newFilter := *f
	newFilter.CaseInsensitive = true
	return &newFilter
}

// Matches checks if the given branch name matches the filter pattern
func (f *BranchFilter) Matches(branchName string) (bool, error) {
	if f.Pattern == "" {
		return true, nil // Empty pattern matches all
	}

	// Handle case insensitive matching
	pattern := f.Pattern
	branch := branchName
	if f.CaseInsensitive {
		pattern = strings.ToLower(pattern)
		branch = strings.ToLower(branch)
	}

	if f.IsRegex {
		return f.matchesRegex(branch, pattern)
	}
	
	// Exact match
	return branch == pattern, nil
}

// matchesRegex performs regex matching with error handling
func (f *BranchFilter) matchesRegex(branchName, pattern string) (bool, error) {
	// Compile regex with case sensitivity flag if needed
	regexPattern := pattern
	if f.CaseInsensitive {
		regexPattern = "(?i)" + pattern
	}

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern '%s': %w", f.Pattern, err)
	}

	return regex.MatchString(branchName), nil
}

// MustMatch is like Matches but panics on error (for testing convenience)
func (f *BranchFilter) MustMatch(branchName string) bool {
	matches, err := f.Matches(branchName)
	if err != nil {
		panic(err)
	}
	return matches
}

// Validate checks if the filter configuration is valid
func (f *BranchFilter) Validate() error {
	if f.Pattern == "" {
		return nil // Empty pattern is valid
	}

	if f.IsRegex {
		_, err := regexp.Compile(f.Pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern '%s': %w", f.Pattern, err)
		}
	}

	return nil
}

// String returns a human-readable description of the filter
func (f *BranchFilter) String() string {
	if f.Pattern == "" {
		return "all branches"
	}

	filterType := "exact"
	if f.IsRegex {
		filterType = "regex"
	}

	caseMode := ""
	if f.CaseInsensitive {
		caseMode = " (case-insensitive)"
	}

	return fmt.Sprintf("%s match: '%s'%s", filterType, f.Pattern, caseMode)
}

// MultiFilter allows filtering with multiple patterns (OR logic)
type MultiFilter struct {
	Filters []BranchFilter `json:"filters"`
}

// NewMultiFilter creates a new multi-pattern filter
func NewMultiFilter(filters ...BranchFilter) *MultiFilter {
	return &MultiFilter{
		Filters: filters,
	}
}

// Matches returns true if any of the contained filters match
func (mf *MultiFilter) Matches(branchName string) (bool, error) {
	if len(mf.Filters) == 0 {
		return true, nil // No filters means match all
	}

	for _, filter := range mf.Filters {
		matches, err := filter.Matches(branchName)
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil // OR logic - any match is sufficient
		}
	}

	return false, nil
}

// Add adds a new filter to the multi-filter
func (mf *MultiFilter) Add(filter BranchFilter) {
	mf.Filters = append(mf.Filters, filter)
}

// Validate checks if all contained filters are valid
func (mf *MultiFilter) Validate() error {
	for i, filter := range mf.Filters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("filter %d: %w", i, err)
		}
	}
	return nil
}

// String returns a description of all filters
func (mf *MultiFilter) String() string {
	if len(mf.Filters) == 0 {
		return "all branches"
	}

	var descriptions []string
	for _, filter := range mf.Filters {
		descriptions = append(descriptions, filter.String())
	}

	return strings.Join(descriptions, " OR ")
}