package errors

import (
	"fmt"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
)

// ErrorType はエラーの種類を定義する
type ErrorType int

const (
	// ErrorTypeGeneral は一般的なエラー
	ErrorTypeGeneral ErrorType = iota
	// ErrorTypeFile はファイル関連のエラー
	ErrorTypeFile
	// ErrorTypeGit はGit関連のエラー
	ErrorTypeGit
	// ErrorTypeCommand はコマンド関連のエラー
	ErrorTypeCommand
	// ErrorTypeData はデータ関連のエラー
	ErrorTypeData
	// ErrorTypeSecurity はセキュリティ関連のエラー
	ErrorTypeSecurity
	// ErrorTypeConfig は設定関連のエラー
	ErrorTypeConfig
	// ErrorTypeNetwork はネットワーク関連のエラー
	ErrorTypeNetwork
)

// FriendlyError はユーザーフレンドリーなエラー
type FriendlyError struct {
	Type         ErrorType
	Key          string
	Args         []interface{}
	Cause        error
	Suggestions  []string
	Command      string
	recoverable  bool
	documentation string
}

// Error は error インターフェースを実装する
func (e *FriendlyError) Error() string {
	return i18n.T(e.Key, e.Args...)
}

// Unwrap は内部エラーを返す
func (e *FriendlyError) Unwrap() error {
	return e.Cause
}

// GetMessage は翻訳されたメッセージを取得する
func (e *FriendlyError) GetMessage() string {
	return i18n.T(e.Key, e.Args...)
}

// GetSuggestions は解決策の提案を取得する
func (e *FriendlyError) GetSuggestions() []string {
	return e.Suggestions
}

// IsRecoverable はエラーが回復可能かどうかを返す
func (e *FriendlyError) IsRecoverable() bool {
	return e.recoverable
}

// GetDocumentation はドキュメントリンクを取得する
func (e *FriendlyError) GetDocumentation() string {
	return e.documentation
}

// NewError は新しいフレンドリーエラーを作成する
func NewError(errorType ErrorType, key string, args ...interface{}) *FriendlyError {
	return &FriendlyError{
		Type: errorType,
		Key:  key,
		Args: args,
	}
}

// WrapError は既存のエラーをラップする
func WrapError(cause error, errorType ErrorType, key string, args ...interface{}) *FriendlyError {
	return &FriendlyError{
		Type:  errorType,
		Key:   key,
		Args:  args,
		Cause: cause,
	}
}

// WithSuggestions は提案を追加する
func (e *FriendlyError) WithSuggestions(suggestions ...string) *FriendlyError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// WithCommand はコマンドコンテキストを追加する
func (e *FriendlyError) WithCommand(command string) *FriendlyError {
	e.Command = command
	return e
}

// WithRecoverable は回復可能フラグを設定する
func (e *FriendlyError) WithRecoverable(recoverable bool) *FriendlyError {
	e.recoverable = recoverable
	return e
}

// WithDocumentation はドキュメントリンクを追加する
func (e *FriendlyError) WithDocumentation(docURL string) *FriendlyError {
	e.documentation = docURL
	return e
}

// ErrorFormatter はエラーのフォーマッター
type ErrorFormatter struct {
	colorEnabled bool
	showCause    bool
	showSuggestions bool
}

// NewErrorFormatter は新しいエラーフォーマッターを作成する
func NewErrorFormatter() *ErrorFormatter {
	return &ErrorFormatter{
		colorEnabled:    true,
		showCause:       true,
		showSuggestions: true,
	}
}

// SetColorEnabled はカラー表示を設定する
func (f *ErrorFormatter) SetColorEnabled(enabled bool) {
	f.colorEnabled = enabled
}

// Format はエラーをフォーマットする
func (f *ErrorFormatter) Format(err error) string {
	if err == nil {
		return ""
	}

	var result strings.Builder
	
	if friendlyErr, ok := err.(*FriendlyError); ok {
		// フレンドリーエラーの場合
		f.formatFriendlyError(&result, friendlyErr)
	} else {
		// 通常のエラーの場合
		f.formatGenericError(&result, err)
	}
	
	return result.String()
}

// formatFriendlyError はフレンドリーエラーをフォーマットする
func (f *ErrorFormatter) formatFriendlyError(result *strings.Builder, err *FriendlyError) {
	// エラータイプのアイコン
	icon := f.getErrorIcon(err.Type)
	
	if f.colorEnabled {
		result.WriteString(f.colorRed(fmt.Sprintf("%s %s: %s", icon, i18n.T("error"), err.GetMessage())))
	} else {
		result.WriteString(fmt.Sprintf("%s %s: %s", icon, i18n.T("error"), err.GetMessage()))
	}
	
	// 原因エラーを表示
	if f.showCause && err.Cause != nil {
		result.WriteString(fmt.Sprintf("\n  %s %s", i18n.T("caused_by", "原因"), err.Cause.Error()))
	}
	
	// 提案を表示
	if f.showSuggestions && len(err.Suggestions) > 0 {
		result.WriteString(fmt.Sprintf("\n\n%s %s:", f.getHintIcon(), i18n.T("suggestions", "解決策")))
		for _, suggestion := range err.Suggestions {
			if f.colorEnabled {
				result.WriteString(fmt.Sprintf("\n  %s %s", f.colorYellow("•"), suggestion))
			} else {
				result.WriteString(fmt.Sprintf("\n  • %s", suggestion))
			}
		}
	}
	
	// コマンド固有のヘルプ
	if err.Command != "" {
		helpHint := f.getCommandHelpHint(err.Command)
		if helpHint != "" {
			result.WriteString(fmt.Sprintf("\n\n%s %s", f.getHintIcon(), helpHint))
		}
	}
	
	// ドキュメントリンク
	if err.documentation != "" {
		result.WriteString(fmt.Sprintf("\n\n%s %s: %s", 
			f.getDocIcon(), i18n.T("documentation", "ドキュメント"), err.documentation))
	}
}

