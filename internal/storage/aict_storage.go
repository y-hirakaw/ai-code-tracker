package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// CheckpointTTL はチェックポイントの有効期限（24時間）。
// stash等で長期間放置されたチェックポイントの誤照合を防止する。
const CheckpointTTL = 24 * time.Hour

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

// lockCheckpointsFile はチェックポイントファイルのアドバイザリロックを取得します。
// SaveCheckpointとrewriteCheckpointsの競合を防止。
func (s *AIctStorage) lockCheckpointsFile() (*os.File, error) {
	lockPath := filepath.Join(s.gitDir, CheckpointsDirName, LatestFileName+".lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, fmt.Errorf("creating lock directory: %w", err)
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening lock file: %w", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("acquiring lock: %w", err)
	}

	return f, nil
}

// unlockCheckpointsFile はアドバイザリロックを解放します。
func unlockCheckpointsFile(f *os.File) {
	if f != nil {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}
}

// SaveCheckpoint appends a checkpoint as a JSONL line to latest.json.
// 旧JSON配列形式のファイルが存在する場合、自動的にJSONL形式にマイグレーションします。
func (s *AIctStorage) SaveCheckpoint(cp *tracker.CheckpointV2) error {
	checkpointsDir := filepath.Join(s.gitDir, CheckpointsDirName)
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return err
	}

	// アドバイザリロック取得（rewriteCheckpointsとの競合防止）
	lockFile, err := s.lockCheckpointsFile()
	if err != nil {
		return fmt.Errorf("acquiring checkpoint lock: %w", err)
	}
	defer unlockCheckpointsFile(lockFile)

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

