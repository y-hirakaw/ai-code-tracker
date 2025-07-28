package cli

import (
	"fmt"
	"path/filepath"

	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/internal/ui"
	"github.com/ai-code-tracker/aict/internal/utils"
)

// InitHandler はinitコマンドを処理する
type InitHandler struct{}

// NewInitHandler は新しいInitHandlerを作成する
func NewInitHandler() *InitHandler {
	return &InitHandler{}
}

// Handle はinitコマンドを実行する
func (h *InitHandler) Handle(args []string) error {
	// 現在のディレクトリがGitリポジトリかチェック
	currentDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return err
	}

	if !utils.IsGitRepository(currentDir) {
		return errors.GitNotRepository()
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
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

// VersionHandler はversionコマンドを処理する
type VersionHandler struct{}

// NewVersionHandler は新しいVersionHandlerを作成する
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// Handle はversionコマンドを実行する
func (h *VersionHandler) Handle(args []string) error {
	fmt.Printf("%s version %s\n", AppName, Version)
	fmt.Println("AI Code Tracker - AIと人間によるコード変更の自動追跡システム")
	return nil
}

// HelpHandler はhelpコマンドを処理する
type HelpHandler struct {
	helpSystem *ui.HelpSystem
}

// NewHelpHandler は新しいHelpHandlerを作成する
func NewHelpHandler(helpSystem *ui.HelpSystem) *HelpHandler {
	return &HelpHandler{
		helpSystem: helpSystem,
	}
}

// Handle はhelpコマンドを実行する
func (h *HelpHandler) Handle(args []string) error {
	if len(args) > 0 {
		// 特定のコマンドのヘルプを表示
		h.helpSystem.ShowCommandHelp(args[0])
	} else {
		// メインヘルプを表示
		h.helpSystem.ShowMainHelp()
	}
	return nil
}

// ConfigHandler はconfigコマンドを処理する
type ConfigHandler struct{}

// NewConfigHandler は新しいConfigHandlerを作成する
func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{}
}

// Handle はconfigコマンドを実行する
func (h *ConfigHandler) Handle(args []string) error {
	fmt.Println("設定機能は今後のバージョンで実装予定です")
	return nil
}