package security

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ExclusionManager は機密ファイルの除外を管理する
type ExclusionManager struct {
	exclusionRules    []ExclusionRule
	exclusionFile     string
	sensitivePatterns []string
	enabled           bool
}

// ExclusionRule は除外ルールを表す
type ExclusionRule struct {
	Pattern     string              `json:"pattern"`
	Type        ExclusionType       `json:"type"`
	Reason      string              `json:"reason"`
	Compiled    *regexp.Regexp      `json:"-"`
	Sensitivity SensitivityLevel    `json:"sensitivity"`
}

// ExclusionType は除外タイプを表す
type ExclusionType int

const (
	ExclusionTypeGlob ExclusionType = iota
	ExclusionTypeRegex
	ExclusionTypeExtension
	ExclusionTypeDirectory
	ExclusionTypeFilename
)

// SensitivityLevel は機密度レベルを表す
type SensitivityLevel int

const (
	SensitivityLow SensitivityLevel = iota
	SensitivityMedium
	SensitivityHigh
	SensitivityCritical
)

// NewExclusionManager は新しい除外マネージャーを作成する
func NewExclusionManager(dataDir string) *ExclusionManager {
	enabled := strings.ToLower(os.Getenv("AICT_ENABLE_EXCLUSIONS")) != "false" // デフォルトで有効

	exclusionFile := filepath.Join(dataDir, ".aict-exclude")
	
	// デフォルトの機密パターン
	sensitivePatterns := []string{
		// 設定ファイル
		"*.env", ".env.*", "config.json", "config.yaml", "config.yml",
		// 認証情報
		"*.key", "*.pem", "*.p12", "*.pfx", "*.crt", "*.cer",
		"id_rsa", "id_dsa", "id_ecdsa", "id_ed25519",
		"*.keystore", "*.jks", "*.p7b",
		// データベース
		"*.db", "*.sqlite", "*.sqlite3", "*.mdb",
		// ログファイル
		"*.log", "*.logs", "log/*", "logs/*",
		// バックアップ
		"*.bak", "*.backup", "*.old", "*.orig",
		// 一時ファイル
		"*.tmp", "*.temp", "temp/*", "tmp/*",
		"*~", "*.swp", "*.swo",
		// ビルド成果物
		"*.exe", "*.dll", "*.so", "*.dylib", "*.a",
		"build/*", "dist/*", "target/*", "bin/*",
		// パッケージ管理
		"node_modules/*", "vendor/*", ".git/*",
		// IDEファイル
		".vscode/*", ".idea/*", "*.iml",
		// OS固有
		".DS_Store", "Thumbs.db", "desktop.ini",
	}

	em := &ExclusionManager{
		exclusionFile:     exclusionFile,
		sensitivePatterns: sensitivePatterns,
		enabled:           enabled,
	}

	// デフォルトルールを設定
	em.setupDefaultRules()

	// カスタムルールをロード
	em.loadCustomRules()

	return em
}

// setupDefaultRules はデフォルトの除外ルールを設定する
func (em *ExclusionManager) setupDefaultRules() {
	defaultRules := []ExclusionRule{
		// 機密ファイル（拡張子ベース）
		{".key", ExclusionTypeExtension, "秘密鍵ファイル", nil, SensitivityCritical},
		{".pem", ExclusionTypeExtension, "証明書ファイル", nil, SensitivityCritical},
		{".env", ExclusionTypeFilename, "環境変数ファイル", nil, SensitivityHigh},
		
		// ログファイル
		{".log", ExclusionTypeExtension, "ログファイル", nil, SensitivityMedium},
		
		// バックアップファイル
		{".bak", ExclusionTypeExtension, "バックアップファイル", nil, SensitivityLow},
		{".backup", ExclusionTypeExtension, "バックアップファイル", nil, SensitivityLow},
		
		// 一時ファイル
		{".tmp", ExclusionTypeExtension, "一時ファイル", nil, SensitivityLow},
		{".temp", ExclusionTypeExtension, "一時ファイル", nil, SensitivityLow},
		
		// データベースファイル
		{".db", ExclusionTypeExtension, "データベースファイル", nil, SensitivityHigh},
		{".sqlite", ExclusionTypeExtension, "SQLiteデータベース", nil, SensitivityHigh},
		
		// ディレクトリ除外
		{"node_modules", ExclusionTypeDirectory, "Node.js依存関係", nil, SensitivityLow},
		{".git", ExclusionTypeDirectory, "Gitメタデータ", nil, SensitivityMedium},
		{"vendor", ExclusionTypeDirectory, "ベンダー依存関係", nil, SensitivityLow},
		
		// 特殊パターン（正規表現）
		{`.*password.*`, ExclusionTypeRegex, "パスワード関連ファイル", nil, SensitivityCritical},
		{`.*secret.*`, ExclusionTypeRegex, "シークレット関連ファイル", nil, SensitivityCritical},
		{`.*credential.*`, ExclusionTypeRegex, "認証情報関連ファイル", nil, SensitivityCritical},
	}

	for _, rule := range defaultRules {
		compiled, err := em.compileRule(rule)
		if err != nil {
			continue // エラーのあるルールはスキップ
		}
		rule.Compiled = compiled
		em.exclusionRules = append(em.exclusionRules, rule)
	}
}

