package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// handleDebug handles the debug command
func handleDebug() error {
	if len(os.Args) < 3 {
		fmt.Println("エラー: debug サブコマンドを指定してください (show, clean, または clear-notes)")
		fmt.Println("使用法:")
		fmt.Println("  aict debug show              # チェックポイント情報を表示")
		fmt.Println("  aict debug clean             # チェックポイントを削除")
		fmt.Println("  aict debug clear-notes       # Git notesのAuthorship Logをクリア")
		return fmt.Errorf("debug subcommand required")
	}

	aictDir := filepath.Join(".git", storage.AictDirName)

	subcommand := os.Args[2]
	switch subcommand {
	case "show":
		return handleDebugShow(aictDir)
	case "clean":
		return handleDebugClean(aictDir)
	case "clear-notes":
		return handleDebugClearNotes()
	default:
		fmt.Printf("エラー: 不明なサブコマンド '%s'\n", subcommand)
		fmt.Println("利用可能なサブコマンド: show, clean, clear-notes")
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

// handleDebugShow displays detailed checkpoint information for debugging
func handleDebugShow(aictDir string) error {
	checkpointsFile := filepath.Join(aictDir, storage.CheckpointsDirName, storage.LatestFileName)

	// Check if checkpoints file exists
	if _, err := os.Stat(checkpointsFile); os.IsNotExist(err) {
		fmt.Println("チェックポイントファイルが存在しません")
		return nil
	}

	// Read checkpoint file
	data, err := os.ReadFile(checkpointsFile)
	if err != nil {
		return fmt.Errorf("チェックポイントファイルの読み込みに失敗しました: %w", err)
	}

	// Parse checkpoint array
	var checkpoints []*tracker.CheckpointV2
	if err := json.Unmarshal(data, &checkpoints); err != nil {
		return fmt.Errorf("チェックポイントのパースに失敗しました: %w", err)
	}

	if len(checkpoints) == 0 {
		fmt.Println("保存されているチェックポイントはありません")
		return nil
	}

	fmt.Printf("=== チェックポイント情報 (%d件) ===\n\n", len(checkpoints))
	fmt.Printf("ファイル: %s\n\n", checkpointsFile)

	// Display each checkpoint
	for i, cp := range checkpoints {
		displayCheckpoint(i+1, cp)
	}

	return nil
}

// displayCheckpoint formats and displays a single checkpoint
func displayCheckpoint(index int, cp *tracker.CheckpointV2) {
	timestamp := cp.Timestamp.Format("2006-01-02 15:04:05")

	fmt.Printf("[%d] チェックポイント\n", index)
	fmt.Printf("  タイムスタンプ: %s\n", timestamp)
	fmt.Printf("  作成者: %s\n", cp.Author)
	fmt.Printf("  種別: %s\n", cp.Type)

	if len(cp.Metadata) > 0 {
		fmt.Println("  メタデータ:")
		for key, value := range cp.Metadata {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}

	fmt.Printf("  変更ファイル数: %d\n", len(cp.Changes))

	if len(cp.Changes) > 0 {
		fmt.Println("  変更内容:")
		for filePath, change := range cp.Changes {
			fmt.Printf("    %s: +%d -%d", filePath, change.Added, change.Deleted)
			if len(change.Lines) > 0 {
				fmt.Printf(" (行範囲: %d個)", len(change.Lines))
			}
			fmt.Println()
		}
	}

	fmt.Println()
}

// handleDebugClean removes all checkpoint data
func handleDebugClean(aictDir string) error {
	checkpointsFile := filepath.Join(aictDir, storage.CheckpointsDirName, storage.LatestFileName)

	// Check if checkpoints file exists
	if _, err := os.Stat(checkpointsFile); os.IsNotExist(err) {
		fmt.Println("チェックポイントファイルが存在しません")
		return nil
	}

	// Count checkpoints
	data, err := os.ReadFile(checkpointsFile)
	if err != nil {
		return fmt.Errorf("チェックポイントファイルの読み込みに失敗しました: %w", err)
	}

	var checkpoints []*tracker.CheckpointV2
	if err := json.Unmarshal(data, &checkpoints); err != nil {
		return fmt.Errorf("チェックポイントのパースに失敗しました: %w", err)
	}

	checkpointCount := len(checkpoints)
	if checkpointCount == 0 {
		fmt.Println("削除するチェックポイントはありません")
		return nil
	}

	// Remove checkpoint file
	if err := os.Remove(checkpointsFile); err != nil {
		return fmt.Errorf("チェックポイントファイルの削除に失敗しました: %w", err)
	}

	fmt.Printf("✅ %d件のチェックポイントを削除しました\n", checkpointCount)
	return nil
}

// handleDebugClearNotes removes all Git notes for authorship tracking
func handleDebugClearNotes() error {
	// Get all aict-related refs
	executor := gitexec.NewExecutor()
	output, err := executor.Run("show-ref")
	if err != nil {
		return fmt.Errorf("Git refsの取得に失敗しました: %w", err)
	}

	var aictRefs []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "aict") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				aictRefs = append(aictRefs, parts[1])
			}
		}
	}

	if len(aictRefs) == 0 {
		fmt.Println("AICT関連のGit notesが存在しません")
		return nil
	}

	fmt.Printf("以下のGit notesを削除します:\n")
	for _, ref := range aictRefs {
		fmt.Printf("  - %s\n", ref)
	}

	// Remove each ref
	removed := 0
	for _, ref := range aictRefs {
		executor := gitexec.NewExecutor()
		_, err := executor.Run("update-ref", "-d", ref)
		if err != nil {
			fmt.Printf("警告: %s の削除に失敗しました: %v\n", ref, err)
		} else {
			removed++
		}
	}

	fmt.Printf("\n✅ %d個のGit notesを削除しました\n", removed)
	if removed > 0 {
		fmt.Println("リモートから削除するには:")
		for _, ref := range aictRefs {
			fmt.Printf("  git push origin :%s\n", ref)
		}
	}
	return nil
}
