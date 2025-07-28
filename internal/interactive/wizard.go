package interactive

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Wizard はインタラクティブな設定ウィザードを提供する
type Wizard struct {
	reader *bufio.Reader
}

// NewWizard は新しいウィザードインスタンスを作成する
func NewWizard() *Wizard {
	return &Wizard{
		reader: bufio.NewReader(os.Stdin),
	}
}

// AskString は文字列入力を求める
func (w *Wizard) AskString(prompt, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	
	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	
	return input
}

// AskBool はYes/No質問を求める
func (w *Wizard) AskBool(prompt string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}
	
	fmt.Printf("%s [%s]: ", prompt, defaultStr)
	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	
	if input == "" {
		return defaultValue
	}
	
	return input == "y" || input == "yes" || input == "はい"
}

// AskInt は整数入力を求める
func (w *Wizard) AskInt(prompt string, defaultValue int, min, max int) int {
	for {
		fmt.Printf("%s [%d]: ", prompt, defaultValue)
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "" {
			return defaultValue
		}
		
		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("❌ 無効な数値です。もう一度入力してください。\n")
			continue
		}
		
		if value < min || value > max {
			fmt.Printf("❌ %d から %d の間で入力してください。\n", min, max)
			continue
		}
		
		return value
	}
}

// AskChoice は選択肢から選択を求める
func (w *Wizard) AskChoice(prompt string, choices []string, defaultIndex int) int {
	fmt.Printf("%s\n", prompt)
	for i, choice := range choices {
		marker := " "
		if i == defaultIndex {
			marker = "*"
		}
		fmt.Printf("%s %d) %s\n", marker, i+1, choice)
	}
	
	for {
		fmt.Printf("選択してください [%d]: ", defaultIndex+1)
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)
		
		if input == "" {
			return defaultIndex
		}
		
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(choices) {
			fmt.Printf("❌ 1 から %d の間で入力してください。\n", len(choices))
			continue
		}
		
		return choice - 1
	}
}

