package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestNewMetricsStorage(t *testing.T) {
	baseDir := "/tmp/test-metrics"
	ms := NewMetricsStorage(baseDir)

	if ms.storage.baseDir != baseDir {
		t.Errorf("Expected baseDir to be '%s', got '%s'", baseDir, ms.storage.baseDir)
	}
}

func TestSaveAndLoadMetrics(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-metrics-storage")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ms := NewMetricsStorage(tmpDir)

	// Create test metrics
	originalMetrics := &tracker.AnalysisResult{
		TotalLines:  1000,
		AILines:     700,
		HumanLines:  300,
		Percentage:  70.0,
		LastUpdated: time.Now(),
	}

	// Save metrics
	err = ms.SaveMetrics(originalMetrics)
	if err != nil {
		t.Fatalf("Failed to save metrics: %v", err)
	}

	// Load metrics
	loadedMetrics, err := ms.LoadMetrics()
	if err != nil {
		t.Fatalf("Failed to load metrics: %v", err)
	}

	// Verify loaded metrics
	if loadedMetrics.TotalLines != originalMetrics.TotalLines {
		t.Errorf("Expected TotalLines %d, got %d", originalMetrics.TotalLines, loadedMetrics.TotalLines)
	}

	if loadedMetrics.AILines != originalMetrics.AILines {
		t.Errorf("Expected AILines %d, got %d", originalMetrics.AILines, loadedMetrics.AILines)
	}

	if loadedMetrics.HumanLines != originalMetrics.HumanLines {
		t.Errorf("Expected HumanLines %d, got %d", originalMetrics.HumanLines, loadedMetrics.HumanLines)
	}

	if loadedMetrics.Percentage != originalMetrics.Percentage {
		t.Errorf("Expected Percentage %.1f, got %.1f", originalMetrics.Percentage, loadedMetrics.Percentage)
	}
}

