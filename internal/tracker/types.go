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
	DefaultAuthor      string            `json:"default_author,omitempty"` // SPEC.md準拠
	AIAgents           []string          `json:"ai_agents,omitempty"`      // SPEC.md準拠
}

// SPEC.md準拠の型定義

// AuthorType represents the type of code author
type AuthorType string

const (
	AuthorTypeHuman AuthorType = "human"
	AuthorTypeAI    AuthorType = "ai"
)

// Change represents file-level changes with line ranges
type Change struct {
	Added   int     `json:"added"`
	Deleted int     `json:"deleted"`
	Lines   [][]int `json:"lines"` // [[start, end], [single], ...]
}

// FileSnapshot represents a snapshot of a file at a specific point in time
type FileSnapshot struct {
	Hash  string `json:"hash"`  // SHA-256 hash of file content
	Lines int    `json:"lines"` // Total number of lines
}

// CheckpointV2 represents a development checkpoint (SPEC.md準拠)
type CheckpointV2 struct {
	Timestamp time.Time             `json:"timestamp"`
	Author    string                `json:"author"`
	Type      AuthorType            `json:"type"`
	Metadata  map[string]string     `json:"metadata,omitempty"`
	Changes   map[string]Change     `json:"changes"`  // filepath -> Change
	Snapshot  map[string]FileSnapshot `json:"snapshot"` // filepath -> FileSnapshot (current state)
}

// AuthorshipLog represents commit-level authorship information
type AuthorshipLog struct {
	Version   string                `json:"version"`
	Commit    string                `json:"commit"`
	Timestamp time.Time             `json:"timestamp"`
	Files     map[string]FileInfo   `json:"files"`
}

// FileInfo contains author information for a single file
type FileInfo struct {
	Authors []AuthorInfo `json:"authors"`
}

// AuthorInfo represents a single author's contribution to a file
type AuthorInfo struct {
	Name     string            `json:"name"`
	Type     AuthorType        `json:"type"`
	Lines    [][]int           `json:"lines"` // [[start, end], ...]
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Report represents generated code generation report
type Report struct {
	Range    string       `json:"range,omitempty"`
	Branch   string       `json:"branch,omitempty"`
	Commits  int          `json:"commits,omitempty"`
	Period   *Period      `json:"period,omitempty"`
	Summary  SummaryStats `json:"summary"`
	ByFile   []FileStats  `json:"by_file,omitempty"`
	ByAuthor []AuthorStats `json:"by_author,omitempty"`
}

// Period represents a time period
type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SummaryStats represents summary statistics
type SummaryStats struct {
	TotalLines   int     `json:"total_lines"`
	AILines      int     `json:"ai_lines"`
	HumanLines   int     `json:"human_lines"`
	AIPercentage float64 `json:"ai_percentage"`
}

// AuthorStats represents statistics per author
type AuthorStats struct {
	Name       string     `json:"name"`
	Type       AuthorType `json:"type"`
	Lines      int        `json:"lines"`
	Percentage float64    `json:"percentage"`
	Commits    int        `json:"commits,omitempty"`
}
