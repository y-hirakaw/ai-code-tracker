package cli

import (
	"fmt"
	"os"

	"github.com/ai-code-tracker/aict/internal/blame"
	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/storage"
)

// BlameHandler はblameコマンドを処理する
type BlameHandler struct{}

// NewBlameHandler は新しいBlameHandlerを作成する
func NewBlameHandler() *BlameHandler {
	return &BlameHandler{}
}

// Handle はblameコマンドを実行する
func (h *BlameHandler) Handle(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrorTypeCommand, "missing_required_argument", "file_path").
			WithSuggestions(
				i18n.T("suggestion_specify_file", "例: aict blame src/main.go"),
			).WithCommand("blame")
	}

	var (
		filePath  = args[0]
		useColor  = true
		showStats = false
		topN      = 0
	)

	// コマンドライン引数をパース
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--no-color":
			useColor = false
		case "--stats":
			showStats = true
		case "--top":
			if i+1 < len(args) {
				if n, err := fmt.Sscanf(args[i+1], "%d", &topN); n == 1 && err == nil {
					i++
				}
			}
		}
	}

	// 現在のディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "directory_access_failed")
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer storage.Close()

	// Blamerを初期化
	blamer := blame.NewBlamer(storage, currentDir)

	// ファイルパスを検証
	if err := blamer.ValidateFilePath(filePath); err != nil {
		return errors.FileNotFound(filePath).WithCommand("blame")
	}

	if showStats || topN > 0 {
		// 統計情報または上位貢献者を表示
		if topN > 0 {
			contributors, err := blamer.GetTopContributors(filePath, topN)
			if err != nil {
				return errors.WrapError(err, errors.ErrorTypeData, "contributor_fetch_failed")
			}

			fmt.Printf("=== %s の上位貢献者 ===\n\n", filePath)
			for i, contributor := range contributors {
				indicator := "👤"
				if contributor.IsAI {
					indicator = "🤖"
				}
				fmt.Printf("%d. %s %s - %d行 (%.1f%%)\n", 
					i+1, indicator, contributor.Name, contributor.Lines, contributor.Percentage)
			}
		} else {
			// 貢献者別統計のみ表示
			contribution, err := blamer.GetFileContribution(filePath)
			if err != nil {
				return errors.WrapError(err, errors.ErrorTypeData, "contributor_fetch_failed")
			}

			fmt.Printf("=== %s の貢献者統計 ===\n\n", filePath)
			for contributor, lines := range contribution {
				fmt.Printf("%-20s: %d行\n", contributor, lines)
			}
		}
	} else {
		// 通常のblame表示
		result, err := blamer.BlameFile(filePath)
		if err != nil {
			return errors.WrapError(err, errors.ErrorTypeGit, "blame_fetch_failed")
		}

		// フォーマットして出力
		output := blamer.FormatBlameOutput(result, useColor)
		fmt.Print(output)
	}

	return nil
}