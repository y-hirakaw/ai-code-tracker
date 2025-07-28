package security

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

func TestNewPrivacyManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewPrivacyManager(tempDir)
	if pm == nil {
		t.Fatal("PrivacyManagerがnilです")
	}

	if pm.dataRetentionDays != 365 {
		t.Errorf("デフォルトのデータ保持期間が期待値と異なります: 期待=365, 実際=%d", pm.dataRetentionDays)
	}
}

func TestPrivacyManager_AnonymizeAuthor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 匿名化を有効にする
	os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	defer os.Unsetenv("AICT_ANONYMIZE_AUTHORS")

	pm := NewPrivacyManager(tempDir)

	tests := []struct {
		name     string
		author   string
		expected string
	}{
		{"人間のユーザー", "john.doe", ""}, // 匿名化されるので空文字列ではない
		{"Claude Code", "Claude Code", "AI Assistant"},
		{"GitHub Copilot", "GitHub Copilot", "AI Assistant"},
		{"claude", "claude", "AI Assistant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.AnonymizeAuthor(tt.author)
			
			if tt.expected == "" {
				// 人間のユーザーは匿名化される（ハッシュ化される）
				if result == tt.author {
					t.Error("人間のユーザーが匿名化されていません")
				}
				if len(result) != 8 {
					t.Errorf("匿名化されたユーザー名の長さが期待値と異なります: 期待=8, 実際=%d", len(result))
				}
			} else {
				// AIアシスタントは統一表記になる
				if result != tt.expected {
					t.Errorf("AIアシスタントの匿名化が期待値と異なります: 期待=%s, 実際=%s", tt.expected, result)
				}
			}
		})
	}
}

func TestPrivacyManager_AnonymizeFilePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// ファイルパスハッシュ化を有効にする
	os.Setenv("AICT_HASH_FILE_PATHS", "true")
	defer os.Unsetenv("AICT_HASH_FILE_PATHS")

	pm := NewPrivacyManager(tempDir)

	tests := []struct {
		name     string
		filePath string
	}{
		{"単純なファイル", "main.go"},
		{"ディレクトリ付き", "src/main.go"},
		{"深い階層", "project/src/components/main.go"},
		{"日本語パス", "プロジェクト/ソース/main.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.AnonymizeFilePath(tt.filePath)
			
			// 元のファイルパスと異なることを確認
			if result == tt.filePath {
				t.Error("ファイルパスが匿名化されていません")
			}
			
			// 拡張子が保持されていることを確認
			if filepath.Ext(result) != filepath.Ext(tt.filePath) {
				t.Errorf("拡張子が保持されていません: 元=%s, 結果=%s", 
					filepath.Ext(tt.filePath), filepath.Ext(result))
			}
		})
	}
}

func TestPrivacyManager_ProcessTimestamp(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	now := time.Now()

	// タイムスタンプ除去を無効にする
	pm1 := NewPrivacyManager(tempDir)
	result1 := pm1.ProcessTimestamp(now)
	if result1 == nil || !result1.Equal(now) {
		t.Error("タイムスタンプが保持されていません")
	}

	// タイムスタンプ除去を有効にする
	os.Setenv("AICT_REMOVE_TIMESTAMPS", "true")
	defer os.Unsetenv("AICT_REMOVE_TIMESTAMPS")

	pm2 := NewPrivacyManager(tempDir)
	result2 := pm2.ProcessTimestamp(now)
	if result2 != nil {
		t.Error("タイムスタンプが除去されていません")
	}
}

func TestPrivacyManager_SanitizeMessage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewPrivacyManager(tempDir)

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"通常のメッセージ", "ファイルを更新しました", "ファイルを更新しました"},
		{"パスワードを含む", "パスワードをmypassword123に設定", "パスワードを[REDACTED]に設定"},
		{"APIキーを含む", "API_KEY=abc123def456", "API_KEY=[REDACTED]"},
		{"秘密を含む", "秘密のトークンは secret123 です", "秘密のトークンは [REDACTED] です"},
		{"複数の機密情報", "パスワード pass123 とキー key456", "パスワード [REDACTED] とキー [REDACTED]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.SanitizeMessage(tt.message)
			if result != tt.expected {
				t.Errorf("メッセージサニタイズが期待値と異なります: 期待=%s, 実際=%s", tt.expected, result)
			}
		})
	}
}

