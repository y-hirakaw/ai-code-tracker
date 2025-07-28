package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
)

// GetCurrentDirectory は現在のディレクトリを取得する共通関数
func GetCurrentDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", errors.WrapError(err, errors.ErrorTypeFile, "directory_access_failed")
	}
	return currentDir, nil
}

// ParseDate は日付文字列を解析する共通関数
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}
	
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, errors.InvalidDateFormat(dateStr)
	}
	return parsedTime, nil
}

// ValidateFilePath はファイルパスの存在を検証する共通関数
func ValidateFilePath(filePath string) error {
	if filePath == "" {
		return errors.NewError(errors.ErrorTypeFile, "empty_file_path")
	}
	
	// ファイルが存在するかチェック
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return errors.FileNotFound(filePath)
	} else if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "file_access_failed")
	}
	
	return nil
}

// EnsureDirectory はディレクトリが存在しない場合作成する
func EnsureDirectory(dirPath string) error {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "directory_creation_failed")
	}
	return nil
}

// JoinPath は安全にパスを結合する
func JoinPath(elements ...string) string {
	return filepath.Join(elements...)
}

// IsGitRepository は指定されたディレクトリがGitリポジトリかチェックする
func IsGitRepository(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// GetHomeDirectory はユーザーのホームディレクトリを取得する
func GetHomeDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.WrapError(err, errors.ErrorTypeFile, "home_directory_access_failed")
	}
	return homeDir, nil
}

// FileExists はファイルが存在するかチェックする
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// CreateFileIfNotExists はファイルが存在しない場合作成する
func CreateFileIfNotExists(filePath string) error {
	if FileExists(filePath) {
		return nil
	}
	
	// ディレクトリを作成
	dir := filepath.Dir(filePath)
	if err := EnsureDirectory(dir); err != nil {
		return err
	}
	
	// ファイルを作成
	file, err := os.Create(filePath)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "file_creation_failed")
	}
	defer file.Close()
	
	return nil
}

// GenerateSessionID はセッション用のランダムIDを生成する
func GenerateSessionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetCurrentTime は現在時刻を取得する
func GetCurrentTime() time.Time {
	return time.Now()
}

// GetCurrentTimeString は現在時刻の文字列を取得する
func GetCurrentTimeString() string {
	return time.Now().Format(time.RFC3339)
}

// ParseFiles はファイルリスト文字列を解析する
func ParseFiles(filesStr string) []string {
	if filesStr == "" {
		return []string{}
	}
	return SplitAndTrim(filesStr, ",")
}

// WriteJSON はJSONファイルを書き込む
func WriteJSON(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "file_creation_failed")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ReadJSON はJSONファイルを読み込む
func ReadJSON(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeFile, "file_open_failed")
	}
	defer file.Close()

	var data map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, errors.WrapError(err, errors.ErrorTypeData, "json_decode_failed")
	}
	return data, nil
}