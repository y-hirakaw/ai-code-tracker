package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

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
		// TTL超過チェックポイントのみ消去（stash保全のため全削除はしない）
		if store != nil {
			_ = store.PurgeExpiredCheckpoints()
		}
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

	// コミット親のファイルハッシュを取得（Phase 2 照合用）
	parentSnapshot := buildParentFileHashes(commitHash, changedFiles)

	// チェックポイントから作成者マッピングを構築
	authorshipMap := authorship.BuildAuthorshipMap(checkpoints, changedFiles, parentSnapshot)

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

	// 使用済みチェックポイントのみ選択的に削除（stash対応）
	consumedTimestamps := collectConsumedTimestamps(authorshipMap)
	// 同じBaseCommitを共有するチェックポイントもペアで消費
	// （例: Developer baseline + AI editのペア）
	expandConsumedByBaseCommit(checkpoints, consumedTimestamps)
	if err := store.RemoveConsumedCheckpoints(consumedTimestamps); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove consumed checkpoints: %v\n", err)
	}
	// 有効期限切れチェックポイントの自動消去
	if err := store.PurgeExpiredCheckpoints(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to purge expired checkpoints: %v\n", err)
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

// buildParentFileHashes はコミット親(HEAD~1)の各ファイルのSHA-256ハッシュを取得します。
// Snapshot ベースの照合（Phase 2）でのみ使用。
// 初回コミットの場合は nil を返します。
func buildParentFileHashes(commitHash string, changedFiles map[string]bool) map[string]string {
	executor := newExecutor()

	// HEAD~1 が存在するか確認
	_, err := executor.Run("rev-parse", commitHash+"~1")
	if err != nil {
		return nil // 初回コミット: Phase 2 無効化
	}

	hashes := make(map[string]string, len(changedFiles))
	for fpath := range changedFiles {
		content, err := executor.Run("show", fmt.Sprintf("%s~1:%s", commitHash, fpath))
		if err != nil {
			continue // 親コミットに存在しないファイル（新規追加）
		}
		hash := sha256.Sum256([]byte(content))
		hashes[fpath] = hex.EncodeToString(hash[:])
	}
	return hashes
}

// collectConsumedTimestamps は authorshipMap で使用されたチェックポイントの
// Timestamp 集合を返します。
func collectConsumedTimestamps(authorMap map[string]*tracker.CheckpointV2) map[time.Time]bool {
	timestamps := make(map[time.Time]bool)
	for _, cp := range authorMap {
		timestamps[cp.Timestamp] = true
	}
	return timestamps
}

// expandConsumedByBaseCommit は消費対象のチェックポイントと同じBaseCommitを
// 共有するチェックポイントも消費対象に追加します。
// これにより、Developer baseline + AI editのペアが一緒に消費されます。
func expandConsumedByBaseCommit(checkpoints []*tracker.CheckpointV2, consumed map[time.Time]bool) {
	// 消費されたチェックポイントのBaseCommit集合を収集
	// 空文字列も有効なグループ（初回コミット前のチェックポイント）
	consumedBases := make(map[string]bool)
	for _, cp := range checkpoints {
		if consumed[cp.Timestamp] {
			consumedBases[cp.BaseCommit] = true
		}
	}
	if len(consumedBases) == 0 {
		return
	}
	// 同じBaseCommitのチェックポイントを消費対象に追加
	for _, cp := range checkpoints {
		if consumedBases[cp.BaseCommit] {
			consumed[cp.Timestamp] = true
		}
	}
}