func TestPrivacyManager_ProcessTrackEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 全ての匿名化機能を有効にする
	os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	os.Setenv("AICT_HASH_FILE_PATHS", "true")
	os.Setenv("AICT_REMOVE_TIMESTAMPS", "true")
	defer func() {
		os.Unsetenv("AICT_ANONYMIZE_AUTHORS")
		os.Unsetenv("AICT_HASH_FILE_PATHS")
		os.Unsetenv("AICT_REMOVE_TIMESTAMPS")
	}()

	pm := NewPrivacyManager(tempDir)

	originalEvent := &types.TrackEvent{
		ID:        "test-001",
		Timestamp: time.Now(),
		EventType: types.EventTypeAI,
		Author:    "john.doe",
		Message:   "パスワードをpassword123に設定しました",
		Files: []types.FileInfo{
			{
				Path:         "src/config.go",
				LinesAdded:   10,
				LinesDeleted:  5,
				LinesModified: 3,
			},
		},
	}

	processed := pm.ProcessTrackEvent(originalEvent)

	// 作成者が匿名化されているかチェック
	if processed.Author == originalEvent.Author {
		t.Error("作成者が匿名化されていません")
	}

	// メッセージがサニタイズされているかチェック
	if processed.Message == originalEvent.Message {
		t.Error("メッセージがサニタイズされていません")
	}

	// タイムスタンプが除去されているかチェック
	if !processed.Timestamp.IsZero() {
		t.Error("タイムスタンプが除去されていません")
	}

	// ファイルパスが匿名化されているかチェック
	if processed.Files[0].Path == originalEvent.Files[0].Path {
		t.Error("ファイルパスが匿名化されていません")
	}

	// IDは変更されないことを確認
	if processed.ID != originalEvent.ID {
		t.Error("IDが変更されました")
	}
}

func TestPrivacyManager_DataRetention(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// データ保持期間を30日に設定
	os.Setenv("AICT_DATA_RETENTION_DAYS", "30")
	defer os.Unsetenv("AICT_DATA_RETENTION_DAYS")

	pm := NewPrivacyManager(tempDir)

	now := time.Now()
	recent := now.Add(-10 * 24 * time.Hour)  // 10日前
	old := now.Add(-40 * 24 * time.Hour)     // 40日前

	if !pm.ShouldRetainData(recent) {
		t.Error("最近のデータが保持されません")
	}

	if pm.ShouldRetainData(old) {
		t.Error("古いデータが保持されています")
	}

	threshold := pm.GetExpiredDataThreshold()
	if threshold.IsZero() {
		t.Error("期限切れデータの閾値が設定されていません")
	}

	// 閾値より前のデータは期限切れ
	if !old.Before(threshold) {
		t.Error("期限切れ閾値が正しく設定されていません")
	}
}

func TestPrivacyManager_GetDataRetentionPolicy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_DATA_RETENTION_DAYS", "90")
	os.Setenv("AICT_AUTO_CLEANUP", "true")
	defer func() {
		os.Unsetenv("AICT_DATA_RETENTION_DAYS")
		os.Unsetenv("AICT_AUTO_CLEANUP")
	}()

	pm := NewPrivacyManager(tempDir)

	policy := pm.GetDataRetentionPolicy()
	if policy == nil {
		t.Fatal("データ保持ポリシーがnilです")
	}

	if retentionDays, ok := policy["retention_days"].(int); !ok || retentionDays != 90 {
		t.Errorf("保持期間が期待値と異なります: 期待=90, 実際=%v", policy["retention_days"])
	}

	if autoCleanup, ok := policy["auto_cleanup"].(bool); !ok || !autoCleanup {
		t.Error("自動クリーンアップが有効になっていません")
	}

	if policyActive, ok := policy["policy_active"].(bool); !ok || !policyActive {
		t.Error("ポリシーがアクティブになっていません")
	}
}

