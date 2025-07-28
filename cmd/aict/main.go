package main

import (
	"os"

	"github.com/y-hirakaw/ai-code-tracker/internal/cli"
)

// main はアプリケーションのエントリーポイント
func main() {
	app := cli.NewApp()
	exitCode := app.Run(os.Args)
	os.Exit(exitCode)
}