func TestLoadMetricsDefault(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-metrics-default")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ms := NewMetricsStorage(tmpDir)

	// Load metrics when file doesn't exist
	metrics, err := ms.LoadMetrics()
	if err != nil {
		t.Fatalf("Failed to load default metrics: %v", err)
	}

	// Verify default values
	if metrics.TotalLines != 0 {
		t.Errorf("Expected default TotalLines 0, got %d", metrics.TotalLines)
	}

	if metrics.AILines != 0 {
		t.Errorf("Expected default AILines 0, got %d", metrics.AILines)
	}

	if metrics.HumanLines != 0 {
		t.Errorf("Expected default HumanLines 0, got %d", metrics.HumanLines)
	}

	if metrics.Percentage != 0 {
		t.Errorf("Expected default Percentage 0, got %.1f", metrics.Percentage)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-config-storage")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ms := NewMetricsStorage(tmpDir)

	// Create test config
	originalConfig := &tracker.Config{
		TargetAIPercentage: 75.0,
		TrackedExtensions:  []string{".go", ".py", ".js"},
		ExcludePatterns:    []string{"test", "vendor"},
		AuthorMappings: map[string]string{
			"AI Bot": "ai",
			"Dev 1":  "human",
		},
	}

	// Save config
	err = ms.SaveConfig(originalConfig)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedConfig, err := ms.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded config
	if loadedConfig.TargetAIPercentage != originalConfig.TargetAIPercentage {
		t.Errorf("Expected TargetAIPercentage %.1f, got %.1f", originalConfig.TargetAIPercentage, loadedConfig.TargetAIPercentage)
	}

	if len(loadedConfig.TrackedExtensions) != len(originalConfig.TrackedExtensions) {
		t.Errorf("Expected %d tracked extensions, got %d", len(originalConfig.TrackedExtensions), len(loadedConfig.TrackedExtensions))
	}

	for i, ext := range loadedConfig.TrackedExtensions {
		if ext != originalConfig.TrackedExtensions[i] {
			t.Errorf("TrackedExtensions[%d]: expected '%s', got '%s'", i, originalConfig.TrackedExtensions[i], ext)
		}
	}

	if len(loadedConfig.ExcludePatterns) != len(originalConfig.ExcludePatterns) {
		t.Errorf("Expected %d exclude patterns, got %d", len(originalConfig.ExcludePatterns), len(loadedConfig.ExcludePatterns))
	}

	if len(loadedConfig.AuthorMappings) != len(originalConfig.AuthorMappings) {
		t.Errorf("Expected %d author mappings, got %d", len(originalConfig.AuthorMappings), len(loadedConfig.AuthorMappings))
	}

	// Verify specific author mapping
	if mapping, exists := loadedConfig.AuthorMappings["AI Bot"]; !exists || mapping != "ai" {
		t.Error("Expected 'AI Bot' to be mapped to 'ai'")
	}
}

func TestLoadConfigDefault(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-config-default")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ms := NewMetricsStorage(tmpDir)

	// Load config when file doesn't exist
	config, err := ms.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	// Verify default values
	if config.TargetAIPercentage != 80.0 {
		t.Errorf("Expected default TargetAIPercentage 80.0, got %.1f", config.TargetAIPercentage)
	}

	expectedExtensions := []string{".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs"}
	if len(config.TrackedExtensions) != len(expectedExtensions) {
		t.Errorf("Expected %d default extensions, got %d", len(expectedExtensions), len(config.TrackedExtensions))
	}

	expectedPatterns := []string{"*_test.go", "*.test.js", "*.spec.ts", "*_generated.go"}
	if len(config.ExcludePatterns) != len(expectedPatterns) {
		t.Errorf("Expected %d default exclude patterns, got %d", len(expectedPatterns), len(config.ExcludePatterns))
	}

	if len(config.AuthorMappings) != 0 {
		t.Errorf("Expected empty author mappings by default, got %d", len(config.AuthorMappings))
	}
}

func TestArchiveMetrics(t *testing.T) {
	// Create temporary directory
	tmpDir := filepath.Join(os.TempDir(), "test-archive-metrics")
	err := os.MkdirAll(tmpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ms := NewMetricsStorage(tmpDir)

	// Create test metrics with specific timestamp
	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	metrics := &tracker.AnalysisResult{
		TotalLines:  500,
		AILines:     400,
		HumanLines:  100,
		Percentage:  80.0,
		LastUpdated: testTime,
	}

	// Archive metrics
	err = ms.ArchiveMetrics(metrics)
	if err != nil {
		t.Fatalf("Failed to archive metrics: %v", err)
	}

	// Verify archived file exists
	expectedFilename := "metrics/archive/20240115_143045.json"
	archivePath := filepath.Join(tmpDir, expectedFilename)

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Errorf("Expected archive file '%s' to exist", expectedFilename)
	}

	// Load archived metrics to verify content
	var archivedMetrics tracker.AnalysisResult
	err = ms.storage.Load(expectedFilename, &archivedMetrics)
	if err != nil {
		t.Fatalf("Failed to load archived metrics: %v", err)
	}

	if archivedMetrics.TotalLines != metrics.TotalLines {
		t.Errorf("Expected archived TotalLines %d, got %d", metrics.TotalLines, archivedMetrics.TotalLines)
	}

	if archivedMetrics.AILines != metrics.AILines {
		t.Errorf("Expected archived AILines %d, got %d", metrics.AILines, archivedMetrics.AILines)
	}
}

func TestGetDefaultConfig(t *testing.T) {
	ms := NewMetricsStorage("/tmp")
	config := ms.getDefaultConfig()

	// Verify default config structure
	if config.TargetAIPercentage != 80.0 {
		t.Errorf("Expected default TargetAIPercentage 80.0, got %.1f", config.TargetAIPercentage)
	}

	// Check that common extensions are included
	commonExtensions := []string{".go", ".py", ".js", ".java"}
	for _, ext := range commonExtensions {
		found := false
		for _, tracked := range config.TrackedExtensions {
			if tracked == ext {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected extension '%s' to be in default tracked extensions", ext)
		}
	}

	// Check that test patterns are excluded
	testPatterns := []string{"*_test.go", "*.test.js"}
	for _, pattern := range testPatterns {
		found := false
		for _, excluded := range config.ExcludePatterns {
			if excluded == pattern {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern '%s' to be in default exclude patterns", pattern)
		}
	}
}