// ShowProgress はプログレスインジケーターを表示する
func (w *Wizard) ShowProgress(message string, duration time.Duration) {
	fmt.Printf("%s", message)
	
	steps := int(duration / (100 * time.Millisecond))
	for i := 0; i < steps; i++ {
		fmt.Print(".")
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println(" ✅")
}

// ShowBanner はウェルカムバナーを表示する
func (w *Wizard) ShowBanner() {
	fmt.Println("🤖 AI Code Tracker セットアップウィザード")
	fmt.Println("=" + strings.Repeat("=", 45))
	fmt.Println("このウィザードでは、AICT の初期設定を行います。")
	fmt.Println("各質問に答えて、プロジェクトに最適な設定を構成しましょう。")
	fmt.Println()
}

// ShowSummary は設定サマリーを表示する
func (w *Wizard) ShowSummary(config map[string]interface{}) {
	fmt.Println("\n📋 設定サマリー")
	fmt.Println("=" + strings.Repeat("=", 25))
	
	for key, value := range config {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	fmt.Println()
	confirmed := w.AskBool("この設定で続行しますか？", true)
	if !confirmed {
		fmt.Println("❌ セットアップがキャンセルされました。")
		os.Exit(1)
	}
}

// InitializationWizard は初期化ウィザードを実行する
func (w *Wizard) InitializationWizard() map[string]interface{} {
	w.ShowBanner()
	
	config := make(map[string]interface{})
	
	// 基本設定
	fmt.Println("📝 基本設定")
	fmt.Println("----------")
	
	authorName := w.AskString("デフォルトの作成者名", "")
	if authorName != "" {
		config["default_author"] = authorName
	}
	
	// セキュリティ設定
	fmt.Println("\n🔒 セキュリティ設定")
	fmt.Println("-----------------")
	
	enableEncryption := w.AskBool("データ暗号化を有効にしますか？", false)
	config["enable_encryption"] = enableEncryption
	
	if enableEncryption {
		fmt.Println("⚠️  暗号化を有効にする場合、AICT_ENCRYPTION_PASSPHRASE 環境変数を設定してください。")
	}
	
	enableAuditLog := w.AskBool("監査ログを有効にしますか？", true)
	config["enable_audit_log"] = enableAuditLog
	
	// プライバシー設定
	fmt.Println("\n🛡️ プライバシー設定")
	fmt.Println("------------------")
	
	anonymizeAuthors := w.AskBool("作成者名を匿名化しますか？", false)
	config["anonymize_authors"] = anonymizeAuthors
	
	retentionDays := w.AskInt("データ保持期間（日数）", 365, 30, 3650)
	config["data_retention_days"] = retentionDays
	
	// Git統合設定
	fmt.Println("\n🔧 Git統合設定")
	fmt.Println("--------------")
	
	setupGitHooks := w.AskBool("Git hooks を自動設定しますか？", true)
	config["setup_git_hooks"] = setupGitHooks
	
	setupClaudeHooks := w.AskBool("Claude Code hooks を自動設定しますか？", true)
	config["setup_claude_hooks"] = setupClaudeHooks
	
	// 統計設定
	fmt.Println("\n📊 統計設定")
	fmt.Println("-----------")
	
	statsModes := []string{
		"基本統計のみ",
		"詳細統計（ファイル別）",
		"完全統計（すべての分析）",
	}
	statsMode := w.AskChoice("統計の詳細レベルを選択してください", statsModes, 1)
	config["stats_mode"] = []string{"basic", "detailed", "full"}[statsMode]
	
	w.ShowSummary(config)
	
	return config
}

// SecurityWizard はセキュリティ設定ウィザードを実行する
func (w *Wizard) SecurityWizard() map[string]interface{} {
	fmt.Println("🔒 AICT セキュリティ設定ウィザード")
	fmt.Println("=" + strings.Repeat("=", 35))
	fmt.Println("セキュリティ機能の設定を行います。")
	fmt.Println()
	
	config := make(map[string]interface{})
	
	// セキュリティレベル選択
	securityLevels := []string{
		"基本 - 最小限のセキュリティ機能",
		"標準 - 推奨されるセキュリティ機能",
		"厳格 - 最大限のセキュリティ機能",
		"カスタム - 個別に設定",
	}
	
	level := w.AskChoice("セキュリティレベルを選択してください", securityLevels, 1)
	
	switch level {
	case 0: // 基本
		config["security_mode"] = "basic"
		config["enable_encryption"] = false
		config["enable_audit_log"] = false
		config["anonymize_authors"] = false
		config["strict_validation"] = false
		
	case 1: // 標準
		config["security_mode"] = "standard"
		config["enable_encryption"] = true
		config["enable_audit_log"] = true
		config["anonymize_authors"] = false
		config["strict_validation"] = false
		
	case 2: // 厳格
		config["security_mode"] = "strict"
		config["enable_encryption"] = true
		config["enable_audit_log"] = true
		config["anonymize_authors"] = true
		config["strict_validation"] = true
		config["hash_file_paths"] = true
		config["data_retention_days"] = 90
		
	case 3: // カスタム
		config["security_mode"] = "custom"
		config["enable_encryption"] = w.AskBool("データ暗号化を有効にしますか？", true)
		config["enable_audit_log"] = w.AskBool("監査ログを有効にしますか？", true)
		config["anonymize_authors"] = w.AskBool("作成者名を匿名化しますか？", false)
		config["strict_validation"] = w.AskBool("厳格な入力検証を有効にしますか？", false)
		
		if w.AskBool("高度なプライバシー機能を設定しますか？", false) {
			config["hash_file_paths"] = w.AskBool("ファイルパスをハッシュ化しますか？", false)
			config["mask_sensitive_data"] = w.AskBool("機密データのマスキングを有効にしますか？", true)
			config["data_retention_days"] = w.AskInt("データ保持期間（日数）", 365, 30, 3650)
		}
	}
	
	// 環境変数設定の提案
	fmt.Println("\n🔧 環境変数設定")
	fmt.Println("---------------")
	fmt.Println("以下の環境変数を設定することをお勧めします：")
	
	if config["enable_encryption"].(bool) {
		fmt.Println("  export AICT_ENCRYPT_DATA=true")
		fmt.Println("  export AICT_ENCRYPTION_PASSPHRASE=\"your-secure-passphrase\"")
	}
	
	if config["enable_audit_log"].(bool) {
		fmt.Println("  export AICT_AUDIT_LOG=true")
	}
	
	if anonymize, ok := config["anonymize_authors"].(bool); ok && anonymize {
		fmt.Println("  export AICT_ANONYMIZE_AUTHORS=true")
	}
	
	fmt.Println()
	
	return config
}

// QuickStartWizard はクイックスタートウィザードを実行する
func (w *Wizard) QuickStartWizard() {
	fmt.Println("🚀 AICT クイックスタート")
	fmt.Println("=" + strings.Repeat("=", 22))
	fmt.Println("数分でAICTを使い始めましょう！")
	fmt.Println()
	
	// プロジェクトタイプの判定
	var projectType string
	if fileExists("go.mod") {
		projectType = "Go"
	} else if fileExists("package.json") {
		projectType = "JavaScript/TypeScript"
	} else if fileExists("requirements.txt") || fileExists("pyproject.toml") {
		projectType = "Python"
	} else if fileExists("Cargo.toml") {
		projectType = "Rust"
	} else {
		projectType = "その他"
	}
	
	fmt.Printf("📂 検出されたプロジェクトタイプ: %s\n", projectType)
	fmt.Println()
	
	// クイック設定
	fmt.Println("⚡ クイック設定（推奨設定を使用）")
	fmt.Println("--------------------------------")
	
	steps := []string{
		"AICT データディレクトリを初期化",
		"基本設定を適用",
		"Git hooks を設定",
		"Claude Code hooks を設定",
		"セキュリティ設定を適用",
		"設定の確認",
	}
	
	for i, step := range steps {
		fmt.Printf("%d. %s...", i+1, step)
		w.ShowProgress("", 500*time.Millisecond)
	}
	
	fmt.Println("\n✅ セットアップ完了！")
	fmt.Println("\n🎉 次のステップ:")
	fmt.Println("  1. Claude Code でコードを編集してみてください")
	fmt.Println("  2. `aict stats` で統計を確認")
	fmt.Println("  3. `aict blame <ファイル名>` でコード属性を確認")
	fmt.Println("\n詳細な使用方法は `aict help` をご覧ください。")
}

// fileExists はファイルが存在するかチェックする
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}