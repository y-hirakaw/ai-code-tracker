package security

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewEncryptionManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewEncryptionManager(tempDir)
	if em == nil {
		t.Fatal("EncryptionManagerがnilです")
	}

	if em.keyFile != filepath.Join(tempDir, ".encryption_key") {
		t.Errorf("keyFileが期待値と異なります: 期待=%s, 実際=%s", filepath.Join(tempDir, ".encryption_key"), em.keyFile)
	}
}

func TestEncryptionManager_InitializeEncryption(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	// 初期化後の状態確認はGetEncryptionStatusで行う
}

func TestEncryptionManager_EncryptDecryptData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 暗号化を有効にする
	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	originalData := []byte("これは機密データです")

	// 暗号化テスト
	encrypted, err := em.EncryptData(originalData)
	if err != nil {
		t.Fatalf("データの暗号化に失敗: %v", err)
	}

	if len(encrypted) == 0 {
		t.Fatal("暗号化されたデータが空です")
	}

	if bytes.Equal(encrypted, originalData) {
		t.Error("データが暗号化されていません")
	}

	// 復号化テスト
	decrypted, err := em.DecryptData(encrypted)
	if err != nil {
		t.Fatalf("データの復号化に失敗: %v", err)
	}

	if !bytes.Equal(decrypted, originalData) {
		t.Errorf("復号化されたデータが元のデータと異なります: 期待=%s, 実際=%s", string(originalData), string(decrypted))
	}
}

func TestEncryptionManager_EmptyData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	// 空のデータのテスト
	emptyData := []byte("")
	encrypted, err := em.EncryptData(emptyData)
	if err != nil {
		t.Fatalf("空データの暗号化に失敗: %v", err)
	}

	decrypted, err := em.DecryptData(encrypted)
	if err != nil {
		t.Fatalf("空データの復号化に失敗: %v", err)
	}

	if !bytes.Equal(decrypted, emptyData) {
		t.Error("空データの暗号化/復号化が正しく動作していません")
	}
}

func TestEncryptionManager_LargeData(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	// 大きなデータのテスト（1MB）
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	encrypted, err := em.EncryptData(largeData)
	if err != nil {
		t.Fatalf("大きなデータの暗号化に失敗: %v", err)
	}

	decrypted, err := em.DecryptData(encrypted)
	if err != nil {
		t.Fatalf("大きなデータの復号化に失敗: %v", err)
	}

	if !bytes.Equal(decrypted, largeData) {
		t.Error("大きなデータの暗号化/復号化が正しく動作していません")
	}
}

func TestEncryptionManager_MultipleEncryptions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	testData := []string{
		"テストデータ1",
		"Test data 2",
		"データ３のテスト",
		"🔐 Encrypted test 🔑",
	}

	// 複数のデータを暗号化/復号化
	for i, data := range testData {
		originalData := []byte(data)

		encrypted, err := em.EncryptData(originalData)
		if err != nil {
			t.Fatalf("データ%dの暗号化に失敗: %v", i+1, err)
		}

		decrypted, err := em.DecryptData(encrypted)
		if err != nil {
			t.Fatalf("データ%dの復号化に失敗: %v", i+1, err)
		}

		if !bytes.Equal(decrypted, originalData) {
			t.Errorf("データ%dの暗号化/復号化が正しく動作していません: 期待=%s, 実際=%s", i+1, string(originalData), string(decrypted))
		}
	}
}

func TestEncryptionManager_InvalidDecryption(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	// 無効なデータの復号化テスト
	invalidData := []byte("これは暗号化されていないデータです")
	_, err = em.DecryptData(invalidData)
	if err == nil {
		t.Error("無効なデータの復号化がエラーになりませんでした")
	}

	// 短すぎるデータの復号化テスト
	tooShortData := []byte("short")
	_, err = em.DecryptData(tooShortData)
	if err == nil {
		t.Error("短すぎるデータの復号化がエラーになりませんでした")
	}
}

func TestEncryptionManager_GetEncryptionStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 暗号化を無効にした場合
	em1 := NewEncryptionManager(tempDir)
	status1 := em1.GetEncryptionStatus()
	if enabled, ok := status1["enabled"].(bool); !ok || enabled {
		t.Error("暗号化無効なのにenabledがtrueです")
	}

	// 暗号化を有効にした場合
	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em2 := NewEncryptionManager(tempDir)
	err = em2.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	status2 := em2.GetEncryptionStatus()
	if enabled, ok := status2["enabled"].(bool); !ok || !enabled {
		t.Error("暗号化有効なのにenabledがfalseです")
	}

	expectedKeys := []string{"enabled", "passphrase_set", "key_file_exists", "encryption_method"}
	for _, key := range expectedKeys {
		if _, exists := status2[key]; !exists {
			t.Errorf("ステータスに%sが含まれていません", key)
		}
	}
}

func TestEncryptionManager_UniqueEncryption(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aict-encryption-test")
	if err != nil {
		t.Fatalf("一時ディレクトリの作成に失敗: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("AICT_ENCRYPT_DATA", "true")
	defer os.Unsetenv("AICT_ENCRYPT_DATA")

	em := NewEncryptionManager(tempDir)

	err = em.InitializeEncryption()
	if err != nil {
		t.Fatalf("暗号化の初期化に失敗: %v", err)
	}

	originalData := []byte("同じデータ")

	// 同じデータを2回暗号化
	encrypted1, err := em.EncryptData(originalData)
	if err != nil {
		t.Fatalf("1回目の暗号化に失敗: %v", err)
	}

	encrypted2, err := em.EncryptData(originalData)
	if err != nil {
		t.Fatalf("2回目の暗号化に失敗: %v", err)
	}

	// 暗号化結果が異なることを確認（ソルトが異なるため）
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("同じデータの暗号化結果が同一です（ソルトが機能していません）")
	}

	// 両方とも正しく復号化できることを確認
	decrypted1, err := em.DecryptData(encrypted1)
	if err != nil {
		t.Fatalf("1回目の復号化に失敗: %v", err)
	}

	decrypted2, err := em.DecryptData(encrypted2)
	if err != nil {
		t.Fatalf("2回目の復号化に失敗: %v", err)
	}

	if !bytes.Equal(decrypted1, originalData) || !bytes.Equal(decrypted2, originalData) {
		t.Error("復号化結果が元のデータと一致しません")
	}
}