package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
)

// ContextHelpProvider はコンテキストに応じたヘルプを提供する
type ContextHelpProvider struct {
	appName string
}

// NewContextHelpProvider は新しいContextHelpProviderを作成する
func NewContextHelpProvider(appName string) *ContextHelpProvider {
	return &ContextHelpProvider{
		appName: appName,
	}
}

// CommandContext は実行コンテキスト情報
type CommandContext struct {
	Command    string
	Args       []string
	Error      error
	ErrorType  errors.ErrorType
	WorkingDir string
	GitRepo    bool
}

// GetContextualHelp はコンテキストに応じたヘルプメッセージを生成する
func (c *ContextHelpProvider) GetContextualHelp(ctx *CommandContext) []string {
	var suggestions []string

	// エラータイプ別の基本的な提案
	switch ctx.ErrorType {
	case errors.ErrorTypeFile:
		suggestions = append(suggestions, c.getFileErrorSuggestions(ctx)...)
	case errors.ErrorTypeGit:
		suggestions = append(suggestions, c.getGitErrorSuggestions(ctx)...)
	case errors.ErrorTypeCommand:
		suggestions = append(suggestions, c.getCommandErrorSuggestions(ctx)...)
	case errors.ErrorTypeData:
		suggestions = append(suggestions, c.getDataErrorSuggestions(ctx)...)
	case errors.ErrorTypeSecurity:
		suggestions = append(suggestions, c.getSecurityErrorSuggestions(ctx)...)
	}

	// コマンド固有の提案
	suggestions = append(suggestions, c.getCommandSpecificSuggestions(ctx)...)

	// 環境固有の提案
	suggestions = append(suggestions, c.getEnvironmentSuggestions(ctx)...)

	return c.deduplicateSuggestions(suggestions)
}

// getFileErrorSuggestions はファイル関連エラーの提案を取得する
func (c *ContextHelpProvider) getFileErrorSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	if strings.Contains(ctx.Error.Error(), "permission denied") {
		suggestions = append(suggestions, i18n.T("suggestion_check_permissions", "ファイルの権限を確認してください"))
	}
	if strings.Contains(ctx.Error.Error(), "no such file") {
		suggestions = append(suggestions, i18n.T("suggestion_check_file_path", "ファイルパスが正しいか確認してください"))
	}
	if strings.Contains(ctx.Error.Error(), "directory") {
		if !ctx.GitRepo {
			suggestions = append(suggestions, i18n.T("suggestion_init_git", "Gitリポジトリを初期化してください: git init"))
		}
		suggestions = append(suggestions, fmt.Sprintf("%s init", c.appName))
	}

	return suggestions
}

// getGitErrorSuggestions はGit関連エラーの提案を取得する
func (c *ContextHelpProvider) getGitErrorSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	if !ctx.GitRepo {
		suggestions = append(suggestions, 
			i18n.T("suggestion_git_init", "現在のディレクトリでGitリポジトリを初期化してください: git init"))
	}

	if strings.Contains(ctx.Error.Error(), "not a git repository") {
		suggestions = append(suggestions, 
			i18n.T("suggestion_cd_git_repo", "Gitリポジトリのルートディレクトリに移動してください"))
	}

	if strings.Contains(ctx.Error.Error(), "git command") {
		suggestions = append(suggestions, 
			i18n.T("suggestion_git_installed", "Gitがインストールされているか確認してください"))
	}

	return suggestions
}

// getCommandErrorSuggestions はコマンド関連エラーの提案を取得する
func (c *ContextHelpProvider) getCommandErrorSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	switch ctx.Command {
	case "track":
		if c.missingRequiredArg(ctx, "--files") {
			suggestions = append(suggestions, 
				i18n.T("suggestion_track_files", "追跡するファイルを指定してください: --files \"*.go\""))
		}
		if c.missingRequiredArg(ctx, "--author") {
			suggestions = append(suggestions, 
				i18n.T("suggestion_track_author", "作成者を指定してください: --author \"Your Name\" または --ai"))
		}
		suggestions = append(suggestions, fmt.Sprintf("%s help track", c.appName))

	case "stats":
		if strings.Contains(ctx.Error.Error(), "date") {
			suggestions = append(suggestions, 
				i18n.T("suggestion_date_format", "日付形式は YYYY-MM-DD を使用してください（例: 2024-01-01）"))
		}
		if strings.Contains(ctx.Error.Error(), "format") {
			suggestions = append(suggestions, 
				i18n.T("suggestion_stats_formats", "利用可能な形式: table, json, summary, daily, files, contributors"))
		}

	case "blame":
		suggestions = append(suggestions, 
			i18n.T("suggestion_blame_file", "ファイルがGitで追跡されているか確認してください"))
		suggestions = append(suggestions, 
			i18n.T("suggestion_blame_path", "相対パスまたは絶対パスでファイルを指定してください"))

	case "setup":
		suggestions = append(suggestions, 
			i18n.T("suggestion_setup_permissions", "hooks設定には書き込み権限が必要です"))
		suggestions = append(suggestions, 
			i18n.T("suggestion_setup_status", "現在の設定状況は --status で確認できます"))

	case "init":
		if strings.Contains(ctx.Error.Error(), "already exists") {
			suggestions = append(suggestions, 
				i18n.T("suggestion_init_force", "既存の設定を上書きするには --force を使用してください"))
		}
	}

	return suggestions
}

