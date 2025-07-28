package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ai-code-tracker/aict/pkg/types"
)

// PrivacyManager はプライバシー設定を管理する
type PrivacyManager struct {
	anonymizeAuthors    bool
	hashFilePaths       bool
	removeTimestamps    bool
	dataRetentionDays   int
	autoCleanup         bool
	sensitivePatterns   []string
	anonymizationSalt   string
}

// NewPrivacyManager は新しいプライバシーマネージャーを作成する
func NewPrivacyManager(dataDir string) *PrivacyManager {
	// 環境変数から設定を読み込み
	anonymizeAuthors := strings.ToLower(os.Getenv("AICT_ANONYMIZE_AUTHORS")) == "true"
	hashFilePaths := strings.ToLower(os.Getenv("AICT_HASH_FILE_PATHS")) == "true"
	removeTimestamps := strings.ToLower(os.Getenv("AICT_REMOVE_TIMESTAMPS")) == "true"
	autoCleanup := strings.ToLower(os.Getenv("AICT_AUTO_CLEANUP")) == "true"

	// データ保持期間のデフォルトは365日
	retentionDays := 365
	if envDays := os.Getenv("AICT_DATA_RETENTION_DAYS"); envDays != "" {
		if days, err := strconv.Atoi(envDays); err == nil && days > 0 {
			retentionDays = days
		}
	}

	// 機密情報のパターン
	sensitivePatterns := []string{
		"password", "passwd", "pwd",
		"secret", "key", "token",
		"api_key", "apikey", "access_key",
		"private_key", "privatekey",
		"credential", "auth",
		"oauth", "bearer",
	}

	// 匿名化用のソルト
	salt := os.Getenv("AICT_ANONYMIZATION_SALT")
	if salt == "" {
		salt = "default-aict-salt-2025" // デフォルトソルト
	}

	return &PrivacyManager{
		anonymizeAuthors:  anonymizeAuthors,
		hashFilePaths:     hashFilePaths,
		removeTimestamps:  removeTimestamps,
		dataRetentionDays: retentionDays,
		autoCleanup:       autoCleanup,
		sensitivePatterns: sensitivePatterns,
		anonymizationSalt: salt,
	}
}

// AnonymizeAuthor は作成者名を匿名化する
func (pm *PrivacyManager) AnonymizeAuthor(author string) string {
	if !pm.anonymizeAuthors {
		return author
	}

	// 既知のAI作成者は匿名化しない
	aiAuthors := []string{
		"Claude Code", "Claude", "claude",
		"GitHub Copilot", "Copilot",
		"AI Assistant", "Assistant",
	}

	for _, aiAuthor := range aiAuthors {
		if strings.Contains(strings.ToLower(author), strings.ToLower(aiAuthor)) {
			return "AI Assistant" // 統一表記
		}
	}

	// 人間の作成者を匿名化
	return pm.hashString(author, "author")
}

// AnonymizeFilePath はファイルパスを匿名化する
func (pm *PrivacyManager) AnonymizeFilePath(filePath string) string {
	if !pm.hashFilePaths {
		return filePath
	}

	// 拡張子は保持
	ext := filepath.Ext(filePath)
	dir := filepath.Dir(filePath)
	base := strings.TrimSuffix(filepath.Base(filePath), ext)

	// ディレクトリ部分のハッシュ化
	hashedDir := ""
	if dir != "." && dir != "/" {
		hashedDir = pm.hashString(dir, "dir") + "/"
	}

	// ファイル名のハッシュ化
	hashedBase := pm.hashString(base, "file")

	return hashedDir + hashedBase + ext
}

// ProcessTimestamp はタイムスタンプを処理する
func (pm *PrivacyManager) ProcessTimestamp(timestamp time.Time) *time.Time {
	if pm.removeTimestamps {
		return nil
	}
	return &timestamp
}

// SanitizeMessage は機密情報を含む可能性のあるメッセージをサニタイズする
func (pm *PrivacyManager) SanitizeMessage(message string) string {
	result := message

	// 機密情報パターンのマスク
	for _, pattern := range pm.sensitivePatterns {
		// 大文字小文字を無視してパターンをチェック
		lowerMessage := strings.ToLower(result)
		lowerPattern := strings.ToLower(pattern)

		if strings.Contains(lowerMessage, lowerPattern) {
			// パターンが見つかった場合、その周辺をマスク
			result = pm.maskSensitiveContent(result, pattern)
		}
	}

	return result
}

