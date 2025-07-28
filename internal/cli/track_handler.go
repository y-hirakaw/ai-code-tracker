package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
	"github.com/y-hirakaw/ai-code-tracker/internal/utils"
	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// TrackHandler はtrackコマンドを処理する
type TrackHandler struct{}

// NewTrackHandler は新しいTrackHandlerを作成する
func NewTrackHandler() *TrackHandler {
	return &TrackHandler{}
}

// Handle はtrackコマンドを実行する
func (h *TrackHandler) Handle(args []string) error {
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
			return errors.NewError(errors.ErrorTypeCommand, "missing_required_option", "--author").
				WithSuggestions(i18n.T("suggestion_specify_author", "--author オプションで作成者を指定してください"))
		}
	}

	if isAI && model == "" {
		model = "claude-code" // デフォルトモデル
	}

	// 現在のディレクトリを取得
	currentDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	// データディレクトリのパス
	dataDir := filepath.Join(currentDir, storage.DefaultDataDir)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		// データディレクトリが存在しない場合は作成
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return errors.WrapError(err, errors.ErrorTypeData, "create_data_directory_failed")
		}
	}

	// DuckDBストレージを使用（periodコマンドと統一）
	config := storage.StorageConfig{
		Type:    storage.StorageTypeDuckDB,
		DataDir: dataDir,
		Debug:   os.Getenv("AICT_DEBUG") == "1",
	}

	store, err := storage.NewAdvancedStorageByType(config)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer store.Close()

	// 既存のJSONLデータをDuckDBに移行
	if err := storage.MigrateJSONLToDuckDB(dataDir, os.Getenv("AICT_DEBUG") == "1"); err != nil {
		fmt.Printf("⚠️  データ移行エラー: %v\n", err)
		// 移行エラーは致命的ではないので続行
	}

	tracker := tracker.NewTracker(store, currentDir)

	// ファイルリストを処理
	var files []string
	if filesStr != "" {
		files = utils.SplitAndTrim(filesStr, ",")
	} else {
		// ファイルが指定されていない場合、変更されたファイルを自動検出
		detectedFiles, err := tracker.DetectChangedFiles()
		if err != nil {
			return errors.WrapError(err, errors.ErrorTypeGit, "git_command_failed", "git diff")
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
		return errors.WrapError(err, errors.ErrorTypeData, "tracking_failed")
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