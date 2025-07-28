package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/internal/tracker"
	"github.com/ai-code-tracker/aict/pkg/types"
)

// ベンチマーク設定
type BenchConfig struct {
	NumEvents     int
	NumFiles      int
	RepoPath      string
	OutputFormat  string
	Concurrent    bool
	MaxGoroutines int
}

// ベンチマーク結果
type BenchResult struct {
	Operation     string
	Duration      time.Duration
	EventsPerSec  float64
	MemoryUsage   uint64
	Success       bool
	Error         error
}

// メモリ使用量の測定
func getMemUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// トラッキングベンチマーク
func benchmarkTracking(config *BenchConfig) *BenchResult {
	start := time.Now()
	startMem := getMemUsage()

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "aict-bench-*")
	if err != nil {
		return &BenchResult{
			Operation: "Tracking",
			Success:   false,
			Error:     err,
		}
	}
	defer os.RemoveAll(tempDir)

	// Gitリポジトリの初期化
	gitDir := filepath.Join(tempDir, ".git")
	os.MkdirAll(gitDir, 0755)

	// ストレージとトラッカーの初期化
	storageManager := storage.NewStorageManager(tempDir)
	trackingManager := tracker.NewTracker(tempDir)

	// イベントの生成と追加
	for i := 0; i < config.NumEvents; i++ {
		event := &types.TrackEvent{
			ID:        fmt.Sprintf("bench-event-%d", i),
			Timestamp: time.Now(),
			EventType: types.EventTypeAI,
			Author:    "Claude Sonnet 4",
			Model:     "claude-sonnet-4",
			Message:   fmt.Sprintf("ベンチマークイベント %d", i),
			Files: []types.FileInfo{
				{
					Path:         fmt.Sprintf("test/file_%d.go", i%config.NumFiles),
					LinesAdded:    10 + i%50,
					LinesDeleted:  i % 20,
					LinesModified: 5 + i%30,
				},
			},
		}

		if config.Concurrent {
			// 並行処理でのトラッキング
			go func(e *types.TrackEvent) {
				trackingManager.Track(e)
			}(event)
		} else {
			// 順次処理でのトラッキング
			if err := trackingManager.Track(event); err != nil {
				return &BenchResult{
					Operation: "Tracking",
					Success:   false,
					Error:     err,
				}
			}
		}
	}

	// 並行処理の場合は少し待機
	if config.Concurrent {
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(start)
	endMem := getMemUsage()
	eventsPerSec := float64(config.NumEvents) / duration.Seconds()

	return &BenchResult{
		Operation:    "Tracking",
		Duration:     duration,
		EventsPerSec: eventsPerSec,
		MemoryUsage:  endMem - startMem,
		Success:      true,
	}
}

// 統計生成ベンチマーク
func benchmarkStatistics(config *BenchConfig) *BenchResult {
	start := time.Now()
	startMem := getMemUsage()

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "aict-stats-bench-*")
	if err != nil {
		return &BenchResult{
			Operation: "Statistics",
			Success:   false,
			Error:     err,
		}
	}
	defer os.RemoveAll(tempDir)

	// ストレージの初期化とデータ準備
	storageManager := storage.NewStorageManager(tempDir)

	// テストデータの準備
	events := make([]*types.TrackEvent, config.NumEvents)
	for i := 0; i < config.NumEvents; i++ {
		events[i] = &types.TrackEvent{
			ID:        fmt.Sprintf("stats-event-%d", i),
			Timestamp: time.Now().Add(-time.Duration(i) * time.Hour),
			EventType: types.EventType(i%3 + 1), // AI, Human, Mixed
			Author:    fmt.Sprintf("Author-%d", i%10),
			Model:     "claude-sonnet-4",
			Message:   fmt.Sprintf("統計ベンチマークイベント %d", i),
			Files: []types.FileInfo{
				{
					Path:          fmt.Sprintf("src/module_%d.go", i%config.NumFiles),
					LinesAdded:    i % 100,
					LinesDeleted:  i % 50,
					LinesModified: i % 75,
				},
			},
		}
		storageManager.SaveEvent(events[i])
	}

	// 統計生成の実行
	stats, err := storageManager.GetStatistics()
	if err != nil {
		return &BenchResult{
			Operation: "Statistics",
			Success:   false,
			Error:     err,
		}
	}

	// 統計の検証
	if stats.TotalEvents != config.NumEvents {
		return &BenchResult{
			Operation: "Statistics",
			Success:   false,
			Error:     fmt.Errorf("期待されるイベント数 %d, 実際 %d", config.NumEvents, stats.TotalEvents),
		}
	}

	duration := time.Since(start)
	endMem := getMemUsage()
	eventsPerSec := float64(config.NumEvents) / duration.Seconds()

	return &BenchResult{
		Operation:    "Statistics",
		Duration:     duration,
		EventsPerSec: eventsPerSec,
		MemoryUsage:  endMem - startMem,
		Success:      true,
	}
}

