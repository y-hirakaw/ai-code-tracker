package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Locale は言語ロケール
type Locale string

const (
	// LocaleJA は日本語
	LocaleJA Locale = "ja"
	// LocaleEN は英語
	LocaleEN Locale = "en"
)

// Messages は翻訳メッセージのマップ
type Messages map[string]string

// I18n は国際化システム
type I18n struct {
	currentLocale Locale
	messages      map[Locale]Messages
	fallback      Locale
}

// NewI18n は新しい国際化システムを作成する
func NewI18n() *I18n {
	i18n := &I18n{
		currentLocale: LocaleJA, // デフォルトは日本語
		messages:      make(map[Locale]Messages),
		fallback:      LocaleEN,
	}
	
	// デフォルトメッセージを読み込み
	i18n.loadDefaultMessages()
	
	// 環境変数から言語設定を読み込み
	if lang := os.Getenv("AICT_LANG"); lang != "" {
		i18n.SetLocale(Locale(lang))
	} else if lang := os.Getenv("LANG"); lang != "" {
		// システムのLANG環境変数から判定
		if strings.HasPrefix(lang, "ja") {
			i18n.SetLocale(LocaleJA)
		} else {
			i18n.SetLocale(LocaleEN)
		}
	}
	
	return i18n
}

// SetLocale は現在のロケールを設定する
func (i *I18n) SetLocale(locale Locale) {
	i.currentLocale = locale
}

// GetLocale は現在のロケールを取得する
func (i *I18n) GetLocale() Locale {
	return i.currentLocale
}

// T は翻訳を取得する（キーと引数を受け取る）
func (i *I18n) T(key string, args ...interface{}) string {
	// 現在のロケールでメッセージを検索
	if messages, exists := i.messages[i.currentLocale]; exists {
		if message, found := messages[key]; found {
			if len(args) > 0 {
				return fmt.Sprintf(message, args...)
			}
			return message
		}
	}
	
	// フォールバック言語で検索
	if i.currentLocale != i.fallback {
		if messages, exists := i.messages[i.fallback]; exists {
			if message, found := messages[key]; found {
				if len(args) > 0 {
					return fmt.Sprintf(message, args...)
				}
				return message
			}
		}
	}
	
	// メッセージが見つからない場合はキーをそのまま返す
	if len(args) > 0 {
		return fmt.Sprintf("%s: %v", key, args)
	}
	return key
}

// LoadMessagesFromFile はファイルから翻訳メッセージを読み込む
func (i *I18n) LoadMessagesFromFile(locale Locale, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("メッセージファイルの読み込みに失敗: %w", err)
	}
	
	var messages Messages
	if err := json.Unmarshal(data, &messages); err != nil {
		return fmt.Errorf("メッセージファイルの解析に失敗: %w", err)
	}
	
	i.messages[locale] = messages
	return nil
}

// LoadMessagesFromDir はディレクトリから翻訳メッセージを読み込む
func (i *I18n) LoadMessagesFromDir(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		
		// ファイル名からロケールを判定 (例: messages.ja.json)
		fileName := strings.TrimSuffix(filepath.Base(path), ".json")
		parts := strings.Split(fileName, ".")
		if len(parts) >= 2 {
			locale := Locale(parts[len(parts)-1])
			return i.LoadMessagesFromFile(locale, path)
		}
		
		return nil
	})
}

