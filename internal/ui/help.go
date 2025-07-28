package ui

import (
	"fmt"
	"os"
)

// HelpSystem は改良されたヘルプシステムを提供する
type HelpSystem struct {
	version string
	appName string
	contextHelp *ContextHelpProvider
}

// NewHelpSystem は新しいヘルプシステムを作成する
func NewHelpSystem(appName, version string) *HelpSystem {
	return &HelpSystem{
		version: version,
		appName: appName,
		contextHelp: NewContextHelpProvider(appName),
	}
}

// ShowMainHelp はメインヘルプを表示する
func (h *HelpSystem) ShowMainHelp() {
	fmt.Printf(`🤖 %s v%s - AI Code Tracker

AIが生成したコードと人間が書いたコードを自動的に区別・追跡します。
Claude Codeとの完全統合により、透明性のある開発プロセスを実現します。

`, h.appName, h.version)

	h.showUsage()
	h.showCommands()
	h.showExamples()
	h.showQuickStart()
}

// showUsage は使用方法を表示する
func (h *HelpSystem) showUsage() {
	fmt.Printf(`📖 使用方法:
  %s <command> [options]

`, h.appName)
}

// showCommands はコマンド一覧を表示する
func (h *HelpSystem) showCommands() {
	fmt.Println("📋 コマンド一覧:")
	
	commands := []struct {
		Name        string
		Description string
		Category    string
		Icon        string
	}{
		{"init", "プロジェクトでAI Code Trackerを初期化", "基本", "🏗️"},
		{"track", "ファイルの変更を手動でトラッキング", "基本", "📝"},
		{"stats", "統計情報を表示", "分析", "📊"},
		{"blame", "ファイルのAI/人間による変更履歴を表示", "分析", "🔍"},
		{"period", "期間別分析を実行", "分析", "📅"},
		{"config", "設定を管理", "設定", "⚙️"},
		{"setup", "Git hooks と Claude Code hooks を自動設定", "設定", "🔧"},
		{"wizard", "インタラクティブセットアップウィザード", "設定", "🧙"},
		{"lang", "言語設定を管理", "設定", "🌐"},
		{"web", "Webダッシュボードを起動", "分析", "🌐"},
		{"security", "セキュリティ機能を管理", "セキュリティ", "🔒"},
		{"version", "バージョン情報を表示", "情報", "ℹ️"},
		{"help", "ヘルプを表示", "情報", "❓"},
	}
	
	// カテゴリ別にグループ化
	categories := map[string][]struct {
		Name        string
		Description string
		Category    string
		Icon        string
	}{
		"基本": {},
		"分析": {},
		"設定": {},
		"セキュリティ": {},
		"情報": {},
	}
	
	for _, cmd := range commands {
		categories[cmd.Category] = append(categories[cmd.Category], cmd)
	}
	
	// 各カテゴリを表示
	for category, cmds := range categories {
		if len(cmds) == 0 {
			continue
		}
		
		fmt.Printf("\n  %s:\n", category)
		for _, cmd := range cmds {
			fmt.Printf("    %s %-12s %s\n", cmd.Icon, cmd.Name, cmd.Description)
		}
	}
	
	fmt.Println()
}

// showExamples は使用例を表示する
func (h *HelpSystem) showExamples() {
	fmt.Println("💡 使用例:")
	
	examples := []struct {
		Command     string
		Description string
		Icon        string
	}{
		{h.appName + " init", "プロジェクトを初期化", "🏗️"},
		{h.appName + " wizard", "インタラクティブセットアップ", "🧙"},
		{h.appName + " track --ai --model claude-sonnet-4 --files \"*.go\"", "AI変更を追跡", "🤖"},
		{h.appName + " track --author \"John Doe\" --files main.go", "人間の変更を追跡", "👤"},
		{h.appName + " stats --format table --since 2024-01-01", "期間別統計", "📊"},
		{h.appName + " period \"Q1 2025\"", "四半期別分析", "📅"},
		{h.appName + " blame src/main.go", "ファイルの変更履歴", "🔍"},
		{h.appName + " security scan", "セキュリティスキャン", "🔒"},
		{h.appName + " setup", "hooks 自動設定", "🔧"},
		{h.appName + " lang ja", "日本語に切り替え", "🌐"},
	}
	
	for _, example := range examples {
		fmt.Printf("  %s %s\n", example.Icon, example.Command)
		fmt.Printf("    → %s\n\n", example.Description)
	}
}

