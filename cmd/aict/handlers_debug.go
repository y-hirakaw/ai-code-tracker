package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// handleDebug handles the debug command
func handleDebug() error {
	if len(os.Args) < 3 {
		fmt.Println("使用法:")
		fmt.Println("  aict debug show              # チェックポイント情報を表示")
		fmt.Println("  aict debug clean             # チェックポイントを削除")
		fmt.Println("  aict debug clear-notes       # Git notesのAuthorship Logをクリア")
		return fmt.Errorf("debug subcommand required (show, clean, clear-notes)")
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "show":
		return handleDebugShow()
	case "clean":
		return handleDebugClean()
	case "clear-notes":
		return handleDebugClearNotes()
	default:
		return fmt.Errorf("unknown debug subcommand: %s (available: show, clean, clear-notes)", subcommand)
	}
}

// handleDebugShow displays detailed checkpoint information for debugging
func handleDebugShow() error {
	store, err := storage.NewAIctStorage()
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		return fmt.Errorf("チェックポイントの読み込みに失敗しました: %w", err)
	}

	if len(checkpoints) == 0 {
		fmt.Println("保存されているチェックポイントはありません")
		return nil
	}

	fmt.Printf("=== チェックポイント情報 (%d件) ===\n\n", len(checkpoints))

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
func handleDebugClean() error {
	store, err := storage.NewAIctStorage()
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		return fmt.Errorf("チェックポイントの読み込みに失敗しました: %w", err)
	}

	if len(checkpoints) == 0 {
		fmt.Println("削除するチェックポイントはありません")
		return nil
	}

	if err := store.ClearCheckpoints(); err != nil {
		return fmt.Errorf("チェックポイントファイルの削除に失敗しました: %w", err)
	}

	fmt.Printf("✅ %d件のチェックポイントを削除しました\n", len(checkpoints))
	return nil
}

// handleDebugClearNotes removes all Git notes for authorship tracking
func handleDebugClearNotes() error {
	// Get all aict-related refs
	executor := newExecutor()
	output, err := executor.Run("show-ref")
	if err != nil {
		return fmt.Errorf("Git refsの取得に失敗しました: %w", err)
	}

	var aictRefs []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		// aict専用のnotes/refsのみ対象（ブランチ名に"aict"を含むものを誤削除しない）
		// refs/aict/* : 直接refとして作成された場合
		// refs/notes/aict* : git notes --ref=aict で作成された場合
		// refs/notes/refs/aict/* : git notes --ref=refs/aict/... で作成された場合
		if strings.HasPrefix(ref, "refs/aict/") ||
			strings.HasPrefix(ref, "refs/notes/aict") ||
			strings.HasPrefix(ref, "refs/notes/refs/aict/") {
			aictRefs = append(aictRefs, ref)
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
		executor := newExecutor()
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
