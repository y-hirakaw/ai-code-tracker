package interfaces

import (
	"time"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// Storage defines the interface for data persistence operations
type Storage interface {
	// Save stores data to the specified filename
	Save(filename string, data interface{}) error
	
	// Load retrieves data from the specified filename
	Load(filename string, data interface{}) error
	
	// Exists checks if a file exists
	Exists(filename string) bool
	
	// Delete removes the specified file
	Delete(filename string) error
	
	// List returns all files matching the pattern
	List(pattern string) ([]string, error)
}

// MetricsStorage defines the interface for metrics-specific operations
type MetricsStorage interface {
	// SaveMetrics saves analysis results
	SaveMetrics(result *tracker.AnalysisResult) error
	
	// LoadMetrics retrieves the latest metrics
	LoadMetrics() (*tracker.AnalysisResult, error)
	
	// SaveConfig saves the configuration
	SaveConfig(config *tracker.Config) error
	
	// LoadConfig retrieves the configuration
	LoadConfig() (*tracker.Config, error)
	
	// ArchiveMetrics creates a timestamped backup
	ArchiveMetrics(timestamp string) error
}

// CheckpointStorage defines the interface for checkpoint operations
type CheckpointStorage interface {
	// RecordCheckpoint saves a new checkpoint
	RecordCheckpoint(checkpoint *tracker.Checkpoint) error
	
	// AppendRecord adds a checkpoint record to JSONL file
	AppendRecord(record *tracker.CheckpointRecord) error
	
	// ReadRecords retrieves all checkpoint records
	ReadRecords() ([]tracker.CheckpointRecord, error)
	
	// GetLatestRecords retrieves records after a specific timestamp
	GetLatestRecords(since time.Time) ([]tracker.CheckpointRecord, error)
}