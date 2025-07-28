package security

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ai-code-tracker/aict/pkg/types"
)

func TestNewSecurityManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	if sm == nil {
		t.Fatal("SecurityManagerがnilです")
	}

	if sm.dataDir != tempDir {
		t.Errorf("dataDirが期待値と異なります: 期待=%s, 実際=%s", tempDir, sm.dataDir)
	}
}

func TestSecurityManager_ProcessTrackEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	event := &types.TrackEvent{
		ID:        "test-event-001",
		Timestamp: time.Now(),
		EventType: types.EventTypeAI,
		Author:    "test-user",
		Message:   "テストメッセージ",
		Files: []types.FileInfo{
			{
				Path:          "test/file.go",
				LinesAdded:    10,
				LinesDeleted:  5,
				LinesModified: 3,
			},
		},
	}

	processed, err := sm.ProcessTrackEvent(event)
	if err != nil {
		t.Fatalf("ProcessTrackEventが失敗: %v", err)
	}

	if processed == nil {
		t.Fatal("処理されたイベントがnilです")
	}

	if processed.ID != event.ID {
		t.Errorf("IDが変更されました: 期待=%s, 実際=%s", event.ID, processed.ID)
	}
}

func TestSecurityManager_EncryptDecryptData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 暗号化を有効にする
	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	originalData := []byte("機密データのテスト")

	// 暗号化テスト
	encrypted, err := sm.EncryptData(originalData)
	if err != nil {
		t.Fatalf("データの暗号化に失敗: %v", err)
	}

	if len(encrypted) == 0 {
		t.Fatal("暗号化されたデータが空です")
	}

	// 復号化テスト
	decrypted, err := sm.DecryptData(encrypted)
	if err != nil {
		t.Fatalf("データの復号化に失敗: %v", err)
	}

	if string(decrypted) != string(originalData) {
		t.Errorf("復号化されたデータが元のデータと異なります: 期待=%s, 実際=%s", string(originalData), string(decrypted))
	}
}

func TestSecurityManager_UpdateSecurityConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	updates := map[string]interface{}{
		"enable_audit_log":      true,
		"encrypt_sensitive_data": true,
	}

	err = sm.UpdateSecurityConfig(updates)
	if err != nil {
		t.Fatalf("設定の更新に失敗: %v", err)
	}

	if !sm.config.EnableAuditLog {
		t.Error("監査ログが有効になっていません")
	}

	if !sm.config.EncryptSensitiveData {
		t.Error("データ暗号化が有効になっていません")
	}
}

func TestSecurityManager_PerformSecurityScan(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	results, err := sm.PerformSecurityScan()
	if err != nil {
		t.Fatalf("セキュリティスキャンに失敗: %v", err)
	}

	if results == nil {
		t.Fatal("スキャン結果がnilです")
	}

	if _, exists := results["score"]; !exists {
		t.Error("スコアが結果に含まれていません")
	}

	if _, exists := results["summary"]; !exists {
		t.Error("サマリーが結果に含まれていません")
	}

	if _, exists := results["issues"]; !exists {
		t.Error("問題一覧が結果に含まれていません")
	}
}

func TestSecurityManager_GetSecurityStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	status := sm.GetSecurityStatus()
	if status == nil {
		t.Fatal("ステータスがnilです")
	}

	expectedKeys := []string{"mode", "encryption", "audit", "privacy", "exclusions", "validation"}
	for _, key := range expectedKeys {
		if _, exists := status[key]; !exists {
			t.Errorf("ステータスに%sが含まれていません", key)
		}
	}
}

func TestSecurityManager_InvalidConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	// 無効な設定値をテスト
	invalidUpdates := map[string]interface{}{
		"mode":          "invalid_mode",
		"max_file_size": -1,
	}

	err = sm.UpdateSecurityConfig(invalidUpdates)
	if err == nil {
		t.Error("無効な設定値が受け入れられました")
	}
}

