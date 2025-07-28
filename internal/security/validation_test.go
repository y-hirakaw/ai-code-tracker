package security

import (
	"os"
	"testing"
)

func TestNewValidationManager(t *testing.T) {
	vm := NewValidationManager()
	if vm == nil {
		t.Fatal("ValidationManagerがnilです")
	}

	if vm.maxPathLength != 4096 {
		t.Errorf("maxPathLengthが期待値と異なります: 期待=4096, 実際=%d", vm.maxPathLength)
	}

	if vm.maxFileNameLength != 255 {
		t.Errorf("maxFileNameLengthが期待値と異なります: 期待=255, 実際=%d", vm.maxFileNameLength)
	}
}

func TestValidationManager_ValidateFilePath(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"正常なパス", "src/main.go", true},
		{"空のパス", "", false},
		{"パストラバーサル", "../../../etc/passwd", false},
		{"NULLバイト", "file\x00.txt", false},
		{"正常な日本語パス", "テスト/ファイル.go", true},
		{"長すぎるパス", string(make([]byte, 5000)), false},
		{"絶対パス（通常モード）", "/absolute/path.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateFilePath(tt.path)
			if tt.expected && err != nil {
				t.Errorf("有効なパスがエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なパスがエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateFilePathStrict(t *testing.T) {
	// ストリクトモードを有効にする
	os.Setenv("AICT_SECURITY_MODE", "strict")
	defer os.Unsetenv("AICT_SECURITY_MODE")

	vm := NewValidationManager()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"正常なGoファイル", "src/main.go", true},
		{"許可されていない拡張子", "test.xyz", false},
		{"絶対パス", "/absolute/path.go", false},
		{"隠しファイル", ".hidden", false},
		{"拡張子なし", "README", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateFilePath(tt.path)
			if tt.expected && err != nil {
				t.Errorf("有効なパスがエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なパスがエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateFileName(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{"正常なファイル名", "main.go", true},
		{"空のファイル名", "", false},
		{"長すぎるファイル名", string(make([]byte, 300)), false},
		{"制御文字を含む", "file\x01.txt", false},
		{"Windows予約名", "CON.txt", false},
		{"日本語ファイル名", "テスト.go", true},
		{"特殊文字", "file-name_123.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateFileName(tt.fileName)
			if tt.expected && err != nil {
				t.Errorf("有効なファイル名がエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なファイル名がエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateFileExtension(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"許可された拡張子", "test.go", true},
		{"許可されていない拡張子", "test.exe", true}, // 通常モードでは許可
		{"拡張子なし", "README", true},               // 通常モードでは許可
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateFileExtension(tt.filePath)
			if tt.expected && err != nil {
				t.Errorf("有効な拡張子がエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効な拡張子がエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateEventData(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected bool
	}{
		{
			"有効なイベントデータ",
			map[string]interface{}{
				"id":         "test-123",
				"timestamp":  "2025-01-01T00:00:00Z",
				"event_type": "ai_edit",
				"author":     "test-user",
				"message":    "テストメッセージ",
			},
			true,
		},
		{
			"IDが不足",
			map[string]interface{}{
				"timestamp":  "2025-01-01T00:00:00Z",
				"event_type": "ai_edit",
				"author":     "test-user",
			},
			false,
		},
		{
			"無効なID",
			map[string]interface{}{
				"id":         "test@123",
				"timestamp":  "2025-01-01T00:00:00Z",
				"event_type": "ai_edit",
				"author":     "test-user",
			},
			false,
		},
		{
			"無効な作成者名",
			map[string]interface{}{
				"id":         "test-123",
				"timestamp":  "2025-01-01T00:00:00Z",
				"event_type": "ai_edit",
				"author":     "test<script>alert('xss')</script>",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateEventData(tt.data)
			if tt.expected && err != nil {
				t.Errorf("有効なイベントデータがエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なイベントデータがエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateID(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		id       string
		expected bool
	}{
		{"有効なID", "test-123", true},
		{"空のID", "", false},
		{"長すぎるID", string(make([]byte, 200)), false},
		{"無効な文字", "test@123", false},
		{"ハイフンとアンダースコア", "test_id-123", true},
		{"数字のみ", "123456", true},
		{"アルファベットのみ", "testid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateID(tt.id)
			if tt.expected && err != nil {
				t.Errorf("有効なIDがエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なIDがエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateAuthorName(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		author   string
		expected bool
	}{
		{"有効な作成者名", "test-user", true},
		{"空の作成者名", "", false},
		{"長すぎる作成者名", string(make([]byte, 300)), false},
		{"危険な文字", "test<script>", false},
		{"日本語の作成者名", "テストユーザー", true},
		{"メールアドレス形式", "user@example.com", true},
		{"NULLバイト", "user\x00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateAuthorName(tt.author)
			if tt.expected && err != nil {
				t.Errorf("有効な作成者名がエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効な作成者名がエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_ValidateMessage(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{"有効なメッセージ", "これはテストメッセージです", true},
		{"空のメッセージ", "", true}, // 空のメッセージは許可
		{"長すぎるメッセージ", string(make([]byte, 2000)), false},
		{"NULLバイト", "message\x00", false},
		{"日本語メッセージ", "日本語のメッセージです", true},
		{"絵文字", "テスト 🚀 メッセージ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateMessage(tt.message)
			if tt.expected && err != nil {
				t.Errorf("有効なメッセージがエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なメッセージがエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_SanitizeFilePath(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"正常なパス", "src/main.go", "src/main.go"},
		{"パストラバーサル", "../../../etc/passwd", "etc/passwd"},
		{"絶対パス", "/absolute/path.go", "absolute/path.go"},
		{"NULLバイト", "file\x00.txt", "file.txt"},
		{"複数のドット", "file/../test.go", "test.go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.SanitizeFilePath(tt.input)
			if result != tt.expected {
				t.Errorf("サニタイズ結果が期待値と異なります: 期待=%s, 実際=%s", tt.expected, result)
			}
		})
	}
}

func TestValidationManager_SanitizeString(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"正常な文字列", "Hello World", "Hello World"},
		{"NULLバイト", "Hello\x00World", "HelloWorld"},
		{"制御文字", "Hello\x01\x02World", "HelloWorld"},
		{"タブと改行", "Hello\tWorld\n", "HelloWorld"},
		{"日本語", "こんにちは世界", "こんにちは世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("サニタイズ結果が期待値と異なります: 期待=%s, 実際=%s", tt.expected, result)
			}
		})
	}
}

func TestValidationManager_ValidateBatchOperation(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		files    []string
		maxFiles int
		expected bool
	}{
		{
			"有効なバッチ",
			[]string{"file1.go", "file2.go", "file3.go"},
			5,
			true,
		},
		{
			"空のリスト",
			[]string{},
			5,
			false,
		},
		{
			"ファイル数制限超過",
			[]string{"file1.go", "file2.go", "file3.go"},
			2,
			false,
		},
		{
			"重複ファイル",
			[]string{"file1.go", "file1.go", "file2.go"},
			5,
			false,
		},
		{
			"無効なパスを含む",
			[]string{"file1.go", "../../../etc/passwd"},
			5,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateBatchOperation(tt.files, tt.maxFiles)
			if tt.expected && err != nil {
				t.Errorf("有効なバッチ操作がエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なバッチ操作がエラーになりませんでした")
			}
		})
	}
}

func TestValidationManager_IsSecureMode(t *testing.T) {
	// 通常モード
	vm1 := NewValidationManager()
	if vm1.IsSecureMode() {
		t.Error("通常モードなのにセキュアモードがtrueです")
	}

	// ストリクトモード
	os.Setenv("AICT_SECURITY_MODE", "strict")
	defer os.Unsetenv("AICT_SECURITY_MODE")

	vm2 := NewValidationManager()
	if !vm2.IsSecureMode() {
		t.Error("ストリクトモードなのにセキュアモードがfalseです")
	}
}

func TestValidationManager_GetValidationRules(t *testing.T) {
	vm := NewValidationManager()

	rules := vm.GetValidationRules()
	if rules == nil {
		t.Fatal("検証ルールがnilです")
	}

	expectedKeys := []string{
		"strict_mode", "max_path_length", "max_filename_length",
		"allowed_extensions", "blocked_patterns", "path_traversal_check",
		"null_byte_check", "control_char_check", "reserved_name_check",
	}

	for _, key := range expectedKeys {
		if _, exists := rules[key]; !exists {
			t.Errorf("検証ルールに%sが含まれていません", key)
		}
	}
}

func TestValidationManager_WindowsReservedNames(t *testing.T) {
	vm := NewValidationManager()

	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM9",
		"LPT1", "LPT2", "LPT9",
	}

	for _, name := range reservedNames {
		t.Run(name, func(t *testing.T) {
			err := vm.ValidateFileName(name)
			if err == nil {
				t.Errorf("Windows予約名%sがエラーになりませんでした", name)
			}

			// 拡張子付きでも検証
			err = vm.ValidateFileName(name + ".txt")
			if err == nil {
				t.Errorf("Windows予約名%s.txtがエラーになりませんでした", name)
			}
		})
	}
}

func TestValidationManager_UTF8Validation(t *testing.T) {
	vm := NewValidationManager()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"有効なUTF-8", "Hello 世界 🌍", true},
		{"ASCII", "Hello World", true},
		{"無効なUTF-8", string([]byte{0xff, 0xfe, 0xfd}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateFilePath(tt.input)
			if tt.expected && err != nil && err.Error() == "ファイルパスに無効なUTF-8文字が含まれています" {
				t.Errorf("有効なUTF-8がエラーになりました: %v", err)
			}
			if !tt.expected && err == nil {
				t.Error("無効なUTF-8がエラーになりませんでした")
			}
		})
	}
}