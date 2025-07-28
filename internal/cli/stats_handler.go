package cli

import (
	"time"

	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/stats"
	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/internal/utils"
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
		format  = "table"
		since   = ""
		until   = ""
		author  = ""
		byFile  = false
		trend   = false
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
		case "--until":
			if i+1 < len(args) {
				until = args[i+1]
				i++
			}
		case "--author":
			if i+1 < len(args) {
				author = args[i+1]
				i++
			}
		case "--by-file":
			byFile = true
		case "--trend":
			trend = true
		}
	}

	// 日付のパース
	var sinceTime, untilTime time.Time
	var err error

	if since != "" {
		sinceTime, err = utils.ParseDate(since)
		if err != nil {
			return err.(*errors.FriendlyError).WithCommand("stats")
		}
	} else {
		// デフォルトは30日前から
		sinceTime = time.Now().AddDate(0, 0, -30)
	}

	if until != "" {
		untilTime, err = utils.ParseDate(until)
		if err != nil {
			return err.(*errors.FriendlyError).WithCommand("stats")
		}
	} else {
		// デフォルトは現在まで
		untilTime = time.Now()
	}

	// ストレージを初期化
	storage, err := storage.NewStorage("")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_initialization_failed")
	}
	defer storage.Close()

	// StatsManagerを初期化
	statsManager := stats.NewStatsManager(storage)

	// フォーマット別処理
	switch format {
	case "daily":
		return h.presenter.ShowDailyStats(statsManager, sinceTime, untilTime)
	case "files":
		return h.presenter.ShowFileStats(statsManager, sinceTime, author)
	case "contributors":
		return h.presenter.ShowContributorStats(statsManager, sinceTime, author)
	}

	// トレンド分析
	if trend {
		return h.presenter.ShowTrendAnalysis(statsManager, sinceTime, untilTime)
	}

	// ファイル別統計
	if byFile {
		return h.presenter.ShowFileStats(statsManager, sinceTime, author)
	}

	// 基本統計情報を取得
	basicStats, err := storage.GetStatistics()
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