package storage

import (
	"fmt"
	"time"

	"github.com/y-hirakawa/ai-code-tracker/internal/tracker"
)

type MetricsStorage struct {
	storage *JSONStorage
}

func NewMetricsStorage(baseDir string) *MetricsStorage {
	return &MetricsStorage{
		storage: NewJSONStorage(baseDir),
	}
}

func (ms *MetricsStorage) SaveMetrics(result *tracker.AnalysisResult) error {
	filename := "metrics/current.json"
	return ms.storage.Save(filename, result)
}

func (ms *MetricsStorage) LoadMetrics() (*tracker.AnalysisResult, error) {
	filename := "metrics/current.json"
	var result tracker.AnalysisResult
	
	if !ms.storage.Exists(filename) {
		return &tracker.AnalysisResult{
			TotalLines:  0,
			AILines:     0,
			HumanLines:  0,
			Percentage:  0,
			LastUpdated: time.Now(),
		}, nil
	}
	
	if err := ms.storage.Load(filename, &result); err != nil {
		return nil, fmt.Errorf("failed to load metrics: %w", err)
	}
	
	return &result, nil
}

func (ms *MetricsStorage) SaveConfig(config *tracker.Config) error {
	filename := "config.json"
	return ms.storage.Save(filename, config)
}

func (ms *MetricsStorage) LoadConfig() (*tracker.Config, error) {
	filename := "config.json"
	var config tracker.Config
	
	if !ms.storage.Exists(filename) {
		return ms.getDefaultConfig(), nil
	}
	
	if err := ms.storage.Load(filename, &config); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	
	return &config, nil
}

func (ms *MetricsStorage) getDefaultConfig() *tracker.Config {
	return &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs"},
		ExcludePatterns:    []string{"*_test.go", "*.test.js", "*.spec.ts", "*_generated.go"},
		AuthorMappings:     make(map[string]string),
	}
}

func (ms *MetricsStorage) ArchiveMetrics(result *tracker.AnalysisResult) error {
	timestamp := result.LastUpdated.Format("20060102_150405")
	filename := fmt.Sprintf("metrics/archive/%s.json", timestamp)
	return ms.storage.Save(filename, result)
}