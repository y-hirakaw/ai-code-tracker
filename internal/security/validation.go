package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ValidationManager は入力検証を管理する
type ValidationManager struct {
	strictMode bool
	maxPathLength int
	maxFileNameLength int
	allowedExtensions []string
	blockedPatterns []*regexp.Regexp
}

// NewValidationManager は新しい検証マネージャーを作成する
func NewValidationManager() *ValidationManager {
	strictMode := strings.ToLower(os.Getenv("AICT_SECURITY_MODE")) == "strict"
	
	// デフォルトの危険パターン
	dangerousPatterns := []string{
		`\.\.`,           // パストラバーサル
		`/\./`,           // カレントディレクトリ参照
		`\\`,             // バックスラッシュ（Windows）
		`[<>:"|?*]`,      // ファイル名に使用できない文字
		`^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])$`, // Windows予約名
		`\x00`,           // NULLバイト
	}
	
	// ストリクトモードでのみ絶対パスを禁止
	if strictMode {
		dangerousPatterns = append(dangerousPatterns, `^/`) // 絶対パス
	}
	
	var compiledPatterns []*regexp.Regexp
	for _, pattern := range dangerousPatterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		}
	}
	
	return &ValidationManager{
		strictMode: strictMode,
		maxPathLength: 4096,
		maxFileNameLength: 255,
		allowedExtensions: []string{
			".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".hpp",
			".rs", ".rb", ".php", ".css", ".html", ".htm", ".xml", ".json",
			".yaml", ".yml", ".toml", ".md", ".txt", ".sh", ".bash",
		},
		blockedPatterns: compiledPatterns,
	}
}

// ValidateFilePath はファイルパスを検証する
func (vm *ValidationManager) ValidateFilePath(path string) error {
	if path == "" {
		return errors.New("ファイルパスが空です")
	}

	// 長さチェック
	if len(path) > vm.maxPathLength {
		return fmt.Errorf("ファイルパスが長すぎます（最大%d文字）", vm.maxPathLength)
	}

	// UTF-8検証
	if !utf8.ValidString(path) {
		return errors.New("ファイルパスに無効なUTF-8文字が含まれています")
	}

	// 正規化
	cleanPath := filepath.Clean(path)
	
	// パストラバーサルチェック
	if strings.Contains(cleanPath, "..") {
		return errors.New("パストラバーサル攻撃を検出しました")
	}

	// 絶対パスチェック
	if filepath.IsAbs(cleanPath) && vm.strictMode {
		return errors.New("絶対パスは許可されていません")
	}

	// 危険パターンチェック
	for _, pattern := range vm.blockedPatterns {
		if pattern.MatchString(path) {
			return fmt.Errorf("危険なパターンを検出しました: %s", pattern.String())
		}
	}

	// ファイル名検証
	fileName := filepath.Base(cleanPath)
	if err := vm.ValidateFileName(fileName); err != nil {
		return fmt.Errorf("ファイル名検証エラー: %w", err)
	}

	// 拡張子チェック（strictモードの場合）
	if vm.strictMode {
		if err := vm.ValidateFileExtension(cleanPath); err != nil {
			return err
		}
	}

	return nil
}

// ValidateFileName はファイル名を検証する
func (vm *ValidationManager) ValidateFileName(fileName string) error {
	if fileName == "" {
		return errors.New("ファイル名が空です")
	}

	// 長さチェック
	if len(fileName) > vm.maxFileNameLength {
		return fmt.Errorf("ファイル名が長すぎます（最大%d文字）", vm.maxFileNameLength)
	}

	// 隠しファイルチェック（先頭が.で始まる）
	if strings.HasPrefix(fileName, ".") && vm.strictMode {
		return errors.New("隠しファイルは許可されていません")
	}

	// 制御文字チェック
	for _, r := range fileName {
		if r < 32 || r == 127 {
			return errors.New("ファイル名に制御文字が含まれています")
		}
	}

	// Windows予約名チェック
	baseName := strings.ToUpper(fileName)
	if idx := strings.Index(baseName, "."); idx > 0 {
		baseName = baseName[:idx]
	}
	
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}
	
	for _, reserved := range reservedNames {
		if baseName == reserved {
			return fmt.Errorf("予約されたファイル名です: %s", reserved)
		}
	}

	return nil
}

// ValidateFileExtension はファイル拡張子を検証する
func (vm *ValidationManager) ValidateFileExtension(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	if ext == "" {
		if vm.strictMode {
			return errors.New("ファイル拡張子が必要です")
		}
		return nil
	}

	// 許可された拡張子かチェック
	for _, allowed := range vm.allowedExtensions {
		if ext == allowed {
			return nil
		}
	}

	if vm.strictMode {
		return fmt.Errorf("許可されていない拡張子です: %s", ext)
	}

	return nil
}

