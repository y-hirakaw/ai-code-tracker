package benchmarks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/git"
	"github.com/y-hirakaw/ai-code-tracker/internal/security"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
	"github.com/y-hirakaw/ai-code-tracker/internal/validation"
)

// BenchmarkJSONStorageV1 benchmarks the original JSON storage
func BenchmarkJSONStorageV1(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_storage_v1")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	storage := storage.NewJSONStorage(tmpDir)
	
	data := &tracker.AnalysisResult{
		TotalLines:  1000,
		AILines:     800,
		HumanLines:  200,
		Percentage:  80.0,
		LastUpdated: time.Now(),
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("test_%d.json", i%100)
		
		// Save
		if err := storage.Save(filename, data); err != nil {
			b.Fatal(err)
		}
		
		// Load
		var loaded tracker.AnalysisResult
		if err := storage.Load(filename, &loaded); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONStorageV2 benchmarks the improved JSON storage
func BenchmarkJSONStorageV2(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_storage_v2")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	storage, err := storage.NewJSONStorageV2(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	
	data := &tracker.AnalysisResult{
		TotalLines:  1000,
		AILines:     800,
		HumanLines:  200,
		Percentage:  80.0,
		LastUpdated: time.Now(),
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("test_%d.json", i%100)
		
		// Save
		if err := storage.Save(filename, data); err != nil {
			b.Fatal(err)
		}
		
		// Load
		var loaded tracker.AnalysisResult
		if err := storage.Load(filename, &loaded); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConfigValidation benchmarks configuration validation
func BenchmarkConfigValidation(b *testing.B) {
	validator := validation.NewConfigValidator()
	
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".js", ".py", ".ts", ".java"},
		ExcludePatterns:    []string{"*_test.go", "vendor/*", "node_modules/*"},
		AuthorMappings: map[string]string{
			"AI Assistant": "ai",
			"Claude":       "ai",
			"Copilot":      "ai",
		},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(config); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSafeCommandValidation benchmarks command validation
func BenchmarkSafeCommandValidation(b *testing.B) {
	executor := security.NewSafeCommandExecutor()
	
	args := []string{"status", "--porcelain", "--untracked-files=no"}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		if err := executor.ValidateCommandArgs(args); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPathValidation benchmarks path validation
func BenchmarkPathValidation(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_path_validation")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	safeOps, err := security.NewSafeFileOperations(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	
	paths := []string{
		"file.json",
		"subdir/file.json",
		"deep/nested/directory/file.json",
		"metrics/current.json",
		"checkpoints/checkpoint_123.json",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		if _, err := safeOps.ValidatePath(path); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGitOperationsContext benchmarks context-aware Git operations
func BenchmarkGitOperationsContext(b *testing.B) {
	analyzer := git.NewContextAwareDiffAnalyzer(30 * time.Second)
	
	// Skip if not in a Git repository
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	if isRepo, _ := analyzer.IsGitRepositoryWithContext(ctx); !isRepo {
		b.Skip("Not in a Git repository")
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		if _, err := analyzer.GetLatestCommitWithContext(ctx); err != nil {
			cancel()
			b.Fatal(err)
		}
		
		cancel()
	}
}

// BenchmarkAnalysisRecords benchmarks record analysis
func BenchmarkAnalysisRecords(b *testing.B) {
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  []string{".go", ".js", ".py"},
		ExcludePatterns:    []string{"*_test.go"},
		AuthorMappings: map[string]string{
			"AI Assistant": "ai",
		},
	}
	
	analyzer := tracker.NewAnalyzer(config)
	
	// Create sample records
	records := make([]tracker.CheckpointRecord, 100)
	baseTime := time.Now().Add(-24 * time.Hour)
	
	for i := 0; i < 100; i++ {
		author := "human"
		if i%3 == 0 {
			author = "AI Assistant"
		}
		
		records[i] = tracker.CheckpointRecord{
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Author:    author,
			Added:     i * 10,
			Deleted:   i * 2,
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		if _, err := analyzer.AnalyzeRecords(records); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLargeConfigValidation benchmarks validation with large configurations
func BenchmarkLargeConfigValidation(b *testing.B) {
	validator := validation.NewConfigValidator()
	
	// Create a configuration with many extensions and patterns
	extensions := make([]string, 50)
	for i := 0; i < 50; i++ {
		extensions[i] = fmt.Sprintf(".ext%d", i)
	}
	
	patterns := make([]string, 100)
	for i := 0; i < 100; i++ {
		patterns[i] = fmt.Sprintf("*_pattern%d.*", i)
	}
	
	mappings := make(map[string]string)
	for i := 0; i < 50; i++ {
		mappings[fmt.Sprintf("Author%d", i)] = "ai"
	}
	
	config := &tracker.Config{
		TargetAIPercentage: 80.0,
		TrackedExtensions:  extensions,
		ExcludePatterns:    patterns,
		AuthorMappings:     mappings,
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(config); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryUsage measures memory allocations
func BenchmarkMemoryUsage(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench_memory")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	storage, err := storage.NewJSONStorageV2(tmpDir)
	if err != nil {
		b.Fatal(err)
	}
	
	data := &tracker.AnalysisResult{
		TotalLines:  1000,
		AILines:     800,
		HumanLines:  200,
		Percentage:  80.0,
		LastUpdated: time.Now(),
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		filename := fmt.Sprintf("memory_test_%d.json", i%10)
		
		if err := storage.Save(filename, data); err != nil {
			b.Fatal(err)
		}
		
		var loaded tracker.AnalysisResult
		if err := storage.Load(filename, &loaded); err != nil {
			b.Fatal(err)
		}
	}
}