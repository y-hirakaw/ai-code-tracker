package security

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewExclusionManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)
	if em == nil {
		t.Fatal("ExclusionManagerがnilです")
	}

	if !em.IsEnabled() {
		t.Error("除外機能がデフォルトで有効になっていません")
	}

	if len(em.exclusionRules) == 0 {
		t.Error("デフォルトの除外ルールが設定されていません")
	}
}

func TestExclusionManager_ShouldExclude(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	tests := []struct {
		name     string
		filePath string
		excluded bool
	}{
		{"通常のGoファイル", "src/main.go", false},
		{"秘密鍵ファイル", "private.key", true},
		{"環境変数ファイル", ".env", true},
		{"ログファイル", "app.log", true},
		{"バックアップファイル", "data.bak", true},
		{"一時ファイル", "temp.tmp", true},
		{"データベースファイル", "app.db", true},
		{"Node.jsモジュール", "node_modules/package/index.js", true},
		{"Gitメタデータ", ".git/config", true},
		{"パスワード関連", "password-config.txt", true},
		{"シークレット関連", "api-secret.json", true},
		{"通常のJavaScriptファイル", "src/app.js", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excluded, reason := em.ShouldExclude(tt.filePath)
			if excluded != tt.excluded {
				if tt.excluded {
					t.Errorf("ファイル%sが除外されませんでした", tt.filePath)
				} else {
					t.Errorf("ファイル%sが除外されました（理由: %s）", tt.filePath, reason)
				}
			}
		})
	}
}

func TestExclusionManager_AnalyzeFileSensitivity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	tests := []struct {
		name        string
		filePath    string
		sensitivity SensitivityLevel
	}{
		{"パスワードファイル", "password.txt", SensitivityCritical},
		{"APIキーファイル", "api_key.json", SensitivityCritical},
		{"秘密設定", "secret-config.yaml", SensitivityCritical},
		{"認証情報", "credentials.json", SensitivityCritical},
		{"設定ファイル", "config.ini", SensitivityHigh},
		{"環境変数", ".env.production", SensitivityHigh},
		{"データベース", "app.db", SensitivityHigh},
		{"SSL証明書", "ssl.cert", SensitivityHigh},
		{"ログファイル", "app.log", SensitivityMedium},
		{"バックアップ", "backup.tar.gz", SensitivityMedium},
		{"キャッシュ", "cache.tmp", SensitivityMedium},
		{"通常のファイル", "main.go", SensitivityLow},
		{"ドキュメント", "README.md", SensitivityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensitivity := em.AnalyzeFileSensitivity(tt.filePath)
			if sensitivity != tt.sensitivity {
				t.Errorf("機密度が期待値と異なります: 期待=%d, 実際=%d", tt.sensitivity, sensitivity)
			}
		})
	}
}

func TestExclusionManager_GetExclusionStats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	filePaths := []string{
		"src/main.go",
		"private.key",
		".env",
		"app.log",
		"README.md",
		"secret-config.yaml",
		"node_modules/package/index.js",
	}

	stats := em.GetExclusionStats(filePaths)
	if stats == nil {
		t.Fatal("除外統計がnilです")
	}

	if totalFiles, ok := stats["total_files"].(int); !ok || totalFiles != len(filePaths) {
		t.Errorf("総ファイル数が期待値と異なります: 期待=%d, 実際=%v", len(filePaths), stats["total_files"])
	}

	expectedKeys := []string{"total_files", "excluded_files", "included_files", "exclusion_reasons", "sensitivity_levels"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("統計に%sが含まれていません", key)
		}
	}

	// 除外されるファイルと含まれるファイルの数をチェック
	excludedFiles, ok1 := stats["excluded_files"].(int)
	includedFiles, ok2 := stats["included_files"].(int)
	if !ok1 || !ok2 {
		t.Fatal("除外/含有ファイル数の型が正しくありません")
	}

	if excludedFiles+includedFiles != len(filePaths) {
		t.Errorf("除外/含有ファイル数の合計が総数と一致しません: 除外=%d, 含有=%d, 総数=%d", 
			excludedFiles, includedFiles, len(filePaths))
	}

	// 除外理由の統計
	exclusionReasons, ok := stats["exclusion_reasons"].(map[string]int)
	if !ok {
		t.Fatal("除外理由統計の型が正しくありません")
	}

	totalExclusionReasons := 0
	for _, count := range exclusionReasons {
		totalExclusionReasons += count
	}

	if totalExclusionReasons != excludedFiles {
		t.Errorf("除外理由の合計が除外ファイル数と一致しません: 理由合計=%d, 除外数=%d", 
			totalExclusionReasons, excludedFiles)
	}
}

