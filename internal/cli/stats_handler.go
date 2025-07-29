package cli

import (
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
)

// StatsHandler はstatsコマンドを処理する
type StatsHandler struct {
	presenter *StatsPresenter
}

// NewStatsHandler は新しいStatsHandlerを作成する
func NewStatsHandler() *StatsHandler {
	return &StatsHandler{
		presenter: NewStatsPresenter(),
	}
}

// Handle はstatsコマンドを実行する
func (h *StatsHandler) Handle(args []string) error {
	var (
		format = "table"
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
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		}
	}

	// DuckDBストレージを直接使用
	dataDir := ".git/ai-tracker"
	duckDB, err := storage.NewDuckDBStorage(dataDir, false)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "duckdb_initialization_failed")
	}
	defer duckDB.Close()

	// 基本統計情報を取得（DuckDBから直接）
	basicStats, err := duckDB.GetStatistics()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "statistics_fetch_failed")
	}

	// 作成者フィルタ処理
	if author != "" {
		h.presenter.ShowAuthorFilter(author)
	}

	// 出力形式に応じて表示
	switch format {
	case "table":
		h.presenter.ShowStatsTable(basicStats)
	case "json":
		h.presenter.ShowStatsJSON(basicStats)
	case "summary":
		h.presenter.ShowStatsSummary(basicStats)
	default:
		return errors.NewError(errors.ErrorTypeCommand, "invalid_output_format", format).
			WithSuggestions(
				i18n.T("suggestion_valid_formats", "有効な形式: table, json, summary, daily, files, contributors"),
			).WithCommand("stats")
	}

	return nil
}