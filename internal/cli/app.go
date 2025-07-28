package cli

import (
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/i18n"
	"github.com/y-hirakaw/ai-code-tracker/internal/ui"
	"github.com/y-hirakaw/ai-code-tracker/internal/utils"
)

const (
	// Version はアプリケーションのバージョン
	Version = "0.2.1"
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
		// コンテキスト情報を作成
		currentDir, _ := utils.GetCurrentDirectory()
		gitRepo := utils.IsGitRepository(currentDir)
		
		ctx := &ui.CommandContext{
			Command:    command,
			Args:       cmdArgs,
			Error:      err,
			WorkingDir: currentDir,
			GitRepo:    gitRepo,
		}
		
		// エラータイプを設定
		if friendlyErr, ok := err.(*errors.FriendlyError); ok {
			ctx.ErrorType = friendlyErr.Type
			ctx.Error = friendlyErr.WithCommand(command)
		} else {
			ctx.ErrorType = errors.ErrorTypeGeneral
			ctx.Error = errors.WrapError(err, errors.ErrorTypeGeneral, "generic_error").WithCommand(command)
		}
		
		// コンテキストアウェアなエラー表示
		a.helpSystem.ShowContextualError(ctx)
		return 1
	}

	return 0
}