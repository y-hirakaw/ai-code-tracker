package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/git"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func handleCommit() {
	// ストレージを初期化
	store, err := storage.NewAIctStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 最新のコミットハッシュを取得
	commitHash, err := getLatestCommitHash()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting commit hash: %v\n", err)
		os.Exit(1)
	}

	// コミットのnumstatを取得
	executor := gitexec.NewExecutor()
	numstatOutput, err := executor.Run("show", "--numstat", "--format=", commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to get numstat for commit %s: %v\n", commitHash, err)
	}

	// numstatから変更されたファイル一覧を取得
	numstatMap, _ := git.ParseNumstat(numstatOutput)
	changedFiles := make(map[string]bool, len(numstatMap))
	for f := range numstatMap {
		changedFiles[f] = true
	}
	if len(changedFiles) == 0 {
		fmt.Println("No tracked files changed in this commit")
		return
	}

	// チェックポイントを読み込み
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checkpoints: %v\n", err)
		os.Exit(1)
	}

	// デバッグ: チェックポイント詳細を出力
	debugf("Loaded %d checkpoints", len(checkpoints))
	for i, cp := range checkpoints {
		debugf("Checkpoint %d: author=%s, type=%s, files=%d", i, cp.Author, cp.Type, len(cp.Changes))
		for filepath := range cp.Changes {
			debugf("  - %s", filepath)
		}
	}

	// 前回コミット（HEAD~1）との完全な差分を取得
	fullDiff, err := getCommitDiff(commitHash)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting commit diff: %v\n", err)
		os.Exit(1)
	}

	// チェックポイントから作成者マッピングを構築
	authorshipMap := buildAuthorshipMap(checkpoints, changedFiles)

	// デバッグ: 作成者マッピングを出力
	debugf("Authorship mapping for %d files:", len(authorshipMap))
	for filepath, cp := range authorshipMap {
		debugf("  %s -> %s (%s)", filepath, cp.Author, cp.Type)
	}

	// 完全な差分情報と作成者情報を統合してAuthorship Logを生成
	log, err := buildAuthorshipLogFromDiff(fullDiff, authorshipMap, commitHash, changedFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building authorship log: %v\n", err)
		os.Exit(1)
	}

	// バリデーション
	if err := authorship.ValidateAuthorshipLog(log); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating authorship log: %v\n", err)
		os.Exit(1)
	}

	// Git notesに保存
	nm := gitnotes.NewNotesManager()
	if err := nm.AddAuthorshipLog(log); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving authorship log: %v\n", err)
		os.Exit(1)
	}

	// チェックポイントをクリア
	if err := store.ClearCheckpoints(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to clear checkpoints: %v\n", err)
	}

	fmt.Println("✓ Authorship log created")
}

// getLatestCommitHash retrieves the latest commit hash
func getLatestCommitHash() (string, error) {
	executor := gitexec.NewExecutor()
	output, err := executor.Run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return output, nil
}

// getCommitDiff gets the complete diff between HEAD~1 and HEAD
// Returns map[filepath]Change with line ranges
func getCommitDiff(commitHash string) (map[string]tracker.Change, error) {
	executor := gitexec.NewExecutor()

	// HEAD~1が存在するかチェック
	_, err := executor.Run("rev-parse", "HEAD~1")
	isInitialCommit := err != nil

	var numstatOutput string
	if isInitialCommit {
		// 初回コミット: 全ファイルが新規追加
		numstatOutput, err = executor.Run("show", "--numstat", "--format=", commitHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get numstat: %w", err)
		}
	} else {
		// 通常のコミット: HEAD~1との差分を取得
		numstatOutput, err = executor.Run("diff", "--numstat", "HEAD~1", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("failed to get diff: %w", err)
		}
	}

	numstatMap, _ := git.ParseNumstat(numstatOutput)
	diffMap := make(map[string]tracker.Change, len(numstatMap))

	for filepath, stats := range numstatMap {
		added := stats[0]
		deleted := stats[1]

		lineRanges := [][]int{}
		if added > 0 {
			lineRanges = append(lineRanges, []int{1, added})
		}

		diffMap[filepath] = tracker.Change{
			Added:   added,
			Deleted: deleted,
			Lines:   lineRanges,
		}
	}

	return diffMap, nil
}

// buildAuthorshipMap builds a map of filepath -> author from checkpoints
func buildAuthorshipMap(checkpoints []*tracker.CheckpointV2, changedFiles map[string]bool) map[string]*tracker.CheckpointV2 {
	authorMap := make(map[string]*tracker.CheckpointV2)

	// 各ファイルについて、最後に変更したチェックポイントを記録
	for _, cp := range checkpoints {
		for filepath := range cp.Changes {
			// changedFilesに含まれるファイルのみ処理
			if changedFiles[filepath] {
				authorMap[filepath] = cp
			}
		}
	}

	return authorMap
}

// buildAuthorshipLogFromDiff creates Authorship Log from diff and authorship mapping
func buildAuthorshipLogFromDiff(
	diffMap map[string]tracker.Change,
	authorMap map[string]*tracker.CheckpointV2,
	commitHash string,
	changedFiles map[string]bool,
) (*tracker.AuthorshipLog, error) {
	log := &tracker.AuthorshipLog{
		Version:   authorship.AuthorshipLogVersion,
		Commit:    commitHash,
		Timestamp: time.Now(),
		Files:     make(map[string]tracker.FileInfo),
	}

	store, err := storage.NewAIctStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	cfg, err := store.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// 各変更ファイルに対してAuthorship情報を生成
	for filepath, change := range diffMap {
		// numstatフィルタリング
		if !changedFiles[filepath] {
			continue
		}

		// 追跡対象の拡張子かチェック
		if !isTrackedFile(filepath, cfg) {
			continue
		}

		// 作成者情報を取得
		var authorName string
		var authorType tracker.AuthorType
		var metadata map[string]string

		if cp, exists := authorMap[filepath]; exists {
			// チェックポイントに記録がある場合
			authorName = cp.Author
			authorType = cp.Type
			metadata = cp.Metadata
		} else {
			// チェックポイントに記録がない場合はデフォルト作成者
			authorName = cfg.DefaultAuthor
			authorType = tracker.AuthorTypeHuman
			metadata = map[string]string{"message": "No checkpoint found, assigned to default author"}
		}

		// FileInfoを作成
		fileInfo := tracker.FileInfo{
			Authors: []tracker.AuthorInfo{
				{
					Name:     authorName,
					Type:     authorType,
					Lines:    change.Lines,
					Metadata: metadata,
				},
			},
		}

		log.Files[filepath] = fileInfo
	}

	return log, nil
}

// isTrackedFile checks if a file should be tracked based on config
func isTrackedFile(filepath string, cfg *tracker.Config) bool {
	// Check tracked extensions
	for _, ext := range cfg.TrackedExtensions {
		if strings.HasSuffix(filepath, ext) {
			// Check exclude patterns
			for _, pattern := range cfg.ExcludePatterns {
				// Simple pattern matching (supports * wildcard)
				if matchesPattern(filepath, pattern) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// matchesPattern performs simple wildcard pattern matching
func matchesPattern(filepath, pattern string) bool {
	// Simple implementation: exact match or suffix match with *
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(filepath, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(filepath, pattern[:len(pattern)-1])
	}
	return filepath == pattern
}
