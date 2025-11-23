package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/authorship"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
)

func handleCommit() {
	// ストレージを初期化
	store, err := storage.NewAIctStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// チェックポイントを読み込み
	checkpoints, err := store.LoadCheckpoints()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading checkpoints: %v\n", err)
		os.Exit(1)
	}

	if len(checkpoints) == 0 {
		// チェックポイントがない場合は何もしない
		return
	}

	// 最新のコミットハッシュを取得
	commitHash, err := getLatestCommitHash()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting commit hash: %v\n", err)
		os.Exit(1)
	}

	// チェックポイント群をAuthorship Logに変換
	log, err := authorship.BuildAuthorshipLog(checkpoints, commitHash)
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
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
