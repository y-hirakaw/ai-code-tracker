package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/internal/tracker"
	"github.com/ai-code-tracker/aict/pkg/types"
)

const (
	// Version はアプリケーションのバージョン
	Version = "0.1.0"
	// AppName はアプリケーション名
	AppName = "aict"
)

// CLI コマンドの定義
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// main はアプリケーションのエントリーポイント
func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	commands := map[string]Command{
		"track": {
			Name:        "track",
			Description: "ファイルの変更を手動でトラッキングする",
			Handler:     handleTrack,
		},
		"init": {
			Name:        "init",
			Description: "プロジェクトでAI Code Trackerを初期化する",
			Handler:     handleInit,
		},
		"stats": {
			Name:        "stats",
			Description: "統計情報を表示する",
			Handler:     handleStats,
		},
		"blame": {
			Name:        "blame",
			Description: "ファイルのAI/人間による変更履歴を表示する",
			Handler:     handleBlame,
		},
		"config": {
			Name:        "config",
			Description: "設定を管理する",
			Handler:     handleConfig,
		},
		"version": {
			Name:        "version",
			Description: "バージョン情報を表示する",
			Handler:     handleVersion,
		},
		"help": {
			Name:        "help",
			Description: "ヘルプを表示する",
			Handler:     func(args []string) error { showHelp(); return nil },
		},
	}

	cmd, exists := commands[command]
	if !exists {
		fmt.Fprintf(os.Stderr, "不明なコマンド: %s\n\n", command)
		showHelp()
		os.Exit(1)
	}

	if err := cmd.Handler(args); err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}
}

// showHelp はヘルプメッセージを表示する
func showHelp() {
	fmt.Printf(`%s v%s - AI Code Tracker

使用方法:
  %s <command> [options]

コマンド:
  init                プロジェクトでAI Code Trackerを初期化
  track               ファイルの変更を手動でトラッキング
    --ai              AI による変更として記録
    --author <name>   作成者を指定
    --model <model>   AI モデルを指定
    --files <files>   追跡するファイルを指定（カンマ区切り）
    --message <msg>   変更の説明
  stats               統計情報を表示
    --format <format> 出力形式 (table|json|summary)
    --since <date>    指定日以降の統計 (YYYY-MM-DD)
    --author <name>   作成者でフィルタ
  blame <file>        ファイルのAI/人間による変更履歴を表示
  config              設定を管理
    --list            現在の設定を表示
    --set <key=value> 設定を変更
  version             バージョン情報を表示
  help                このヘルプを表示

例:
  %s init
  %s track --ai --model claude-sonnet-4 --files "*.go" --message "AI によるリファクタリング"
  %s track --author "John Doe" --files main.go --message "バグ修正"
  %s stats --format table --since 2024-01-01
  %s blame src/main.go
`, AppName, Version, AppName, AppName, AppName, AppName, AppName, AppName)
}

// handleInit はプロジェクトの初期化を処理する
func handleInit(args []string) error {
	// 現在のディレクトリがGitリポジトリかチェック
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("現在のディレクトリの取得に失敗しました: %w", err)
	}

	if !tracker.IsGitRepo(currentDir) {
		return fmt.Errorf("現在のディレクトリはGitリポジトリではありません")
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("ストレージの初期化に失敗しました: %w", err)
	}
	defer storage.Close()

	fmt.Println("AI Code Tracker を初期化しました")
	fmt.Printf("データディレクトリ: %s\n", filepath.Join(currentDir, ".git/ai-tracker"))
	fmt.Println("\n次のステップ:")
	fmt.Println("1. `aict track` でファイルの変更を追跡")
	fmt.Println("2. `aict stats` で統計情報を確認")
	fmt.Println("3. Git hooks の自動設定は今後のバージョンで対応予定")

	return nil
}

// handleTrack はファイルのトラッキングを処理する
func handleTrack(args []string) error {
	var (
		isAI        = false
		author      = ""
		model       = ""
		filesStr    = ""
		message     = ""
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--ai":
			isAI = true
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		case "--model":
			if i+1 < len(args) {
				model = args[i+1]
				i++
			}
		case "--files":
			if i+1 < len(args) {
				filesStr = args[i+1]
				i++
			}
		case "--message":
			if i+1 < len(args) {
				message = args[i+1]
				i++
			}
		}
	}

	// パラメータの検証
	if author == "" {
		if isAI {
			author = "Claude Code"
		} else {
			return fmt.Errorf("--author が必須です")
		}
	}

	if isAI && model == "" {
		model = "claude-code" // デフォルトモデル
	}

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("現在のディレクトリの取得に失敗しました: %w", err)
	}

	// ストレージとトラッカーを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("ストレージの初期化に失敗しました: %w", err)
	}
	defer storage.Close()

	tracker := tracker.NewTracker(storage, currentDir)

	// ファイルリストを処理
	var files []string
	if filesStr != "" {
		files = strings.Split(filesStr, ",")
		for i, file := range files {
			files[i] = strings.TrimSpace(file)
		}
	} else {
		// ファイルが指定されていない場合、変更されたファイルを自動検出
		detectedFiles, err := tracker.DetectChangedFiles()
		if err != nil {
			return fmt.Errorf("変更ファイルの検出に失敗しました: %w", err)
		}
		files = detectedFiles
	}

	if len(files) == 0 {
		fmt.Println("追跡するファイルがありません")
		return nil
	}

	// イベントタイプを決定
	eventType := types.EventTypeHuman
	if isAI {
		eventType = types.EventTypeAI
	}

	// トラッキングを実行
	err = tracker.TrackFileChanges(eventType, author, model, files, message)
	if err != nil {
		return fmt.Errorf("トラッキングに失敗しました: %w", err)
	}

	fmt.Printf("✓ %d個のファイルの変更を追跡しました\n", len(files))
	for _, file := range files {
		fmt.Printf("  - %s\n", file)
	}
	fmt.Printf("作成者: %s\n", author)
	if isAI {
		fmt.Printf("モデル: %s\n", model)
	}
	if message != "" {
		fmt.Printf("メッセージ: %s\n", message)
	}

	return nil
}

