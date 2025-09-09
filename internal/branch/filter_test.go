package branch

import (
	"testing"
)

func TestBranchFilter_ExactMatch(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		branchName string
		want       bool
	}{
		{
			name:       "exact match - main",
			pattern:    "main",
			branchName: "main",
			want:       true,
		},
		{
			name:       "exact match - feature branch",
			pattern:    "feature/ui-improve",
			branchName: "feature/ui-improve",
			want:       true,
		},
		{
			name:       "no match - different branch",
			pattern:    "main",
			branchName: "develop",
			want:       false,
		},
		{
			name:       "no match - partial match",
			pattern:    "feature",
			branchName: "feature/ui-improve",
			want:       false,
		},
		{
			name:       "empty pattern matches all",
			pattern:    "",
			branchName: "any-branch",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewExactFilter(tt.pattern)
			got, err := filter.Matches(tt.branchName)
			if err != nil {
				t.Errorf("BranchFilter.Matches() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("BranchFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBranchFilter_RegexMatch(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		branchName string
		want       bool
		wantError  bool
	}{
		{
			name:       "feature prefix regex",
			pattern:    "^feature/",
			branchName: "feature/ui-improve",
			want:       true,
		},
		{
			name:       "feature prefix regex - no match",
			pattern:    "^feature/",
			branchName: "main",
			want:       false,
		},
		{
			name:       "hotfix or bugfix regex",
			pattern:    "(hotfix|bugfix)/",
			branchName: "hotfix/critical-bug",
			want:       true,
		},
		{
			name:       "hotfix or bugfix regex - bugfix match",
			pattern:    "(hotfix|bugfix)/",
			branchName: "bugfix/login-issue",
			want:       true,
		},
		{
			name:       "hotfix or bugfix regex - no match",
			pattern:    "(hotfix|bugfix)/",
			branchName: "feature/new-ui",
			want:       false,
		},
		{
			name:       "release version regex",
			pattern:    "^release/v[0-9]+\\.[0-9]+$",
			branchName: "release/v1.0",
			want:       true,
		},
		{
			name:       "release version regex - no match",
			pattern:    "^release/v[0-9]+\\.[0-9]+$",
			branchName: "release/v1.0.1",
			want:       false,
		},
		{
			name:       "invalid regex",
			pattern:    "[invalid",
			branchName: "any-branch",
			want:       false,
			wantError:  true,
		},
		{
			name:       "empty regex matches all",
			pattern:    "",
			branchName: "any-branch",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewRegexFilter(tt.pattern)
			got, err := filter.Matches(tt.branchName)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("BranchFilter.Matches() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("BranchFilter.Matches() error = %v", err)
				return
			}
			
			if got != tt.want {
				t.Errorf("BranchFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBranchFilter_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		branchName string
		isRegex    bool
		want       bool
	}{
		{
			name:       "exact match - case insensitive",
			pattern:    "MAIN",
			branchName: "main",
			isRegex:    false,
			want:       true,
		},
		{
			name:       "regex match - case insensitive",
			pattern:    "^FEATURE/",
			branchName: "feature/ui-improve",
			isRegex:    true,
			want:       true,
		},
		{
			name:       "no match - case insensitive",
			pattern:    "DEVELOP",
			branchName: "main",
			isRegex:    false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewBranchFilter(tt.pattern, tt.isRegex).WithCaseInsensitive()
			got, err := filter.Matches(tt.branchName)
			if err != nil {
				t.Errorf("BranchFilter.Matches() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("BranchFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBranchFilter_Validate(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		isRegex   bool
		wantError bool
	}{
		{
			name:      "valid exact pattern",
			pattern:   "main",
			isRegex:   false,
			wantError: false,
		},
		{
			name:      "valid regex pattern",
			pattern:   "^feature/.*",
			isRegex:   true,
			wantError: false,
		},
		{
			name:      "invalid regex pattern",
			pattern:   "[invalid",
			isRegex:   true,
			wantError: true,
		},
		{
			name:      "empty pattern is valid",
			pattern:   "",
			isRegex:   true,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewBranchFilter(tt.pattern, tt.isRegex)
			err := filter.Validate()
			
			if tt.wantError && err == nil {
				t.Errorf("BranchFilter.Validate() expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("BranchFilter.Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestBranchFilter_String(t *testing.T) {
	tests := []struct {
		name     string
		filter   *BranchFilter
		expected string
	}{
		{
			name:     "exact match",
			filter:   NewExactFilter("main"),
			expected: "exact match: 'main'",
		},
		{
			name:     "regex match",
			filter:   NewRegexFilter("^feature/"),
			expected: "regex match: '^feature/'",
		},
		{
			name:     "case insensitive exact",
			filter:   NewExactFilter("MAIN").WithCaseInsensitive(),
			expected: "exact match: 'MAIN' (case-insensitive)",
		},
		{
			name:     "empty pattern",
			filter:   NewExactFilter(""),
			expected: "all branches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.String()
			if got != tt.expected {
				t.Errorf("BranchFilter.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMultiFilter(t *testing.T) {
	// Create multiple filters
	filter1 := NewExactFilter("main")
	filter2 := NewRegexFilter("^feature/")
	filter3 := NewRegexFilter("^hotfix/")

	multiFilter := NewMultiFilter(*filter1, *filter2, *filter3)

	tests := []struct {
		name       string
		branchName string
		want       bool
	}{
		{
			name:       "matches first filter",
			branchName: "main",
			want:       true,
		},
		{
			name:       "matches second filter",
			branchName: "feature/ui-improve",
			want:       true,
		},
		{
			name:       "matches third filter",
			branchName: "hotfix/critical-bug",
			want:       true,
		},
		{
			name:       "no match",
			branchName: "develop",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := multiFilter.Matches(tt.branchName)
			if err != nil {
				t.Errorf("MultiFilter.Matches() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("MultiFilter.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiFilter_Empty(t *testing.T) {
	multiFilter := NewMultiFilter()

	// Empty multi-filter should match all branches
	got, err := multiFilter.Matches("any-branch")
	if err != nil {
		t.Errorf("MultiFilter.Matches() error = %v", err)
		return
	}
	if !got {
		t.Errorf("MultiFilter.Matches() = %v, want true for empty filter", got)
	}
}

func TestMultiFilter_Validate(t *testing.T) {
	// Create multi-filter with one invalid regex
	validFilter := NewExactFilter("main")
	invalidFilter := NewRegexFilter("[invalid")

	multiFilter := NewMultiFilter(*validFilter, *invalidFilter)

	err := multiFilter.Validate()
	if err == nil {
		t.Errorf("MultiFilter.Validate() expected error for invalid regex, got nil")
	}
}

func TestBranchFilter_MustMatch(t *testing.T) {
	filter := NewExactFilter("main")

	// Should not panic for valid match
	result := filter.MustMatch("main")
	if !result {
		t.Errorf("BranchFilter.MustMatch() = %v, want true", result)
	}

	// Should not panic for valid non-match
	result = filter.MustMatch("develop")
	if result {
		t.Errorf("BranchFilter.MustMatch() = %v, want false", result)
	}

	// Test panic with invalid regex (commented out to avoid actual panic in test)
	// invalidFilter := NewRegexFilter("[invalid")
	// This would panic: invalidFilter.MustMatch("any")
}