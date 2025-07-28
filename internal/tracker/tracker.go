package tracker

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
)

// Tracker はファイル変更の追跡とイベントの記録を管理する
type Tracker struct {
	storage       storage.StorageInterface
	gitRepo       string
	lastEventTime time.Time
	duplicateWindow time.Duration
}

// DiffResult はファイルの差分情報を表す
type DiffResult struct {
	FilePath      string
	LinesAdded    int
	LinesModified int
	LinesDeleted  int
	ContentHash   string
}

// NewTracker は新しいTrackerインスタンスを作成する
func NewTracker(storageInstance storage.StorageInterface, gitRepo string) *Tracker {
	return &Tracker{
		storage:         storageInstance,
		gitRepo:         gitRepo,
		duplicateWindow: 5 * time.Second, // 5秒以内の重複イベントを防ぐ
	}
}

// TrackFileChanges はファイルの変更を追跡し、イベントを記録する
func (t *Tracker) TrackFileChanges(eventType types.EventType, author string, model string, files []string, message string) error {
	// 重複イベントチェック
	if t.isDuplicateEvent() {
		return nil // 重複イベントをスキップ
	}

	// ファイル変更情報を収集
	var fileInfos []types.FileInfo
	for _, filePath := range files {
		diffResult, err := t.calculateFileDiff(filePath)
		if err != nil {
			return fmt.Errorf("ファイル %s の差分計算に失敗しました: %w", filePath, err)
		}

		fileInfo := types.FileInfo{
			Path:          diffResult.FilePath,
			LinesAdded:    diffResult.LinesAdded,
			LinesModified: diffResult.LinesModified,
			LinesDeleted:  diffResult.LinesDeleted,
			Hash:          diffResult.ContentHash,
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	// トラッキングイベントを作成
	now := time.Now()
	event := &types.TrackEvent{
		ID:        types.GenerateEventID(now, eventType, author),
		Timestamp: now,
		EventType: eventType,
		Author:    author,
		Model:     model,
		Files:     fileInfos,
		Message:   message,
	}

	// イベントを保存
	if err := t.storage.StoreTrackEvent(event); err != nil {
		return fmt.Errorf("イベントの保存に失敗しました: %w", err)
	}

	// 最後のイベント時刻を更新
	t.lastEventTime = now

	return nil
}

// TrackCommit はGitコミットを追跡する
func (t *Tracker) TrackCommit(commitHash string, author string, message string) error {
	// コミットの変更ファイルを取得
	changedFiles, err := t.getCommitChangedFiles(commitHash)
	if err != nil {
		return fmt.Errorf("コミット %s の変更ファイル取得に失敗しました: %w", commitHash, err)
	}

	// 各ファイルの差分を計算
	var fileInfos []types.FileInfo
	for _, filePath := range changedFiles {
		diffResult, err := t.calculateCommitFileDiff(commitHash, filePath)
		if err != nil {
			// ファイルが削除された場合など、エラーを無視して続行
			continue
		}

		fileInfo := types.FileInfo{
			Path:          diffResult.FilePath,
			LinesAdded:    diffResult.LinesAdded,
			LinesModified: diffResult.LinesModified,
			LinesDeleted:  diffResult.LinesDeleted,
			Hash:          diffResult.ContentHash,
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	// コミットイベントを作成
	commitTime, err := t.getCommitTime(commitHash)
	if err != nil {
		commitTime = time.Now()
	}

	event := &types.TrackEvent{
		ID:         types.GenerateEventID(commitTime, types.EventTypeCommit, author),
		Timestamp:  commitTime,
		EventType:  types.EventTypeCommit,
		Author:     author,
		CommitHash: commitHash,
		Files:      fileInfos,
		Message:    message,
	}

	// Claude Codeによるコミットかどうかを判定
	if t.isClaudeCodeCommit(message) {
		event.EventType = types.EventTypeAI
		event.Model = t.detectClaudeModel(message)
		event.Author = "Claude Code"
	}

	return t.storage.StoreTrackEvent(event)
}

// calculateFileDiff は現在のファイルと前回の状態との差分を計算する
func (t *Tracker) calculateFileDiff(filePath string) (*DiffResult, error) {
	// ファイルが存在するかチェック
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &DiffResult{
			FilePath:    filePath,
			LinesDeleted: 0, // ファイル削除は別途処理
		}, nil
	}

	// 現在のファイル内容を読み取り
	currentContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み取りエラー: %w", err)
	}

	// ファイルハッシュを計算
	hash := sha256.Sum256(currentContent)
	contentHash := fmt.Sprintf("%x", hash)

	// Git diffを使用して変更を検出
	linesAdded, linesModified, linesDeleted, err := t.calculateGitDiff(filePath)
	if err != nil {
		// Gitで追跡されていないファイルの場合、全行を追加として扱う
		lines := strings.Count(string(currentContent), "\n") + 1
		return &DiffResult{
			FilePath:    filePath,
			LinesAdded:  lines,
			ContentHash: contentHash,
		}, nil
	}

	return &DiffResult{
		FilePath:      filePath,
		LinesAdded:    linesAdded,
		LinesModified: linesModified,
		LinesDeleted:  linesDeleted,
		ContentHash:   contentHash,
	}, nil
}

// calculateGitDiff はGitを使用してファイルの差分を計算する
func (t *Tracker) calculateGitDiff(filePath string) (added, modified, deleted int, err error) {
	// git diff --numstat を使用して統計を取得
	cmd := exec.Command("git", "diff", "--numstat", "HEAD", "--", filePath)
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, err
	}

	// 出力をパース: "追加行数\t削除行数\tファイル名"
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return 0, 0, 0, nil // 変更なし
	}

	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 && strings.HasSuffix(parts[2], filePath) {
			if parts[0] != "-" {
				if a, err := strconv.Atoi(parts[0]); err == nil {
					added = a
				}
			}
			if parts[1] != "-" {
				if d, err := strconv.Atoi(parts[1]); err == nil {
					deleted = d
				}
			}
			// 修正行数は簡単な推定（実際の修正は追加+削除の組み合わせ）
			if added > 0 && deleted > 0 {
				if added < deleted {
					modified = added
					added = 0
				} else {
					modified = deleted
					deleted = 0
				}
			}
			break
		}
	}

	return added, modified, deleted, nil
}

