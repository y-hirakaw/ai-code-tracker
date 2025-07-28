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
		preEdit     = false
		postEdit    = false
		sessionID   = ""
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
		case "--pre-edit":
			preEdit = true
		case "--post-edit":
			postEdit = true
		case "--session":
			if i+1 < len(args) {
				sessionID = args[i+1]
				i++
			}
		}
	}

	// pre-edit/post-editモードの処理分岐
	if preEdit {
		return h.handlePreEdit(filesStr, sessionID)
	} else if postEdit {
		return h.handlePostEdit(filesStr, sessionID, isAI, author, model, message)
	}

	// 通常のtrackモード（既存の動作）
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

// handlePreEdit は編集前の状態を記録する
func (h *TrackHandler) handlePreEdit(filesStr, sessionID string) error {
	if filesStr == "" {
		return errors.NewError(errors.ErrorTypeCommand, "missing_required_option", "--files").
			WithSuggestions("--pre-edit モードでは --files オプションが必要です")
	}

	// セッションIDを生成（未指定の場合）
	if sessionID == "" {
		sessionID = utils.GenerateSessionID()
	}

	// 現在のディレクトリを取得
	currentDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	// データディレクトリのパス
	dataDir := filepath.Join(currentDir, storage.DefaultDataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "create_data_directory_failed")
	}

	// セッションディレクトリを作成
	sessionDir := filepath.Join(dataDir, "sessions")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "create_session_directory_failed")
	}

	// 編集前のファイル状態を記録
	preEditPath := filepath.Join(sessionDir, sessionID+"-pre.json")
	files := utils.ParseFiles(filesStr)
	
	preEditData := map[string]interface{}{
		"session_id": sessionID,
		"timestamp":  utils.GetCurrentTimeString(),
		"files":      make(map[string]interface{}),
	}

	for _, file := range files {
		if content, err := os.ReadFile(file); err == nil {
			preEditData["files"].(map[string]interface{})[file] = map[string]interface{}{
				"content":  string(content),
				"size":     len(content),
				"exists":   true,
			}
		} else {
			preEditData["files"].(map[string]interface{})[file] = map[string]interface{}{
				"exists": false,
			}
		}
	}

	// セッション情報を保存
	if err := utils.WriteJSON(preEditPath, preEditData); err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "save_pre_edit_state_failed")
	}

	fmt.Printf("📝 編集前状態を記録しました (セッション: %s)\n", sessionID)
	return nil
}

// handlePostEdit は編集後の変更を記録する
func (h *TrackHandler) handlePostEdit(filesStr, sessionID string, isAI bool, author, model, message string) error {
	if filesStr == "" {
		return errors.NewError(errors.ErrorTypeCommand, "missing_required_option", "--files").
			WithSuggestions("--post-edit モードでは --files オプションが必要です")
	}

	if sessionID == "" {
		return errors.NewError(errors.ErrorTypeCommand, "missing_required_option", "--session").
			WithSuggestions("--post-edit モードでは --session オプションが必要です")
	}

	// デフォルト値設定
	if author == "" {
		if isAI {
			author = "Claude Code"
		} else {
			return errors.NewError(errors.ErrorTypeCommand, "missing_required_option", "--author").
				WithSuggestions("--author オプションで作成者を指定してください")
		}
	}

	if isAI && model == "" {
		model = "claude-code"
	}

	// 現在のディレクトリを取得
	currentDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	dataDir := filepath.Join(currentDir, storage.DefaultDataDir)
	sessionDir := filepath.Join(dataDir, "sessions")
	preEditPath := filepath.Join(sessionDir, sessionID+"-pre.json")

	// 編集前状態を読み込み
	_, err = utils.ReadJSON(preEditPath)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "load_pre_edit_state_failed").
			WithSuggestions(fmt.Sprintf("セッション %s の編集前状態が見つかりません", sessionID))
	}

	// 通常のトラッキング処理を実行
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
	}

	// トラッカーを初期化
	tracker := tracker.NewTracker(store, currentDir)

	// イベントタイプを決定
	eventType := types.EventTypeHuman
	if isAI {
		eventType = types.EventTypeAI
	}

	// ファイル処理
	files := utils.ParseFiles(filesStr)
	if len(files) == 0 {
		fmt.Println("追跡するファイルがありません")
		return nil
	}

	// トラッキングを実行
	err = tracker.TrackFileChanges(eventType, author, model, files, message)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "tracking_failed")
	}


	// セッションファイルをクリーンアップ
	os.Remove(preEditPath)

	fmt.Printf("✅ 編集後変更を記録しました (セッション: %s)\n📊 %d個のファイルを処理\n", 
		sessionID, len(files))

	return nil
}