// handleStats は統計情報の表示を処理する
func handleStats(args []string) error {
	var (
		format = "table"
		since  = ""
		author = ""
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--since":
			if i+1 < len(args) {
				since = args[i+1]
				i++
			}
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		}
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return fmt.Errorf("ストレージの初期化に失敗しました: %w", err)
	}
	defer storage.Close()

	// 統計情報を取得
	stats, err := storage.GetStatistics()
	if err != nil {
		return fmt.Errorf("統計情報の取得に失敗しました: %w", err)
	}

	// フィルタリング処理（簡単な実装）
	if since != "" || author != "" {
		fmt.Printf("注意: フィルタリング機能は今後のバージョンで実装予定です\n\n")
	}

	// 出力形式に応じて表示
	switch format {
	case "table":
		showStatsTable(stats)
	case "json":
		showStatsJSON(stats)
	case "summary":
		showStatsSummary(stats)
	default:
		return fmt.Errorf("不明な出力形式: %s", format)
	}

	return nil
}

// showStatsTable はテーブル形式で統計を表示する
func showStatsTable(stats *types.Statistics) {
	fmt.Println("=== AI Code Tracker 統計情報 ===\n")
	
	fmt.Printf("%-20s: %d\n", "総イベント数", stats.TotalEvents)
	fmt.Printf("%-20s: %d (%.1f%%)\n", "AI イベント", stats.AIEvents, stats.AIPercentage())
	fmt.Printf("%-20s: %d (%.1f%%)\n", "人間 イベント", stats.HumanEvents, stats.HumanPercentage())
	fmt.Printf("%-20s: %d\n", "コミット イベント", stats.CommitEvents)
	fmt.Println()
	
	fmt.Printf("%-20s: %d\n", "追加行数", stats.TotalLinesAdded)
	fmt.Printf("%-20s: %d\n", "変更行数", stats.TotalLinesModified)
	fmt.Printf("%-20s: %d\n", "削除行数", stats.TotalLinesDeleted)
	fmt.Printf("%-20s: %d\n", "総変更行数", stats.TotalChanges())
	fmt.Println()
	
	if stats.FirstEvent != nil {
		fmt.Printf("%-20s: %s\n", "最初のイベント", stats.FirstEvent.Format("2006-01-02 15:04:05"))
	}
	if stats.LastEvent != nil {
		fmt.Printf("%-20s: %s\n", "最後のイベント", stats.LastEvent.Format("2006-01-02 15:04:05"))
	}
}

// showStatsJSON はJSON形式で統計を表示する
func showStatsJSON(stats *types.Statistics) {
	fmt.Printf(`{
  "total_events": %d,
  "ai_events": %d,
  "human_events": %d,
  "commit_events": %d,
  "ai_percentage": %.1f,
  "human_percentage": %.1f,
  "total_lines_added": %d,
  "total_lines_modified": %d,
  "total_lines_deleted": %d,
  "total_changes": %d`,
		stats.TotalEvents,
		stats.AIEvents,
		stats.HumanEvents,
		stats.CommitEvents,
		stats.AIPercentage(),
		stats.HumanPercentage(),
		stats.TotalLinesAdded,
		stats.TotalLinesModified,
		stats.TotalLinesDeleted,
		stats.TotalChanges())

	if stats.FirstEvent != nil {
		fmt.Printf(`,
  "first_event": "%s"`, stats.FirstEvent.Format("2006-01-02T15:04:05Z07:00"))
	}
	if stats.LastEvent != nil {
		fmt.Printf(`,
  "last_event": "%s"`, stats.LastEvent.Format("2006-01-02T15:04:05Z07:00"))
	}

	fmt.Println("\n}")
}

// showStatsSummary はサマリー形式で統計を表示する
func showStatsSummary(stats *types.Statistics) {
	fmt.Println("📊 AI Code Tracker サマリー")
	fmt.Println(strings.Repeat("=", 30))
	
	if stats.TotalEvents == 0 {
		fmt.Println("まだイベントが記録されていません")
		return
	}
	
	fmt.Printf("🤖 AI によるコード: %.1f%% (%d イベント)\n", stats.AIPercentage(), stats.AIEvents)
	fmt.Printf("👤 人間によるコード: %.1f%% (%d イベント)\n", stats.HumanPercentage(), stats.HumanEvents)
	fmt.Printf("📝 総変更行数: %d 行\n", stats.TotalChanges())
	
	if stats.FirstEvent != nil && stats.LastEvent != nil {
		duration := stats.LastEvent.Sub(*stats.FirstEvent)
		fmt.Printf("📅 追跡期間: %d 日間\n", int(duration.Hours()/24))
	}
}

// handleBlame はファイルのblame情報を表示する（今後実装）
func handleBlame(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("ファイルパスが必要です")
	}

	fmt.Printf("Blame機能は今後のバージョンで実装予定です\n")
	fmt.Printf("対象ファイル: %s\n", args[0])
	
	return nil
}

// handleConfig は設定管理を処理する（今後実装）
func handleConfig(args []string) error {
	fmt.Println("設定機能は今後のバージョンで実装予定です")
	return nil
}

// handleVersion はバージョン情報を表示する
func handleVersion(args []string) error {
	fmt.Printf("%s version %s\n", AppName, Version)
	fmt.Println("AI Code Tracker - AIと人間によるコード変更の自動追跡システム")
	return nil
}