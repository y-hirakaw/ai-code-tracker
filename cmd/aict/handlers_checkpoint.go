package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func handleCheckpoint() {
	fs := flag.NewFlagSet("checkpoint", flag.ExitOnError)
	author := fs.String("author", "", "作成者名（デフォルト: config.default_author）")
	model := fs.String("model", "", "AIモデル名（AIエージェントの場合）")
	message := fs.String("message", "", "メモ（オプション）")
	fs.Parse(os.Args[2:])

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
			cmd := exec.Command("git", "config", "user.name")
			output, err := cmd.Output()
			if err == nil {
				authorName = strings.TrimSpace(string(output))
			} else {
				fmt.Fprintf(os.Stderr, "Error: Author name not specified and default_author not configured\n")
				fmt.Fprintf(os.Stderr, "Use --author flag or configure default_author\n")
				os.Exit(1)
			}
		}
	}

	// 作成者タイプを判定
	authorType := tracker.AuthorTypeHuman
	if isAIAgent(authorName, config.AIAgents) {
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

	if len(changes) == 0 {
		if lastCheckpoint == nil {
			fmt.Println("✓ Initial checkpoint created (baseline snapshot)")
		} else {
			fmt.Println("No changes detected since last checkpoint")
		}
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

	// Git管理下のファイル一覧を取得（追跡されているファイル + 変更されたファイル）
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list git files: %w", err)
	}

	// 拡張子マップを作成
	extMap := make(map[string]bool)
	for _, ext := range trackedExtensions {
		extMap[ext] = true
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
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

// getDetailedDiff gets detailed diff information for a file
func getDetailedDiff(filepath string) (added, deleted int, lineRanges [][]int, err error) {
	// git diff --numstat HEAD でファイルの追加・削除行数を取得
	cmd := exec.Command("git", "diff", "--numstat", "HEAD", "--", filepath)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, nil, err
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return 0, 0, [][]int{}, nil
	}

	parts := strings.Fields(line)
	if len(parts) < 3 {
		return 0, 0, nil, fmt.Errorf("invalid numstat output")
	}

	// バイナリファイルチェック
	if parts[0] == "-" || parts[1] == "-" {
		return 0, 0, nil, fmt.Errorf("binary file")
	}

	added, _ = strconv.Atoi(parts[0])
	deleted, _ = strconv.Atoi(parts[1])

	// 行範囲を取得
	lineRanges, err = getLineRanges(filepath)
	if err != nil {
		lineRanges = [][]int{}
	}

	return added, deleted, lineRanges, nil
}

// getLineRanges extracts line ranges from git diff output
func getLineRanges(filepath string) ([][]int, error) {
	cmd := exec.Command("git", "diff", "--unified=0", "HEAD", "--", filepath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ranges [][]int

	// @@ -1,2 +3,4 @@ 形式の行範囲を解析
	for _, line := range strings.Split(string(output), "\n") {
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

// isAIAgent checks if author is an AI agent
func isAIAgent(author string, aiAgents []string) bool {
	for _, agent := range aiAgents {
		if author == agent {
			return true
		}
	}
	return false
}