// migrateToJSONLIfNeeded は旧JSON配列形式のチェックポイントファイルを
// JSONL（1行1JSON）形式にマイグレーションします。
// SaveCheckpointのロック内で呼ばれるため、呼び出し元がロックを保持している前提です。
// ファイルが存在しない・空・既にJSONL形式の場合は何もしません。
// tmp+renameパターンでクラッシュ安全性を確保しています。
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

	// JSONL形式でtmp+renameパターンで安全に書き直し
	var buf bytes.Buffer
	for _, cp := range checkpoints {
		line, err := json.Marshal(cp)
		if err != nil {
			return err
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	tmpFile := path + ".migrate.tmp"
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpFile, path); err != nil {
		os.Remove(tmpFile)
		return err
	}
	return nil
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

// RemoveConsumedCheckpoints は照合で使用されたチェックポイントのみを削除し、
// 未使用のチェックポイントを残します（stash退避中の変更の保全用）。
// 同じBaseCommitを共有するペア（Developer baseline + AI edit）も一緒に消費します。
// Load→Process→Rewrite全体をロック保護してTOCTOU競合を防止します。
func (s *AIctStorage) RemoveConsumedCheckpoints(consumedTimestamps map[time.Time]bool) error {
	if len(consumedTimestamps) == 0 {
		return nil
	}

	// ロック取得（Load→Rewrite全体を保護）
	lockFile, err := s.lockCheckpointsFile()
	if err != nil {
		return fmt.Errorf("acquiring checkpoint lock: %w", err)
	}
	defer unlockCheckpointsFile(lockFile)

	checkpoints, err := s.LoadCheckpoints()
	if err != nil {
		return err
	}

	// 同じBaseCommitを共有するチェックポイントもペアで消費
	expandConsumedByBaseCommit(checkpoints, consumedTimestamps)

	var remaining []*tracker.CheckpointV2
	for _, cp := range checkpoints {
		if !consumedTimestamps[cp.Timestamp] {
			remaining = append(remaining, cp)
		}
	}

	if len(remaining) == 0 {
		return s.clearCheckpointsLocked()
	}

	return s.rewriteCheckpointsLocked(remaining)
}

// expandConsumedByBaseCommit は消費対象のチェックポイントと同じBaseCommitを
// 共有し、かつファイルパスが重複するチェックポイントも消費対象に追加します。
// これにより、Developer baseline + AI editのペアが一緒に消費されます。
// 同じBaseCommitでもファイルパスが重複しないチェックポイント（別のstashセッション由来）は残します。
func expandConsumedByBaseCommit(checkpoints []*tracker.CheckpointV2, consumed map[time.Time]bool) {
	// 消費済みチェックポイントの情報を収集
	type baseGroup struct {
		baseCommit string
		files      map[string]bool // 消費済みCPのChangesファイルパス集合
	}
	groups := make(map[string]*baseGroup) // BaseCommit -> group
	for _, cp := range checkpoints {
		if !consumed[cp.Timestamp] {
			continue
		}
		g, ok := groups[cp.BaseCommit]
		if !ok {
			g = &baseGroup{baseCommit: cp.BaseCommit, files: make(map[string]bool)}
			groups[cp.BaseCommit] = g
		}
		for fpath := range cp.Changes {
			g.files[fpath] = true
		}
		// Snapshotのファイルパスも考慮（Changes空のbaselineチェックポイント用）
		for fpath := range cp.Snapshot {
			g.files[fpath] = true
		}
	}
	if len(groups) == 0 {
		return
	}

	// 同じBaseCommitかつファイルパスが重複するチェックポイントを消費対象に追加
	for _, cp := range checkpoints {
		if consumed[cp.Timestamp] {
			continue
		}
		g, ok := groups[cp.BaseCommit]
		if !ok {
			continue
		}
		if hasFileOverlap(cp, g.files) {
			consumed[cp.Timestamp] = true
		}
	}
}

// hasFileOverlap はチェックポイントのファイルパスが対象ファイル集合と重複するか判定します。
// ChangesもSnapshotも空のチェックポイント（初回ベースライン等）は常にマッチします。
func hasFileOverlap(cp *tracker.CheckpointV2, targetFiles map[string]bool) bool {
	// 空のチェックポイントはベースラインマーカーなので常にペア消費対象
	if len(cp.Changes) == 0 && len(cp.Snapshot) == 0 {
		return true
	}
	for fpath := range cp.Changes {
		if targetFiles[fpath] {
			return true
		}
	}
	for fpath := range cp.Snapshot {
		if targetFiles[fpath] {
			return true
		}
	}
	return false
}

// PurgeExpiredCheckpoints はTTLを超えた古いチェックポイントを削除します。
// Load→Process→Rewrite全体をロック保護してTOCTOU競合を防止します。
func (s *AIctStorage) PurgeExpiredCheckpoints(ttl time.Duration) error {
	if ttl <= 0 {
		ttl = CheckpointTTL
	}
	effectiveTTL := ttl

	// ロック取得（Load→Rewrite全体を保護）
	lockFile, err := s.lockCheckpointsFile()
	if err != nil {
		return fmt.Errorf("acquiring checkpoint lock: %w", err)
	}
	defer unlockCheckpointsFile(lockFile)

	checkpoints, err := s.LoadCheckpoints()
	if err != nil {
		return err
	}
	if len(checkpoints) == 0 {
		return nil
	}

	now := time.Now()
	var valid []*tracker.CheckpointV2
	for _, cp := range checkpoints {
		if now.Sub(cp.Timestamp) < effectiveTTL {
			valid = append(valid, cp)
		}
	}

	if len(valid) == len(checkpoints) {
		return nil // 全て有効期限内
	}

	if len(valid) == 0 {
		return s.clearCheckpointsLocked()
	}

	return s.rewriteCheckpointsLocked(valid)
}

// rewriteCheckpoints はチェックポイントリストをJSONL形式で書き直します。
// アドバイザリロック + 一時ファイル + rename パターンでクラッシュ安全性を確保。
func (s *AIctStorage) rewriteCheckpoints(checkpoints []*tracker.CheckpointV2) error {
	lockFile, err := s.lockCheckpointsFile()
	if err != nil {
		return fmt.Errorf("acquiring checkpoint lock: %w", err)
	}
	defer unlockCheckpointsFile(lockFile)

	return s.rewriteCheckpointsLocked(checkpoints)
}

// rewriteCheckpointsLocked はロック保持済みの状態でチェックポイントを書き直します。
// 呼び出し元がロックを保持していることが前提です。
func (s *AIctStorage) rewriteCheckpointsLocked(checkpoints []*tracker.CheckpointV2) error {
	checkpointsDir := filepath.Join(s.gitDir, CheckpointsDirName)
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return err
	}

	checkpointsFile := filepath.Join(checkpointsDir, LatestFileName)
	tmpFile := checkpointsFile + ".tmp"

	var buf bytes.Buffer
	for _, cp := range checkpoints {
		line, err := json.Marshal(cp)
		if err != nil {
			return fmt.Errorf("marshal checkpoint: %w", err)
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, checkpointsFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// clearCheckpointsLocked はロック保持済みの状態でチェックポイントを削除します。
func (s *AIctStorage) clearCheckpointsLocked() error {
	checkpointsFile := filepath.Join(s.gitDir, CheckpointsDirName, LatestFileName)
	err := os.Remove(checkpointsFile)
	if os.IsNotExist(err) {
		return nil
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

	// バリデーション
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validateConfig はConfig値の妥当性を検証します。
func validateConfig(cfg *tracker.Config) error {
	if cfg.TargetAIPercentage < 0 || cfg.TargetAIPercentage > 100 {
		return fmt.Errorf("target_ai_percentage must be between 0 and 100, got %.1f", cfg.TargetAIPercentage)
	}

	if len(cfg.TrackedExtensions) == 0 {
		return fmt.Errorf("tracked_extensions must not be empty")
	}

	if cfg.DefaultAuthor == "" {
		return fmt.Errorf("default_author must not be empty")
	}

	if cfg.CheckpointTTLHours < 0 {
		return fmt.Errorf("checkpoint_ttl_hours must be >= 0, got %d", cfg.CheckpointTTLHours)
	}

	return nil
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