// showQuickStart はクイックスタートガイドを表示する
func (h *HelpSystem) showQuickStart() {
	fmt.Println("🚀 クイックスタート:")
	fmt.Printf("  1. %s init                    # プロジェクトを初期化\n", h.appName)
	fmt.Printf("  2. %s setup                   # hooks を自動設定\n", h.appName)
	fmt.Printf("  3. Claude Code でコードを編集     # 自動追跡が開始されます\n")
	fmt.Printf("  4. %s stats                   # 統計を確認\n", h.appName)
	fmt.Println()
	fmt.Printf("より詳細な設定は '%s wizard' をお試しください。\n", h.appName)
	fmt.Println()
}

// ShowCommandHelp は特定のコマンドのヘルプを表示する
func (h *HelpSystem) ShowCommandHelp(command string) {
	switch command {
	case "init":
		h.showInitHelp()
	case "track":
		h.showTrackHelp()
	case "stats":
		h.showStatsHelp()
	case "blame":
		h.showBlameHelp()
	case "period":
		h.showPeriodHelp()
	case "config":
		h.showConfigHelp()
	case "setup":
		h.showSetupHelp()
	case "wizard":
		h.showWizardHelp()
	case "lang":
		h.showLangHelp()
	case "web":
		h.showWebHelp()
	case "security":
		h.showSecurityHelp()
	default:
		fmt.Printf("❌ 不明なコマンド: %s\n", command)
		fmt.Printf("利用可能なコマンドは '%s help' をご覧ください。\n", h.appName)
	}
}