// formatGenericError は通常のエラーをフォーマットする
func (f *ErrorFormatter) formatGenericError(result *strings.Builder, err error) {
	icon := f.getErrorIcon(ErrorTypeGeneral)
	
	if f.colorEnabled {
		result.WriteString(f.colorRed(fmt.Sprintf("%s %s: %s", icon, i18n.T("error"), err.Error())))
	} else {
		result.WriteString(fmt.Sprintf("%s %s: %s", icon, i18n.T("error"), err.Error()))
	}
}

// getErrorIcon はエラータイプに応じたアイコンを返す
func (f *ErrorFormatter) getErrorIcon(errorType ErrorType) string {
	switch errorType {
	case ErrorTypeFile:
		return "📁"
	case ErrorTypeGit:
		return "🔧"
	case ErrorTypeCommand:
		return "⚙️"
	case ErrorTypeData:
		return "📊"
	case ErrorTypeSecurity:
		return "🔒"
	case ErrorTypeConfig:
		return "🛠️"
	case ErrorTypeNetwork:
		return "🌐"
	default:
		return "❌"
	}
}

// getHintIcon はヒントアイコンを返す
func (f *ErrorFormatter) getHintIcon() string {
	return "💡"
}

// getDocIcon はドキュメントアイコンを返す
func (f *ErrorFormatter) getDocIcon() string {
	return "📖"
}

// colorRed は文字列を赤色にする
func (f *ErrorFormatter) colorRed(text string) string {
	if !f.colorEnabled {
		return text
	}
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}

// colorYellow は文字列を黄色にする
func (f *ErrorFormatter) colorYellow(text string) string {
	if !f.colorEnabled {
		return text
	}
	return fmt.Sprintf("\033[33m%s\033[0m", text)
}

// getCommandHelpHint はコマンド固有のヘルプヒントを返す
func (f *ErrorFormatter) getCommandHelpHint(command string) string {
	switch command {
	case "track":
		return i18n.T("help_hint_track")
	case "stats":
		return i18n.T("help_hint_date_format")
	case "blame":
		return i18n.T("help_hint_git_tracking")
	case "init":
		return i18n.T("help_hint_force_option")
	default:
		return i18n.T("help_hint_general")
	}
}

// 便利な関数群

// FileNotFound はファイルが見つからないエラーを作成する
func FileNotFound(filePath string) *FriendlyError {
	return NewError(ErrorTypeFile, "file_not_found", filePath).
		WithSuggestions(
			i18n.T("suggestion_check_file_path", "ファイルパスを確認してください"),
			i18n.T("suggestion_check_working_directory", "作業ディレクトリを確認してください"),
		).
		WithRecoverable(true)
}

// GitNotRepository はGitリポジトリではないエラーを作成する
func GitNotRepository() *FriendlyError {
	return NewError(ErrorTypeGit, "not_git_repository").
		WithSuggestions(
			i18n.T("suggestion_git_init", "`git init` でリポジトリを初期化してください"),
			i18n.T("suggestion_change_directory", "Gitリポジトリのディレクトリに移動してください"),
		).
		WithRecoverable(true)
}

// UnknownCommand は不明なコマンドエラーを作成する
func UnknownCommand(command string) *FriendlyError {
	return NewError(ErrorTypeCommand, "unknown_command", command).
		WithSuggestions(
			i18n.T("help_hint_general"),
			i18n.T("suggestion_check_spelling", "コマンドのスペルを確認してください"),
		).
		WithRecoverable(true)
}

// InvalidDateFormat は無効な日付形式エラーを作成する
func InvalidDateFormat(dateStr string) *FriendlyError {
	return NewError(ErrorTypeCommand, "invalid_date_format", dateStr).
		WithSuggestions(
			i18n.T("help_hint_date_format"),
			i18n.T("suggestion_date_example", "例: 2024-01-01"),
		).
		WithRecoverable(true)
}

// NoTrackingData はトラッキングデータなしエラーを作成する
func NoTrackingData() *FriendlyError {
	return NewError(ErrorTypeData, "no_tracking_data").
		WithSuggestions(
			i18n.T("suggestion_run_init", "`aict init` を実行してください"),
			i18n.T("suggestion_track_files", "ファイルを追跡してみてください"),
		).
		WithRecoverable(true)
}

// SecurityScanFailed はセキュリティスキャン失敗エラーを作成する
func SecurityScanFailed(cause error) *FriendlyError {
	return WrapError(cause, ErrorTypeSecurity, "security_scan_failed").
		WithSuggestions(
			i18n.T("suggestion_check_permissions", "ファイル権限を確認してください"),
			i18n.T("suggestion_run_as_admin", "管理者権限で実行してみてください"),
		).
		WithRecoverable(true)
}

// ConfigNotFound は設定ファイルが見つからないエラーを作成する
func ConfigNotFound() *FriendlyError {
	return NewError(ErrorTypeConfig, "config_not_found").
		WithSuggestions(
			i18n.T("suggestion_run_wizard", "`aict wizard` で設定を作成してください"),
			i18n.T("suggestion_run_init", "`aict init` を実行してください"),
		).
		WithRecoverable(true)
}

// Global formatter instance
var globalFormatter *ErrorFormatter

// InitializeFormatter はグローバルなエラーフォーマッターを初期化する
func InitializeFormatter() {
	globalFormatter = NewErrorFormatter()
}

// FormatError はグローバルなエラーフォーマット関数
func FormatError(err error) string {
	if globalFormatter == nil {
		InitializeFormatter()
	}
	return globalFormatter.Format(err)
}