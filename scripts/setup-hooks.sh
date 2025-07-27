#!/bin/bash

# AI Code Tracker (AICT) Hooks セットアップスクリプト
# このスクリプトは Git hooks と Claude Code hooks を自動設定します

set -e  # エラーが発生した場合は即座に終了

# カラー設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# ログ関数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $1"
}

# 設定変数
CLAUDE_CONFIG_DIR="$HOME/.claude"
CLAUDE_HOOKS_FILE="$CLAUDE_CONFIG_DIR/hooks-aict.json"
CURRENT_DIR=$(pwd)

# Git リポジトリの確認
check_git_repo() {
    log_step "Git リポジトリの確認中..."
    
    if ! git rev-parse --git-dir &> /dev/null; then
        log_error "現在のディレクトリはGitリポジトリではありません"
        log_info "Gitリポジトリのルートディレクトリで実行してください"
        exit 1
    fi
    
    local git_root=$(git rev-parse --show-toplevel)
    log_success "Gitリポジトリが見つかりました: $git_root"
    
    # Gitリポジトリのルートに移動
    cd "$git_root"
}

# AICT コマンドの確認
check_aict_command() {
    log_step "AICT コマンドの確認中..."
    
    if ! command -v aict &> /dev/null; then
        log_error "aict コマンドが見つかりません"
        log_info "先に AICT をインストールしてください:"
        log_info "  bash scripts/install.sh"
        exit 1
    fi
    
    local version=$(aict version 2>/dev/null || echo "unknown")
    log_success "AICT コマンドが見つかりました (version: $version)"
}

# AICT の初期化
initialize_aict() {
    log_step "AICT の初期化中..."
    
    if [ -d ".git/ai-tracker" ]; then
        log_info "AICT は既に初期化されています"
    else
        log_info "AICT を初期化中..."
        aict init
        log_success "AICT が初期化されました"
    fi
}

# Claude Code の確認
check_claude_code() {
    log_step "Claude Code の確認中..."
    
    if command -v claude &> /dev/null; then
        log_success "Claude Code が見つかりました"
        return 0
    else
        log_warning "Claude Code が見つかりません"
        log_info "Claude Code hooks の設定をスキップします"
        return 1
    fi
}

# Git hooks の設定
setup_git_hooks() {
    log_step "Git hooks の設定中..."
    
    # aict setup コマンドを使用してGit hooksを設定
    if aict setup --git-hooks; then
        log_success "Git hooks が設定されました"
    else
        log_error "Git hooks の設定に失敗しました"
        return 1
    fi
    
    # 設定状況の確認
    local status=$(aict setup --status 2>/dev/null | grep -i "git hooks" || echo "unknown")
    log_info "Git hooks 状況: $status"
}

# Claude Code hooks の設定
setup_claude_hooks() {
    log_step "Claude Code hooks の設定中..."
    
    # Claude Code の存在確認
    if ! check_claude_code; then
        return 1
    fi
    
    # Claude設定ディレクトリの作成
    mkdir -p "$CLAUDE_CONFIG_DIR"
    
    # aict setup コマンドを使用してClaude Code hooksを設定
    if aict setup --claude-hooks; then
        log_success "Claude Code hooks が設定されました"
    else
        log_error "Claude Code hooks の設定に失敗しました"
        return 1
    fi
    
    # 設定状況の確認
    local status=$(aict setup --status 2>/dev/null | grep -i "claude hooks" || echo "unknown")
    log_info "Claude Code hooks 状況: $status"
    
    # 設定ファイルの確認
    if [ -f "$CLAUDE_HOOKS_FILE" ]; then
        log_success "Claude Code hooks 設定ファイルが作成されました: $CLAUDE_HOOKS_FILE"
    else
        log_warning "Claude Code hooks 設定ファイルが見つかりません"
    fi
}

# hooks 設定の確認
verify_hooks_setup() {
    log_step "hooks 設定の確認中..."
    
    echo ""
    log_info "📊 hooks 設定状況:"
    aict setup --status
    echo ""
}

# テスト実行
test_hooks() {
    log_step "hooks のテスト中..."
    
    # テスト用のファイルを作成
    local test_file="aict-hooks-test.txt"
    echo "# AICT Hooks テスト" > "$test_file"
    echo "作成日時: $(date)" >> "$test_file"
    
    log_info "テスト用ファイルを作成しました: $test_file"
    
    # 手動でトラッキングをテスト
    if aict track --author "Hooks Test" --files "$test_file" --message "hooks セットアップテスト"; then
        log_success "手動トラッキングが成功しました"
    else
        log_warning "手動トラッキングでエラーが発生しました"
    fi
    
    # Gitコミットのテスト（hooks が動作するかテスト）
    git add "$test_file"
    if git commit -m "AICT hooks セットアップテスト

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>"; then
        log_success "Git コミットが成功しました（post-commit hook が実行されました）"
    else
        log_warning "Git コミットでエラーが発生しました"
    fi
    
    # 統計の確認
    echo ""
    log_info "📈 現在の統計:"
    aict stats --format summary
    
    # テストファイルのクリーンアップ
    rm -f "$test_file"
    log_info "テストファイルを削除しました"
}

