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
	Commit    string    `json:"commit,omitempty"`
	Added     int       `json:"added"`   // Total added lines across all files
	Deleted   int       `json:"deleted"` // Total deleted lines across all files
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