// ファイルI/Oベンチマーク
func benchmarkFileIO(config *BenchConfig) *BenchResult {
	start := time.Now()
	startMem := getMemUsage()

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "aict-io-bench-*")
	if err != nil {
		return &BenchResult{
			Operation: "FileIO",
			Success:   false,
			Error:     err,
		}
	}
	defer os.RemoveAll(tempDir)

	storageManager := storage.NewStorageManager(tempDir)

	// 大量のイベントを書き込み、読み込み
	writeStart := time.Now()
	for i := 0; i < config.NumEvents; i++ {
		event := &types.TrackEvent{
			ID:        fmt.Sprintf("io-event-%d", i),
			Timestamp: time.Now(),
			EventType: types.EventTypeAI,
			Author:    "Benchmark Test",
			Message:   fmt.Sprintf("I/Oベンチマークイベント %d", i),
		}
		
		if err := storageManager.SaveEvent(event); err != nil {
			return &BenchResult{
				Operation: "FileIO",
				Success:   false,
				Error:     fmt.Errorf("書き込みエラー: %v", err),
			}
		}
	}
	writeTime := time.Since(writeStart)

	// 読み込みベンチマーク
	readStart := time.Now()
	events, err := storageManager.GetEvents(nil)
	if err != nil {
		return &BenchResult{
			Operation: "FileIO",
			Success:   false,
			Error:     fmt.Errorf("読み込みエラー: %v", err),
		}
	}
	readTime := time.Since(readStart)

	if len(events) != config.NumEvents {
		return &BenchResult{
			Operation: "FileIO",
			Success:   false,
			Error:     fmt.Errorf("期待されるイベント数 %d, 実際 %d", config.NumEvents, len(events)),
		}
	}

	totalDuration := time.Since(start)
	endMem := getMemUsage()
	eventsPerSec := float64(config.NumEvents*2) / totalDuration.Seconds() // 読み書き両方

	fmt.Printf("  - 書き込み時間: %v (%.0f events/sec)\n", writeTime, float64(config.NumEvents)/writeTime.Seconds())
	fmt.Printf("  - 読み込み時間: %v (%.0f events/sec)\n", readTime, float64(config.NumEvents)/readTime.Seconds())

	return &BenchResult{
		Operation:    "FileIO",
		Duration:     totalDuration,
		EventsPerSec: eventsPerSec,
		MemoryUsage:  endMem - startMem,
		Success:      true,
	}
}

// ベンチマーク結果の表示
func printBenchResults(results []*BenchResult) {
	fmt.Println("\n📊 ベンチマーク結果")
	fmt.Println("==========================================")
	
	for _, result := range results {
		if !result.Success {
			fmt.Printf("❌ %s: FAILED - %v\n", result.Operation, result.Error)
			continue
		}

		fmt.Printf("✅ %s:\n", result.Operation)
		fmt.Printf("  - 実行時間: %v\n", result.Duration)
		fmt.Printf("  - スループット: %.0f events/sec\n", result.EventsPerSec)
		fmt.Printf("  - メモリ使用量: %.2f MB\n", float64(result.MemoryUsage)/1024/1024)
	}
}