func TestExclusionManager_AddCustomRule(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	initialRuleCount := len(em.exclusionRules)

	err = em.AddCustomRule("*.custom", "カスタムファイル", ExclusionTypeGlob, SensitivityMedium)
	if err != nil {
		t.Fatalf("カスタムルールの追加に失敗: %v", err)
	}

	if len(em.exclusionRules) != initialRuleCount+1 {
		t.Errorf("ルール数が期待値と異なります: 期待=%d, 実際=%d", 
			initialRuleCount+1, len(em.exclusionRules))
	}

	// 追加したルールが動作するかテスト
	excluded, reason := em.ShouldExclude("test.custom")
	if !excluded {
		t.Error("追加したカスタムルールが動作していません")
	}

	if reason != "カスタムファイル" {
		t.Errorf("除外理由が期待値と異なります: 期待=カスタムファイル, 実際=%s", reason)
	}
}

func TestExclusionManager_CustomRulesFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// カスタム除外ファイルを作成
	excludeFile := filepath.Join(tempDir, ".aict-exclude")
	excludeContent := `# カスタム除外ルール
*.secret
test-* glob:テストファイル
sensitive regex:機密ファイル
*.doc ext:ドキュメントファイル
backup dir:バックアップディレクトリ`

	err = os.WriteFile(excludeFile, []byte(excludeContent), 0600)
	if err != nil {
		t.Fatalf("除外ファイルの作成に失敗: %v", err)
	}

	em := NewExclusionManager(tempDir)

	tests := []struct {
		name     string
		filePath string
		excluded bool
		reason   string
	}{
		{"秘密ファイル", "data.secret", true, "カスタムルール"},
		{"テストファイル", "test-data.txt", true, "テストファイル"},
		{"機密ファイル", "sensitive-info.json", true, "機密ファイル"},
		{"ドキュメント", "manual.doc", true, "ドキュメントファイル"},
		{"バックアップディレクトリ", "backup/file.txt", true, "バックアップディレクトリ"},
		{"通常ファイル", "normal.txt", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excluded, reason := em.ShouldExclude(tt.filePath)
			if excluded != tt.excluded {
				if tt.excluded {
					t.Errorf("ファイル%sが除外されませんでした", tt.filePath)
				} else {
					t.Errorf("ファイル%sが除外されました（理由: %s）", tt.filePath, reason)
				}
			}

			if tt.excluded && reason != tt.reason {
				t.Errorf("除外理由が期待値と異なります: 期待=%s, 実際=%s", tt.reason, reason)
			}
		})
	}
}

