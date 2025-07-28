package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/pkg/types"
)

// SecurityManager は全体のセキュリティ機能を統合管理する
type SecurityManager struct {
	dataDir           string
	config            *SecurityConfig
	encryptionManager *EncryptionManager
	auditManager      *AuditManager
	validationManager *ValidationManager
	privacyManager    *PrivacyManager
	exclusionManager  *ExclusionManager
}

// SecurityConfig はセキュリティ設定を表す
type SecurityConfig struct {
	Mode                   SecurityMode          `json:"mode"`
	FilePermissions        string                `json:"file_permissions"`
	DirectoryPermissions   string                `json:"directory_permissions"`
	EnableAuditLog         bool                  `json:"enable_audit_log"`
	EncryptSensitiveData   bool                  `json:"encrypt_sensitive_data"`
	ValidateFilePaths      bool                  `json:"validate_file_paths"`
	RestrictFileAccess     bool                  `json:"restrict_file_access"`
	AutoBackup             bool                  `json:"auto_backup"`
	MaxFileSize            int64                 `json:"max_file_size"`
	MaxEventsPerDay        int                   `json:"max_events_per_day"`
	RequireIntegrity       bool                  `json:"require_integrity"`
	Privacy                PrivacyConfig         `json:"privacy"`
	Exclusions             ExclusionConfig       `json:"exclusions"`
	Alerts                 AlertConfig           `json:"alerts"`
}

// SecurityMode はセキュリティモードを表す
type SecurityMode int

const (
	SecurityModeBasic SecurityMode = iota
	SecurityModeStandard
	SecurityModeStrict
	SecurityModeMaximum
)

// PrivacyConfig はプライバシー設定を表す
type PrivacyConfig struct {
	AnonymizeAuthors    bool `json:"anonymize_authors"`
	HashFilePaths       bool `json:"hash_file_paths"`
	DataRetentionDays   int  `json:"data_retention_days"`
	AutoCleanup         bool `json:"auto_cleanup"`
	RemoveTimestamps    bool `json:"remove_timestamps"`
}

// ExclusionConfig は除外設定を表す
type ExclusionConfig struct {
	Enabled             bool     `json:"enabled"`
	CustomRulesFile     string   `json:"custom_rules_file"`
	ExcludeSensitive    bool     `json:"exclude_sensitive"`
	ExcludeLargeFiles   bool     `json:"exclude_large_files"`
	ExcludedExtensions  []string `json:"excluded_extensions"`
}

// AlertConfig はアラート設定を表す
type AlertConfig struct {
	Enabled                bool     `json:"enabled"`
	SecurityEventThreshold int      `json:"security_event_threshold"`
	FailureThreshold       int      `json:"failure_threshold"`
	NotificationMethods    []string `json:"notification_methods"`
}

// NewSecurityManager は新しいセキュリティマネージャーを作成する
func NewSecurityManager(dataDir string) (*SecurityManager, error) {
	sm := &SecurityManager{
		dataDir: dataDir,
	}

	// 設定をロード
	if err := sm.loadConfig(); err != nil {
		return nil, fmt.Errorf("設定のロードに失敗: %w", err)
	}

	// 各マネージャーを初期化
	sm.encryptionManager = NewEncryptionManager(dataDir)
	sm.auditManager = NewAuditManager(dataDir)
	sm.validationManager = NewValidationManager()
	sm.privacyManager = NewPrivacyManager(dataDir)
	sm.exclusionManager = NewExclusionManager(dataDir)

	// セキュリティ機能を初期化
	if err := sm.initialize(); err != nil {
		return nil, fmt.Errorf("セキュリティ機能の初期化に失敗: %w", err)
	}

	return sm, nil
}

// loadConfig は設定ファイルからセキュリティ設定をロードする
func (sm *SecurityManager) loadConfig() error {
	configFile := filepath.Join(sm.dataDir, "security-config.json")
	
	// デフォルト設定
	sm.config = sm.getDefaultConfig()

	// 設定ファイルが存在する場合は読み込み
	if data, err := os.ReadFile(configFile); err == nil {
		var config SecurityConfig
		if err := json.Unmarshal(data, &config); err == nil {
			sm.config = &config
		}
	}

	// 環境変数で設定を上書き
	sm.overrideFromEnv()

	return nil
}

