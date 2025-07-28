package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionManager はデータ暗号化を管理する
type EncryptionManager struct {
	enabled    bool
	passphrase string
	keyFile    string
}

// NewEncryptionManager は新しい暗号化マネージャーを作成する
func NewEncryptionManager(dataDir string) *EncryptionManager {
	enabled := os.Getenv("AICT_ENCRYPT_DATA") == "true"
	keyFile := filepath.Join(dataDir, ".encryption_key")
	
	return &EncryptionManager{
		enabled: enabled,
		keyFile: keyFile,
	}
}

// IsEnabled は暗号化が有効かどうかを返す
func (em *EncryptionManager) IsEnabled() bool {
	return em.enabled
}

// InitializeEncryption は暗号化を初期化する
func (em *EncryptionManager) InitializeEncryption() error {
	if !em.enabled {
		return nil
	}

	// パスフレーズの取得または生成
	if err := em.setupPassphrase(); err != nil {
		return fmt.Errorf("パスフレーズの設定に失敗: %w", err)
	}

	return nil
}

// setupPassphrase はパスフレーズを設定する
func (em *EncryptionManager) setupPassphrase() error {
	// 環境変数からパスフレーズを取得
	if passphrase := os.Getenv("AICT_ENCRYPTION_PASSPHRASE"); passphrase != "" {
		em.passphrase = passphrase
		return nil
	}

	// キーファイルから読み込み
	if data, err := os.ReadFile(em.keyFile); err == nil {
		em.passphrase = string(data)
		return nil
	}

	// 新しいパスフレーズを生成
	passphrase, err := em.generatePassphrase()
	if err != nil {
		return fmt.Errorf("パスフレーズの生成に失敗: %w", err)
	}

	em.passphrase = passphrase

	// キーファイルに保存
	if err := os.WriteFile(em.keyFile, []byte(passphrase), 0600); err != nil {
		return fmt.Errorf("キーファイルの保存に失敗: %w", err)
	}

	return nil
}

// generatePassphrase はランダムなパスフレーズを生成する
func (em *EncryptionManager) generatePassphrase() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// deriveKey はパスフレーズから暗号化キーを導出する
func (em *EncryptionManager) deriveKey(salt []byte) []byte {
	return pbkdf2.Key([]byte(em.passphrase), salt, 10000, 32, sha256.New)
}

// EncryptData はデータを暗号化する
func (em *EncryptionManager) EncryptData(data []byte) ([]byte, error) {
	if !em.enabled {
		return data, nil
	}

	if em.passphrase == "" {
		return nil, errors.New("暗号化が有効ですが、パスフレーズが設定されていません")
	}

	// ソルトを生成
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("ソルトの生成に失敗: %w", err)
	}

	// キーを導出
	key := em.deriveKey(salt)

	// AES暗号化ブロックを作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES暗号化ブロックの作成に失敗: %w", err)
	}

	// GCMモードを使用
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCMモードの作成に失敗: %w", err)
	}

	// ナンスを生成
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("ナンスの生成に失敗: %w", err)
	}

	// データを暗号化
	encrypted := gcm.Seal(nonce, nonce, data, nil)

	// ソルト + 暗号化データを結合
	result := make([]byte, len(salt)+len(encrypted))
	copy(result, salt)
	copy(result[len(salt):], encrypted)

	return result, nil
}

// DecryptData はデータを復号化する
func (em *EncryptionManager) DecryptData(encryptedData []byte) ([]byte, error) {
	if !em.enabled {
		return encryptedData, nil
	}

	if em.passphrase == "" {
		return nil, errors.New("暗号化が有効ですが、パスフレーズが設定されていません")
	}

	if len(encryptedData) < 16 {
		return nil, errors.New("暗号化データが短すぎます")
	}

	// ソルトを抽出
	salt := encryptedData[:16]
	encrypted := encryptedData[16:]

	// キーを導出
	key := em.deriveKey(salt)

	// AES復号化ブロックを作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("AES復号化ブロックの作成に失敗: %w", err)
	}

	// GCMモードを使用
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("GCMモードの作成に失敗: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, errors.New("暗号化データが短すぎます")
	}

	// ナンスとデータを分離
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	// データを復号化
	data, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("データの復号化に失敗: %w", err)
	}

	return data, nil
}

// EncryptFile はファイルを暗号化する
func (em *EncryptionManager) EncryptFile(filePath string) error {
	if !em.enabled {
		return nil
	}

	// ファイルを読み込み
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("ファイルの読み込みに失敗: %w", err)
	}

	// データを暗号化
	encryptedData, err := em.EncryptData(data)
	if err != nil {
		return fmt.Errorf("データの暗号化に失敗: %w", err)
	}

	// 暗号化されたデータを書き込み
	if err := os.WriteFile(filePath, encryptedData, 0600); err != nil {
		return fmt.Errorf("暗号化ファイルの書き込みに失敗: %w", err)
	}

	return nil
}

// DecryptFile はファイルを復号化する
func (em *EncryptionManager) DecryptFile(filePath string) ([]byte, error) {
	if !em.enabled {
		// 暗号化が無効の場合は通常のファイル読み込み
		return os.ReadFile(filePath)
	}

	// 暗号化されたファイルを読み込み
	encryptedData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("暗号化ファイルの読み込みに失敗: %w", err)
	}

	// データを復号化
	data, err := em.DecryptData(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("データの復号化に失敗: %w", err)
	}

	return data, nil
}

// IsFileEncrypted はファイルが暗号化されているかチェックする
func (em *EncryptionManager) IsFileEncrypted(filePath string) (bool, error) {
	if !em.enabled {
		return false, nil
	}

	// ファイルサイズをチェック
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	if info.Size() < 16 {
		return false, nil
	}

	// ファイルの先頭を読んでソルトの存在をチェック
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	header := make([]byte, 32)
	n, err := file.Read(header)
	if err != nil {
		return false, err
	}

	// 暗号化されたファイルかどうかの簡易判定
	// （実際の判定はより厳密に行う）
	return n >= 32, nil
}

// GetEncryptionStatus は暗号化の状況を返す
func (em *EncryptionManager) GetEncryptionStatus() map[string]interface{} {
	status := map[string]interface{}{
		"enabled":           em.enabled,
		"passphrase_set":    em.passphrase != "",
		"key_file_exists":   false,
		"encryption_method": "AES-256-GCM",
	}

	// キーファイルの存在確認
	if _, err := os.Stat(em.keyFile); err == nil {
		status["key_file_exists"] = true
	}

	return status
}

// RotateEncryptionKey は暗号化キーをローテーションする
func (em *EncryptionManager) RotateEncryptionKey() error {
	if !em.enabled {
		return errors.New("暗号化が無効です")
	}

	// 新しいパスフレーズを生成
	newPassphrase, err := em.generatePassphrase()
	if err != nil {
		return fmt.Errorf("新しいパスフレーズの生成に失敗: %w", err)
	}

	// 古いパスフレーズをバックアップ
	oldPassphrase := em.passphrase

	// 新しいパスフレーズを設定
	em.passphrase = newPassphrase

	// キーファイルを更新
	if err := os.WriteFile(em.keyFile, []byte(newPassphrase), 0600); err != nil {
		// 失敗した場合は元に戻す
		em.passphrase = oldPassphrase
		return fmt.Errorf("新しいキーファイルの保存に失敗: %w", err)
	}

	return nil
}