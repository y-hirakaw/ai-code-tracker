package tracker

import "time"

// Legacy Checkpoint struct (for backward compatibility)
type Checkpoint struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Author      string                 `json:"author"`
	CommitHash  string                 `json:"commit_hash,omitempty"`
	Files       map[string]FileContent `json:"files"`
	NumstatData map[string][2]int      `json:"numstat_data,omitempty"` // [added, deleted] lines from HEAD
}

// CheckpointRecord is the new lightweight format for JSONL storage
type CheckpointRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Branch    string    `json:"branch,omitempty"`    // Branch name where changes occurred
	Commit    string    `json:"commit,omitempty"`
	Added     int       `json:"added"`   // Total added lines across all files
	Deleted   int       `json:"deleted"` // Total deleted lines across all files
}

// GetBranch returns the branch name, defaulting to "main" for backward compatibility
func (r *CheckpointRecord) GetBranch() string {
	if r.Branch == "" {
		return "main"
	}
	return r.Branch
}

// HasBranchInfo returns true if the record contains explicit branch information
func (r *CheckpointRecord) HasBranchInfo() bool {
	return r.Branch != ""
}

// GetDisplayBranch returns branch name for display purposes
func (r *CheckpointRecord) GetDisplayBranch() string {
	branch := r.GetBranch()
	if branch == "main" && !r.HasBranchInfo() {
		return "main (inferred)"
	}
	return branch
}

type FileContent struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

type AnalysisResult struct {
	TotalLines  int       `json:"total_lines"`
	AILines     int       `json:"ai_lines"`
	HumanLines  int       `json:"human_lines"`
	Percentage  float64   `json:"percentage"`
	LastUpdated time.Time `json:"last_updated"`
}

type FileStats struct {
	Path       string `json:"path"`
	TotalLines int    `json:"total_lines"`
	AILines    int    `json:"ai_lines"`
	HumanLines int    `json:"human_lines"`
}

type Config struct {
	TargetAIPercentage float64           `json:"target_ai_percentage"`
	TrackedExtensions  []string          `json:"tracked_extensions"`
	ExcludePatterns    []string          `json:"exclude_patterns"`
	AuthorMappings     map[string]string `json:"author_mappings"`
}