// maskSensitiveContent は機密コンテンツをマスクする
func (pm *PrivacyManager) maskSensitiveContent(text, pattern string) string {
	// パターンの前後の単語もマスクする
	words := strings.Fields(text)
	var maskedWords []string
	skipNext := false

	for i, word := range words {
		if skipNext {
			maskedWords = append(maskedWords, "[REDACTED]")
			skipNext = false
			continue
		}

		lowerWord := strings.ToLower(word)
		if strings.Contains(lowerWord, strings.ToLower(pattern)) {
			// パターンを含む単語をマスク
			maskedWords = append(maskedWords, "[REDACTED]")
			
			// 次の単語も値の可能性があるためマスク
			if i+1 < len(words) {
				skipNext = true
			}
		} else {
			maskedWords = append(maskedWords, word)
		}
	}

	return strings.Join(maskedWords, " ")
}

// ProcessTrackEvent はトラッキングイベントをプライバシー設定に従って処理する
func (pm *PrivacyManager) ProcessTrackEvent(event *types.TrackEvent) *types.TrackEvent {
	processed := *event // コピーを作成

	// 作成者の匿名化
	processed.Author = pm.AnonymizeAuthor(processed.Author)

	// メッセージのサニタイズ
	processed.Message = pm.SanitizeMessage(processed.Message)

	// タイムスタンプの処理
	if pm.removeTimestamps {
		processed.Timestamp = time.Time{} // ゼロ値に設定
	}

	// ファイル情報の処理
	for i, fileInfo := range processed.Files {
		processed.Files[i].Path = pm.AnonymizeFilePath(fileInfo.Path)
	}

	return &processed
}

// hashString は文字列をハッシュ化する
func (pm *PrivacyManager) hashString(input, context string) string {
	// コンテキストとソルトを含めてハッシュ化
	data := fmt.Sprintf("%s:%s:%s", context, pm.anonymizationSalt, input)
	hash := sha256.Sum256([]byte(data))
	
	// 8文字の短縮ハッシュを使用（可読性のため）
	return hex.EncodeToString(hash[:])[:8]
}

// ShouldRetainData はデータを保持すべきかどうかを判定する
func (pm *PrivacyManager) ShouldRetainData(timestamp time.Time) bool {
	if pm.dataRetentionDays <= 0 {
		return true // 無制限保持
	}

	retentionDuration := time.Duration(pm.dataRetentionDays) * 24 * time.Hour
	return time.Since(timestamp) <= retentionDuration
}

// GetExpiredDataThreshold は期限切れデータの閾値を返す
func (pm *PrivacyManager) GetExpiredDataThreshold() time.Time {
	if pm.dataRetentionDays <= 0 {
		return time.Time{} // 無制限の場合はゼロ値
	}

	retentionDuration := time.Duration(pm.dataRetentionDays) * 24 * time.Hour
	return time.Now().Add(-retentionDuration)
}

// IsAutoCleanupEnabled は自動クリーンアップが有効かどうかを返す
func (pm *PrivacyManager) IsAutoCleanupEnabled() bool {
	return pm.autoCleanup
}

// GetDataRetentionPolicy はデータ保持ポリシーを返す
func (pm *PrivacyManager) GetDataRetentionPolicy() map[string]interface{} {
	threshold := pm.GetExpiredDataThreshold()
	var thresholdStr string
	if !threshold.IsZero() {
		thresholdStr = threshold.Format("2006-01-02")
	} else {
		thresholdStr = "unlimited"
	}

	return map[string]interface{}{
		"retention_days":    pm.dataRetentionDays,
		"auto_cleanup":      pm.autoCleanup,
		"expire_threshold":  thresholdStr,
		"policy_active":     pm.dataRetentionDays > 0,
	}
}

// GeneratePrivacyReport はプライバシー設定のレポートを生成する
func (pm *PrivacyManager) GeneratePrivacyReport() map[string]interface{} {
	return map[string]interface{}{
		"anonymization": map[string]interface{}{
			"anonymize_authors":   pm.anonymizeAuthors,
			"hash_file_paths":     pm.hashFilePaths,
			"remove_timestamps":   pm.removeTimestamps,
			"mask_sensitive_data": len(pm.sensitivePatterns) > 0,
		},
		"data_retention": pm.GetDataRetentionPolicy(),
		"sensitive_patterns": map[string]interface{}{
			"enabled":       len(pm.sensitivePatterns) > 0,
			"pattern_count": len(pm.sensitivePatterns),
			"patterns":      pm.sensitivePatterns,
		},
		"security_features": map[string]interface{}{
			"salted_hashing":     true,
			"context_aware_hash": true,
			"ai_author_preserve": true,
		},
	}
}

