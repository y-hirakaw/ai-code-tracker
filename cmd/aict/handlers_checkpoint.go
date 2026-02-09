package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func handleCheckpoint() {
	fs := flag.NewFlagSet("checkpoint", flag.ExitOnError)
	author := fs.String("author", "", "作成者名（デフォルト: config.default_author）")
	model := fs.String("model", "", "AIモデル名（AIエージェントの場合）")
	message := fs.String("message", "", "メモ（オプション）")
	fs.Parse(os.Args[2:])

	// Gitリポジトリのルートディレクトリに移動
	executor := gitexec.NewExecutor()
	repoRoot, err := executor.Run("rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Not in a git repository\n")
		os.Exit(1)
	}
	if err := os.Chdir(repoRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to change directory to %s: %v\n", repoRoot, err)
		os.Exit(1)
	}

	// ストレージを初期化
	store, err := storage.NewAIctStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 設定を読み込み
	config, err := store.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run 'aict init' first\n")
		os.Exit(1)
	}

	// 作成者名を決定
	authorName := *author
	if authorName == "" {
		if config.DefaultAuthor != "" {
			authorName = config.DefaultAuthor
		} else {
			// git config から取得を試みる
			output, err := executor.Run("config", "user.name")
			if err == nil {
				authorName = output
			} else {
				fmt.Fprintf(os.Stderr, "Error: Author name not specified and default_author not configured\n")
				fmt.Fprintf(os.Stderr, "Use --author flag or configure default_author\n")
				os.Exit(1)
			}
		}
	}

	// 作成者タイプを判定
	authorType := tracker.AuthorTypeHuman
	if tracker.IsAIAgent(authorName, config.AIAgents, config.AuthorMappings) {
		authorType = tracker.AuthorTypeAI
	}

	// 前回のチェックポイントを読み込む
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checkpoints: %v\n", err)
		os.Exit(1)
	}

	var lastCheckpoint *tracker.CheckpointV2
	if len(checkpoints) > 0 {
		lastCheckpoint = checkpoints[len(checkpoints)-1]
	}

	// 現在のスナップショットを作成
	currentSnapshot, err := captureSnapshot(config.TrackedExtensions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error capturing snapshot: %v\n", err)
		os.Exit(1)
	}

	// 前回のチェックポイントとの差分を検出
	changes, err := detectChangesFromSnapshot(lastCheckpoint, currentSnapshot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting changes: %v\n", err)
		os.Exit(1)
	}

	// 変更がない場合でもチェックポイントを記録（初回やbaseline）
	if len(changes) == 0 {
		if lastCheckpoint == nil {
			// 初回チェックポイント: 前回コミットから差分なし = baseline
			fmt.Fprintf(os.Stderr, "[DEBUG] Initial checkpoint: author=%s, files=0\n", authorName)
			fmt.Println("✓ Initial checkpoint created (baseline, no changes since last commit)")
		} else {
			// 2回目以降: 前回チェックポイントから差分なし
			fmt.Fprintf(os.Stderr, "[DEBUG] Checkpoint: author=%s, files=0 (no changes)\n", authorName)
			fmt.Println("✓ Checkpoint created (no changes since last checkpoint)")
		}
	} else {
		// 変更がある場合
		fmt.Fprintf(os.Stderr, "[DEBUG] Checkpoint: author=%s, files=%d, changes=%v\n", authorName, len(changes), getFileList(changes))
	}

	// チェックポイントを作成
	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    authorName,
		Type:      authorType,
		Metadata:  make(map[string]string),
		Changes:   changes,
		Snapshot:  currentSnapshot,
	}

	// メタデータを追加
	if *model != "" {
		checkpoint.Metadata["model"] = *model
	}
	if *message != "" {
		checkpoint.Metadata["message"] = *message
	}

	// チェックポイントを保存
	if err := store.SaveCheckpoint(checkpoint); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving checkpoint: %v\n", err)
		os.Exit(1)
	}

	// 変更行数をカウント
	totalAdded := 0
	totalFiles := 0
	for _, change := range changes {
		totalAdded += change.Added
		totalFiles++
	}

	fmt.Printf("✓ Checkpoint created (%s, %d files, %d lines added)\n", authorName, totalFiles, totalAdded)
}

// captureSnapshot creates a snapshot of all tracked files in working directory
func captureSnapshot(trackedExtensions []string) (map[string]tracker.FileSnapshot, error) {
	snapshot := make(map[string]tracker.FileSnapshot)

	// Git管理下のファイル一覧を取得（追跡されているファイル + 未追跡の新規ファイル）
	executor := gitexec.NewExecutor()
	output, err := executor.Run("ls-files", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("failed to list git files: %w", err)
	}

	// 拡張子マップを作成
	extMap := make(map[string]bool)
	for _, ext := range trackedExtensions {
		extMap[ext] = true
	}

	files := strings.Split(output, "\n")
	for _, filepath := range files {
		if filepath == "" {
			continue
		}

		// 拡張子チェック
		ext := ""
		if idx := strings.LastIndex(filepath, "."); idx != -1 {
			ext = filepath[idx:]
		}
		if !extMap[ext] {
			continue
		}

		// 作業ディレクトリのファイル内容を読み込み（コミット済みでなくても良い）
		content, err := os.ReadFile(filepath)
		if err != nil {
			continue // ファイルが読めない場合はスキップ
		}

		// ハッシュ計算
		hash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(hash[:])

		// 行数カウント
		lines := len(strings.Split(string(content), "\n"))

		snapshot[filepath] = tracker.FileSnapshot{
			Hash:  hashStr,
			Lines: lines,
		}
	}

	return snapshot, nil
}