func TestPrivacyManager_GeneratePrivacyReport(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	os.Setenv("AICT_HASH_FILE_PATHS", "true")
	defer func() {
		os.Unsetenv("AICT_ANONYMIZE_AUTHORS")
		os.Unsetenv("AICT_HASH_FILE_PATHS")
	}()

	pm := NewPrivacyManager(tempDir)

	report := pm.GeneratePrivacyReport()
	if report == nil {
		t.Fatal("プライバシーレポートがnilです")
	}

	expectedSections := []string{"anonymization", "data_retention", "sensitive_patterns", "security_features"}
	for _, section := range expectedSections {
		if _, exists := report[section]; !exists {
			t.Errorf("レポートに%sセクションが含まれていません", section)
		}
	}

	// 匿名化設定の確認
	if anonymization, ok := report["anonymization"].(map[string]interface{}); ok {
		if anonymizeAuthors, ok := anonymization["anonymize_authors"].(bool); !ok || !anonymizeAuthors {
			t.Error("作成者匿名化設定が正しく反映されていません")
		}
		if hashFilePaths, ok := anonymization["hash_file_paths"].(bool); !ok || !hashFilePaths {
			t.Error("ファイルパスハッシュ化設定が正しく反映されていません")
		}
	} else {
		t.Error("匿名化セクションが正しい形式ではありません")
	}
}

func TestPrivacyManager_ValidatePrivacySettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewPrivacyManager(tempDir)

	// 有効な設定
	err = pm.ValidatePrivacySettings()
	if err != nil {
		t.Errorf("有効な設定でエラーが発生しました: %v", err)
	}

	// 無効な設定: 負の保持期間
	pm.dataRetentionDays = -1
	err = pm.ValidatePrivacySettings()
	if err == nil {
		t.Error("負の保持期間がエラーになりませんでした")
	}

	// 無効な設定: 短すぎる保持期間
	pm.dataRetentionDays = 3
	err = pm.ValidatePrivacySettings()
	if err == nil {
		t.Error("短すぎる保持期間がエラーになりませんでした")
	}

	// 無効な設定: 空のソルト
	pm.dataRetentionDays = 365
	pm.anonymizationSalt = ""
	err = pm.ValidatePrivacySettings()
	if err == nil {
		t.Error("空のソルトがエラーになりませんでした")
	}

	// 無効な設定: 短すぎるソルト
	pm.anonymizationSalt = "short"
	err = pm.ValidatePrivacySettings()
	if err == nil {
		t.Error("短すぎるソルトがエラーになりませんでした")
	}
}

func TestPrivacyManager_UpdatePrivacySettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pm := NewPrivacyManager(tempDir)

	updates := map[string]interface{}{
		"anonymize_authors":    true,
		"hash_file_paths":      true,
		"remove_timestamps":    true,
		"data_retention_days":  180,
		"auto_cleanup":         true,
	}

	err = pm.UpdatePrivacySettings(updates)
	if err != nil {
		t.Fatalf("設定の更新に失敗: %v", err)
	}

	if !pm.anonymizeAuthors {
		t.Error("作成者匿名化設定が更新されていません")
	}

	if !pm.hashFilePaths {
		t.Error("ファイルパスハッシュ化設定が更新されていません")
	}

	if !pm.removeTimestamps {
		t.Error("タイムスタンプ除去設定が更新されていません")
	}

	if pm.dataRetentionDays != 180 {
		t.Errorf("データ保持期間が更新されていません: 期待=180, 実際=%d", pm.dataRetentionDays)
	}

	if !pm.autoCleanup {
		t.Error("自動クリーンアップ設定が更新されていません")
	}

	// 無効な設定の更新
	invalidUpdates := map[string]interface{}{
		"data_retention_days": -1,
	}

	err = pm.UpdatePrivacySettings(invalidUpdates)
	if err == nil {
		t.Error("無効な設定の更新がエラーになりませんでした")
	}
}

