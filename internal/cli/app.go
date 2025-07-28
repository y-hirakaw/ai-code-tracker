package cli

import (
	"fmt"
	"os"

	"github.com/ai-code-tracker/aict/internal/errors"
	"github.com/ai-code-tracker/aict/internal/i18n"
	"github.com/ai-code-tracker/aict/internal/ui"
)

const (
	// Version はアプリケーションのバージョン
	Version = "0.1.0"
	// AppName はアプリケーション名
	AppName = "aict"
)

// App はCLIアプリケーションを表す
type App struct {
	helpSystem     *ui.HelpSystem
	commandHandler *CommandHandler
}

// NewApp は新しいCLIアプリケーションを作成する
func NewApp() *App {
	// i18nシステムを初期化
	i18n.Initialize()
	
	// エラーフォーマッターを初期化
	errors.InitializeFormatter()
	
	// ヘルプシステムを初期化
	helpSystem := ui.NewHelpSystem(AppName, Version)
	
	// コマンドハンドラーを初期化
	commandHandler := NewCommandHandler(helpSystem)
	
	return &App{
		helpSystem:     helpSystem,
		commandHandler: commandHandler,
	}
}

// Run はCLIアプリケーションを実行する
func (a *App) Run(args []string) int {
	if len(args) < 2 {
		a.helpSystem.ShowMainHelp()
		return 1
	}

	command := args[1]
	cmdArgs := args[2:]

	// コマンドを実行
	if err := a.commandHandler.Execute(command, cmdArgs); err != nil {
		// エラーが既にFriendlyErrorの場合はそのまま使用、そうでなければラップ
		if friendlyErr, ok := err.(*errors.FriendlyError); ok {
			fmt.Fprint(os.Stderr, errors.FormatError(friendlyErr.WithCommand(command)))
		} else {
			friendlyErr := errors.WrapError(err, errors.ErrorTypeGeneral, "generic_error").WithCommand(command)
			fmt.Fprint(os.Stderr, errors.FormatError(friendlyErr))
		}
		return 1
	}

	return 0
}