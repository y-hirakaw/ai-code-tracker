package main

import (
	"fmt"
	"os"
)

const version = "1.4.2"

// exitFunc is used to mock os.Exit in tests
var exitFunc = os.Exit

// debugEnabled controls debug output via AICT_DEBUG environment variable
var debugEnabled = os.Getenv("AICT_DEBUG") != ""

// debugf prints debug messages to stderr when AICT_DEBUG is set
func debugf(format string, args ...interface{}) {
	if debugEnabled {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		exitFunc(1)
	}

	command := os.Args[1]

	var err error
	switch command {
	case "init":
		err = handleInitV2()
	case "checkpoint":
		err = handleCheckpoint()
	case "commit":
		err = handleCommit()
	case "report":
		err = handleRangeReport()
	case "sync":
		err = handleSync()
	case "setup-hooks":
		err = handleSetupHooksV2()
	case "debug":
		err = handleDebug()
	case "version", "--version", "-v":
		fmt.Printf("AI Code Tracker (aict) version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		exitFunc(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		exitFunc(1)
	}
}





func printUsage() {
	fmt.Printf("AI Code Tracker (aict) v%s - Track AI vs Human code contributions\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  aict init                    Initialize tracking (.git/aict/ directory)")
	fmt.Println("  aict checkpoint [options]    Record development checkpoint")
	fmt.Println("    --author <name>            Author name (required)")
	fmt.Println("    --model <model>            AI model name (for AI agents)")
	fmt.Println("    --message <msg>            Optional message")
	fmt.Println("  aict commit                  Generate Authorship Log from checkpoints")
	fmt.Println("  aict report [options]        Show code generation statistics")
	fmt.Println("    --range <range>            Commit range (e.g., 'origin/main..HEAD')")
	fmt.Println("    --since <date>             Show commits since date (e.g., '7d', '2w', '1m')")
	fmt.Println("    --format <format>          Output format: table or json (default: table)")
	fmt.Println("  aict sync [push|fetch]       Sync authorship logs with remote")
	fmt.Println("  aict setup-hooks             Setup Claude Code and Git hooks")
	fmt.Println("  aict debug [show|clean|clear-notes]  Debug and cleanup commands")
	fmt.Println("    show                       Display all checkpoint details")
	fmt.Println("    clean                      Remove all checkpoint data")
	fmt.Println("    clear-notes                Remove all Git notes (authorship logs)")
	fmt.Println("  aict version                 Show version information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  aict init")
	fmt.Println("  aict setup-hooks")
	fmt.Println("  aict checkpoint --author \"Developer\"")
	fmt.Println("  aict report --range origin/main..HEAD")
	fmt.Println("  aict report --since 7d        # 7 days ago")
	fmt.Println("  aict report --since 2w        # 2 weeks ago")
	fmt.Println("  aict report --since yesterday")
	fmt.Println("  aict sync push")
	fmt.Println("  aict debug show               # Show checkpoint details")
	fmt.Println("  aict debug clean              # Clean checkpoints")
	fmt.Println("  aict debug clear-notes        # Clear Git notes")
}

func getGitUserName() string {
	executor := newExecutor()
	output, err := executor.Run("config", "user.name")
	if err != nil {
		return ""
	}
	return output
}