// loadDefaultMessages はデフォルトの翻訳メッセージを読み込む
func (i *I18n) loadDefaultMessages() {
	// 日本語メッセージ
	i.messages[LocaleJA] = Messages{
		// 一般的なメッセージ
		"error":            "エラー",
		"warning":          "警告", 
		"info":             "情報",
		"success":          "成功",
		"failed":           "失敗",
		"completed":        "完了",
		"cancelled":        "キャンセル",
		"yes":              "はい",
		"no":               "いいえ",
		"unknown":          "不明",
		"loading":          "読み込み中",
		"processing":       "処理中",
		"please_wait":      "お待ちください",
		
		// エラー関連
		"caused_by":            "原因",
		"suggestions":          "解決策",
		"documentation":        "ドキュメント",
		"generic_error":        "システムエラーが発生しました",
		"missing_required_argument": "必須引数が不足しています: %s",
		"invalid_output_format": "無効な出力形式です: %s",
		"directory_access_failed": "ディレクトリへのアクセスに失敗しました",
		"tracking_failed":      "トラッキングに失敗しました",
		"statistics_fetch_failed": "統計情報の取得に失敗しました",
		"contributor_fetch_failed": "貢献者情報の取得に失敗しました",
		"blame_fetch_failed":   "blame情報の取得に失敗しました",
		
		// ファイル関連
		"file_not_found":           "ファイルが見つかりません: %s",
		"file_not_exists":          "ファイルが存在しません: %s",
		"file_permission_denied":   "ファイルへのアクセス権限がありません: %s",
		"directory_not_found":      "ディレクトリが見つかりません: %s",
		"invalid_file_path":        "無効なファイルパスです: %s",
		"file_too_large":           "ファイルサイズが大きすぎます: %s",
		
		// Git関連
		"not_git_repository":       "Gitリポジトリではありません",
		"git_command_failed":       "Gitコマンドが失敗しました: %s",
		"no_git_history":           "Gitの履歴がありません",
		"git_file_not_tracked":     "ファイルがGitで追跡されていません: %s",
		
		// コマンド関連
		"unknown_command":          "不明なコマンド: %s",
		"invalid_option":           "無効なオプション: %s",
		"missing_required_option":  "必須オプションが不足しています: %s",
		"invalid_date_format":      "日付の形式が不正です (YYYY-MM-DD): %s",
		"invalid_number":           "無効な数値です: %s",
		
		// データ関連
		"no_tracking_data":         "トラッキングデータがありません",
		"invalid_tracking_data":    "無効なトラッキングデータです",
		"data_corruption":          "データが破損している可能性があります",
		"storage_initialization_failed": "ストレージの初期化に失敗しました",
		
		// セキュリティ関連
		"security_scan_failed":     "セキュリティスキャンに失敗しました",
		"encryption_failed":        "暗号化に失敗しました",
		"decryption_failed":        "復号化に失敗しました",
		"invalid_passphrase":       "パスフレーズが無効です",
		"security_violation":       "セキュリティ違反を検出しました",
		
		// 設定関連
		"config_not_found":         "設定ファイルが見つかりません",
		"invalid_config":           "設定が無効です",
		"config_save_failed":       "設定の保存に失敗しました",
		
		// ヘルプメッセージ
		"help_command_not_found":   "コマンド '%s' のヘルプが見つかりません",
		"help_hint_general":        "'aict help' で利用可能なコマンドを確認できます",
		"help_hint_track":          "'aict help track' でtrackコマンドの詳細な使用方法を確認できます",
		"help_hint_date_format":    "有効な日付形式は YYYY-MM-DD です（例: 2024-01-01）",
		"help_hint_git_tracking":   "ファイルがGitで追跡されているか確認してください",
		"help_hint_force_option":   "既存の設定がある場合は --force オプションを使用してください",
		
		// 提案メッセージ
		"suggestion_specify_author":   "--author オプションで作成者を指定してください",
		"suggestion_specify_file":     "例: aict blame src/main.go",
		"suggestion_valid_formats":    "有効な形式: table, json, summary, daily, files, contributors",
		"suggestion_check_file_path":  "ファイルパスを確認してください",
		"suggestion_check_working_directory": "作業ディレクトリを確認してください",
		"suggestion_git_init":         "`git init` でリポジトリを初期化してください",
		"suggestion_change_directory": "Gitリポジトリのディレクトリに移動してください",
		"suggestion_check_spelling":   "コマンドのスペルを確認してください",
		"suggestion_date_example":     "例: 2024-01-01",
		"suggestion_run_init":         "`aict init` を実行してください",
		"suggestion_track_files":      "ファイルを追跡してみてください",
		"suggestion_check_permissions": "ファイル権限を確認してください",
		"suggestion_run_as_admin":     "管理者権限で実行してみてください",
		"suggestion_run_wizard":       "`aict wizard` で設定を作成してください",
		
		// 成功メッセージ
		"initialization_success":   "AI Code Tracker を初期化しました",
		"tracking_success":         "%d個のファイルの変更を追跡しました",
		"setup_success":            "Hooks設定が完了しました",
		"wizard_success":           "設定ウィザードが完了しました",
		"security_config_success":  "セキュリティ設定が正常に適用されました",
		
		// プログレスメッセージ
		"initializing":             "初期化中",
		"setting_up_hooks":         "Hooks設定中",
		"applying_security":        "セキュリティ設定適用中",
		"scanning_files":           "ファイルスキャン中",
		"generating_stats":         "統計生成中",
	}
	
	// 英語メッセージ
	i.messages[LocaleEN] = Messages{
		// General messages
		"error":            "Error",
		"warning":          "Warning",
		"info":             "Info", 
		"success":          "Success",
		"failed":           "Failed",
		"completed":        "Completed",
		"cancelled":        "Cancelled",
		"yes":              "Yes",
		"no":               "No",
		"unknown":          "Unknown",
		"loading":          "Loading",
		"processing":       "Processing",
		"please_wait":      "Please wait",
		
		// Error related
		"caused_by":            "Caused by",
		"suggestions":          "Suggestions",
		"documentation":        "Documentation", 
		"generic_error":        "System error occurred",
		"missing_required_argument": "Missing required argument: %s",
		"invalid_output_format": "Invalid output format: %s",
		"directory_access_failed": "Failed to access directory",
		"tracking_failed":      "Tracking failed",
		"statistics_fetch_failed": "Failed to fetch statistics",
		"contributor_fetch_failed": "Failed to fetch contributor information",
		"blame_fetch_failed":   "Failed to fetch blame information",
		
		// File related
		"file_not_found":           "File not found: %s",
		"file_not_exists":          "File does not exist: %s", 
		"file_permission_denied":   "Permission denied for file: %s",
		"directory_not_found":      "Directory not found: %s",
		"invalid_file_path":        "Invalid file path: %s",
		"file_too_large":           "File too large: %s",
		
		// Git related
		"not_git_repository":       "Not a Git repository",
		"git_command_failed":       "Git command failed: %s",
		"no_git_history":           "No Git history found",
		"git_file_not_tracked":     "File not tracked by Git: %s",
		
		// Command related
		"unknown_command":          "Unknown command: %s",
		"invalid_option":           "Invalid option: %s",
		"missing_required_option":  "Missing required option: %s",
		"invalid_date_format":      "Invalid date format (YYYY-MM-DD): %s",
		"invalid_number":           "Invalid number: %s",
		
		// Data related
		"no_tracking_data":         "No tracking data available",
		"invalid_tracking_data":    "Invalid tracking data",
		"data_corruption":          "Data may be corrupted",
		"storage_initialization_failed": "Failed to initialize storage",
		
		// Security related
		"security_scan_failed":     "Security scan failed",
		"encryption_failed":        "Encryption failed",
		"decryption_failed":        "Decryption failed",
		"invalid_passphrase":       "Invalid passphrase",
		"security_violation":       "Security violation detected",
		
		// Configuration related
		"config_not_found":         "Configuration file not found",
		"invalid_config":           "Invalid configuration",
		"config_save_failed":       "Failed to save configuration",
		
		// Help messages
		"help_command_not_found":   "Help for command '%s' not found",
		"help_hint_general":        "Use 'aict help' to see available commands",
		"help_hint_track":          "Use 'aict help track' for detailed usage of track command",
		"help_hint_date_format":    "Valid date format is YYYY-MM-DD (e.g., 2024-01-01)",
		"help_hint_git_tracking":   "Make sure the file is tracked by Git",
		"help_hint_force_option":   "Use --force option if there are existing settings",
		
		// Suggestion messages
		"suggestion_specify_author":   "Specify author with --author option",
		"suggestion_specify_file":     "Example: aict blame src/main.go",
		"suggestion_valid_formats":    "Valid formats: table, json, summary, daily, files, contributors",
		"suggestion_check_file_path":  "Check the file path",
		"suggestion_check_working_directory": "Check the working directory",
		"suggestion_git_init":         "Initialize repository with `git init`",
		"suggestion_change_directory": "Change to Git repository directory",
		"suggestion_check_spelling":   "Check command spelling",
		"suggestion_date_example":     "Example: 2024-01-01",
		"suggestion_run_init":         "Run `aict init`",
		"suggestion_track_files":      "Try tracking some files",
		"suggestion_check_permissions": "Check file permissions",
		"suggestion_run_as_admin":     "Try running with administrator privileges",
		"suggestion_run_wizard":       "Create configuration with `aict wizard`",
		
		// Success messages
		"initialization_success":   "AI Code Tracker initialized",
		"tracking_success":         "Tracked changes in %d file(s)",
		"setup_success":            "Hooks setup completed",
		"wizard_success":           "Configuration wizard completed",
		"security_config_success":  "Security configuration applied successfully",
		
		// Progress messages
		"initializing":             "Initializing",
		"setting_up_hooks":         "Setting up hooks",
		"applying_security":        "Applying security settings",
		"scanning_files":           "Scanning files",
		"generating_stats":         "Generating statistics",
	}
}

// GetAvailableLocales は利用可能なロケール一覧を返す
func (i *I18n) GetAvailableLocales() []Locale {
	locales := make([]Locale, 0, len(i.messages))
	for locale := range i.messages {
		locales = append(locales, locale)
	}
	return locales
}

// ValidateLocale はロケールが有効かどうかを確認する
func (i *I18n) ValidateLocale(locale Locale) bool {
	_, exists := i.messages[locale]
	return exists
}

// Global instance
var globalI18n *I18n

// Initialize はグローバルなi18nシステムを初期化する
func Initialize() {
	globalI18n = NewI18n()
}

// T はグローバルな翻訳関数
func T(key string, args ...interface{}) string {
	if globalI18n == nil {
		Initialize()
	}
	return globalI18n.T(key, args...)
}

// SetLocale はグローバルなロケールを設定する
func SetLocale(locale Locale) {
	if globalI18n == nil {
		Initialize()
	}
	globalI18n.SetLocale(locale)
}

// GetLocale はグローバルなロケールを取得する
func GetLocale() Locale {
	if globalI18n == nil {
		Initialize()
	}
	return globalI18n.GetLocale()
}