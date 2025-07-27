#!/bin/bash

# AI Code Tracker (AICT) インストールスクリプト
# このスクリプトは AICT をシステムにインストールし、必要な設定を行います

set -e  # エラーが発生した場合は即座に終了

# カラー設定
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 設定変数
AICT_VERSION="v1.0.0"
AICT_BINARY="aict"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.aict"
TEMP_DIR="/tmp/aict-install"

# プラットフォーム検出
detect_platform() {
    local os=$(uname -s)
    local arch=$(uname -m)
    
    case "$os" in
        Darwin)
            platform="darwin"
            ;;
        Linux)
            platform="linux"
            ;;
        *)
            log_error "サポートされていないOS: $os"
            exit 1
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            arch="amd64"
            ;;
        arm64|aarch64)
            arch="arm64"
            ;;
        *)
            log_error "サポートされていないアーキテクチャ: $arch"
            exit 1
            ;;
    esac
    
    log_info "検出されたプラットフォーム: ${platform}-${arch}"
}

# 依存関係チェック
check_dependencies() {
    log_info "依存関係をチェック中..."
    
    # Gitの確認
    if ! command -v git &> /dev/null; then
        log_error "Git が見つかりません。Git をインストールしてください。"
        exit 1
    fi
    
    local git_version=$(git --version | awk '{print $3}')
    log_success "Git $git_version が見つかりました"
    
    # Goの確認（ビルド用）
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}')
        log_success "Go $go_version が見つかりました"
    else
        log_warning "Go が見つかりません。ソースからビルドする場合は Go 1.19+ をインストールしてください。"
    fi
    
    # Claude Code の確認
    if command -v claude &> /dev/null; then
        log_success "Claude Code が見つかりました"
    else
        log_warning "Claude Code が見つかりません。hooks 統合を使用する場合は Claude Code をインストールしてください。"
    fi
}

# バイナリのダウンロードまたはビルド
install_binary() {
    log_info "AICT バイナリをインストール中..."
    
    # 一時ディレクトリの作成
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # ローカルビルドの確認
    if [ -f "$(dirname "$0")/../bin/aict" ]; then
        log_info "ローカルビルドされたバイナリを使用します"
        cp "$(dirname "$0")/../bin/aict" "$TEMP_DIR/$AICT_BINARY"
    elif [ -f "$(dirname "$0")/../main.go" ] && command -v go &> /dev/null; then
        log_info "ソースからビルド中..."
        cd "$(dirname "$0")/.."
        go build -o "$TEMP_DIR/$AICT_BINARY" ./cmd/aict/
        cd "$TEMP_DIR"
    else
        log_error "バイナリまたはソースが見つかりません"
        log_info "以下のいずれかを実行してください:"
        log_info "1. プロジェクトディレクトリで 'make build' を実行"
        log_info "2. Go をインストールしてソースからビルド"
        exit 1
    fi
    
    # バイナリの確認
    if [ ! -f "$AICT_BINARY" ]; then
        log_error "バイナリファイルが見つかりません"
        exit 1
    fi
    
    # 実行権限の付与
    chmod +x "$AICT_BINARY"
    
    # バージョン確認
    local version=$(./"$AICT_BINARY" version 2>/dev/null || echo "unknown")
    log_success "AICT バイナリを準備しました (version: $version)"
}

# システムインストール
install_to_system() {
    log_info "システムにインストール中..."
    
    # インストールディレクトリの権限確認
    if [ ! -w "$INSTALL_DIR" ]; then
        log_warning "$INSTALL_DIR への書き込み権限がありません。sudo を使用します。"
        sudo cp "$TEMP_DIR/$AICT_BINARY" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$AICT_BINARY"
    else
        cp "$TEMP_DIR/$AICT_BINARY" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$AICT_BINARY"
    fi
    
    log_success "AICT を $INSTALL_DIR にインストールしました"
}

# 設定ディレクトリの作成
setup_config() {
    log_info "設定ディレクトリを作成中..."
    
    mkdir -p "$CONFIG_DIR"
    
    # デフォルト設定ファイルの作成
    cat > "$CONFIG_DIR/config.json" << EOF
{
  "version": "1.0.0",
  "created": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "settings": {
    "default_model": "claude-sonnet-4",
    "auto_detect_claude": true,
    "show_statistics": true,
    "color_output": true,
    "debug_mode": false
  },
  "paths": {
    "data_dir": ".git/ai-tracker",
    "hooks_config": "~/.claude/hooks-aict.json"
  }
}
EOF
    
    log_success "設定ディレクトリを $CONFIG_DIR に作成しました"
}

