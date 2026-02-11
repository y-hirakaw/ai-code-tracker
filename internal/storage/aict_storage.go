package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

const (
	AictDirName        = "aict"
	CheckpointsDirName = "checkpoints"
	LatestFileName     = "latest.json"
	ConfigFileName     = "config.json"
)

// AIctStorage manages .git/aict/ directory
type AIctStorage struct {
	gitDir string // .git/aict/
}

// NewAIctStorage creates a new AIctStorage instance
func NewAIctStorage() (*AIctStorage, error) {
	// 1. .git ディレクトリを検出
	gitDir, err := findGitDir()
	if err != nil {
		return nil, err
	}

	// 2. .git/aict/ を作成
	aictDir := filepath.Join(gitDir, AictDirName)
	if err := os.MkdirAll(aictDir, 0755); err != nil {
		return nil, err
	}

	return &AIctStorage{gitDir: aictDir}, nil
}

// SaveCheckpoint appends a checkpoint as a JSONL line to latest.json.
// 旧JSON配列形式のファイルが存在する場合、自動的にJSONL形式にマイグレーションします。
func (s *AIctStorage) SaveCheckpoint(cp *tracker.CheckpointV2) error {
	checkpointsDir := filepath.Join(s.gitDir, CheckpointsDirName)
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return err
	}

	checkpointsFile := filepath.Join(checkpointsDir, LatestFileName)

	// 旧JSON配列形式の場合、JSONL形式にマイグレーション
	if err := migrateToJSONLIfNeeded(checkpointsFile); err != nil {
		return fmt.Errorf("failed to migrate checkpoint format: %w", err)
	}

	// 単一チェックポイントをコンパクトJSONにシリアライズ
	data, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// ファイルに追記（O_APPENDは小さな書き込みに対してアトミック）
	f, err := os.OpenFile(checkpointsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

// LoadCheckpoints loads all checkpoints from latest.json.
// JSON配列（旧形式）とJSONL（新形式）の両方を自動判別して読み込みます。
func (s *AIctStorage) LoadCheckpoints() ([]*tracker.CheckpointV2, error) {
	checkpointsFile := filepath.Join(s.gitDir, CheckpointsDirName, LatestFileName)
	return loadCheckpointsFromFile(checkpointsFile)
}

// loadCheckpointsFromFile reads checkpoints from a file, auto-detecting format.
func loadCheckpointsFromFile(path string) ([]*tracker.CheckpointV2, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*tracker.CheckpointV2{}, nil
		}
		return nil, err
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return []*tracker.CheckpointV2{}, nil
	}

	// 旧形式判定: '[' で始まる場合はJSON配列
	if data[0] == '[' {
		var checkpoints []*tracker.CheckpointV2
		if err := json.Unmarshal(data, &checkpoints); err != nil {
			return nil, err
		}
		return checkpoints, nil
	}

	// JSONL形式: 1行1JSONオブジェクト（不正な行はスキップ）
	var checkpoints []*tracker.CheckpointV2
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var cp tracker.CheckpointV2
		if err := json.Unmarshal(line, &cp); err != nil {
			log.Printf("Warning: skipping invalid JSONL line in checkpoints: %v", err)
			continue
		}
		checkpoints = append(checkpoints, &cp)
	}
	return checkpoints, nil
}

// migrateToJSONLIfNeeded converts a legacy JSON array file to JSONL format.
func migrateToJSONLIfNeeded(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '[' {
		return nil // 既にJSONLまたは空ファイル
	}

	// 旧JSON配列をパース
	var checkpoints []*tracker.CheckpointV2
	if err := json.Unmarshal(data, &checkpoints); err != nil {
		return err
	}

	// JSONL形式で書き直し
	var buf bytes.Buffer
	for _, cp := range checkpoints {
		line, err := json.Marshal(cp)
		if err != nil {
			return err
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// ClearCheckpoints removes all checkpoints
func (s *AIctStorage) ClearCheckpoints() error {
	checkpointsFile := filepath.Join(s.gitDir, CheckpointsDirName, LatestFileName)
	err := os.Remove(checkpointsFile)
	if os.IsNotExist(err) {
		return nil // Already cleared
	}
	return err
}

// SaveConfig saves config.json
func (s *AIctStorage) SaveConfig(cfg *tracker.Config) error {
	configFile := filepath.Join(s.gitDir, ConfigFileName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

// LoadConfig loads config.json
func (s *AIctStorage) LoadConfig() (*tracker.Config, error) {
	configFile := filepath.Join(s.gitDir, ConfigFileName)
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg tracker.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// AuthorMappingsの初期化（nil書き込み防止）
	if cfg.AuthorMappings == nil {
		cfg.AuthorMappings = make(map[string]string)
	}

	return &cfg, nil
}

// GetAictDir returns the .git/aict directory path
func (s *AIctStorage) GetAictDir() string {
	return s.gitDir
}

// findGitDir finds .git directory from current directory
func findGitDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return gitDir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf(".git directory not found")
		}
		dir = parent
	}
}
