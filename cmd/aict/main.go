package main

import (
	"os"

	"github.com/ai-code-tracker/aict/internal/cli"
)

// main はアプリケーションのエントリーポイント
func main() {
	app := cli.NewApp()
	exitCode := app.Run(os.Args)
	os.Exit(exitCode)
}