# PATH の確認と案内
check_path() {
    log_info "PATH の確認中..."
    
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        log_success "$INSTALL_DIR は既に PATH に含まれています"
    else
        log_warning "$INSTALL_DIR が PATH に含まれていません"
        log_info "以下のコマンドを実行して PATH に追加してください:"
        echo ""
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""
        log_info "永続的に設定するには、以下を ~/.bashrc または ~/.zshrc に追加してください:"
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# インストール完了後のテスト
test_installation() {
    log_info "インストールをテスト中..."
    
    if command -v aict &> /dev/null; then
        local version=$(aict version 2>/dev/null || echo "unknown")
        log_success "AICT がインストールされました (version: $version)"
        
        # ヘルプの表示テスト
        if aict help &> /dev/null; then
            log_success "基本コマンドが正常に動作します"
        else
            log_warning "基本コマンドの実行でエラーが発生しました"
        fi
    else
        log_error "AICT コマンドが見つかりません"
        log_info "PATH の設定を確認してください"
        return 1
    fi
}

# クリーンアップ
cleanup() {
    log_info "一時ファイルをクリーンアップ中..."
    rm -rf "$TEMP_DIR"
    log_success "クリーンアップが完了しました"
}

# 使用方法の表示
show_usage() {
    cat << EOF
AI Code Tracker (AICT) インストールスクリプト

使用方法:
  $0 [オプション]

オプション:
  -h, --help     このヘルプを表示
  -v, --verbose  詳細な出力を表示
  --install-dir DIR  インストールディレクトリを指定 (デフォルト: $INSTALL_DIR)
  --config-dir DIR   設定ディレクトリを指定 (デフォルト: $CONFIG_DIR)
  --no-hooks     hooks 設定をスキップ
  --force        既存のインストールを上書き

例:
  $0                           # 標準インストール
  $0 --install-dir ~/bin       # ユーザーディレクトリにインストール
  $0 --no-hooks               # hooks 設定なしでインストール

EOF
}

# 次の手順の案内
show_next_steps() {
    echo ""
    log_success "✅ AICT のインストールが完了しました！"
    echo ""
    log_info "次の手順:"
    echo "  1. 新しいターミナルを開くか、以下を実行してください:"
    echo "     source ~/.bashrc  # または source ~/.zshrc"
    echo ""
    echo "  2. Git リポジトリで AICT を初期化してください:"
    echo "     cd /path/to/your/git/repo"
    echo "     aict init"
    echo ""
    echo "  3. hooks を設定してください (任意):"
    echo "     aict setup"
    echo ""
    echo "  4. 使用方法を確認してください:"
    echo "     aict help"
    echo ""
    log_info "詳細な使用方法については、プロジェクトのドキュメントを参照してください。"
}

# メイン関数
main() {
    local setup_hooks=true
    local force_install=false
    local verbose=false
    
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
            --install-dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --config-dir)
                CONFIG_DIR="$2"
                shift 2
                ;;
            --no-hooks)
                setup_hooks=false
                shift
                ;;
            --force)
                force_install=true
                shift
                ;;
            *)
                log_error "不明なオプション: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    echo "🚀 AI Code Tracker (AICT) インストーラー"
    echo "========================================"
    echo ""
    
    # 既存インストールの確認
    if command -v aict &> /dev/null && [ "$force_install" = false ]; then
        local existing_version=$(aict version 2>/dev/null || echo "unknown")
        log_warning "AICT は既にインストールされています (version: $existing_version)"
        read -p "上書きしますか？ (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "インストールをキャンセルしました"
            exit 0
        fi
    fi
    
    # インストール手順の実行
    detect_platform
    check_dependencies
    install_binary
    install_to_system
    setup_config
    check_path
    test_installation
    
    # hooks セットアップの案内
    if [ "$setup_hooks" = true ]; then
        echo ""
        log_info "hooks 設定スクリプトを実行しますか？"
        read -p "Git と Claude Code の hooks を設定しますか？ (Y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
            if [ -f "$(dirname "$0")/setup-hooks.sh" ]; then
                bash "$(dirname "$0")/setup-hooks.sh"
            else
                log_info "後で 'aict setup' コマンドで hooks を設定できます"
            fi
        fi
    fi
    
    cleanup
    show_next_steps
}

# スクリプトのエラーハンドリング
trap cleanup EXIT

# メイン関数の実行
main "$@"