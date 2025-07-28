package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/storage"
	"github.com/ai-code-tracker/aict/internal/utils"
)

// PeriodHandler は期間別分析コマンドを処理する
type PeriodHandler struct{}

// NewPeriodHandler は新しい PeriodHandler を作成する
func NewPeriodHandler() *PeriodHandler {
	return &PeriodHandler{}
}

// Handle は period コマンドを実行する
func (h *PeriodHandler) Handle(args []string) error {
	if len(args) == 0 {
		return errors.NewError(errors.ErrorTypeCommand, "missing_period_expression")
	}

	// 現在のディレクトリを取得
	currentDir, err := utils.GetCurrentDirectory()
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeGeneral, "get_current_directory_failed")
	}

	// Gitリポジトリかチェック
	if !utils.IsGitRepository(currentDir) {
		return errors.GitNotRepository()
	}

	// データディレクトリのパス
	dataDir := filepath.Join(currentDir, storage.DefaultDataDir)
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return errors.NoTrackingData()
	}

	// DuckDBストレージを使用
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

	// 接続テスト
	if err := store.TestConnection(); err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "storage_connection_failed")
	}

	// 期間表現を解析
	periodExpr := strings.Join(args, " ")
	startDate, endDate, err := storage.ParsePeriodExpression(periodExpr)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeCommand, "invalid_period_expression")
	}

	// 期間分析を実行
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	analysis, err := store.GetPeriodAnalysis(ctx, startDate, endDate)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "period_analysis_failed")
	}

	// エクスポート形式をチェック
	if hasExportFlag(args) {
		return h.handleExport(analysis, args)
	}

	// 分析結果を表示
	h.presentAnalysis(analysis)

	return nil
}

// presentAnalysis は期間分析結果を表示する
func (h *PeriodHandler) presentAnalysis(analysis *storage.PeriodAnalysis) {
	fmt.Printf("📊 %s\n", i18n.T("period_analysis_title"))
	fmt.Printf("📅 %s: %s - %s\n", 
		i18n.T("period"),
		analysis.StartDate.Format("2006-01-02"),
		analysis.EndDate.Format("2006-01-02"))
	
	durationDays := int(analysis.EndDate.Sub(analysis.StartDate).Hours() / 24)
	fmt.Printf("⏱️  %s: %d %s (%s: %d %s)\n",
		i18n.T("duration_days"),
		durationDays,
		i18n.T("days"),
		i18n.T("active_days"),
		analysis.ActiveDays,
		i18n.T("days"))
	
	fmt.Println("\n" + strings.Repeat("─", 50))
	
	// 全体統計
	fmt.Printf("📈 %s\n", i18n.T("overall_statistics"))
	fmt.Printf("  📝 %s: %s\n", i18n.T("total_lines"), formatNumber(analysis.TotalLines))
	fmt.Printf("  🤖 %s: %s (%.1f%%)\n", i18n.T("ai_lines"), formatNumber(analysis.AILines), analysis.AIPercentage)
	fmt.Printf("  👤 %s: %s (%.1f%%)\n", i18n.T("human_lines"), formatNumber(analysis.HumanLines), 100.0-analysis.AIPercentage)
	fmt.Printf("  📁 %s: %d\n", i18n.T("files_edited"), analysis.FileCount)
	fmt.Printf("  🔄 %s: %d\n", i18n.T("sessions"), analysis.SessionCount)
	
	// トップファイル（上位5つ）
	if len(analysis.FileBreakdown) > 0 {
		fmt.Printf("\n📂 %s\n", i18n.T("top_files"))
		count := len(analysis.FileBreakdown)
		if count > 5 {
			count = 5
		}
		
		for i := 0; i < count; i++ {
			file := analysis.FileBreakdown[i]
			fmt.Printf("  %d. %s\n", i+1, file.FilePath)
			fmt.Printf("     %s: %s (AI: %.1f%%) [%s]\n",
				i18n.T("lines"),
				formatNumber(file.TotalLines),
				file.AIPercentage,
				file.Language)
		}
		
		if len(analysis.FileBreakdown) > 5 {
			fmt.Printf("     ... %s %d %s\n", i18n.T("and"), len(analysis.FileBreakdown)-5, i18n.T("more_files"))
		}
	}
	
	// 言語統計
	if len(analysis.LanguageStats) > 0 {
		fmt.Printf("\n🏷️  %s\n", i18n.T("language_statistics"))
		for _, lang := range analysis.LanguageStats {
			fmt.Printf("  %s: %s %s (AI: %.1f%%, %d %s)\n",
				lang.Language,
				formatNumber(lang.TotalLines),
				i18n.T("lines"),
				lang.AIPercentage,
				lang.FileCount,
				i18n.T("files"))
		}
	}
	
	// 貢献者統計
	if len(analysis.ContributorStats) > 0 {
		fmt.Printf("\n👥 %s\n", i18n.T("contributor_statistics"))
		for _, contributor := range analysis.ContributorStats {
			authorType := "👤"
			if contributor.AuthorType == "ai" {
				authorType = "🤖"
			}
			
			fmt.Printf("  %s %s: %s %s (%.1f%%, %.1f %s/day)\n",
				authorType,
				contributor.Author,
				formatNumber(contributor.Lines),
				i18n.T("lines"),
				contributor.Percentage,
				contributor.AvgLinesPerDay,
				i18n.T("lines"))
		}
	}
}

// hasExportFlag は引数にエクスポートフラグが含まれているかチェック
func hasExportFlag(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "--export") || arg == "-e" {
			return true
		}
	}
	return false
}

// handleExport はエクスポート処理を行う
func (h *PeriodHandler) handleExport(analysis *storage.PeriodAnalysis, args []string) error {
	// 今回はエクスポート機能の基本実装
	// 次のタスクで詳細に実装
	fmt.Printf("🚧 %s\n", i18n.T("export_feature_coming_soon"))
	h.presentAnalysis(analysis)
	return nil
}

// formatNumber は数値をカンマ区切りでフォーマットする
func formatNumber(n int) string {
	str := strconv.Itoa(n)
	if len(str) <= 3 {
		return str
	}
	
	var result []string
	for i, char := range reverse(str) {
		if i > 0 && i%3 == 0 {
			result = append(result, ",")
		}
		result = append(result, string(char))
	}
	
	return reverse(strings.Join(result, ""))
}

// reverse は文字列を逆順にする
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}