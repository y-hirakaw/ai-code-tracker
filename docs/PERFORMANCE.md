# AICT パフォーマンスガイド

## 📊 パフォーマンス目標

AICT は以下のパフォーマンス目標を設定しています：

### 🎯 応答時間目標
- **トラッキング処理**: 100ms 以内
- **統計生成**: 500ms 以内
- **blame表示**: 500ms 以内
- **ファイルI/O**: 1秒以内

### 📈 スループット目標
- **トラッキング**: 100 events/sec 以上
- **統計計算**: 200 events/sec 以上
- **ファイルI/O**: 50 events/sec 以上

## 🔧 パフォーマンステスト

### ベンチマーク実行方法

```bash
# 基本パフォーマンステスト
make bench-performance

# 詳細なベンチマーク
go run cmd/aict-bench/main.go --all --events=10000

# 特定機能のベンチマーク
go run cmd/aict-bench/main.go --tracking --events=5000
go run cmd/aict-bench/main.go --statistics --events=5000
go run cmd/aict-bench/main.go --fileio --events=5000

# 並行処理ベンチマーク
go run cmd/aict-bench/main.go --all --concurrent --events=1000
```

### E2Eパフォーマンステスト

```bash
# 統合テスト（パフォーマンス含む）
make test-integration

# 特定のE2Eパフォーマンステスト
go test -v ./test/integration/ -run TestE2EPerformance
```

## 📉 パフォーマンス最適化

### 1. ファイルI/O最適化

#### 現在の実装
- **JSONL形式**: 行指向で追記効率が良い
- **原子書き込み**: 一時ファイル経由で安全性確保
- **インデックス**: 高速検索のためのメタデータ

#### 最適化手法
```go
// バッファ付きファイル書き込み
func optimizedWrite(filename string, data []byte) error {
    file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // バッファリング
    writer := bufio.NewWriter(file)
    defer writer.Flush()
    
    _, err = writer.Write(data)
    return err
}
```

### 2. メモリ使用量最適化

#### オブジェクトプール
```go
var eventPool = sync.Pool{
    New: func() interface{} {
        return &types.TrackEvent{}
    },
}

func getEvent() *types.TrackEvent {
    return eventPool.Get().(*types.TrackEvent)
}

func putEvent(event *types.TrackEvent) {
    // リセット処理
    *event = types.TrackEvent{}
    eventPool.Put(event)
}
```

#### ストリーミング処理
```go
func processLargeDataset(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        // 行ごとに処理（メモリ効率）
        line := scanner.Text()
        processEvent(line)
    }
    
    return scanner.Err()
}
```

### 3. 並行処理最適化

#### Worker Pool パターン
```go
func processEventsParallel(events []*types.TrackEvent) {
    const numWorkers = 4
    eventChan := make(chan *types.TrackEvent, len(events))
    resultChan := make(chan error, len(events))
    
    // Worker起動
    for i := 0; i < numWorkers; i++ {
        go worker(eventChan, resultChan)
    }
    
    // イベント送信
    for _, event := range events {
        eventChan <- event
    }
    close(eventChan)
    
    // 結果収集
    for i := 0; i < len(events); i++ {
        if err := <-resultChan; err != nil {
            log.Printf("処理エラー: %v", err)
        }
    }
}
```

### 4. キャッシュ戦略

#### 統計キャッシュ
```go
type StatsCache struct {
    cache map[string]*types.Statistics
    mutex sync.RWMutex
    ttl   time.Duration
}

func (sc *StatsCache) GetStats(key string) (*types.Statistics, bool) {
    sc.mutex.RLock()
    defer sc.mutex.RUnlock()
    
    stats, exists := sc.cache[key]
    return stats, exists
}
```

## 📊 パフォーマンス監視

### 1. プロファイリング

```bash
# CPU プロファイル
go test -cpuprofile=cpu.prof -bench=. ./...

# メモリプロファイル
go test -memprofile=mem.prof -bench=. ./...

# プロファイル解析
go tool pprof cpu.prof
go tool pprof mem.prof
```

### 2. リアルタイム監視

```go
// パフォーマンスメトリクス
type PerformanceMetrics struct {
    TrackingDuration time.Duration
    StatsDuration    time.Duration
    MemoryUsage      uint64
    EventsProcessed  int64
}

func measurePerformance(operation func()) PerformanceMetrics {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    startMem := m.Alloc
    
    start := time.Now()
    operation()
    duration := time.Since(start)
    
    runtime.ReadMemStats(&m)
    endMem := m.Alloc
    
    return PerformanceMetrics{
        TrackingDuration: duration,
        MemoryUsage:      endMem - startMem,
    }
}
```

### 3. 継続的なパフォーマンステスト

```bash
# CI/CDでのパフォーマンスチェック
#!/bin/bash
echo "パフォーマンスベンチマーク実行中..."
go run cmd/aict-bench/main.go --all --events=1000

if [ $? -eq 0 ]; then
    echo "✅ パフォーマンス目標クリア"
else
    echo "❌ パフォーマンス目標未達"
    exit 1
fi
```

## 🎯 ベンチマーク結果の例

### 理想的な結果
```
📊 ベンチマーク結果
==========================================
✅ Tracking:
  - 実行時間: 85ms
  - スループット: 118 events/sec
  - メモリ使用量: 2.5 MB

✅ Statistics:
  - 実行時間: 450ms
  - スループット: 222 events/sec
  - メモリ使用量: 1.8 MB

✅ FileIO:
  - 実行時間: 890ms
  - スループット: 56 events/sec
  - メモリ使用量: 3.2 MB

🎉 全てのパフォーマンス目標をクリアしました！
```

### 最適化が必要な場合
```
📊 ベンチマーク結果
==========================================
❌ Tracking:
  - 実行時間: 150ms (目標: 100ms) ❌
  - スループット: 67 events/sec (目標: 100) ❌
  - メモリ使用量: 5.2 MB

⚠️ 一部のパフォーマンス目標を下回りました。最適化が必要です。
```

## 🚀 パフォーマンス改善の手順

### 1. 問題の特定
```bash
# プロファイルの取得
go test -bench=. -cpuprofile=cpu.prof ./internal/tracker/

# ボトルネックの分析
go tool pprof cpu.prof
```

### 2. 最適化の実装
- ホットスポットの特定
- アルゴリズムの改善
- データ構造の最適化
- 並行処理の導入

### 3. 効果の測定
```bash
# 最適化前後の比較
go run cmd/aict-bench/main.go --all --events=1000 > before.txt
# (最適化実装)
go run cmd/aict-bench/main.go --all --events=1000 > after.txt

# 結果比較
diff before.txt after.txt
```

## 📋 パフォーマンス設定

### 環境変数での調整
```bash
# 並行処理数の調整
export AICT_MAX_WORKERS=8

# バッファサイズの調整
export AICT_BUFFER_SIZE=4096

# キャッシュサイズの調整
export AICT_CACHE_SIZE=1000
```

### 設定ファイルでの調整
```json
{
  "performance": {
    "max_workers": 4,
    "buffer_size": 4096,
    "cache_ttl": "5m",
    "batch_size": 100,
    "enable_compression": false
  }
}
```

このパフォーマンスガイドに従って、AICT の性能を継続的に監視・改善していくことで、大規模なプロジェクトでも安定した動作を確保できます。