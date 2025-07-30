package tracker

import "time"

type Checkpoint struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Author    string                 `json:"author"`
	Files     map[string]FileContent `json:"files"`
}

type FileContent struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

type AnalysisResult struct {
	TotalLines    int     `json:"total_lines"`
	BaselineLines int     `json:"baseline_lines"`
	AILines       int     `json:"ai_lines"`
	HumanLines    int     `json:"human_lines"`
	Percentage    float64 `json:"percentage"`
	LastUpdated   time.Time `json:"last_updated"`
}

type FileStats struct {
	Path       string `json:"path"`
	TotalLines int    `json:"total_lines"`
	AILines    int    `json:"ai_lines"`
	HumanLines int    `json:"human_lines"`
}

type Config struct {
	TargetAIPercentage float64  `json:"target_ai_percentage"`
	TrackedExtensions  []string `json:"tracked_extensions"`
	ExcludePatterns    []string `json:"exclude_patterns"`
	AuthorMappings     map[string]string `json:"author_mappings"`
}