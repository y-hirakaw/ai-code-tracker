package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
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
	aictDir := filepath.Join(gitDir, "aict")
	if err := os.MkdirAll(aictDir, 0755); err != nil {
		return nil, err
	}

	return &AIctStorage{gitDir: aictDir}, nil
}

// SaveCheckpoint appends a checkpoint to latest.json
func (s *AIctStorage) SaveCheckpoint(cp *tracker.CheckpointV2) error {
	// .git/aict/checkpoints/latest.json に追記（配列形式）
	checkpointsDir := filepath.Join(s.gitDir, "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return err
	}

	checkpointsFile := filepath.Join(checkpointsDir, "latest.json")

	// 既存のチェックポイントを読み込み
	checkpoints, _ := s.LoadCheckpoints()

	// 新しいチェックポイントを追加
	checkpoints = append(checkpoints, cp)

	// JSON配列として保存
	data, err := json.MarshalIndent(checkpoints, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(checkpointsFile, data, 0644)
}

// LoadCheckpoints loads all checkpoints from latest.json
func (s *AIctStorage) LoadCheckpoints() ([]*tracker.CheckpointV2, error) {
	checkpointsFile := filepath.Join(s.gitDir, "checkpoints", "latest.json")

	data, err := os.ReadFile(checkpointsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*tracker.CheckpointV2{}, nil
		}
		return nil, err
	}

	var checkpoints []*tracker.CheckpointV2
	if err := json.Unmarshal(data, &checkpoints); err != nil {
		return nil, err
	}

	return checkpoints, nil
}

// ClearCheckpoints removes all checkpoints
func (s *AIctStorage) ClearCheckpoints() error {
	checkpointsFile := filepath.Join(s.gitDir, "checkpoints", "latest.json")
	err := os.Remove(checkpointsFile)
	if os.IsNotExist(err) {
		return nil // Already cleared
	}
	return err
}

// SaveConfig saves config.json
func (s *AIctStorage) SaveConfig(cfg *tracker.Config) error {
	configFile := filepath.Join(s.gitDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

// LoadConfig loads config.json
func (s *AIctStorage) LoadConfig() (*tracker.Config, error) {
	configFile := filepath.Join(s.gitDir, "config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg tracker.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
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