// getDefaultConfig はデフォルト設定を返す
func (sm *SecurityManager) getDefaultConfig() *SecurityConfig {
	return &SecurityConfig{
		Mode:                   SecurityModeStandard,
		FilePermissions:        "600",
		DirectoryPermissions:   "700",
		EnableAuditLog:         false,
		EncryptSensitiveData:   false,
		ValidateFilePaths:      true,
		RestrictFileAccess:     true,
		AutoBackup:             false,
		MaxFileSize:            10 * 1024 * 1024, // 10MB
		MaxEventsPerDay:        10000,
		RequireIntegrity:       false,
		Privacy: PrivacyConfig{
			AnonymizeAuthors:  false,
			HashFilePaths:     false,
			DataRetentionDays: 365,
			AutoCleanup:       false,
			RemoveTimestamps:  false,
		},
		Exclusions: ExclusionConfig{
			Enabled:            true,
			ExcludeSensitive:   true,
			ExcludeLargeFiles:  true,
			ExcludedExtensions: []string{".tmp", ".log", ".bak"},
		},
		Alerts: AlertConfig{
			Enabled:                false,
			SecurityEventThreshold: 10,
			FailureThreshold:       5,
			NotificationMethods:    []string{"log"},
		},
	}
}

// overrideFromEnv は環境変数から設定を上書きする
func (sm *SecurityManager) overrideFromEnv() {
	// セキュリティモード
	if mode := os.Getenv("AICT_SECURITY_MODE"); mode != "" {
		switch strings.ToLower(mode) {
		case "basic":
			sm.config.Mode = SecurityModeBasic
		case "standard":
			sm.config.Mode = SecurityModeStandard
		case "strict":
			sm.config.Mode = SecurityModeStrict
		case "maximum":
			sm.config.Mode = SecurityModeMaximum
		}
	}

	// 監査ログ
	if os.Getenv("AICT_AUDIT_LOG") == "true" {
		sm.config.EnableAuditLog = true
	}

	// データ暗号化
	if os.Getenv("AICT_ENCRYPT_DATA") == "true" {
		sm.config.EncryptSensitiveData = true
	}

	// その他の設定...
}

// initialize はセキュリティ機能を初期化する
func (sm *SecurityManager) initialize() error {
	// ディレクトリ権限の設定
	if err := sm.setupDirectoryPermissions(); err != nil {
		return fmt.Errorf("ディレクトリ権限の設定に失敗: %w", err)
	}

	// 暗号化の初期化
	if sm.config.EncryptSensitiveData {
		if err := sm.encryptionManager.InitializeEncryption(); err != nil {
			return fmt.Errorf("暗号化の初期化に失敗: %w", err)
		}
	}

	// 監査ログの初期化
	if sm.config.EnableAuditLog {
		sm.auditManager.LogEvent("security_init", "system", "initialize", true, 
			map[string]interface{}{
				"security_mode": sm.config.Mode,
				"timestamp":     time.Now(),
			}, nil)
	}

	return nil
}

// setupDirectoryPermissions はディレクトリ権限を設定する
func (sm *SecurityManager) setupDirectoryPermissions() error {
	// データディレクトリの権限設定
	if err := os.Chmod(sm.dataDir, 0700); err != nil {
		return fmt.Errorf("データディレクトリの権限設定に失敗: %w", err)
	}

	return nil
}

// ProcessTrackEvent はトラッキングイベントをセキュリティ処理する
func (sm *SecurityManager) ProcessTrackEvent(event *types.TrackEvent) (*types.TrackEvent, error) {
	// 入力検証
	eventData := map[string]interface{}{
		"id":         event.ID,
		"timestamp":  event.Timestamp,
		"event_type": event.EventType,
		"author":     event.Author,
		"message":    event.Message,
	}

	if err := sm.validationManager.ValidateEventData(eventData); err != nil {
		sm.logSecurityEvent("validation_failed", fmt.Sprintf("イベント検証失敗: %v", err), "medium")
		return nil, fmt.Errorf("イベント検証エラー: %w", err)
	}

	// ファイルパスの検証
	for _, fileInfo := range event.Files {
		if err := sm.validationManager.ValidateFilePath(fileInfo.Path); err != nil {
			sm.logSecurityEvent("path_validation_failed", 
				fmt.Sprintf("ファイルパス検証失敗: %s", fileInfo.Path), "high")
			return nil, fmt.Errorf("ファイルパス検証エラー: %w", err)
		}

		// 除外チェック
		if excluded, reason := sm.exclusionManager.ShouldExclude(fileInfo.Path); excluded {
			sm.logSecurityEvent("file_excluded", 
				fmt.Sprintf("ファイルが除外されました: %s (理由: %s)", fileInfo.Path, reason), "low")
			return nil, fmt.Errorf("ファイルが除外されています: %s", reason)
		}
	}

	// プライバシー処理
	processedEvent := sm.privacyManager.ProcessTrackEvent(event)

	// 監査ログ
	if sm.config.EnableAuditLog {
		sm.auditManager.LogDataOperation("track_event", "events", 1, true, nil)
	}

	return processedEvent, nil
}

