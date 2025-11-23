package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
)

func handleSync() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: aict sync [push|fetch]")
		os.Exit(1)
	}

	subcommand := os.Args[2]

	switch subcommand {
	case "push":
		handleSyncPush()
	case "fetch":
		handleSyncFetch()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		fmt.Println("Usage: aict sync [push|fetch]")
		os.Exit(1)
	}
}

func handleSyncPush() {
	// refs/aict/authorship/* をリモートにpush
	refspec := gitnotes.AuthorshipNotesRef + "/*:" + gitnotes.AuthorshipNotesRef + "/*"
	cmd := exec.Command("git", "push", "origin", refspec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error pushing authorship logs: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Println("✓ Authorship logs pushed to remote")
}

func handleSyncFetch() {
	// リモートから refs/aict/authorship/* をfetch
	refspec := gitnotes.AuthorshipNotesRef + "/*:" + gitnotes.AuthorshipNotesRef + "/*"
	cmd := exec.Command("git", "fetch", "origin", refspec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching authorship logs: %v\n%s\n", err, output)
		os.Exit(1)
	}

	fmt.Println("✓ Authorship logs fetched from remote")
}