func TestExclusionManager_GenerateExclusionReport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	report := em.GenerateExclusionReport()
	if report == nil {
		t.Fatal("除外レポートがnilです")
	}

	expectedKeys := []string{"enabled", "rules_count", "exclusion_file", "rules_by_type", "rules_by_sensitivity", "sample_patterns"}
	for _, key := range expectedKeys {
		if _, exists := report[key]; !exists {
			t.Errorf("レポートに%sが含まれていません", key)
		}
	}

	if enabled, ok := report["enabled"].(bool); !ok || !enabled {
		t.Error("除外機能が有効と報告されていません")
	}

	if rulesCount, ok := report["rules_count"].(int); !ok || rulesCount == 0 {
		t.Error("ルール数が0と報告されています")
	}

	// タイプ別統計の確認
	rulesByType, ok := report["rules_by_type"].(map[string]int)
	if !ok {
		t.Fatal("タイプ別統計の型が正しくありません")
	}

	totalByType := 0
	for _, count := range rulesByType {
		totalByType += count
	}

	if totalByType != report["rules_count"].(int) {
		t.Errorf("タイプ別統計の合計がルール数と一致しません: 統計合計=%d, ルール数=%d", 
			totalByType, report["rules_count"].(int))
	}

	// 機密度別統計の確認
	rulesBySensitivity, ok := report["rules_by_sensitivity"].(map[string]int)
	if !ok {
		t.Fatal("機密度別統計の型が正しくありません")
	}

	totalBySensitivity := 0
	for _, count := range rulesBySensitivity {
		totalBySensitivity += count
	}

	if totalBySensitivity != report["rules_count"].(int) {
		t.Errorf("機密度別統計の合計がルール数と一致しません: 統計合計=%d, ルール数=%d", 
			totalBySensitivity, report["rules_count"].(int))
	}
}

func TestExclusionManager_Disabled(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 除外機能を無効にする
	os.Setenv("AICT_ENABLE_EXCLUSIONS", "false")
	defer os.Unsetenv("AICT_ENABLE_EXCLUSIONS")

	em := NewExclusionManager(tempDir)

	if em.IsEnabled() {
		t.Error("除外機能が無効になっていません")
	}

	// 無効状態では全てのファイルが含まれる
	excluded, _ := em.ShouldExclude("private.key")
	if excluded {
		t.Error("除外機能が無効なのにファイルが除外されました")
	}
}

func TestExclusionManager_RuleTypes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	// 各タイプのルールを追加
	rules := []struct {
		pattern string
		ruleType ExclusionType
		testPath string
		expected bool
	}{
		{"*.test", ExclusionTypeGlob, "file.test", true},
		{"test.*", ExclusionTypeRegex, "test.anything", true},
		{".tmp", ExclusionTypeExtension, "file.tmp", true},
		{"cache", ExclusionTypeDirectory, "cache/file.txt", true},
		{"README", ExclusionTypeFilename, "docs/README", true},
	}

	for i, rule := range rules {
		t.Run(fmt.Sprintf("RuleType_%d", i), func(t *testing.T) {
			err := em.AddCustomRule(rule.pattern, "テストルール", rule.ruleType, SensitivityLow)
			if err != nil {
				t.Fatalf("ルールの追加に失敗: %v", err)
			}

			excluded, _ := em.ShouldExclude(rule.testPath)
			if excluded != rule.expected {
				t.Errorf("ルールタイプ%dが期待通りに動作していません: パターン=%s, テストパス=%s, 期待=%t, 実際=%t", 
					rule.ruleType, rule.pattern, rule.testPath, rule.expected, excluded)
			}
		})
	}
}

func TestExclusionManager_EdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-exclusion-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewExclusionManager(tempDir)

	tests := []struct {
		name     string
		filePath string
		excluded bool
	}{
		{"空のパス", "", false},
		{"ルートパス", "/", false},
		{"カレントディレクトリ", ".", false},
		{"親ディレクトリ", "..", false},
		{"隠しファイル", ".hidden", false},
		{"非常に長いパス", "a/" + string(make([]byte, 1000)), false},
		{"特殊文字を含むパス", "file with spaces.txt", false},
		{"Unicode文字", "ファイル名.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excluded, _ := em.ShouldExclude(tt.filePath)
			if excluded != tt.excluded {
				t.Errorf("エッジケース%sが期待通りに処理されませんでした: 期待=%t, 実際=%t", 
					tt.filePath, tt.excluded, excluded)
			}
		})
	}
}