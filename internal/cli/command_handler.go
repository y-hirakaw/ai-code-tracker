package cli

import (
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/ui"
)

// Command はCLIコマンドを定義する
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// CommandHandler はコマンドの管理と実行を行う
type CommandHandler struct {
	commands   map[string]Command
	helpSystem *ui.HelpSystem
	
	// 各コマンドハンドラー
	initHandler    *InitHandler
	trackHandler   *TrackHandler
	statsHandler   *StatsHandler
	blameHandler   *BlameHandler
	setupHandler   *SetupHandler
	configHandler  *ConfigHandler
	wizardHandler  *WizardHandler
	langHandler    *LangHandler
	webHandler     *WebHandler
	periodHandler  *PeriodHandler
	versionHandler *VersionHandler
	helpHandler    *HelpHandler
}

// NewCommandHandler は新しいコマンドハンドラーを作成する
func NewCommandHandler(helpSystem *ui.HelpSystem) *CommandHandler {
	ch := &CommandHandler{
		helpSystem: helpSystem,
		commands:   make(map[string]Command),
	}
	
	// 各ハンドラーを初期化
	ch.initHandler = NewInitHandler()
	ch.trackHandler = NewTrackHandler()
	ch.statsHandler = NewStatsHandler()
	ch.blameHandler = NewBlameHandler()
	ch.setupHandler = NewSetupHandler()
	ch.configHandler = NewConfigHandler()
	ch.wizardHandler = NewWizardHandler(helpSystem)
	ch.langHandler = NewLangHandler()
	ch.webHandler = NewWebHandler()
	ch.periodHandler = NewPeriodHandler()
	ch.versionHandler = NewVersionHandler()
	ch.helpHandler = NewHelpHandler(helpSystem)
	
	// コマンドを登録
	ch.registerCommands()
	
	return ch
}

// registerCommands はコマンドを登録する
func (ch *CommandHandler) registerCommands() {
	ch.commands = map[string]Command{
		"track": {
			Name:        "track",
			Description: "ファイルの変更を手動でトラッキングする",
			Handler:     ch.trackHandler.Handle,
		},
		"init": {
			Name:        "init",
			Description: "プロジェクトでAI Code Trackerを初期化する",
			Handler:     ch.initHandler.Handle,
		},
		"stats": {
			Name:        "stats",
			Description: "統計情報を表示する",
			Handler:     ch.statsHandler.Handle,
		},
		"blame": {
			Name:        "blame",
			Description: "ファイルのAI/人間による変更履歴を表示する",
			Handler:     ch.blameHandler.Handle,
		},
		"config": {
			Name:        "config",
			Description: "設定を管理する",
			Handler:     ch.configHandler.Handle,
		},
		"setup": {
			Name:        "setup",
			Description: "Git hooks と Claude Code hooks を自動設定する",
			Handler:     ch.setupHandler.Handle,
		},
		"wizard": {
			Name:        "wizard",
			Description: "インタラクティブセットアップウィザードを実行する",
			Handler:     ch.wizardHandler.Handle,
		},
		"lang": {
			Name:        "lang",
			Description: "言語設定を管理する",
			Handler:     ch.langHandler.Handle,
		},
		"web": {
			Name:        "web",
			Description: "Webダッシュボードを起動する",
			Handler:     ch.webHandler.Handle,
		},
		"period": {
			Name:        "period",
			Description: "期間別分析を実行する",
			Handler:     ch.periodHandler.Handle,
		},
		"version": {
			Name:        "version",
			Description: "バージョン情報を表示する",
			Handler:     ch.versionHandler.Handle,
		},
		"help": {
			Name:        "help",
			Description: "ヘルプを表示する",
			Handler:     ch.helpHandler.Handle,
		},
	}
}

// Execute はコマンドを実行する
func (ch *CommandHandler) Execute(command string, args []string) error {
	cmd, exists := ch.commands[command]
	if !exists {
		return errors.UnknownCommand(command)
	}

	return cmd.Handler(args)
}

// GetCommands は登録されているコマンド一覧を取得する
func (ch *CommandHandler) GetCommands() map[string]Command {
	return ch.commands
}