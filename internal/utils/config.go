package utils

import (
	"encoding/json"
	"os"

	"github.com/ai-code-tracker/aict/internal/errors"
)

// ConfigManager は設定管理の共通インターフェース
type ConfigManager struct {
	configPath string
}

// NewConfigManager は新しいConfigManagerを作成する
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

// LoadConfig は指定されたパスから設定を読み込む
func (c *ConfigManager) LoadConfig(config interface{}) error {
	if !FileExists(c.configPath) {
		// 設定ファイルが存在しない場合は空の設定として扱う
		return nil
	}
	
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "config_read_failed")
	}
	
	if err := json.Unmarshal(data, config); err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "config_parse_failed")
	}
	
	return nil
}

// SaveConfig は設定を指定されたパスに保存する
func (c *ConfigManager) SaveConfig(config interface{}) error {
	// ディレクトリを作成
	if err := EnsureDirectory(JoinPath(c.configPath, "..")); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeData, "config_marshal_failed")
	}
	
	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "config_write_failed")
	}
	
	return nil
}

// BackupConfig は設定ファイルをバックアップする
func (c *ConfigManager) BackupConfig() error {
	if !FileExists(c.configPath) {
		return nil // バックアップするファイルが存在しない
	}
	
	backupPath := c.configPath + ".backup"
	
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "config_backup_read_failed")
	}
	
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "config_backup_write_failed")
	}
	
	return nil
}

// RestoreConfig はバックアップから設定を復元する
func (c *ConfigManager) RestoreConfig() error {
	backupPath := c.configPath + ".backup"
	
	if !FileExists(backupPath) {
		return errors.NewError(errors.ErrorTypeFile, "backup_not_found")
	}
	
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "backup_read_failed")
	}
	
	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return errors.WrapError(err, errors.ErrorTypeFile, "config_restore_failed")
	}
	
	return nil
}

// ConfigExists は設定ファイルが存在するかチェックする
func (c *ConfigManager) ConfigExists() bool {
	return FileExists(c.configPath)
}

// GetConfigPath は設定ファイルのパスを取得する
func (c *ConfigManager) GetConfigPath() string {
	return c.configPath
}

// DefaultConfig は共通のデフォルト設定構造体
type DefaultConfig struct {
	Version         string            `json:"version"`
	Language        string            `json:"language"`
	DefaultAuthor   string            `json:"default_author"`
	EnableDebug     bool              `json:"enable_debug"`
	LogLevel        string            `json:"log_level"`
	CustomSettings  map[string]string `json:"custom_settings"`
}

// NewDefaultConfig はデフォルト設定を作成する
func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		Version:        "0.1.0",
		Language:       "ja",
		DefaultAuthor:  "",
		EnableDebug:    false,
		LogLevel:       "info",
		CustomSettings: make(map[string]string),
	}
}

// GetEnvironmentOverrides は環境変数から設定上書きを取得する
func GetEnvironmentOverrides() map[string]string {
	overrides := make(map[string]string)
	
	// 一般的な環境変数をチェック
	envVars := map[string]string{
		"AICT_LANGUAGE":      "language",
		"AICT_AUTHOR":        "default_author",
		"AICT_DEBUG":         "enable_debug",
		"AICT_LOG_LEVEL":     "log_level",
	}
	
	for envVar, configKey := range envVars {
		if value := os.Getenv(envVar); value != "" {
			overrides[configKey] = value
		}
	}
	
	return overrides
}

// ApplyEnvironmentOverrides は環境変数による設定上書きを適用する
func ApplyEnvironmentOverrides(config *DefaultConfig) {
	overrides := GetEnvironmentOverrides()
	
	if lang, ok := overrides["language"]; ok {
		config.Language = lang
	}
	if author, ok := overrides["default_author"]; ok {
		config.DefaultAuthor = author
	}
	if debug, ok := overrides["enable_debug"]; ok {
		config.EnableDebug = (debug == "true" || debug == "1")
	}
	if logLevel, ok := overrides["log_level"]; ok {
		config.LogLevel = logLevel
	}
}