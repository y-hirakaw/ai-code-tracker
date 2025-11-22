package storage

import (
	"fmt"
	"time"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/interfaces"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
	"github.com/y-hirakaw/ai-code-tracker/internal/validation"
)

// Ensure MetricsStorageV2 implements the MetricsStorage interface
var _ interfaces.MetricsStorage = (*MetricsStorageV2)(nil)

// MetricsStorageV2 is an improved version with validation and error handling
type MetricsStorageV2 struct {
	storage   interfaces.Storage
	validator *validation.ConfigValidator
}

// NewMetricsStorageV2 creates a new improved metrics storage instance
func NewMetricsStorageV2(storage interfaces.Storage) *MetricsStorageV2 {
	return &MetricsStorageV2{
		storage:   storage,
		validator: validation.NewConfigValidator(),
	}
}

// SaveMetrics saves analysis results with validation
func (ms *MetricsStorageV2) SaveMetrics(result *tracker.AnalysisResult) error {
	if result == nil {
		return errors.NewValidationError("SaveMetrics", "result", "analysis result cannot be nil")
	}
	
	// Validate percentages
	if result.Percentage < 0 || result.Percentage > 100 {
		return errors.NewValidationError(
			"SaveMetrics",
			"Percentage",
			fmt.Sprintf("percentage must be between 0 and 100, got: %.2f", result.Percentage),
		)
	}
	
	// Validate line counts
	if result.TotalLines < 0 || result.AILines < 0 || result.HumanLines < 0 {
		return errors.NewValidationError("SaveMetrics", "lines", "line counts cannot be negative")
	}
	
	// Update timestamp if not set
	if result.LastUpdated.IsZero() {
		result.LastUpdated = time.Now()
	}
	
	filename := "metrics/current.json"
	return ms.storage.Save(filename, result)
}

// LoadMetrics retrieves the latest metrics with defaults
func (ms *MetricsStorageV2) LoadMetrics() (*tracker.AnalysisResult, error) {
	filename := "metrics/current.json"
	var result tracker.AnalysisResult
	
	if !ms.storage.Exists(filename) {
		// Return default metrics
		return &tracker.AnalysisResult{
			TotalLines:  0,
			AILines:     0,
			HumanLines:  0,
			Percentage:  0,
			LastUpdated: time.Now(),
		}, nil
	}
	
	if err := ms.storage.Load(filename, &result); err != nil {
		return nil, errors.NewStorageError("LoadMetrics", filename, err)
	}
	
	// Validate loaded data
	if result.Percentage < 0 || result.Percentage > 100 {
		// Fix invalid data
		result.Percentage = 0
	}
	
	return &result, nil
}

// SaveConfig saves configuration with validation
func (ms *MetricsStorageV2) SaveConfig(config *tracker.Config) error {
	// Validate configuration
	if err := ms.validator.Validate(config); err != nil {
		return err
	}
	
	filename := "config.json"
	return ms.storage.Save(filename, config)
}

// LoadConfig retrieves configuration with defaults
func (ms *MetricsStorageV2) LoadConfig() (*tracker.Config, error) {
	filename := "config.json"
	
	if !ms.storage.Exists(filename) {
		// Create default config file if it doesn't exist
		defaultConfig := GetDefaultConfig()
		if err := ms.SaveConfig(defaultConfig); err != nil {
			return nil, errors.NewConfigError("LoadConfig", "failed to create default config", err)
		}
		return defaultConfig, nil
	}
	
	var config tracker.Config
	if err := ms.storage.Load(filename, &config); err != nil {
		return nil, errors.NewConfigError("LoadConfig", "failed to load configuration", err)
	}
	
	// Validate loaded configuration
	if err := ms.validator.Validate(&config); err != nil {
		// Return default on validation failure
		return GetDefaultConfig(), nil
	}
	
	return &config, nil
}

// ArchiveMetrics creates a timestamped backup of current metrics
func (ms *MetricsStorageV2) ArchiveMetrics(timestamp string) error {
	// Load current metrics
	current, err := ms.LoadMetrics()
	if err != nil {
		return errors.NewStorageError("ArchiveMetrics", "current", err)
	}
	
	// Validate timestamp format
	if timestamp == "" {
		timestamp = time.Now().Format("20060102_150405")
	}
	
	// Save to archive
	archiveFilename := fmt.Sprintf("metrics/archive/%s.json", timestamp)
	if err := ms.storage.Save(archiveFilename, current); err != nil {
		return errors.NewStorageError("ArchiveMetrics", archiveFilename, err)
	}
	
	return nil
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *tracker.Config {
	return &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions: []string{
			".go", ".py", ".js", ".ts", ".java", ".cs", ".cpp", ".c", ".h",
			".rb", ".php", ".swift", ".kt", ".rs", ".scala", ".r", ".m",
		},
		ExcludePatterns: []string{
			"*_test.go", "*.test.js", "*.spec.ts", "*_test.py",
			"vendor/*", "node_modules/*", ".git/*", "dist/*", "build/*",
		},
		AuthorMappings: map[string]string{
			"AI Assistant": "ai",
			"Claude":       "ai",
			"GitHub Copilot": "ai",
		},
	}
}