// compileRule はルールをコンパイルする
func (em *ExclusionManager) compileRule(rule ExclusionRule) (*regexp.Regexp, error) {
	switch rule.Type {
	case ExclusionTypeRegex:
		return regexp.Compile(rule.Pattern)
	case ExclusionTypeGlob:
		// Globパターンを正規表現に変換
		pattern := strings.ReplaceAll(rule.Pattern, "*", ".*")
		pattern = strings.ReplaceAll(pattern, "?", ".")
		return regexp.Compile("^" + pattern + "$")
	case ExclusionTypeExtension:
		// 拡張子パターン
		escaped := regexp.QuoteMeta(rule.Pattern)
		return regexp.Compile(`.*\.` + strings.TrimPrefix(escaped, `\.`) + `$`)
	case ExclusionTypeDirectory:
		// ディレクトリパターン
		escaped := regexp.QuoteMeta(rule.Pattern)
		return regexp.Compile(`(^|/)` + escaped + `(/.*)?$`)
	case ExclusionTypeFilename:
		// ファイル名パターン
		escaped := regexp.QuoteMeta(rule.Pattern)
		return regexp.Compile(`(^|/)` + escaped + `$`)
	default:
		return nil, fmt.Errorf("未知の除外タイプ: %d", rule.Type)
	}
}

// loadCustomRules はカスタム除外ルールをロードする
func (em *ExclusionManager) loadCustomRules() {
	file, err := os.Open(em.exclusionFile)
	if err != nil {
		return // ファイルが存在しない場合はスキップ
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// コメント行や空行をスキップ
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// ルールをパース
		rule, err := em.parseCustomRule(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "除外ルール解析エラー（行%d）: %v\n", lineNum, err)
			continue
		}

		// ルールをコンパイル
		compiled, err := em.compileRule(rule)
		if err != nil {
			fmt.Fprintf(os.Stderr, "除外ルールコンパイルエラー（行%d）: %v\n", lineNum, err)
			continue
		}

		rule.Compiled = compiled
		em.exclusionRules = append(em.exclusionRules, rule)
	}
}

// parseCustomRule はカスタムルールをパースする
func (em *ExclusionManager) parseCustomRule(line string) (ExclusionRule, error) {
	// 基本形式: pattern [type:reason]
	parts := strings.SplitN(line, " ", 2)
	pattern := parts[0]
	
	rule := ExclusionRule{
		Pattern:     pattern,
		Type:        ExclusionTypeGlob, // デフォルト
		Reason:      "カスタムルール",
		Sensitivity: SensitivityMedium,
	}

	// オプション部分の解析
	if len(parts) > 1 {
		options := parts[1]
		if strings.Contains(options, "regex:") {
			rule.Type = ExclusionTypeRegex
			rule.Reason = strings.TrimPrefix(options, "regex:")
		} else if strings.Contains(options, "ext:") {
			rule.Type = ExclusionTypeExtension
			rule.Reason = strings.TrimPrefix(options, "ext:")
		} else if strings.Contains(options, "dir:") {
			rule.Type = ExclusionTypeDirectory
			rule.Reason = strings.TrimPrefix(options, "dir:")
		} else {
			rule.Reason = options
		}
	}

	return rule, nil
}

// ShouldExclude はファイルが除外されるべきかどうかを判定する
func (em *ExclusionManager) ShouldExclude(filePath string) (bool, string) {
	if !em.enabled {
		return false, ""
	}

	// 正規化されたパスを使用
	normalizedPath := filepath.Clean(filePath)

	// 各ルールをチェック
	for _, rule := range em.exclusionRules {
		if rule.Compiled != nil && rule.Compiled.MatchString(normalizedPath) {
			return true, rule.Reason
		}
	}

	// ファイル名のみでもチェック
	fileName := filepath.Base(normalizedPath)
	for _, rule := range em.exclusionRules {
		if rule.Type == ExclusionTypeFilename && rule.Compiled != nil {
			if rule.Compiled.MatchString(fileName) {
				return true, rule.Reason
			}
		}
	}

	return false, ""
}

// AnalyzeFileSensitivity はファイルの機密度を分析する
func (em *ExclusionManager) AnalyzeFileSensitivity(filePath string) SensitivityLevel {
	normalizedPath := strings.ToLower(filepath.Clean(filePath))
	
	// 最高機密度のパターンをチェック
	criticalPatterns := []string{
		"password", "passwd", "secret", "key", "token", "credential",
		"private", "confidential", "api_key", "access_key",
	}
	
	for _, pattern := range criticalPatterns {
		if strings.Contains(normalizedPath, pattern) {
			return SensitivityCritical
		}
	}

	// 高機密度のパターンをチェック
	highPatterns := []string{
		"config", "env", "database", "db", "auth", "cert", "ssl",
	}
	
	for _, pattern := range highPatterns {
		if strings.Contains(normalizedPath, pattern) {
			return SensitivityHigh
		}
	}

	// 中程度の機密度（ログ、バックアップなど）
	mediumPatterns := []string{
		"log", "backup", "cache", "session",
	}
	
	for _, pattern := range mediumPatterns {
		if strings.Contains(normalizedPath, pattern) {
			return SensitivityMedium
		}
	}

	return SensitivityLow
}