// パフォーマンス目標の確認
func checkPerformanceTargets(results []*BenchResult) bool {
	fmt.Println("\n🎯 パフォーマンス目標チェック")
	fmt.Println("==========================================")
	
	targets := map[string]struct {
		MaxDuration time.Duration
		MinThroughput float64
	}{
		"Tracking":   {MaxDuration: 100 * time.Millisecond, MinThroughput: 100}, // 100ms以内、100 events/sec以上
		"Statistics": {MaxDuration: 500 * time.Millisecond, MinThroughput: 200}, // 500ms以内、200 events/sec以上
		"FileIO":     {MaxDuration: 1 * time.Second, MinThroughput: 50},          // 1s以内、50 events/sec以上
	}

	allPassed := true
	for _, result := range results {
		if !result.Success {
			continue
		}

		target, exists := targets[result.Operation]
		if !exists {
			continue
		}

		durationOK := result.Duration <= target.MaxDuration
		throughputOK := result.EventsPerSec >= target.MinThroughput

		status := "✅"
		if !durationOK || !throughputOK {
			status = "❌"
			allPassed = false
		}

		fmt.Printf("%s %s:\n", status, result.Operation)
		fmt.Printf("  - 実行時間: %v (目標: %v) %s\n", 
			result.Duration, target.MaxDuration, 
			map[bool]string{true: "✅", false: "❌"}[durationOK])
		fmt.Printf("  - スループット: %.0f events/sec (目標: %.0f) %s\n", 
			result.EventsPerSec, target.MinThroughput,
			map[bool]string{true: "✅", false: "❌"}[throughputOK])
	}

	return allPassed
}

func main() {
	var (
		numEvents     = flag.Int("events", 1000, "ベンチマーク用イベント数")
		numFiles      = flag.Int("files", 100, "ベンチマーク用ファイル数")
		outputFormat  = flag.String("format", "console", "出力形式 (console|json)")
		concurrent    = flag.Bool("concurrent", false, "並行処理を有効化")
		maxGoroutines = flag.Int("goroutines", runtime.NumCPU(), "最大並行処理数")
		all           = flag.Bool("all", false, "全てのベンチマークを実行")
		tracking      = flag.Bool("tracking", false, "トラッキングベンチマークを実行")
		statistics    = flag.Bool("statistics", false, "統計ベンチマークを実行")
		fileio        = flag.Bool("fileio", false, "ファイルI/Oベンチマークを実行")
	)
	flag.Parse()

	config := &BenchConfig{
		NumEvents:     *numEvents,
		NumFiles:      *numFiles,
		OutputFormat:  *outputFormat,
		Concurrent:    *concurrent,
		MaxGoroutines: *maxGoroutines,
	}

	fmt.Printf("🚀 AICT パフォーマンスベンチマーク\n")
	fmt.Printf("イベント数: %d, ファイル数: %d\n", config.NumEvents, config.NumFiles)
	fmt.Printf("並行処理: %v\n", config.Concurrent)
	fmt.Printf("CPU数: %d\n", runtime.NumCPU())
	fmt.Println("==========================================")

	var results []*BenchResult

	// ベンチマークの実行
	if *all || *tracking {
		fmt.Println("🔄 トラッキングベンチマーク実行中...")
		results = append(results, benchmarkTracking(config))
	}

	if *all || *statistics {
		fmt.Println("📈 統計ベンチマーク実行中...")
		results = append(results, benchmarkStatistics(config))
	}

	if *all || *fileio {
		fmt.Println("💾 ファイルI/Oベンチマーク実行中...")
		results = append(results, benchmarkFileIO(config))
	}

	if len(results) == 0 {
		log.Fatal("ベンチマークが指定されていません。-all または個別のフラグを使用してください。")
	}

	// 結果の表示
	printBenchResults(results)

	// パフォーマンス目標の確認
	if checkPerformanceTargets(results) {
		fmt.Println("\n🎉 全てのパフォーマンス目標をクリアしました！")
		os.Exit(0)
	} else {
		fmt.Println("\n⚠️  一部のパフォーマンス目標を下回りました。最適化が必要です。")
		os.Exit(1)
	}
}