func TestPrivacyManager_GetPrivacyImpact(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 全てのプライバシー機能を有効にする
	os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	os.Setenv("AICT_HASH_FILE_PATHS", "true")
	os.Setenv("AICT_REMOVE_TIMESTAMPS", "true")
	os.Setenv("AICT_DATA_RETENTION_DAYS", "90")
	defer func() {
		os.Unsetenv("AICT_ANONYMIZE_AUTHORS")
		os.Unsetenv("AICT_HASH_FILE_PATHS")
		os.Unsetenv("AICT_REMOVE_TIMESTAMPS")
		os.Unsetenv("AICT_DATA_RETENTION_DAYS")
	}()

	pm := NewPrivacyManager(tempDir)

	impact := pm.GetPrivacyImpact()
	if impact == nil {
		t.Fatal("プライバシー影響分析がnilです")
	}

	expectedSections := []string{"data_protection_level", "reversibility", "functionality_impact", "compliance"}
	for _, section := range expectedSections {
		if _, exists := impact[section]; !exists {
			t.Errorf("影響分析に%sセクションが含まれていません", section)
		}
	}

	// 保護レベルの確認
	if protectionLevel, ok := impact["data_protection_level"].(string); !ok || protectionLevel != "high" {
		t.Errorf("データ保護レベルが期待値と異なります: 期待=high, 実際=%v", impact["data_protection_level"])
	}

	// 可逆性の確認
	if reversibility, ok := impact["reversibility"].(map[string]bool); ok {
		if reversibility["author_anonymization"] {
			t.Error("作成者匿名化が可逆と判定されています")
		}
		if reversibility["file_path_hashing"] {
			t.Error("ファイルパスハッシュ化が可逆と判定されています")
		}
	} else {
		t.Error("可逆性セクションが正しい形式ではありません")
	}
}

func TestPrivacyManager_EnvironmentVariables(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-privacy-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 環境変数を設定
	os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	os.Setenv("AICT_HASH_FILE_PATHS", "true")
	os.Setenv("AICT_REMOVE_TIMESTAMPS", "true")
	os.Setenv("AICT_AUTO_CLEANUP", "true")
	os.Setenv("AICT_DATA_RETENTION_DAYS", "180")
	os.Setenv("AICT_ANONYMIZATION_SALT", "custom-salt-for-testing")

	defer func() {
		os.Unsetenv("AICT_ANONYMIZE_AUTHORS")
		os.Unsetenv("AICT_HASH_FILE_PATHS")
		os.Unsetenv("AICT_REMOVE_TIMESTAMPS")
		os.Unsetenv("AICT_AUTO_CLEANUP")
		os.Unsetenv("AICT_DATA_RETENTION_DAYS")
		os.Unsetenv("AICT_ANONYMIZATION_SALT")
	}()

	pm := NewPrivacyManager(tempDir)

	if !pm.anonymizeAuthors {
		t.Error("環境変数による作成者匿名化設定が反映されていません")
	}

	if !pm.hashFilePaths {
		t.Error("環境変数によるファイルパスハッシュ化設定が反映されていません")
	}

	if !pm.removeTimestamps {
		t.Error("環境変数によるタイムスタンプ除去設定が反映されていません")
	}

	if !pm.autoCleanup {
		t.Error("環境変数による自動クリーンアップ設定が反映されていません")
	}

	if pm.dataRetentionDays != 180 {
		t.Errorf("環境変数によるデータ保持期間設定が反映されていません: 期待=180, 実際=%d", pm.dataRetentionDays)
	}

	if pm.anonymizationSalt != "custom-salt-for-testing" {
		t.Errorf("環境変数による匿名化ソルト設定が反映されていません: 期待=custom-salt-for-testing, 実際=%s", pm.anonymizationSalt)
	}
}