// GetExclusionStats は除外統計を取得する
func (em *ExclusionManager) GetExclusionStats(filePaths []string) map[string]interface{} {
	stats := map[string]interface{}{
		"total_files":    len(filePaths),
		"excluded_files": 0,
		"included_files": 0,
		"exclusion_reasons": make(map[string]int),
		"sensitivity_levels": make(map[string]int),
	}

	reasons := stats["exclusion_reasons"].(map[string]int)
	sensitivities := stats["sensitivity_levels"].(map[string]int)

	for _, filePath := range filePaths {
		excluded, reason := em.ShouldExclude(filePath)
		sensitivity := em.AnalyzeFileSensitivity(filePath)

		if excluded {
			stats["excluded_files"] = stats["excluded_files"].(int) + 1
			reasons[reason]++
		} else {
			stats["included_files"] = stats["included_files"].(int) + 1
		}

		// 機密度統計
		switch sensitivity {
		case SensitivityCritical:
			sensitivities["critical"]++
		case SensitivityHigh:
			sensitivities["high"]++
		case SensitivityMedium:
			sensitivities["medium"]++
		case SensitivityLow:
			sensitivities["low"]++
		}
	}

	return stats
}

// AddCustomRule はカスタム除外ルールを追加する
func (em *ExclusionManager) AddCustomRule(pattern, reason string, ruleType ExclusionType, sensitivity SensitivityLevel) error {
	rule := ExclusionRule{
		Pattern:     pattern,
		Type:        ruleType,
		Reason:      reason,
		Sensitivity: sensitivity,
	}

	// ルールをコンパイル
	compiled, err := em.compileRule(rule)
	if err != nil {
		return fmt.Errorf("ルールのコンパイルに失敗: %w", err)
	}

	rule.Compiled = compiled
	em.exclusionRules = append(em.exclusionRules, rule)

	// ファイルに保存
	return em.saveCustomRule(rule)
}

// saveCustomRule はカスタムルールをファイルに保存する
func (em *ExclusionManager) saveCustomRule(rule ExclusionRule) error {
	file, err := os.OpenFile(em.exclusionFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("除外ファイルのオープンに失敗: %w", err)
	}
	defer file.Close()

	// ルールの形式を決定
	var typePrefix string
	switch rule.Type {
	case ExclusionTypeRegex:
		typePrefix = "regex:"
	case ExclusionTypeExtension:
		typePrefix = "ext:"
	case ExclusionTypeDirectory:
		typePrefix = "dir:"
	default:
		typePrefix = ""
	}

	line := fmt.Sprintf("%s %s%s\n", rule.Pattern, typePrefix, rule.Reason)
	
	_, err = file.WriteString(line)
	return err
}

// IsEnabled は除外機能が有効かどうかを返す
func (em *ExclusionManager) IsEnabled() bool {
	return em.enabled
}

// GetExclusionRules は除外ルールの一覧を返す
func (em *ExclusionManager) GetExclusionRules() []ExclusionRule {
	return em.exclusionRules
}

// GenerateExclusionReport は除外レポートを生成する
func (em *ExclusionManager) GenerateExclusionReport() map[string]interface{} {
	report := map[string]interface{}{
		"enabled":      em.enabled,
		"rules_count":  len(em.exclusionRules),
		"exclusion_file": em.exclusionFile,
		"rules_by_type": make(map[string]int),
		"rules_by_sensitivity": make(map[string]int),
		"sample_patterns": []string{},
	}

	typeCount := report["rules_by_type"].(map[string]int)
	sensitivityCount := report["rules_by_sensitivity"].(map[string]int)
	
	var samplePatterns []string

	for i, rule := range em.exclusionRules {
		// タイプ別カウント
		switch rule.Type {
		case ExclusionTypeGlob:
			typeCount["glob"]++
		case ExclusionTypeRegex:
			typeCount["regex"]++
		case ExclusionTypeExtension:
			typeCount["extension"]++
		case ExclusionTypeDirectory:
			typeCount["directory"]++
		case ExclusionTypeFilename:
			typeCount["filename"]++
		}

		// 機密度別カウント
		switch rule.Sensitivity {
		case SensitivityCritical:
			sensitivityCount["critical"]++
		case SensitivityHigh:
			sensitivityCount["high"]++
		case SensitivityMedium:
			sensitivityCount["medium"]++
		case SensitivityLow:
			sensitivityCount["low"]++
		}

		// サンプルパターン（最初の5つ）
		if i < 5 {
			samplePatterns = append(samplePatterns, rule.Pattern)
		}
	}

	report["sample_patterns"] = samplePatterns
	return report
}