package main

import (
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

	// 前回のチェックポイント以降の変更を検出
	changes, err := detectChanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting changes: %v\n", err)
		os.Exit(1)
	}

	if len(changes) == 0 {
		fmt.Println("No changes detected")
		return
	}

	// チェックポイントを作成
	checkpoint := &tracker.CheckpointV2{
		Timestamp: time.Now(),
		Author:    authorName,
		Type:      authorType,
		Metadata:  make(map[string]string),
		Changes:   changes,
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

// detectChanges detects file changes since last checkpoint
func detectChanges() (map[string]tracker.Change, error) {
	// git diff --unified=0 --numstat HEAD で変更を取得
	cmd := exec.Command("git", "diff", "--unified=0", "--numstat", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	changes := make(map[string]tracker.Change)

	// 各ファイルの変更を解析
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		addedStr := parts[0]
		deletedStr := parts[1]
		filepath := parts[2]

		// バイナリファイルはスキップ
		if addedStr == "-" || deletedStr == "-" {
			continue
		}

		added, _ := strconv.Atoi(addedStr)
		deleted, _ := strconv.Atoi(deletedStr)

		// 行範囲を取得
		lineRanges, err := getLineRanges(filepath)
		if err != nil {
			// エラーがあっても継続（行範囲なしで記録）
			lineRanges = [][]int{}
		}

		changes[filepath] = tracker.Change{
			Added:   added,
			Deleted: deleted,
			Lines:   lineRanges,
		}
	}

	return changes, nil
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