func (h *HelpSystem) showInitHelp() {
	fmt.Printf(`🏗️ %s init - プロジェクト初期化

説明:
  現在のディレクトリでAI Code Trackerを初期化します。
  .git/ai-tracker ディレクトリとデータファイルを作成します。

使用方法:
  %s init [options]

オプション:
  --force      既存の設定を上書きして初期化
  --security   セキュリティ機能を有効にして初期化
  --wizard     インタラクティブセットアップを実行

例:
  %s init                    # 基本初期化
  %s init --force            # 強制初期化
  %s init --wizard           # ウィザード実行

`, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showTrackHelp() {
	fmt.Printf(`📝 %s track - 変更追跡

説明:
  ファイルの変更を手動でトラッキングします。
  通常はClaude Code hooksにより自動的に実行されます。

使用方法:
  %s track [options]

必須オプション:
  --files <files>     追跡するファイル（カンマ区切り）
  --author <name>     作成者名 OR --ai フラグ

オプション:
  --ai                AI による変更として記録
  --model <model>     AI モデル名（--ai 使用時）
  --message <msg>     変更の説明
  --session <id>      セッションID

例:
  %s track --ai --model claude-sonnet-4 --files "src/*.go" --message "リファクタリング"
  %s track --author "John Doe" --files main.go --message "バグ修正"

`, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showStatsHelp() {
	fmt.Printf(`📊 %s stats - 統計表示

説明:
  プロジェクトのAI/人間によるコード統計を表示します。

使用方法:
  %s stats [options]

オプション:
  --format <format>   出力形式 (table|json|summary|daily|files|contributors)
  --since <date>      指定日以降の統計 (YYYY-MM-DD)
  --until <date>      指定日まで統計 (YYYY-MM-DD)
  --author <name>     作成者でフィルタ
  --by-file           ファイル別統計を表示
  --trend             トレンド分析を表示
  --top <N>           上位N件を表示

例:
  %s stats                                    # 基本統計
  %s stats --format json                     # JSON形式
  %s stats --since 2024-01-01 --until 2024-01-31  # 期間指定
  %s stats --by-file --top 10                # ファイル別上位10件
  %s stats --trend --author claude           # Claudeのトレンド

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showBlameHelp() {
	fmt.Printf(`🔍 %s blame - 変更履歴

説明:
  ファイルのAI/人間による変更履歴を行単位で表示します。
  Git blameを拡張してAI貢献度を可視化します。

使用方法:
  %s blame <file> [options]

オプション:
  --no-color      カラー表示を無効化
  --stats         貢献者統計のみ表示
  --top <N>       上位N名の貢献者を表示
  --format <fmt>  出力形式 (default|compact|detailed)

例:
  %s blame src/main.go                    # 基本blame表示
  %s blame --stats src/main.go            # 統計のみ
  %s blame --top 5 src/main.go            # 上位5名
  %s blame --no-color --format compact src/main.go  # 簡潔表示

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showConfigHelp() {
	fmt.Printf(`⚙️ %s config - 設定管理

説明:
  AICT の設定を表示・変更します。

使用方法:
  %s config [options]

オプション:
  --list              現在の設定を表示
  --set <key=value>   設定を変更
  --get <key>         特定の設定値を取得
  --reset             設定をリセット
  --export            設定をファイルにエクスポート
  --import <file>     設定をファイルからインポート

設定項目:
  default_author        デフォルト作成者名
  enable_encryption     データ暗号化
  enable_audit_log      監査ログ
  anonymize_authors     作成者匿名化
  data_retention_days   データ保持期間

例:
  %s config --list                           # 設定一覧
  %s config --set default_author="John Doe"  # 作成者設定
  %s config --get enable_encryption          # 暗号化設定確認

`, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showPeriodHelp() {
	fmt.Printf(`📅 %s period - 期間別分析

説明:
  指定した期間におけるAI/人間のコード貢献度を詳細に分析します。
  四半期、月、日付などの柔軟な期間指定に対応しています。

使用方法:
  %s period <period_expression>

期間表現:
  四半期:     Q1 2025, Q2 2024, q3 2023, q4 2022
  年:        this year, last year, 2024
  日付:      2025-07-28, 2025/07/28
  月:        2025-07, 2024-12
  月名:      Jan-Mar 2024, Apr-Jun 2025
  相対:      last 3 months, last month

出力内容:
  • 全体統計 (AI/人間コード行数、割合)
  • 上位ファイル別分析
  • 言語別統計
  • 貢献者別統計
  • アクティブ日数

例:
  %s period "Q1 2025"              # 2025年第1四半期
  %s period "2025-07-28"           # 特定の日
  %s period "this year"            # 今年
  %s period "last 3 months"        # 過去3ヶ月
  %s period "2024-12"              # 2024年12月

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showSetupHelp() {
	fmt.Printf(`🔧 %s setup - hooks 設定

説明:
  Git hooks と Claude Code hooks を自動設定します。

使用方法:
  %s setup [options]

オプション:
  --git-hooks         Git hooks のみを設定
  --claude-hooks      Claude Code hooks のみを設定
  --remove            hooks を削除
  --status            hooks の設定状況を表示
  --force             既存のhooksを上書き

例:
  %s setup                    # 全てのhooksを設定
  %s setup --git-hooks        # Git hooksのみ
  %s setup --status           # 設定状況確認
  %s setup --remove           # hooks削除

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showWizardHelp() {
	fmt.Printf(`🧙 %s wizard - セットアップウィザード

説明:
  インタラクティブなセットアップウィザードを実行します。
  初回セットアップや設定変更に最適です。

使用方法:
  %s wizard [type]

ウィザードタイプ:
  init        初期化ウィザード（デフォルト）
  security    セキュリティ設定ウィザード
  quickstart  クイックスタートウィザード

例:
  %s wizard                   # 初期化ウィザード
  %s wizard security          # セキュリティ設定
  %s wizard quickstart        # クイックスタート

`, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showLangHelp() {
	fmt.Printf(`🌐 %s lang - 言語設定管理

説明:
  表示言語を動的に切り替えます。
  設定は一時的または永続的に保存できます。

使用方法:
  %s lang [options] [language_code]

オプション:
  --list              利用可能な言語を表示
  --set <code>        言語を設定 (ja|en)
  --persistent        設定を永続化

引数:
  language_code       言語コード (ja または en)

例:
  %s lang                    # 現在の言語を表示
  %s lang --list             # 利用可能な言語一覧
  %s lang ja                 # 日本語に切り替え
  %s lang en                 # 英語に切り替え
  %s lang ja --persistent    # 日本語に設定して永続化

環境変数:
  AICT_LANGUAGE       デフォルト言語 (ja|en)

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showWebHelp() {
	fmt.Printf(`🌐 %s web - Webダッシュボード

説明:
  ブラウザベースのリアルタイムダッシュボードを起動します。
  AI/人間のコード統計、ファイル分析、タイムラインなどを視覚的に表示します。

使用方法:
  %s web [options]

オプション:
  -p, --port <port>     サーバーポート（デフォルト: 8080）
  -l, --lang <lang>     表示言語（ja|en、デフォルト: ja）
  -d, --debug          デバッグモードを有効化
      --data <dir>     データディレクトリを指定
      --no-browser     ブラウザを自動で開かない

機能:
  • リアルタイム統計更新
  • 多言語対応インターフェース
  • レスポンシブデザイン
  • インタラクティブチャート
  • ファイル別詳細分析
  • 貢献者別統計
  • タイムライン表示

例:
  %s web                          # デフォルト設定で起動
  %s web -p 3000                  # ポート3000で起動
  %s web -l en --debug            # 英語+デバッグモードで起動
  %s web --no-browser             # ブラウザを開かずに起動

アクセス:
  http://localhost:8080           # デフォルトURL

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) showSecurityHelp() {
	fmt.Printf(`🔒 %s security - セキュリティ管理

説明:
  セキュリティ機能の管理とスキャンを実行します。

使用方法:
  %s security <subcommand> [options]

サブコマンド:
  scan        セキュリティスキャン実行
  status      セキュリティ状況確認
  config      セキュリティ設定管理
  audit       監査ログ管理

オプション（scan）:
  --check <type>      特定の項目をチェック (permissions|encryption|audit)
  --output <file>     レポートをファイルに出力
  --format <fmt>      出力形式 (text|json)

オプション（audit）:
  --show              監査ログを表示
  --filter <filter>   ログをフィルタ
  --since <date>      指定日以降のログ

例:
  %s security scan                    # セキュリティスキャン
  %s security status                  # 状況確認
  %s security audit --show            # 監査ログ表示
  %s security scan --check permissions --output report.json

`, h.appName, h.appName, h.appName, h.appName, h.appName, h.appName)
}

// ShowError はエラーメッセージを表示する
func (h *HelpSystem) ShowError(err error, command string) {
	fmt.Fprintf(os.Stderr, "❌ エラー: %v\n", err)
	
	// コマンド固有のヘルプ提案
	switch command {
	case "track":
		fmt.Fprintf(os.Stderr, "\n💡 ヒント: `%s help track` でtrack コマンドの詳細な使用方法を確認できます。\n", h.appName)
	case "stats":
		fmt.Fprintf(os.Stderr, "\n💡 ヒント: 有効な日付形式は YYYY-MM-DD です（例: 2024-01-01）。\n")
	case "blame":
		fmt.Fprintf(os.Stderr, "\n💡 ヒント: ファイルがGitで追跡されているか確認してください。\n")
	case "init":
		fmt.Fprintf(os.Stderr, "\n💡 ヒント: 既存の設定がある場合は --force オプションを使用してください。\n")
	default:
		fmt.Fprintf(os.Stderr, "\n💡 ヒント: `%s help` で利用可能なコマンドを確認できます。\n", h.appName)
	}
}

// ShowContextualError はコンテキストアウェアなエラー表示を提供する
func (h *HelpSystem) ShowContextualError(ctx *CommandContext) {
	h.contextHelp.ShowContextualError(ctx)
}

// GetQuickHelp は簡潔なヘルプを取得する
func (h *HelpSystem) GetQuickHelp(command string) string {
	return h.contextHelp.GetQuickHelp(command)
}

// ShowWarning は警告メッセージを表示する
func (h *HelpSystem) ShowWarning(message string) {
	fmt.Printf("⚠️  警告: %s\n", message)
}

// ShowSuccess は成功メッセージを表示する
func (h *HelpSystem) ShowSuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// ShowInfo は情報メッセージを表示する
func (h *HelpSystem) ShowInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// GenerateCompletionScript はシェル補完スクリプトを生成する
func (h *HelpSystem) GenerateCompletionScript(shell string) {
	switch shell {
	case "bash":
		h.generateBashCompletion()
	case "zsh":
		h.generateZshCompletion()
	default:
		fmt.Printf("❌ サポートされていないシェル: %s\n", shell)
		fmt.Println("サポートされているシェル: bash, zsh")
	}
}

func (h *HelpSystem) generateBashCompletion() {
	fmt.Printf(`#!/bin/bash

_%s_completion() {
    local cur prev words cword
    _init_completion || return

    local commands="init track stats blame config setup wizard security version help"
    local track_options="--ai --author --model --files --message --session"
    local stats_options="--format --since --until --author --by-file --trend --top"
    local blame_options="--no-color --stats --top --format"
    local config_options="--list --set --get --reset --export --import"
    local setup_options="--git-hooks --claude-hooks --remove --status --force"
    local security_commands="scan status config audit"

    if [[ ${cword} == 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- ${cur}))
        return 0
    fi

    case ${words[1]} in
        track)
            COMPREPLY=($(compgen -W "${track_options}" -- ${cur}))
            ;;
        stats)
            COMPREPLY=($(compgen -W "${stats_options}" -- ${cur}))
            ;;
        blame)
            COMPREPLY=($(compgen -W "${blame_options}" -- ${cur}))
            ;;
        config)
            COMPREPLY=($(compgen -W "${config_options}" -- ${cur}))
            ;;
        setup)
            COMPREPLY=($(compgen -W "${setup_options}" -- ${cur}))
            ;;
        security)
            if [[ ${cword} == 2 ]]; then
                COMPREPLY=($(compgen -W "${security_commands}" -- ${cur}))
            fi
            ;;
    esac
}

complete -F _%s_completion %s
`, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) generateZshCompletion() {
	fmt.Printf(`#compdef %s

_%s() {
    local line state

    _arguments -C \
        "1: :->commands" \
        "*: :->args"

    case $state in
        commands)
            _values 'commands' \
                'init[プロジェクト初期化]' \
                'track[変更追跡]' \
                'stats[統計表示]' \
                'blame[変更履歴]' \
                'config[設定管理]' \
                'setup[hooks設定]' \
                'wizard[セットアップウィザード]' \
                'security[セキュリティ管理]' \
                'version[バージョン表示]' \
                'help[ヘルプ表示]'
            ;;
        args)
            case $line[1] in
                track)
                    _arguments \
                        '--ai[AI変更として記録]' \
                        '--author[作成者名]:author:' \
                        '--model[モデル名]:model:' \
                        '--files[ファイル]:files:' \
                        '--message[メッセージ]:message:'
                    ;;
                stats)
                    _arguments \
                        '--format[出力形式]:format:(table json summary daily files contributors)' \
                        '--since[開始日]:date:' \
                        '--until[終了日]:date:' \
                        '--author[作成者]:author:' \
                        '--by-file[ファイル別]' \
                        '--trend[トレンド]' \
                        '--top[上位N件]:number:'
                    ;;
                blame)
                    _arguments \
                        '--no-color[カラー無効]' \
                        '--stats[統計のみ]' \
                        '--top[上位N名]:number:' \
                        '--format[形式]:format:(default compact detailed)' \
                        '*:file:_files'
                    ;;
                security)
                    _values 'security commands' \
                        'scan[セキュリティスキャン]' \
                        'status[状況確認]' \
                        'config[設定管理]' \
                        'audit[監査ログ]'
                    ;;
            esac
            ;;
    esac
}

_%s "$@"
`, h.appName, h.appName, h.appName)
}

func (h *HelpSystem) ShowTips() {
	tips := []string{
		"💡 Claude Code で編集すると自動的にAI変更が追跡されます",
		"💡 `aict stats --trend` でAI使用率の変化を確認できます",
		"💡 `aict blame <file>` で各行の作成者（AI/人間）を確認できます",
		"💡 セキュリティ機能は `aict wizard security` で簡単に設定できます",
		"💡 `aict setup --status` でhooksの設定状況を確認できます",
		"💡 統計データは `.git/ai-tracker/` に保存されます",
		"💡 環境変数でデフォルト設定をカスタマイズできます",
		"💡 `aict config --export` で設定をバックアップできます",
	}
	
	fmt.Println("🎯 便利なTips:")
	for _, tip := range tips {
		fmt.Printf("  %s\n", tip)
	}
	fmt.Println()
}