package security

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewAuditManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 監査ログを有効にする
	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)
	if am == nil {
		t.Fatal("AuditManagerがnilです")
	}

	if !am.IsEnabled() {
		t.Error("監査ログが有効になっていません")
	}
}

func TestAuditManager_LogEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	details := map[string]interface{}{
		"test_key": "test_value",
		"count":    42,
	}

	am.LogEvent("test_event", "test_resource", "test_action", true, details, nil)

	// ログファイルが作成されているかチェック
	if _, err := os.Stat(am.logFile); os.IsNotExist(err) {
		t.Error("監査ログファイルが作成されていません")
	}

	// ログ内容の確認
	logs, err := am.GetAuditLogs(10, nil)
	if err != nil {
		t.Fatalf("監査ログの取得に失敗: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("期待されるログ数と異なります: 期待=1, 実際=%d", len(logs))
	}

	log := logs[0]
	if log.Event != "test_event" {
		t.Errorf("イベント名が期待値と異なります: 期待=test_event, 実際=%s", log.Event)
	}

	if log.Resource != "test_resource" {
		t.Errorf("リソース名が期待値と異なります: 期待=test_resource, 実際=%s", log.Resource)
	}

	if log.Action != "test_action" {
		t.Errorf("アクション名が期待値と異なります: 期待=test_action, 実際=%s", log.Action)
	}

	if !log.Success {
		t.Error("成功フラグがfalseです")
	}
}

func TestAuditManager_LogFileAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	am.LogFileAccess("/path/to/file.go", "read", true, nil)

	logs, err := am.GetAuditLogs(10, nil)
	if err != nil {
		t.Fatalf("監査ログの取得に失敗: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("期待されるログ数と異なります: 期待=1, 実際=%d", len(logs))
	}

	log := logs[0]
	if log.Event != "file_access" {
		t.Errorf("イベント名が期待値と異なります: 期待=file_access, 実際=%s", log.Event)
	}

	if filePath, ok := log.Details["file_path"].(string); !ok || filePath != "/path/to/file.go" {
		t.Errorf("ファイルパスが期待値と異なります: 期待=/path/to/file.go, 実際=%v", log.Details["file_path"])
	}
}

func TestAuditManager_LogDataOperation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	am.LogDataOperation("create", "events", 5, true, nil)

	logs, err := am.GetAuditLogs(10, nil)
	if err != nil {
		t.Fatalf("監査ログの取得に失敗: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("期待されるログ数と異なります: 期待=1, 実際=%d", len(logs))
	}

	log := logs[0]
	if log.Event != "data_operation" {
		t.Errorf("イベント名が期待値と異なります: 期待=data_operation, 実際=%s", log.Event)
	}

	if recordCount, ok := log.Details["record_count"].(int); ok && recordCount == 5 {
		// テストが期待通りに動作していることを確認
		t.Logf("レコード数が正常に記録されました: %d", recordCount)
	} else {
		t.Errorf("レコード数が期待値と異なります: 期待=5, 実際=%v", log.Details["record_count"])
	}
}

func TestAuditManager_LogSecurityEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	details := map[string]interface{}{
		"ip_address": "192.168.1.100",
	}

	am.LogSecurityEvent("unauthorized_access", "不正アクセスの試行", "high", details)

	logs, err := am.GetAuditLogs(10, nil)
	if err != nil {
		t.Fatalf("監査ログの取得に失敗: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("期待されるログ数と異なります: 期待=1, 実際=%d", len(logs))
	}

	log := logs[0]
	if log.Event != "security_event" {
		t.Errorf("イベント名が期待値と異なります: 期待=security_event, 実際=%s", log.Event)
	}

	if severity, ok := log.Details["severity"].(string); !ok || severity != "high" {
		t.Errorf("深刻度が期待値と異なります: 期待=high, 実際=%v", log.Details["severity"])
	}

	if description, ok := log.Details["description"].(string); !ok || description != "不正アクセスの試行" {
		t.Errorf("説明が期待値と異なります: 期待=不正アクセスの試行, 実際=%v", log.Details["description"])
	}
}

func TestAuditManager_FilterLogs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	// 複数のログエントリを作成
	am.LogEvent("event1", "resource1", "action1", true, nil, nil)
	am.LogEvent("event2", "resource2", "action2", false, nil, nil)
	am.LogEvent("event1", "resource3", "action1", true, nil, nil)

	// イベントタイプでフィルタ
	filter := map[string]interface{}{
		"event": "event1",
	}

	filteredLogs, err := am.GetAuditLogs(10, filter)
	if err != nil {
		t.Fatalf("フィルタされたログの取得に失敗: %v", err)
	}

	if len(filteredLogs) != 2 {
		t.Fatalf("フィルタ結果の数が期待値と異なります: 期待=2, 実際=%d", len(filteredLogs))
	}

	for _, log := range filteredLogs {
		if log.Event != "event1" {
			t.Errorf("フィルタが正しく動作していません: %s", log.Event)
		}
	}

	// 成功/失敗でフィルタ
	successFilter := map[string]interface{}{
		"success": false,
	}

	failedLogs, err := am.GetAuditLogs(10, successFilter)
	if err != nil {
		t.Fatalf("失敗ログの取得に失敗: %v", err)
	}

	if len(failedLogs) != 1 {
		t.Fatalf("失敗ログ数が期待値と異なります: 期待=1, 実際=%d", len(failedLogs))
	}

	if failedLogs[0].Success {
		t.Error("失敗ログのフィルタが正しく動作していません")
	}
}