// ValidateEventData はイベントデータを検証する
func (vm *ValidationManager) ValidateEventData(data map[string]interface{}) error {
	// 必須フィールドチェック
	requiredFields := []string{"id", "timestamp", "event_type", "author"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("必須フィールドが不足しています: %s", field)
		}
	}

	// IDの検証
	if id, ok := data["id"].(string); ok {
		if err := vm.ValidateID(id); err != nil {
			return fmt.Errorf("ID検証エラー: %w", err)
		}
	}

	// 作成者名の検証
	if author, ok := data["author"].(string); ok {
		if err := vm.ValidateAuthorName(author); err != nil {
			return fmt.Errorf("作成者名検証エラー: %w", err)
		}
	}

	// メッセージの検証
	if message, ok := data["message"].(string); ok {
		if err := vm.ValidateMessage(message); err != nil {
			return fmt.Errorf("メッセージ検証エラー: %w", err)
		}
	}

	return nil
}

// ValidateID はIDを検証する
func (vm *ValidationManager) ValidateID(id string) error {
	if id == "" {
		return errors.New("IDが空です")
	}

	// 長さチェック
	if len(id) > 128 {
		return errors.New("IDが長すぎます（最大128文字）")
	}

	// 使用可能文字チェック（英数字、ハイフン、アンダースコア）
	validID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validID.MatchString(id) {
		return errors.New("IDに無効な文字が含まれています")
	}

	return nil
}

// ValidateAuthorName は作成者名を検証する
func (vm *ValidationManager) ValidateAuthorName(author string) error {
	if author == "" {
		return errors.New("作成者名が空です")
	}

	// 長さチェック
	if len(author) > 256 {
		return errors.New("作成者名が長すぎます（最大256文字）")
	}

	// UTF-8検証
	if !utf8.ValidString(author) {
		return errors.New("作成者名に無効なUTF-8文字が含まれています")
	}

	// 危険な文字チェック
	dangerousChars := []string{"<", ">", "\"", "'", "&", "\x00"}
	for _, char := range dangerousChars {
		if strings.Contains(author, char) {
			return fmt.Errorf("作成者名に危険な文字が含まれています: %s", char)
		}
	}

	return nil
}

// ValidateMessage はメッセージを検証する
func (vm *ValidationManager) ValidateMessage(message string) error {
	// 長さチェック
	if len(message) > 1024 {
		return errors.New("メッセージが長すぎます（最大1024文字）")
	}

	// UTF-8検証
	if !utf8.ValidString(message) {
		return errors.New("メッセージに無効なUTF-8文字が含まれています")
	}

	// NULLバイトチェック
	if strings.Contains(message, "\x00") {
		return errors.New("メッセージにNULLバイトが含まれています")
	}

	return nil
}

// SanitizeFilePath はファイルパスをサニタイズする
func (vm *ValidationManager) SanitizeFilePath(path string) string {
	// 危険な文字を除去
	cleaned := strings.ReplaceAll(path, "..", "")
	cleaned = strings.ReplaceAll(cleaned, "\x00", "")
	
	// 先頭の/を除去（相対パスに変換）
	cleaned = strings.TrimPrefix(cleaned, "/")
	
	// 正規化
	cleaned = filepath.Clean(cleaned)
	
	return cleaned
}

// SanitizeString は文字列をサニタイズする
func (vm *ValidationManager) SanitizeString(input string) string {
	// NULLバイトを除去
	sanitized := strings.ReplaceAll(input, "\x00", "")
	
	// 制御文字を除去
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 && r != 127 {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// ValidateBatchOperation はバッチ操作を検証する
func (vm *ValidationManager) ValidateBatchOperation(files []string, maxFiles int) error {
	if len(files) == 0 {
		return errors.New("ファイルリストが空です")
	}

	if len(files) > maxFiles {
		return fmt.Errorf("ファイル数が多すぎます（最大%d個）", maxFiles)
	}

	// 各ファイルパスを検証
	for i, file := range files {
		if err := vm.ValidateFilePath(file); err != nil {
			return fmt.Errorf("ファイル%d検証エラー (%s): %w", i+1, file, err)
		}
	}

	// 重複チェック
	seen := make(map[string]bool)
	for _, file := range files {
		normalized := filepath.Clean(file)
		if seen[normalized] {
			return fmt.Errorf("重複したファイルパス: %s", file)
		}
		seen[normalized] = true
	}

	return nil
}

// IsSecureMode はセキュアモードかどうかを返す
func (vm *ValidationManager) IsSecureMode() bool {
	return vm.strictMode
}

// GetValidationRules は検証ルールを返す
func (vm *ValidationManager) GetValidationRules() map[string]interface{} {
	return map[string]interface{}{
		"strict_mode":          vm.strictMode,
		"max_path_length":      vm.maxPathLength,
		"max_filename_length":  vm.maxFileNameLength,
		"allowed_extensions":   vm.allowedExtensions,
		"blocked_patterns":     len(vm.blockedPatterns),
		"path_traversal_check": true,
		"null_byte_check":      true,
		"control_char_check":   true,
		"reserved_name_check":  true,
	}
}