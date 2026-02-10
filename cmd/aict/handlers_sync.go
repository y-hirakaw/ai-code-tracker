package main

import (
	"fmt"
	"os"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
)

func handleSync() error {
	if len(os.Args) < 3 {
		fmt.Println("Usage: aict sync [push|fetch]")
		return fmt.Errorf("sync subcommand required")
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "push":
		return handleSyncPush()
	case "fetch":
		return handleSyncFetch()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		fmt.Println("Usage: aict sync [push|fetch]")
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func handleSyncPush() error {
	// refs/aict/authorship/* をリモートにpush
	refspec := gitnotes.AuthorshipNotesRef + "/*:" + gitnotes.AuthorshipNotesRef + "/*"
	executor := gitexec.NewExecutor()
	_, err := executor.Run("push", "origin", refspec)
	if err != nil {
		return fmt.Errorf("pushing authorship logs: %w", err)
	}

	fmt.Println("✓ Authorship logs pushed to remote")
	return nil
}

func handleSyncFetch() error {
	// リモートから refs/aict/authorship/* をfetch
	refspec := gitnotes.AuthorshipNotesRef + "/*:" + gitnotes.AuthorshipNotesRef + "/*"
	executor := gitexec.NewExecutor()
	_, err := executor.Run("fetch", "origin", refspec)
	if err != nil {
		return fmt.Errorf("fetching authorship logs: %w", err)
	}

	fmt.Println("✓ Authorship logs fetched from remote")
	return nil
}