// calculateCommitFileDiff はコミット内の特定ファイルの差分を計算する
func (t *Tracker) calculateCommitFileDiff(commitHash string, filePath string) (*DiffResult, error) {
	// git show --numstat でコミットの統計を取得
	cmd := exec.Command("git", "show", "--numstat", commitHash, "--", filePath)
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var added, deleted int
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			if parts[0] != "-" {
				if a, err := strconv.Atoi(parts[0]); err == nil {
					added = a
				}
			}
			if parts[1] != "-" {
				if d, err := strconv.Atoi(parts[1]); err == nil {
					deleted = d
				}
			}
			break
		}
	}

	// ファイル内容のハッシュを取得
	contentHash, _ := t.getFileHashAtCommit(commitHash, filePath)

	return &DiffResult{
		FilePath:    filePath,
		LinesAdded:  added,
		LinesDeleted: deleted,
		ContentHash: contentHash,
	}, nil
}

// getCommitChangedFiles はコミットで変更されたファイルのリストを取得する
func (t *Tracker) getCommitChangedFiles(commitHash string) ([]string, error) {
	cmd := exec.Command("git", "show", "--name-only", "--pretty=format:", commitHash)
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// getCommitTime はコミットの時刻を取得する
func (t *Tracker) getCommitTime(commitHash string) (time.Time, error) {
	cmd := exec.Command("git", "show", "-s", "--format=%ct", commitHash)
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

// getFileHashAtCommit は特定コミットでのファイルハッシュを取得する
func (t *Tracker) getFileHashAtCommit(commitHash string, filePath string) (string, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", commitHash, filePath))
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(output)
	return fmt.Sprintf("%x", hash), nil
}

// isClaudeCodeCommit はClaude Codeによるコミットかどうかを判定する
func (t *Tracker) isClaudeCodeCommit(message string) bool {
	claudeCodeIndicators := []string{
		"🤖 Generated with [Claude Code]",
		"Co-Authored-By: Claude",
		"claude.ai/code",
	}

	messageLower := strings.ToLower(message)
	for _, indicator := range claudeCodeIndicators {
		if strings.Contains(messageLower, strings.ToLower(indicator)) {
			return true
		}
	}

	return false
}

// detectClaudeModel はコミットメッセージからClaudeモデルを検出する
func (t *Tracker) detectClaudeModel(message string) string {
	// デフォルトはClaude Code
	defaultModel := "claude-code"

	// コミットメッセージからモデル情報を抽出する簡単なパターンマッチング
	modelPatterns := map[string]string{
		"claude-sonnet-4": "claude-sonnet-4",
		"claude-opus-4":   "claude-opus-4",
		"sonnet-4":        "claude-sonnet-4",
		"opus-4":          "claude-opus-4",
		// 後方互換性のため古いモデル名も残す
		"claude-3-opus":   "claude-3-opus",
		"claude-3-sonnet": "claude-3-sonnet",
		"claude-3-haiku":  "claude-3-haiku",
	}

	messageLower := strings.ToLower(message)
	for pattern, model := range modelPatterns {
		if strings.Contains(messageLower, pattern) {
			return model
		}
	}

	return defaultModel
}

// isDuplicateEvent は重複イベントかどうかをチェックする
func (t *Tracker) isDuplicateEvent() bool {
	if t.lastEventTime.IsZero() {
		return false
	}

	timeSinceLastEvent := time.Since(t.lastEventTime)
	return timeSinceLastEvent < t.duplicateWindow
}

// GetRepoRoot はGitリポジトリのルートパスを取得する
func (t *Tracker) GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("Gitリポジトリルートの取得に失敗しました: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsGitRepo は指定されたディレクトリがGitリポジトリかどうかをチェックする
func IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// DetectChangedFiles は未コミットの変更ファイルを検出する
func (t *Tracker) DetectChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = t.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("変更ファイルの検出に失敗しました: %w", err)
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			files = append(files, line)
		}
	}

	// 新規ファイルも追加
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = t.gitRepo

	output, err = cmd.Output()
	if err == nil {
		scanner = bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				files = append(files, line)
			}
		}
	}

	return files, nil
}