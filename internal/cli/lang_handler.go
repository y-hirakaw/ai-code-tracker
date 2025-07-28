package cli

import (
	"fmt"

	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/utils"
)

// LangHandler は言語設定コマンドを処理する
type LangHandler struct{}

// NewLangHandler は新しいLangHandlerを作成する
func NewLangHandler() *LangHandler {
	return &LangHandler{}
}

// Handle は言語設定コマンドを実行する
func (h *LangHandler) Handle(args []string) error {
	if len(args) == 0 {
		return h.showCurrentLanguage()
	}

	var (
		setLang    = ""
		listLangs  = false
		persistent = false
	)

	// コマンドライン引数をパース
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--set":
			if i+1 < len(args) {
				setLang = args[i+1]
				i++
			}
		case "--list":
			listLangs = true
		case "--persistent":
			persistent = true
		default:
			// 最初の引数を言語コードとして扱う
			if setLang == "" && !listLangs {
				setLang = args[i]
			}
		}
	}

	// 使用可能言語の一覧表示
	if listLangs {
		return h.listAvailableLanguages()
	}

	// 言語設定
	if setLang != "" {
		return h.setLanguage(setLang, persistent)
	}

	// 現在の言語を表示
	return h.showCurrentLanguage()
}

// showCurrentLanguage は現在の言語設定を表示する
func (h *LangHandler) showCurrentLanguage() error {
	currentLang := i18n.GetLocale()
	
	fmt.Printf("🌐 %s: %s (%s)\n", 
		i18n.T("current_language", "現在の言語"),
		h.getLanguageName(string(currentLang)),
		currentLang)

	// 言語の変更方法を表示
	fmt.Printf("\n%s:\n", i18n.T("change_language_help", "言語を変更するには"))
	fmt.Printf("  aict lang ja          # %s\n", i18n.T("set_japanese", "日本語に設定"))
	fmt.Printf("  aict lang en          # %s\n", i18n.T("set_english", "英語に設定"))
	fmt.Printf("  aict lang --list      # %s\n", i18n.T("list_languages", "利用可能な言語を表示"))
	fmt.Printf("  aict lang ja --persistent  # %s\n", i18n.T("set_persistent", "設定を永続化"))

	return nil
}

// listAvailableLanguages は利用可能な言語一覧を表示する
func (h *LangHandler) listAvailableLanguages() error {
	fmt.Printf("🌐 %s:\n\n", i18n.T("available_languages", "利用可能な言語"))

	languages := []struct {
		Code string
		Name string
		Native string
	}{
		{"ja", "Japanese", "日本語"},
		{"en", "English", "English"},
	}

	currentLang := string(i18n.GetLocale())

	for _, lang := range languages {
		marker := "  "
		if lang.Code == currentLang {
			marker = "✓ "
		}
		
		fmt.Printf("%s%-4s - %s (%s)\n", 
			marker, lang.Code, lang.Name, lang.Native)
	}

	fmt.Printf("\n%s:\n", i18n.T("usage_examples", "使用例"))
	fmt.Printf("  aict lang ja    # %s\n", i18n.T("switch_to_japanese", "日本語に切り替え"))
	fmt.Printf("  aict lang en    # %s\n", i18n.T("switch_to_english", "英語に切り替え"))

	return nil
}

// setLanguage は言語を設定する
func (h *LangHandler) setLanguage(langCode string, persistent bool) error {
	// 言語コードの検証
	if !h.isValidLanguageCode(langCode) {
		return errors.NewError(errors.ErrorTypeCommand, "invalid_language_code", langCode).
			WithSuggestions(i18n.T("suggestion_valid_languages", "利用可能な言語: ja (日本語), en (English)"))
	}

	// 現在の言語と同じかチェック
	currentLang := string(i18n.GetLocale())
	if langCode == currentLang {
		fmt.Printf("ℹ️ %s %s\n", 
			i18n.T("already_set", "既に設定されています:"),
			h.getLanguageName(langCode))
		return nil
	}

	// 言語を設定
	locale := i18n.Locale(langCode)
	i18n.SetLocale(locale)

	// 永続化する場合は設定ファイルに保存
	if persistent {
		if err := h.savePersistentLanguage(langCode); err != nil {
			// エラーが発生しても言語設定は成功しているので警告として表示
			fmt.Printf("⚠️ %s: %v\n", 
				i18n.T("persistent_save_failed", "設定の永続化に失敗しました"), err)
			fmt.Printf("💡 %s\n", 
				i18n.T("env_var_alternative", "代替として環境変数 AICT_LANGUAGE を使用できます"))
		} else {
			fmt.Printf("✅ %s %s (%s)\n",
				i18n.T("language_set_persistent", "言語を永続的に設定しました:"),
				h.getLanguageName(langCode), langCode)
			return nil
		}
	}

	// 一時的な設定
	fmt.Printf("✅ %s %s (%s)\n",
		i18n.T("language_set_temporary", "言語を一時的に設定しました:"),
		h.getLanguageName(langCode), langCode)

	if !persistent {
		fmt.Printf("💡 %s\n", 
			i18n.T("persistent_hint", "永続化するには --persistent オプションを使用してください"))
	}

	return nil
}

// savePersistentLanguage は言語設定を永続化する
func (h *LangHandler) savePersistentLanguage(langCode string) error {
	homeDir, err := utils.GetHomeDirectory()
	if err != nil {
		return err
	}

	configPath := utils.JoinPath(homeDir, ".aict", "config.json")
	configManager := utils.NewConfigManager(configPath)

	// 既存の設定を読み込み
	config := utils.NewDefaultConfig()
	if err := configManager.LoadConfig(config); err != nil {
		return err
	}

	// 言語設定を更新
	config.Language = langCode

	// 設定を保存
	if err := configManager.SaveConfig(config); err != nil {
		return err
	}

	return nil
}

// isValidLanguageCode は言語コードが有効かチェックする
func (h *LangHandler) isValidLanguageCode(langCode string) bool {
	validCodes := []string{"ja", "en"}
	for _, code := range validCodes {
		if code == langCode {
			return true
		}
	}
	return false
}

// getLanguageName は言語コードから言語名を取得する
func (h *LangHandler) getLanguageName(langCode string) string {
	names := map[string]string{
		"ja": i18n.T("language_japanese", "日本語"),
		"en": i18n.T("language_english", "English"),
	}
	
	if name, exists := names[langCode]; exists {
		return name
	}
	return langCode
}