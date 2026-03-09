package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
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
		if store != nil && cfg != nil {
			_ = store.PurgeExpiredCheckpoints(cfg.GetCheckpointTTL())
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
	if err := store.RemoveConsumedCheckpoints(consumedTimestamps); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove consumed checkpoints: %v\n", err)
	}
	// 有効期限切れチェックポイントの自動消去
	if err := store.PurgeExpiredCheckpoints(cfg.GetCheckpointTTL()); err != nil {
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

// maxBlobSize はbuildParentFileHashesで処理する最大ファイルサイズ（10MB）。
// これを超えるファイル（バイナリ等）はスキップしてメモリ圧迫を防止。
const maxBlobSize = 10 * 1024 * 1024

// buildParentFileHashes はコミット親(HEAD~1)の各ファイルのSHA-256ハッシュを取得します。
// git ls-tree + git cat-file --batch で2プロセスにバッチ化（N+1問題の解消）。
// 初回コミットの場合は nil を返します。
func buildParentFileHashes(commitHash string, changedFiles map[string]bool) map[string]string {
	executor := newExecutor()

	// HEAD~1 が存在するか確認
	_, err := executor.Run("rev-parse", commitHash+"~1")
	if err != nil {
		return nil // 初回コミット: Phase 2 無効化
	}

	// Step 1: 親コミットのls-treeでblob SHA一覧を取得（1プロセス）
	fileList := make([]string, 0, len(changedFiles))
	for f := range changedFiles {
		fileList = append(fileList, f)
	}
	lsTreeArgs := append([]string{"ls-tree", commitHash + "~1", "--"}, fileList...)
	lsTreeOutput, err := executor.Run(lsTreeArgs...)
	if err != nil {
		debugf("ls-tree failed, falling back: %v", err)
		return nil
	}

	// ls-tree出力をパース: "100644 blob <sha>\t<path>"
	type blobEntry struct {
		sha  string
		path string
	}
	var entries []blobEntry
	blobToFiles := make(map[string][]string) // 同一内容の複数ファイル対応
	for _, line := range strings.Split(lsTreeOutput, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		fields := strings.Fields(parts[0])
		if len(fields) != 3 || fields[1] != "blob" {
			continue
		}
		blobSHA := fields[2]
		filePath := parts[1]
		entries = append(entries, blobEntry{sha: blobSHA, path: filePath})
		blobToFiles[blobSHA] = append(blobToFiles[blobSHA], filePath)
	}

	if len(entries) == 0 {
		return make(map[string]string)
	}

	// Step 2: cat-file --batchで全blobの内容を一括取得（1プロセス）
	// 重複blob SHAを排除
	seen := make(map[string]bool)
	var uniqueSHAs []string
	for _, e := range entries {
		if !seen[e.sha] {
			seen[e.sha] = true
			uniqueSHAs = append(uniqueSHAs, e.sha)
		}
	}

	stdinStr := strings.Join(uniqueSHAs, "\n") + "\n"
	batchOutput, err := executor.RunWithStdin(stdinStr, "cat-file", "--batch")
	if err != nil {
		debugf("cat-file --batch failed, falling back: %v", err)
		return nil
	}

	// cat-file --batch出力をパースしてSHA-256を計算
	// 形式: "<sha> <type> <size>\n<content (size bytes)>\n"
	hashes := make(map[string]string, len(entries))
	remaining := batchOutput
	for _, blobSHA := range uniqueSHAs {
		headerEnd := strings.Index(remaining, "\n")
		if headerEnd == -1 {
			break
		}
		header := remaining[:headerEnd]
		remaining = remaining[headerEnd+1:]

		headerFields := strings.Fields(header)
		if len(headerFields) < 3 || headerFields[1] == "missing" {
			continue
		}

		size, err := strconv.Atoi(headerFields[2])
		if err != nil {
			continue
		}

		// 大ファイルはスキップ
		if size > maxBlobSize {
			if len(remaining) > size {
				remaining = remaining[size+1:]
			}
			continue
		}

		if len(remaining) < size+1 {
			break
		}
		content := remaining[:size]
		remaining = remaining[size+1:] // skip trailing LF

		hash := sha256.Sum256([]byte(content))
		hashStr := hex.EncodeToString(hash[:])

		for _, filePath := range blobToFiles[blobSHA] {
			hashes[filePath] = hashStr
		}
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