# 使用方法の表示
show_usage() {
    cat << EOF
AI Code Tracker (AICT) Hooks セットアップスクリプト

使用方法:
  $0 [オプション]

オプション:
  -h, --help          このヘルプを表示
  -v, --verbose       詳細な出力を表示
  --git-only          Git hooks のみ設定
  --claude-only       Claude Code hooks のみ設定
  --no-test          テストを実行しない
  --status-only      現在の設定状況のみ表示
  --remove           hooks を削除

例:
  $0                  # 全ての hooks を設定
  $0 --git-only       # Git hooks のみ設定
  $0 --claude-only    # Claude Code hooks のみ設定
  $0 --status-only    # 現在の状況を確認
  $0 --remove         # hooks を削除

EOF
}

# hooks の削除
remove_hooks() {
    log_step "hooks の削除中..."
    
    if aict setup --remove; then
        log_success "hooks が削除されました"
    else
        log_error "hooks の削除に失敗しました"
        return 1
    fi
    
    verify_hooks_setup
}

# 設定完了後の案内
show_completion_guide() {
    echo ""
    log_success "🎉 AICT hooks のセットアップが完了しました！"
    echo ""
    log_info "📋 次の手順:"
    echo "  1. 通常通り開発を続けてください"
    echo "  2. Claude Code でコードを編集すると自動的にトラッキングされます"
    echo "  3. Git commit 時にも自動的にトラッキングされます"
    echo "  4. 統計情報を確認するには: aict stats"
    echo "  5. blame情報を確認するには: aict blame <file>"
    echo ""
    log_info "🔧 便利なコマンド:"
    echo "  aict stats --format summary    # 簡単な統計表示"
    echo "  aict stats --format daily      # 日次統計表示"
    echo "  aict blame --stats <file>      # ファイルの貢献者統計"
    echo "  aict setup --status            # hooks 設定状況確認"
    echo ""
    log_info "❓ 問題が発生した場合:"
    echo "  - hooks の状況確認: aict setup --status"
    echo "  - hooks の再設定: bash scripts/setup-hooks.sh"
    echo "  - hooks の削除: aict setup --remove"
}

# メイン関数
main() {
    local setup_git=true
    local setup_claude=true
    local run_test=true
    local verbose=false
    local status_only=false
    local remove_hooks=false
    
    # コマンドライン引数の解析
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -v|--verbose)
                verbose=true
                set -x  # デバッグ出力を有効化
                shift
                ;;
            --git-only)
                setup_git=true
                setup_claude=false
                shift
                ;;
            --claude-only)
                setup_git=false
                setup_claude=true
                shift
                ;;
            --no-test)
                run_test=false
                shift
                ;;
            --status-only)
                status_only=true
                setup_git=false
                setup_claude=false
                run_test=false
                shift
                ;;
            --remove)
                remove_hooks=true
                setup_git=false
                setup_claude=false
                run_test=false
                shift
                ;;
            *)
                log_error "不明なオプション: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    echo "🔗 AI Code Tracker (AICT) Hooks セットアップ"
    echo "============================================="
    echo ""
    
    # 基本チェック
    check_git_repo
    check_aict_command
    
    # 削除モード
    if [ "$remove_hooks" = true ]; then
        remove_hooks
        exit 0
    fi
    
    # 状況確認のみ
    if [ "$status_only" = true ]; then
        verify_hooks_setup
        exit 0
    fi
    
    # AICT の初期化
    initialize_aict
    
    # hooks のセットアップ
    local setup_success=true
    
    if [ "$setup_git" = true ]; then
        if ! setup_git_hooks; then
            setup_success=false
        fi
    fi
    
    if [ "$setup_claude" = true ]; then
        if ! setup_claude_hooks; then
            log_warning "Claude Code hooks の設定に失敗しましたが、続行します"
        fi
    fi
    
    # 設定確認
    verify_hooks_setup
    
    # テストの実行
    if [ "$run_test" = true ] && [ "$setup_success" = true ]; then
        echo ""
        read -p "hooks のテストを実行しますか？ (Y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
            test_hooks
        fi
    fi
    
    # 完了案内
    if [ "$setup_success" = true ]; then
        show_completion_guide
    else
        log_error "一部の設定に失敗しました。手動で設定を確認してください。"
        exit 1
    fi
    
    # 元のディレクトリに戻る
    cd "$CURRENT_DIR"
}

# メイン関数の実行
main "$@"