// EncryptData はデータを暗号化する
func (sm *SecurityManager) EncryptData(data []byte) ([]byte, error) {
	if !sm.config.EncryptSensitiveData {
		return data, nil
	}

	encrypted, err := sm.encryptionManager.EncryptData(data)
	if err != nil {
		sm.logSecurityEvent("encryption_failed", "データ暗号化に失敗", "high")
		return nil, err
	}

	if sm.config.EnableAuditLog {
		sm.auditManager.LogEvent("data_encryption", "data", "encrypt", true, 
			map[string]interface{}{
				"data_size": len(data),
			}, nil)
	}

	return encrypted, nil
}

// DecryptData はデータを復号化する
func (sm *SecurityManager) DecryptData(encryptedData []byte) ([]byte, error) {
	if !sm.config.EncryptSensitiveData {
		return encryptedData, nil
	}

	decrypted, err := sm.encryptionManager.DecryptData(encryptedData)
	if err != nil {
		sm.logSecurityEvent("decryption_failed", "データ復号化に失敗", "high")
		return nil, err
	}

	if sm.config.EnableAuditLog {
		sm.auditManager.LogEvent("data_decryption", "data", "decrypt", true, 
			map[string]interface{}{
				"encrypted_size": len(encryptedData),
				"decrypted_size": len(decrypted),
			}, nil)
	}

	return decrypted, nil
}

// logSecurityEvent はセキュリティイベントをログに記録する
func (sm *SecurityManager) logSecurityEvent(eventType, description, severity string) {
	if sm.config.EnableAuditLog {
		sm.auditManager.LogSecurityEvent(eventType, description, severity, nil)
	}

	// アラート処理
	if sm.config.Alerts.Enabled {
		sm.processSecurityAlert(eventType, description, severity)
	}
}

// processSecurityAlert はセキュリティアラートを処理する
func (sm *SecurityManager) processSecurityAlert(eventType, description, severity string) {
	// アラート条件をチェック
	if severity == "high" || severity == "critical" {
		// 重要なセキュリティイベントの場合は即座にアラート
		sm.sendAlert(fmt.Sprintf("セキュリティアラート: %s - %s", eventType, description))
	}
}

// sendAlert はアラートを送信する
func (sm *SecurityManager) sendAlert(message string) {
	// ログに記録
	fmt.Fprintf(os.Stderr, "🚨 SECURITY ALERT: %s\n", message)
	
	// 設定に基づいて通知方法を決定
	for _, method := range sm.config.Alerts.NotificationMethods {
		switch method {
		case "log":
			// 既にログに記録済み
		case "stderr":
			fmt.Fprintf(os.Stderr, "ALERT: %s\n", message)
		// 他の通知方法（メール、Slack等）は将来実装
		}
	}
}

// GetSecurityStatus はセキュリティ状況を取得する
func (sm *SecurityManager) GetSecurityStatus() map[string]interface{} {
	status := map[string]interface{}{
		"mode":           sm.config.Mode,
		"encryption":     sm.encryptionManager.GetEncryptionStatus(),
		"audit":          sm.auditManager.GetAuditStatus(),
		"privacy":        sm.privacyManager.GeneratePrivacyReport(),
		"exclusions":     sm.exclusionManager.GenerateExclusionReport(),
		"validation":     sm.validationManager.GetValidationRules(),
		"config_file":    filepath.Join(sm.dataDir, "security-config.json"),
		"last_updated":   time.Now(),
	}

	return status
}

// UpdateSecurityConfig はセキュリティ設定を更新する
func (sm *SecurityManager) UpdateSecurityConfig(updates map[string]interface{}) error {
	// 設定の検証
	if err := sm.validateConfigUpdates(updates); err != nil {
		return fmt.Errorf("設定検証エラー: %w", err)
	}

	// 設定を適用
	oldConfig := *sm.config
	if err := sm.applyConfigUpdates(updates); err != nil {
		// 失敗した場合は元に戻す
		sm.config = &oldConfig
		return fmt.Errorf("設定適用エラー: %w", err)
	}

	// 設定ファイルに保存
	if err := sm.saveConfig(); err != nil {
		return fmt.Errorf("設定保存エラー: %w", err)
	}

	// 監査ログ
	if sm.config.EnableAuditLog {
		sm.auditManager.LogConfigChange("security_config", 
			fmt.Sprintf("%+v", oldConfig), fmt.Sprintf("%+v", sm.config), true, nil)
	}

	return nil
}