// detectChangesFromSnapshot detects changes between two snapshots
func detectChangesFromSnapshot(lastCheckpoint *tracker.CheckpointV2, currentSnapshot map[string]tracker.FileSnapshot) (map[string]tracker.Change, error) {
	changes := make(map[string]tracker.Change)

	// 初回チェックポイントの場合は変更なし
	if lastCheckpoint == nil {
		return changes, nil
	}

	lastSnapshot := lastCheckpoint.Snapshot

	// 変更・追加されたファイルを検出
	for filepath, currentFile := range currentSnapshot {
		lastFile, existed := lastSnapshot[filepath]

		if !existed {
			// 新規ファイル
			changes[filepath] = tracker.Change{
				Added:   currentFile.Lines,
				Deleted: 0,
				Lines:   [][]int{{1, currentFile.Lines}},
			}
		} else if currentFile.Hash != lastFile.Hash {
			// ファイルが変更された場合、git diffで詳細を取得
			added, deleted, lineRanges, err := getDetailedDiff(filepath)
			if err != nil {
				// エラーがある場合は簡易的に行数の差分で計算
				if currentFile.Lines > lastFile.Lines {
					changes[filepath] = tracker.Change{
						Added:   currentFile.Lines - lastFile.Lines,
						Deleted: 0,
						Lines:   [][]int{},
					}
				} else if currentFile.Lines < lastFile.Lines {
					changes[filepath] = tracker.Change{
						Added:   0,
						Deleted: lastFile.Lines - currentFile.Lines,
						Lines:   [][]int{},
					}
				}
			} else {
				changes[filepath] = tracker.Change{
					Added:   added,
					Deleted: deleted,
					Lines:   lineRanges,
				}
			}
		}
	}

	// 削除されたファイルを検出
	for filepath, lastFile := range lastSnapshot {
		if _, exists := currentSnapshot[filepath]; !exists {
			changes[filepath] = tracker.Change{
				Added:   0,
				Deleted: lastFile.Lines,
				Lines:   [][]int{},
			}
		}
	}

	return changes, nil
}

// getDetailedDiff gets detailed diff information for a file by comparing file content directly
func getDetailedDiff(filepath string) (added, deleted int, lineRanges [][]int, err error) {
	// 作業ディレクトリの現在のファイル内容を取得
	currentContent, err := os.ReadFile(filepath)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to read current file: %w", err)
	}

	// HEADのファイル内容を取得（git show HEAD:filepath）
	executor := gitexec.NewExecutor()
	headContentStr, err := executor.Run("show", fmt.Sprintf("HEAD:%s", filepath))
	if err != nil {
		// HEADに存在しない（新規ファイル）の場合
		currentLines := strings.Split(strings.TrimSpace(string(currentContent)), "\n")
		lineCount := len(currentLines)
		return lineCount, 0, [][]int{{1, lineCount}}, nil
	}

	// 両方の内容を行単位で比較
	currentLines := strings.Split(strings.TrimSpace(string(currentContent)), "\n")
	headLines := strings.Split(headContentStr, "\n")

	// 簡易的なdiff計算（追加・削除行数）
	currentLineCount := len(currentLines)
	headLineCount := len(headLines)

	if currentLineCount > headLineCount {
		added = currentLineCount - headLineCount
		deleted = 0
	} else if currentLineCount < headLineCount {
		added = 0
		deleted = headLineCount - currentLineCount
	} else {
		// 行数が同じでも内容が変更されている可能性があるため、
		// 変更された行をカウント
		changedLines := 0
		for i := 0; i < currentLineCount && i < headLineCount; i++ {
			if currentLines[i] != headLines[i] {
				changedLines++
			}
		}
		if changedLines > 0 {
			// 簡易的に変更行数を追加として扱う
			added = changedLines
			deleted = 0
		}
	}

	// 行範囲を取得（git diffを使用）
	lineRanges, err = getLineRangesFromDiff(filepath)
	if err != nil {
		// エラー時は簡易的な範囲を返す
		if added > 0 {
			lineRanges = [][]int{{1, currentLineCount}}
		} else {
			lineRanges = [][]int{}
		}
	}

	return added, deleted, lineRanges, nil
}

// getLineRangesFromDiff extracts line ranges using git diff
func getLineRangesFromDiff(filepath string) ([][]int, error) {
	executor := gitexec.NewExecutor()
	output, err := executor.Run("diff", "--unified=0", "HEAD", "--", filepath)
	if err != nil {
		return nil, err
	}

	var ranges [][]int

	// @@ -1,2 +3,4 @@ 形式の行範囲を解析
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "@@") {
			continue
		}

		// +3,4 の部分を抽出
		parts := strings.Split(line, "@@")
		if len(parts) < 2 {
			continue
		}

		rangePart := strings.TrimSpace(parts[1])
		plusIdx := strings.Index(rangePart, "+")
		if plusIdx == -1 {
			continue
		}

		rangeStr := strings.Fields(rangePart[plusIdx+1:])[0]
		rangeNums := strings.Split(rangeStr, ",")

		if len(rangeNums) == 1 {
			// 単一行: +10
			lineNum, err := strconv.Atoi(rangeNums[0])
			if err == nil && lineNum > 0 {
				ranges = append(ranges, []int{lineNum})
			}
		} else if len(rangeNums) == 2 {
			// 範囲: +10,5 (10行目から5行)
			start, err1 := strconv.Atoi(rangeNums[0])
			count, err2 := strconv.Atoi(rangeNums[1])
			if err1 == nil && err2 == nil && start > 0 && count > 0 {
				ranges = append(ranges, []int{start, start + count - 1})
			}
		}
	}

	return ranges, nil
}


// getFileList returns a list of filenames from changes map
func getFileList(changes map[string]tracker.Change) []string {
	files := make([]string, 0, len(changes))
	for filepath := range changes {
		files = append(files, filepath)
	}
	return files
}