func TestSecurityManager_FileExclusion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	// 除外されるべきファイルでイベントを作成
	event := &types.TrackEvent{
		ID:        "test-exclude-001",
		Timestamp: time.Now(),
		EventType: types.EventTypeAI,
		Author:    "test-user",
		Message:   "テストメッセージ",
		Files: []types.FileInfo{
			{
				Path:         "secret.key", // 除外されるべき
				LinesAdded:   1,
				LinesDeleted:  0,
				LinesModified: 0,
			},
		},
	}

	_, err = sm.ProcessTrackEvent(event)
	if err == nil {
		t.Error("除外されるべきファイルが処理されました")
	}
}

func TestSecurityManager_IntegrationWithComponents(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 全機能を有効にする
	os.Setenv("AICT_AUDIT_LOG", "true")
	os.Setenv("AICT_ENCRYPT_DATA", "true")
	os.Setenv("AICT_ANONYMIZE_AUTHORS", "true")
	defer func() {
		os.Unsetenv("AICT_AUDIT_LOG")
		os.Unsetenv("AICT_ENCRYPT_DATA")
		os.Unsetenv("AICT_ANONYMIZE_AUTHORS")
	}()

	sm, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	// 各コンポーネントが正しく初期化されているかテスト
	if sm.GetEncryptionManager() == nil {
		t.Error("EncryptionManagerが初期化されていません")
	}

	if sm.GetAuditManager() == nil {
		t.Error("AuditManagerが初期化されていません")
	}

	if sm.GetValidationManager() == nil {
		t.Error("ValidationManagerが初期化されていません")
	}

	if sm.GetPrivacyManager() == nil {
		t.Error("PrivacyManagerが初期化されていません")
	}

	if sm.GetExclusionManager() == nil {
		t.Error("ExclusionManagerが初期化されていません")
	}

	// 統合テスト: 正常なイベント処理
	event := &types.TrackEvent{
		ID:        "integration-test-001",
		Timestamp: time.Now(),
		EventType: types.EventTypeAI,
		Author:    "human-user",
		Message:   "統合テストメッセージ",
		Files: []types.FileInfo{
			{
				Path:         "src/main.go",
				LinesAdded:   20,
				LinesDeleted:  5,
				LinesModified: 10,
			},
		},
	}

	processed, err := sm.ProcessTrackEvent(event)
	if err != nil {
		t.Fatalf("統合テストが失敗: %v", err)
	}

	// プライバシー設定により作成者が匿名化されているかチェック
	if processed.Author == event.Author {
		t.Error("作成者が匿名化されていません")
	}

	// セキュリティスキャンテスト
	scanResults, err := sm.PerformSecurityScan()
	if err != nil {
		t.Fatalf("統合セキュリティスキャンが失敗: %v", err)
	}

	if score, ok := scanResults["score"].(int); !ok || score < 0 {
		t.Error("セキュリティスコアが無効です")
	}
}

func TestSecurityConfigSaveLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-security-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 最初のマネージャーで設定を保存
	sm1, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	updates := map[string]interface{}{
		"enable_audit_log":      true,
		"encrypt_sensitive_data": true,
	}

	err = sm1.UpdateSecurityConfig(updates)
	if err != nil {
		t.Fatalf("設定の更新に失敗: %v", err)
	}

	// 新しいマネージャーで設定をロード
	sm2, err := NewSecurityManager(tempDir)
	if err != nil {
		t.Fatalf("SecurityManagerの作成に失敗: %v", err)
	}

	// 設定が保持されているかチェック
	if !sm2.config.EnableAuditLog {
		t.Error("監査ログ設定が保持されていません")
	}

	if !sm2.config.EncryptSensitiveData {
		t.Error("暗号化設定が保持されていません")
	}

	// 設定ファイルが存在するかチェック
	configPath := filepath.Join(tempDir, "security-config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("設定ファイルが作成されていません")
	}
}