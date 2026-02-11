package main

import (
	"fmt"
	"os"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/git"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func handleCommit() error {
	// ストレージと設定を読み込み
	store, cfg, err := loadStorageAndConfig()
	if err != nil {
		return err
	}

	// 最新のコミットハッシュを取得
	commitHash, err := getLatestCommitHash()
	if err != nil {
		return fmt.Errorf("getting commit hash: %w", err)
	}

	// コミットのnumstatを取得
	executor := newExecutor()
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
		return nil
	}

	// チェックポイントを読み込み
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		return fmt.Errorf("loading checkpoints: %w", err)
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
		return fmt.Errorf("getting commit diff: %w", err)
	}

	// チェックポイントから作成者マッピングを構築
	authorshipMap := authorship.BuildAuthorshipMap(checkpoints, changedFiles)

	// デバッグ: 作成者マッピングを出力
	debugf("Authorship mapping for %d files:", len(authorshipMap))
	for filepath, cp := range authorshipMap {
		debugf("  %s -> %s (%s)", filepath, cp.Author, cp.Type)
	}

	// 完全な差分情報と作成者情報を統合してAuthorship Logを生成
	log, err := authorship.BuildAuthorshipLogFromDiff(fullDiff, authorshipMap, commitHash, changedFiles, cfg)
	if err != nil {
		return fmt.Errorf("building authorship log: %w", err)
	}

	// バリデーション
	if err := authorship.ValidateAuthorshipLog(log); err != nil {
		return fmt.Errorf("validating authorship log: %w", err)
	}

	// Git notesに保存
	nm := gitnotes.NewNotesManager()
	if err := nm.AddAuthorshipLog(log); err != nil {
		return fmt.Errorf("saving authorship log: %w", err)
	}

	// チェックポイントをクリア
	if err := store.ClearCheckpoints(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to clear checkpoints: %v\n", err)
	}

	fmt.Println("✓ Authorship log created")
	return nil
}

// getLatestCommitHash は最新のコミットハッシュを取得します
func getLatestCommitHash() (string, error) {
	executor := newExecutor()
	output, err := executor.Run("rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return output, nil
}

// getCommitDiff はHEAD~1とHEADの間の完全なdiffを取得します
// 戻り値: map[filepath]Change (行範囲付き)
func getCommitDiff(commitHash string) (map[string]tracker.Change, error) {
	executor := newExecutor()

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

	for fpath, stats := range numstatMap {
		added := stats[0]
		deleted := stats[1]

		lineRanges := [][]int{}
		if added > 0 {
			lineRanges = append(lineRanges, []int{1, added})
		}

		diffMap[fpath] = tracker.Change{
			Added:   added,
			Deleted: deleted,
			Lines:   lineRanges,
		}
	}

	return diffMap, nil
}