// getDataErrorSuggestions はデータ関連エラーの提案を取得する
func (c *ContextHelpProvider) getDataErrorSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	if strings.Contains(ctx.Error.Error(), "storage") {
		suggestions = append(suggestions, fmt.Sprintf("%s init", c.appName))
		suggestions = append(suggestions, 
			i18n.T("suggestion_storage_permissions", "データディレクトリの権限を確認してください"))
	}

	if strings.Contains(ctx.Error.Error(), "statistics") {
		suggestions = append(suggestions, 
			i18n.T("suggestion_no_data", "データが存在しません。まずコードの変更を追跡してください"))
		suggestions = append(suggestions, fmt.Sprintf("%s track --help", c.appName))
	}

	return suggestions
}

// getSecurityErrorSuggestions はセキュリティ関連エラーの提案を取得する
func (c *ContextHelpProvider) getSecurityErrorSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	suggestions = append(suggestions, 
		i18n.T("suggestion_security_scan", "セキュリティスキャンを実行してください"))
	suggestions = append(suggestions, fmt.Sprintf("%s security status", c.appName))

	return suggestions
}

// getCommandSpecificSuggestions はコマンド固有の提案を取得する
func (c *ContextHelpProvider) getCommandSpecificSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	// 共通的な提案
	if ctx.Command != "help" {
		suggestions = append(suggestions, fmt.Sprintf("%s help %s", c.appName, ctx.Command))
	}

	// より詳細なヘルプが利用可能
	suggestions = append(suggestions, fmt.Sprintf("%s help", c.appName))

	return suggestions
}

// getEnvironmentSuggestions は環境固有の提案を取得する
func (c *ContextHelpProvider) getEnvironmentSuggestions(ctx *CommandContext) []string {
	suggestions := []string{}

	// 環境変数の提案
	if ctx.Command == "setup" || ctx.Command == "config" {
		suggestions = append(suggestions, 
			i18n.T("suggestion_env_vars", "環境変数でデフォルト設定をカスタマイズできます"))
	}

	// デバッグモードの提案
	suggestions = append(suggestions, 
		i18n.T("suggestion_debug_mode", "詳細なログは AICT_DEBUG=1 で有効化できます"))

	return suggestions
}

// missingRequiredArg は必須引数が不足しているかチェック
func (c *ContextHelpProvider) missingRequiredArg(ctx *CommandContext, arg string) bool {
	return strings.Contains(ctx.Error.Error(), arg) && 
		   strings.Contains(ctx.Error.Error(), "missing")
}

// deduplicateSuggestions は重複する提案を除去する
func (c *ContextHelpProvider) deduplicateSuggestions(suggestions []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, suggestion := range suggestions {
		if !seen[suggestion] && suggestion != "" {
			seen[suggestion] = true
			result = append(result, suggestion)
		}
	}

	return result
}

// ShowContextualError はコンテキストに応じたエラーメッセージを表示する
func (c *ContextHelpProvider) ShowContextualError(ctx *CommandContext) {
	fmt.Fprintf(os.Stderr, "❌ %s: %v\n", 
		i18n.T("error_occurred", "エラーが発生しました"), ctx.Error)

	suggestions := c.GetContextualHelp(ctx)
	if len(suggestions) > 0 {
		fmt.Fprintf(os.Stderr, "\n💡 %s:\n", 
			i18n.T("suggestions", "解決方法の提案"))
		
		for i, suggestion := range suggestions {
			if i < 3 { // 最大3つの提案を表示
				fmt.Fprintf(os.Stderr, "   • %s\n", suggestion)
			}
		}
		
		if len(suggestions) > 3 {
			fmt.Fprintf(os.Stderr, "   %s %d %s\n",
				i18n.T("and_more", "他"), len(suggestions)-3, 
				i18n.T("more_suggestions", "個の提案があります"))
		}
	}
}

// GetQuickHelp は簡潔なヘルプメッセージを取得する
func (c *ContextHelpProvider) GetQuickHelp(command string) string {
	quickHelps := map[string]string{
		"init":    i18n.T("quick_help_init", "プロジェクトを初期化します"),
		"track":   i18n.T("quick_help_track", "ファイルの変更を追跡します"),
		"stats":   i18n.T("quick_help_stats", "統計情報を表示します"),
		"blame":   i18n.T("quick_help_blame", "ファイルの変更履歴を表示します"),
		"setup":   i18n.T("quick_help_setup", "hooks を自動設定します"),
		"config":  i18n.T("quick_help_config", "設定を管理します"),
		"wizard":  i18n.T("quick_help_wizard", "セットアップウィザードを実行します"),
		"version": i18n.T("quick_help_version", "バージョン情報を表示します"),
		"help":    i18n.T("quick_help_help", "ヘルプを表示します"),
	}

	if help, exists := quickHelps[command]; exists {
		return help
	}
	return i18n.T("quick_help_unknown", "詳細は help コマンドを参照してください")
}