// validateConfigUpdates は設定更新を検証する
func (sm *SecurityManager) validateConfigUpdates(updates map[string]interface{}) error {
	// 基本検証ロジック
	for key, value := range updates {
		switch key {
		case "mode":
			if mode, ok := value.(string); ok {
				validModes := []string{"basic", "standard", "strict", "maximum"}
				valid := false
				for _, validMode := range validModes {
					if mode == validMode {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("無効なセキュリティモード: %s", mode)
				}
			}
		case "max_file_size":
			if size, ok := value.(int64); ok && size <= 0 {
				return fmt.Errorf("ファイルサイズ制限は正の値である必要があります")
			}
		}
	}

	return nil
}

// applyConfigUpdates は設定更新を適用する
func (sm *SecurityManager) applyConfigUpdates(updates map[string]interface{}) error {
	// 設定の適用ロジック
	for key, value := range updates {
		switch key {
		case "enable_audit_log":
			if enable, ok := value.(bool); ok {
				sm.config.EnableAuditLog = enable
			}
		case "encrypt_sensitive_data":
			if encrypt, ok := value.(bool); ok {
				sm.config.EncryptSensitiveData = encrypt
			}
		// 他の設定項目...
		}
	}

	return nil
}

// saveConfig は設定をファイルに保存する
func (sm *SecurityManager) saveConfig() error {
	configFile := filepath.Join(sm.dataDir, "security-config.json")
	
	data, err := json.MarshalIndent(sm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("設定のシリアライズに失敗: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("設定ファイルの書き込みに失敗: %w", err)
	}

	return nil
}

// PerformSecurityScan はセキュリティスキャンを実行する
func (sm *SecurityManager) PerformSecurityScan() (map[string]interface{}, error) {
	results := map[string]interface{}{
		"timestamp": time.Now(),
		"summary":   map[string]interface{}{},
		"issues":    []map[string]interface{}{},
		"score":     0,
	}

	summary := results["summary"].(map[string]interface{})
	var issues []map[string]interface{}

	// ファイル権限チェック
	if permIssues := sm.checkFilePermissions(); len(permIssues) > 0 {
		issues = append(issues, permIssues...)
		summary["file_permission_issues"] = len(permIssues)
	}

	// 設定チェック
	if configIssues := sm.checkSecurityConfig(); len(configIssues) > 0 {
		issues = append(issues, configIssues...)
		summary["config_issues"] = len(configIssues)
	}

	// スコア計算
	totalIssues := len(issues)
	maxScore := 100
	score := maxScore - (totalIssues * 10)
	if score < 0 {
		score = 0
	}

	results["issues"] = issues
	results["score"] = score
	summary["total_issues"] = totalIssues

	// 監査ログ
	if sm.config.EnableAuditLog {
		sm.auditManager.LogEvent("security_scan", "system", "scan", true,
			map[string]interface{}{
				"issues_found": totalIssues,
				"score":        score,
			}, nil)
	}

	return results, nil
}

// checkFilePermissions はファイル権限をチェックする
func (sm *SecurityManager) checkFilePermissions() []map[string]interface{} {
	var issues []map[string]interface{}

	// データディレクトリの権限チェック
	if info, err := os.Stat(sm.dataDir); err == nil {
		mode := info.Mode().Perm()
		if mode != 0700 {
			issues = append(issues, map[string]interface{}{
				"type":        "file_permission",
				"severity":    "medium",
				"description": fmt.Sprintf("データディレクトリの権限が緩すぎます: %o", mode),
				"path":        sm.dataDir,
				"expected":    "700",
				"actual":      fmt.Sprintf("%o", mode),
			})
		}
	}

	return issues
}

// checkSecurityConfig はセキュリティ設定をチェックする
func (sm *SecurityManager) checkSecurityConfig() []map[string]interface{} {
	var issues []map[string]interface{}

	// 基本的な設定チェック
	if !sm.config.EnableAuditLog {
		issues = append(issues, map[string]interface{}{
			"type":        "config",
			"severity":    "low",
			"description": "監査ログが無効になっています",
			"recommendation": "セキュリティ強化のため監査ログを有効にすることを推奨します",
		})
	}

	if !sm.config.EncryptSensitiveData {
		issues = append(issues, map[string]interface{}{
			"type":        "config",
			"severity":    "medium",
			"description": "データ暗号化が無効になっています",
			"recommendation": "機密データ保護のため暗号化を有効にすることを推奨します",
		})
	}

	return issues
}

// GetSecurityManager はセキュリティマネージャーの各コンポーネントへのアクセスを提供する
func (sm *SecurityManager) GetEncryptionManager() *EncryptionManager {
	return sm.encryptionManager
}

func (sm *SecurityManager) GetAuditManager() *AuditManager {
	return sm.auditManager
}

func (sm *SecurityManager) GetValidationManager() *ValidationManager {
	return sm.validationManager
}

func (sm *SecurityManager) GetPrivacyManager() *PrivacyManager {
	return sm.privacyManager
}

func (sm *SecurityManager) GetExclusionManager() *ExclusionManager {
	return sm.exclusionManager
}