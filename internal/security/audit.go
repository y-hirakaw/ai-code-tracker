package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuditEvent は監査ログのイベントを表す
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Event     string                 `json:"event"`
	User      string                 `json:"user"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Success   bool                   `json:"success"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
	ClientIP  string                 `json:"client_ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
}

// AuditManager は監査ログを管理する
type AuditManager struct {
	enabled bool
	logFile string
	user    string
}

// NewAuditManager は新しい監査マネージャーを作成する
func NewAuditManager(dataDir string) *AuditManager {
	enabled := os.Getenv("AICT_AUDIT_LOG") == "true"
	logFile := filepath.Join(dataDir, "audit.jsonl")
	
	// 現在のユーザーを取得
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}
	
	return &AuditManager{
		enabled: enabled,
		logFile: logFile,
		user:    user,
	}
}

// IsEnabled は監査ログが有効かどうかを返す
func (am *AuditManager) IsEnabled() bool {
	return am.enabled
}

// LogEvent は監査イベントをログに記録する
func (am *AuditManager) LogEvent(event, resource, action string, success bool, details map[string]interface{}, err error) {
	if !am.enabled {
		return
	}

	auditEvent := AuditEvent{
		Timestamp: time.Now(),
		Event:     event,
		User:      am.user,
		Resource:  resource,
		Action:    action,
		Success:   success,
		Details:   details,
	}

	if err != nil {
		auditEvent.Error = err.Error()
	}

	// 環境変数から追加情報を取得
	if clientIP := os.Getenv("AICT_CLIENT_IP"); clientIP != "" {
		auditEvent.ClientIP = clientIP
	}
	if userAgent := os.Getenv("AICT_USER_AGENT"); userAgent != "" {
		auditEvent.UserAgent = userAgent
	}

	am.writeAuditLog(auditEvent)
}

// LogFileAccess はファイルアクセスをログに記録する
func (am *AuditManager) LogFileAccess(filePath, action string, success bool, err error) {
	details := map[string]interface{}{
		"file_path": filePath,
	}
	
	if !success && err != nil {
		details["error_type"] = fmt.Sprintf("%T", err)
	}

	am.LogEvent("file_access", filePath, action, success, details, err)
}

// LogDataOperation はデータ操作をログに記録する
func (am *AuditManager) LogDataOperation(operation, resource string, recordCount int, success bool, err error) {
	details := map[string]interface{}{
		"operation":    operation,
		"record_count": recordCount,
	}

	am.LogEvent("data_operation", resource, operation, success, details, err)
}

// LogAuthentication は認証イベントをログに記録する
func (am *AuditManager) LogAuthentication(method, result string, success bool, err error) {
	details := map[string]interface{}{
		"auth_method": method,
		"result":      result,
	}

	am.LogEvent("authentication", "system", "authenticate", success, details, err)
}

// LogConfigChange は設定変更をログに記録する
func (am *AuditManager) LogConfigChange(configType, oldValue, newValue string, success bool, err error) {
	details := map[string]interface{}{
		"config_type": configType,
		"old_value":   oldValue,
		"new_value":   newValue,
	}

	am.LogEvent("config_change", configType, "modify", success, details, err)
}

// LogHookExecution はhook実行をログに記録する
func (am *AuditManager) LogHookExecution(hookType, hookPath string, exitCode int, success bool, err error) {
	details := map[string]interface{}{
		"hook_type": hookType,
		"hook_path": hookPath,
		"exit_code": exitCode,
	}

	am.LogEvent("hook_execution", hookPath, "execute", success, details, err)
}

// LogSecurityEvent はセキュリティイベントをログに記録する
func (am *AuditManager) LogSecurityEvent(eventType, description string, severity string, details map[string]interface{}) {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["severity"] = severity
	details["description"] = description

	am.LogEvent("security_event", "system", eventType, true, details, nil)
}

// writeAuditLog は監査ログをファイルに書き込む
func (am *AuditManager) writeAuditLog(event AuditEvent) {
	// ログディレクトリを作成
	if err := os.MkdirAll(filepath.Dir(am.logFile), 0700); err != nil {
		// ログ書き込みでエラーが発生した場合はstderrに出力
		fmt.Fprintf(os.Stderr, "監査ログディレクトリの作成に失敗: %v\n", err)
		return
	}

	// JSONLINEとしてシリアライズ
	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "監査ログのシリアライズに失敗: %v\n", err)
		return
	}

	// ファイルに追記
	file, err := os.OpenFile(am.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "監査ログファイルのオープンに失敗: %v\n", err)
		return
	}
	defer file.Close()

	// データを書き込み
	if _, err := file.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "監査ログの書き込みに失敗: %v\n", err)
	}
}

// GetAuditLogs は監査ログを取得する
func (am *AuditManager) GetAuditLogs(limit int, filter map[string]interface{}) ([]AuditEvent, error) {
	if !am.enabled {
		return nil, fmt.Errorf("監査ログが無効です")
	}

	file, err := os.Open(am.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []AuditEvent{}, nil
		}
		return nil, fmt.Errorf("監査ログファイルのオープンに失敗: %w", err)
	}
	defer file.Close()

	var events []AuditEvent
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var event AuditEvent
		if err := decoder.Decode(&event); err != nil {
			continue // 不正な行はスキップ
		}

		// フィルタ適用
		if am.matchesFilter(event, filter) {
			events = append(events, event)
		}

		// 制限チェック
		if limit > 0 && len(events) >= limit {
			break
		}
	}

	return events, nil
}

// matchesFilter はイベントがフィルタにマッチするかチェックする
func (am *AuditManager) matchesFilter(event AuditEvent, filter map[string]interface{}) bool {
	if filter == nil {
		return true
	}

	for key, value := range filter {
		switch key {
		case "event":
			if event.Event != value.(string) {
				return false
			}
		case "user":
			if event.User != value.(string) {
				return false
			}
		case "action":
			if event.Action != value.(string) {
				return false
			}
		case "success":
			if event.Success != value.(bool) {
				return false
			}
		case "since":
			if since, ok := value.(time.Time); ok && event.Timestamp.Before(since) {
				return false
			}
		case "until":
			if until, ok := value.(time.Time); ok && event.Timestamp.After(until) {
				return false
			}
		}
	}

	return true
}

// GetAuditSummary は監査ログのサマリーを取得する
func (am *AuditManager) GetAuditSummary(since time.Time) (map[string]interface{}, error) {
	if !am.enabled {
		return nil, fmt.Errorf("監査ログが無効です")
	}

	filter := map[string]interface{}{
		"since": since,
	}

	events, err := am.GetAuditLogs(0, filter)
	if err != nil {
		return nil, err
	}

	summary := map[string]interface{}{
		"total_events":    len(events),
		"success_count":   0,
		"failure_count":   0,
		"event_types":     make(map[string]int),
		"users":           make(map[string]int),
		"most_active_user": "",
		"period_start":    since,
		"period_end":      time.Now(),
	}

	userCounts := make(map[string]int)
	maxUserCount := 0

	for _, event := range events {
		if event.Success {
			summary["success_count"] = summary["success_count"].(int) + 1
		} else {
			summary["failure_count"] = summary["failure_count"].(int) + 1
		}

		// イベントタイプごとのカウント
		eventTypes := summary["event_types"].(map[string]int)
		eventTypes[event.Event]++

		// ユーザーごとのカウント
		userCounts[event.User]++
		if userCounts[event.User] > maxUserCount {
			maxUserCount = userCounts[event.User]
			summary["most_active_user"] = event.User
		}
	}

	summary["users"] = userCounts
	return summary, nil
}

// RotateAuditLog は監査ログをローテーションする
func (am *AuditManager) RotateAuditLog() error {
	if !am.enabled {
		return fmt.Errorf("監査ログが無効です")
	}

	// 現在のログファイルをバックアップ
	timestamp := time.Now().Format("20060102-150405")
	backupFile := fmt.Sprintf("%s.%s", am.logFile, timestamp)

	if err := os.Rename(am.logFile, backupFile); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("監査ログのバックアップに失敗: %w", err)
		}
	}

	// 新しいログファイルの開始ログ
	am.LogEvent("audit_log", am.logFile, "rotate", true, map[string]interface{}{
		"backup_file": backupFile,
	}, nil)

	return nil
}

// GetAuditStatus は監査機能の状況を返す
func (am *AuditManager) GetAuditStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":      am.enabled,
		"log_file":     am.logFile,
		"current_user": am.user,
		"file_exists":  false,
		"file_size":    int64(0),
	}

	// ログファイルの状況確認
	if info, err := os.Stat(am.logFile); err == nil {
		status["file_exists"] = true
		status["file_size"] = info.Size()
		status["last_modified"] = info.ModTime()
	}

	return status
}