func TestAuditManager_GetAuditSummary(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	os.Setenv("USER", "test-user")
	defer func() {
		os.Unsetenv("AICT_AUDIT_LOG")
		os.Unsetenv("USER")
	}()

	am := NewAuditManager(tempDir)

	// 複数のログエントリを作成
	am.LogEvent("event1", "resource1", "action1", true, nil, nil)
	am.LogEvent("event2", "resource2", "action2", false, nil, nil)
	am.LogEvent("event1", "resource3", "action1", true, nil, nil)

	since := time.Now().Add(-1 * time.Hour)
	summary, err := am.GetAuditSummary(since)
	if err != nil {
		t.Fatalf("サマリーの取得に失敗: %v", err)
	}

	if totalEvents, ok := summary["total_events"].(int); !ok || totalEvents != 3 {
		t.Errorf("総イベント数が期待値と異なります: 期待=3, 実際=%v", summary["total_events"])
	}

	if successCount, ok := summary["success_count"].(int); !ok || successCount != 2 {
		t.Errorf("成功数が期待値と異なります: 期待=2, 実際=%v", summary["success_count"])
	}

	if failureCount, ok := summary["failure_count"].(int); !ok || failureCount != 1 {
		t.Errorf("失敗数が期待値と異なります: 期待=1, 実際=%v", summary["failure_count"])
	}

	if mostActiveUser, ok := summary["most_active_user"].(string); !ok || mostActiveUser != "test-user" {
		t.Errorf("最もアクティブなユーザーが期待値と異なります: 期待=test-user, 実際=%v", summary["most_active_user"])
	}
}

func TestAuditManager_RotateAuditLog(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	// 初期ログエントリを作成
	am.LogEvent("initial_event", "resource", "action", true, nil, nil)

	// ローテーション実行
	err = am.RotateAuditLog()
	if err != nil {
		t.Fatalf("ログローテーションに失敗: %v", err)
	}

	// 新しいログエントリを作成
	am.LogEvent("after_rotation", "resource", "action", true, nil, nil)

	// 現在のログファイルに新しいエントリのみが含まれているかチェック
	logs, err := am.GetAuditLogs(10, nil)
	if err != nil {
		t.Fatalf("監査ログの取得に失敗: %v", err)
	}

	// ローテーション後のログには2つのエントリがあるはず（ローテーションイベント + 新しいイベント）
	if len(logs) < 1 {
		t.Errorf("ローテーション後のログ数が期待値より少ないです: 実際=%d", len(logs))
	}

	// 最新のログがafter_rotationイベントかチェック
	found := false
	for _, log := range logs {
		if log.Event == "after_rotation" {
			found = true
			break
		}
	}

	if !found {
		t.Error("ローテーション後の新しいイベントが見つかりません")
	}
}

func TestAuditManager_GetAuditStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	os.Setenv("USER", "test-user")
	defer func() {
		os.Unsetenv("AICT_AUDIT_LOG")
		os.Unsetenv("USER")
	}()

	am := NewAuditManager(tempDir)

	status := am.GetAuditStatus()
	if status == nil {
		t.Fatal("ステータスがnilです")
	}

	if enabled, ok := status["enabled"].(bool); !ok || !enabled {
		t.Error("監査ログが有効になっていません")
	}

	if currentUser, ok := status["current_user"].(string); !ok || currentUser != "test-user" {
		t.Errorf("現在のユーザーが期待値と異なります: 期待=test-user, 実際=%v", status["current_user"])
	}

	// ログファイルを作成後のステータス
	am.LogEvent("test", "resource", "action", true, nil, nil)

	status = am.GetAuditStatus()
	if fileExists, ok := status["file_exists"].(bool); !ok || !fileExists {
		t.Error("ファイル存在フラグがfalseです")
	}

	if fileSize, ok := status["file_size"].(int64); !ok || fileSize == 0 {
		t.Error("ファイルサイズが0です")
	}
}

func TestAuditManager_Disabled(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 監査ログを無効にする
	os.Setenv("AICT_AUDIT_LOG", "false")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	if am.IsEnabled() {
		t.Error("監査ログが無効になっていません")
	}

	// 無効状態でのログ操作
	am.LogEvent("test_event", "resource", "action", true, nil, nil)

	// ログファイルが作成されていないことを確認
	if _, err := os.Stat(am.logFile); !os.IsNotExist(err) {
		t.Error("監査ログが無効なのにファイルが作成されています")
	}

	// 無効状態でのログ取得
	_, err = am.GetAuditLogs(10, nil)
	if err == nil {
		t.Error("監査ログが無効なのにログ取得がエラーになりませんでした")
	}
}

func TestAuditManager_JSONLFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-audit-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_AUDIT_LOG", "true")
	defer os.Unsetenv("AICT_AUDIT_LOG")

	am := NewAuditManager(tempDir)

	details := map[string]interface{}{
		"test_key": "test_value",
		"count":    42,
	}

	am.LogEvent("format_test", "resource", "action", true, details, nil)

	// ログファイルを直接読んで形式をチェック
	data, err := os.ReadFile(am.logFile)
	if err != nil {
		t.Fatalf("ログファイルの読み込みに失敗: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("期待されるログ行数と異なります: 期待=1, 実際=%d", len(lines))
	}

	// JSONパース可能かチェック
	var event AuditEvent
	err = json.Unmarshal([]byte(lines[0]), &event)
	if err != nil {
		t.Fatalf("ログのJSON解析に失敗: %v", err)
	}

	if event.Event != "format_test" {
		t.Errorf("イベント名が期待値と異なります: 期待=format_test, 実際=%s", event.Event)
	}

	if testKey, ok := event.Details["test_key"].(string); !ok || testKey != "test_value" {
		t.Errorf("詳細情報が期待値と異なります: 期待=test_value, 実際=%v", event.Details["test_key"])
	}
}