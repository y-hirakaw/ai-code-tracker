package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ai-code-tracker/aict/pkg/types"
)

// MigrateJSONLToDuckDB は既存のJSONLデータをDuckDBに移行する
func MigrateJSONLToDuckDB(dataDir string, debug bool) error {
	jsonlPath := filepath.Join(dataDir, "events.jsonl")
	
	// JSONLファイルの存在チェック
	if _, err := os.Stat(jsonlPath); os.IsNotExist(err) {
		// JSONLファイルが存在しない場合は移行不要
		return nil
	}

	fmt.Println("🔄 JSONLデータをDuckDBに移行中...")

	// DuckDBストレージを初期化
	config := StorageConfig{
		Type:    StorageTypeDuckDB,
		DataDir: dataDir,
		Debug:   debug,
	}

	duckDB, err := NewAdvancedStorageByType(config)
	if err != nil {
		return fmt.Errorf("DuckDB初期化エラー: %w", err)
	}
	defer duckDB.Close()

	// 既存データをチェック
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stats, err := duckDB.GetBasicStats(ctx)
	if err != nil {
		return fmt.Errorf("統計取得エラー: %w", err)
	}

	if stats.TotalEvents > 0 {
		fmt.Printf("ℹ️  DuckDBには既に %d 件のイベントが存在します。スキップします。\n", stats.TotalEvents)
		return nil
	}

	// JSONLファイルを読み込み
	file, err := os.Open(jsonlPath)
	if err != nil {
		return fmt.Errorf("JSONLファイル読み込みエラー: %w", err)
	}
	defer file.Close()

	var migratedCount int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event types.TrackEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			fmt.Printf("⚠️  イベント解析エラーをスキップ: %v\n", err)
			continue
		}

		// DuckDBに保存
		if err := duckDB.StoreTrackEvent(&event); err != nil {
			fmt.Printf("⚠️  イベント保存エラーをスキップ: %v\n", err)
			continue
		}

		migratedCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("JSONLファイル読み込みエラー: %w", err)
	}

	fmt.Printf("✅ %d 件のイベントをDuckDBに移行完了\n", migratedCount)

	// バックアップファイルを作成
	backupPath := jsonlPath + ".backup." + time.Now().Format("20060102_150405")
	if err := os.Rename(jsonlPath, backupPath); err != nil {
		fmt.Printf("⚠️  バックアップファイル作成エラー: %v\n", err)
	} else {
		fmt.Printf("📁 JSONLファイルを %s にバックアップしました\n", backupPath)
	}

	return nil
}