// ValidatePrivacySettings はプライバシー設定を検証する
func (pm *PrivacyManager) ValidatePrivacySettings() error {
	// データ保持期間の検証
	if pm.dataRetentionDays < 0 {
		return fmt.Errorf("データ保持期間は0以上である必要があります: %d", pm.dataRetentionDays)
	}

	// 過度に短い保持期間の警告
	if pm.dataRetentionDays > 0 && pm.dataRetentionDays < 7 {
		return fmt.Errorf("データ保持期間が短すぎます（最低7日推奨）: %d日", pm.dataRetentionDays)
	}

	// ソルトの検証
	if pm.anonymizationSalt == "" {
		return fmt.Errorf("匿名化ソルトが設定されていません")
	}

	if len(pm.anonymizationSalt) < 8 {
		return fmt.Errorf("匿名化ソルトが短すぎます（最低8文字）")
	}

	return nil
}

// UpdatePrivacySettings はプライバシー設定を更新する
func (pm *PrivacyManager) UpdatePrivacySettings(settings map[string]interface{}) error {
	if anonymize, ok := settings["anonymize_authors"].(bool); ok {
		pm.anonymizeAuthors = anonymize
	}

	if hashPaths, ok := settings["hash_file_paths"].(bool); ok {
		pm.hashFilePaths = hashPaths
	}

	if removeTime, ok := settings["remove_timestamps"].(bool); ok {
		pm.removeTimestamps = removeTime
	}

	if retention, ok := settings["data_retention_days"].(int); ok {
		if retention < 0 {
			return fmt.Errorf("データ保持期間は0以上である必要があります")
		}
		pm.dataRetentionDays = retention
	}

	if cleanup, ok := settings["auto_cleanup"].(bool); ok {
		pm.autoCleanup = cleanup
	}

	// 設定の妥当性を検証
	return pm.ValidatePrivacySettings()
}

// GetPrivacyImpact はプライバシー設定の影響を分析する
func (pm *PrivacyManager) GetPrivacyImpact() map[string]interface{} {
	impact := map[string]interface{}{
		"data_protection_level": "medium",
		"reversibility":         map[string]bool{},
		"functionality_impact":  map[string]string{},
		"compliance":           map[string]bool{},
	}

	reversibility := impact["reversibility"].(map[string]bool)
	functionality := impact["functionality_impact"].(map[string]string)
	compliance := impact["compliance"].(map[string]bool)

	// 可逆性の分析
	reversibility["author_anonymization"] = false // ハッシュ化は不可逆
	reversibility["file_path_hashing"] = false    // ハッシュ化は不可逆
	reversibility["timestamp_removal"] = false    // 削除は不可逆
	reversibility["message_sanitization"] = false // マスクは不可逆

	// 機能への影響
	if pm.anonymizeAuthors {
		functionality["blame_analysis"] = "制限あり（匿名化されたデータでの分析）"
	}
	if pm.hashFilePaths {
		functionality["file_tracking"] = "制限あり（ハッシュ化されたパスでの追跡）"
	}
	if pm.removeTimestamps {
		functionality["temporal_analysis"] = "無効（時系列分析不可）"
	}

	// コンプライアンス
	compliance["gdpr_article_17"] = pm.dataRetentionDays > 0 // 削除する権利
	compliance["gdpr_article_25"] = pm.anonymizeAuthors     // プライバシー・バイ・デザイン
	compliance["data_minimization"] = len(pm.sensitivePatterns) > 0

	// 保護レベルの決定
	protectionScore := 0
	if pm.anonymizeAuthors {
		protectionScore++
	}
	if pm.hashFilePaths {
		protectionScore++
	}
	if pm.removeTimestamps {
		protectionScore++
	}
	if pm.dataRetentionDays > 0 && pm.dataRetentionDays <= 365 {
		protectionScore++
	}

	switch protectionScore {
	case 4:
		impact["data_protection_level"] = "high"
	case 2, 3:
		impact["data_protection_level"] = "medium"
	default:
		impact["data_protection_level"] = "low"
	}

	return impact
}