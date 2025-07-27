#!/bin/bash

# AI Code Tracker (AICT) セキュリティスキャンスクリプト
# このスクリプトは AICT のセキュリティ状況を包括的にチェックします

set -e

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

log_check() {
    echo -e "${PURPLE}[CHECK]${NC} $1"
}

# セキュリティスキャン結果
SECURITY_ISSUES=0
PRIVACY_ISSUES=0
TOTAL_CHECKS=0

# 結果の記録
record_issue() {
    local type=$1
    local message=$2
    
    case $type in
        "security")
            SECURITY_ISSUES=$((SECURITY_ISSUES + 1))
            log_error "セキュリティ: $message"
            ;;
        "privacy")
            PRIVACY_ISSUES=$((PRIVACY_ISSUES + 1))
            log_warning "プライバシー: $message"
            ;;
        "info")
            log_info "$message"
            ;;
    esac
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

# 1. ファイル権限チェック
check_file_permissions() {
    log_check "ファイル権限をチェック中..."
    
    # .git/ai-tracker ディレクトリの存在確認
    if [ ! -d ".git/ai-tracker" ]; then
        record_issue "info" ".git/ai-tracker ディレクトリが見つかりません（未初期化の可能性）"
        return
    fi
    
    # ディレクトリ権限チェック
    local dir_perms=$(stat -f "%A" .git/ai-tracker 2>/dev/null || stat -c "%a" .git/ai-tracker 2>/dev/null)
    if [ "$dir_perms" != "700" ]; then
        record_issue "security" ".git/ai-tracker ディレクトリの権限が緩すぎます: $dir_perms (推奨: 700)"
    else
        log_success "ディレクトリ権限: OK ($dir_perms)"
    fi
    
    # JSONL ファイル権限チェック
    for file in .git/ai-tracker/*.jsonl; do
        if [ -f "$file" ]; then
            local file_perms=$(stat -f "%A" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
            if [ "$file_perms" != "600" ]; then
                record_issue "security" "ファイル権限が緩すぎます: $file ($file_perms, 推奨: 600)"
            else
                log_success "ファイル権限: OK $(basename "$file") ($file_perms)"
            fi
        fi
    done
}

# 2. データ漏洩チェック
check_data_leakage() {
    log_check "データ漏洩の可能性をチェック中..."
    
    # 機密情報パターンの検索
    local sensitive_patterns=(
        "password"
        "api_key"
        "secret"
        "token"
        "credential"
        "private_key"
    )
    
    for pattern in "${sensitive_patterns[@]}"; do
        if [ -d ".git/ai-tracker" ]; then
            local matches=$(grep -r -i "$pattern" .git/ai-tracker/ 2>/dev/null | wc -l)
            if [ "$matches" -gt 0 ]; then
                record_issue "privacy" "機密情報の可能性がある文字列を検出: $pattern ($matches 件)"
            fi
        fi
    done
    
    # 大きなファイルサイズの警告
    if [ -d ".git/ai-tracker" ]; then
        local large_files=$(find .git/ai-tracker -name "*.jsonl" -size +10M 2>/dev/null)
        if [ -n "$large_files" ]; then
            record_issue "privacy" "大容量ファイルを検出（個人情報の過度な収集の可能性）: $large_files"
        fi
    fi
}

# 3. hooks セキュリティチェック
check_hooks_security() {
    log_check "Git hooks のセキュリティをチェック中..."
    
    # post-commit hook の存在と権限
    if [ -f ".git/hooks/post-commit" ]; then
        local hook_perms=$(stat -f "%A" .git/hooks/post-commit 2>/dev/null || stat -c "%a" .git/hooks/post-commit 2>/dev/null)
        if [ "$hook_perms" != "755" ]; then
            record_issue "security" "post-commit hook の権限が不適切: $hook_perms (推奨: 755)"
        else
            log_success "hooks 権限: OK ($hook_perms)"
        fi
        
        # hook の内容チェック
        if grep -q "AICT" .git/hooks/post-commit; then
            log_success "AICT hooks が正常に設定されています"
        else
            record_issue "security" "post-commit hook に AICT の設定が見つかりません"
        fi
        
        # 危険なコマンドのチェック
        local dangerous_commands=("rm -rf" "sudo" "curl" "wget" "ssh")
        for cmd in "${dangerous_commands[@]}"; do
            if grep -q "$cmd" .git/hooks/post-commit; then
                record_issue "security" "hooks に危険なコマンドを検出: $cmd"
            fi
        done
    else
        record_issue "info" "post-commit hook が設定されていません"
    fi
}

# 4. Claude Code hooks チェック
check_claude_hooks() {
    log_check "Claude Code hooks のセキュリティをチェック中..."
    
    local claude_config="$HOME/.claude/hooks-aict.json"
    if [ -f "$claude_config" ]; then
        # ファイル権限チェック
        local config_perms=$(stat -f "%A" "$claude_config" 2>/dev/null || stat -c "%a" "$claude_config" 2>/dev/null)
        if [ "$config_perms" != "600" ]; then
            record_issue "security" "Claude hooks 設定ファイルの権限が緩すぎます: $config_perms (推奨: 600)"
        else
            log_success "Claude hooks 権限: OK ($config_perms)"
        fi
        
        # JSON 形式の検証
        if ! python3 -m json.tool "$claude_config" > /dev/null 2>&1; then
            record_issue "security" "Claude hooks 設定ファイルの JSON 形式が不正です"
        else
            log_success "Claude hooks 設定形式: OK"
        fi
        
        # 危険な設定のチェック
        if grep -q "shell" "$claude_config"; then
            record_issue "security" "Claude hooks にシェル実行の設定が含まれています"
        fi
    else
        record_issue "info" "Claude Code hooks が設定されていません"
    fi
}

# 5. ネットワークセキュリティチェック
check_network_security() {
    log_check "ネットワークセキュリティをチェック中..."
    
    # AICT バイナリの外部通信チェック（静的解析）
    local aict_binary="./bin/aict"
    if [ -f "$aict_binary" ]; then
        # 外部通信関連の文字列を検索
        local network_patterns=("http://" "https://" "tcp://" "udp://" "ftp://")
        for pattern in "${network_patterns[@]}"; do
            if strings "$aict_binary" 2>/dev/null | grep -q "$pattern"; then
                record_issue "security" "バイナリに外部通信の可能性を検出: $pattern"
            fi
        done
        
        log_success "ネットワーク通信: 外部通信は検出されませんでした"
    else
        record_issue "info" "AICT バイナリが見つかりません"
    fi
}

# 6. 依存関係セキュリティチェック
check_dependencies() {
    log_check "依存関係のセキュリティをチェック中..."
    
    # Go modules の脆弱性チェック（govulncheck があれば）
    if command -v govulncheck &> /dev/null; then
        log_info "govulncheck で脆弱性をスキャン中..."
        if govulncheck ./... 2>/dev/null; then
            log_success "依存関係: 既知の脆弱性は見つかりませんでした"
        else
            record_issue "security" "依存関係に脆弱性が見つかりました"
        fi
    else
        record_issue "info" "govulncheck がインストールされていません（推奨）"
    fi
    
    # go.mod の直接依存関係チェック
    if [ -f "go.mod" ]; then
        local external_deps=$(grep -c "require" go.mod || echo "0")
        if [ "$external_deps" -gt 0 ]; then
            log_info "外部依存関係が $external_deps 個見つかりました"
            grep "require" go.mod | while read -r line; do
                log_info "  $line"
            done
        else
            log_success "外部依存関係なし（セキュリティ的に理想）"
        fi
    fi
}

# 7. プライバシー設定チェック
check_privacy_settings() {
    log_check "プライバシー設定をチェック中..."
    
    # 設定ファイルの確認
    local config_file="$HOME/.aict/config.json"
    if [ -f "$config_file" ]; then
        # プライバシー関連設定の確認
        if grep -q "anonymize" "$config_file"; then
            log_success "匿名化設定が見つかりました"
        else
            record_issue "privacy" "匿名化設定が見つかりません"
        fi
        
        if grep -q "retention" "$config_file"; then
            log_success "データ保持期間設定が見つかりました"
        else
            record_issue "privacy" "データ保持期間設定が見つかりません"
        fi
    else
        record_issue "info" "AICT 設定ファイルが見つかりません"
    fi
    
    # 収集データの分析
    if [ -d ".git/ai-tracker" ]; then
        local total_events=$(find .git/ai-tracker -name "*.jsonl" -exec cat {} \; 2>/dev/null | wc -l)
        if [ "$total_events" -gt 1000 ]; then
            record_issue "privacy" "大量のトラッキングデータを検出 ($total_events イベント)"
        else
            log_info "トラッキングデータ: $total_events イベント"
        fi
    fi
}

# 8. 自動セキュリティ修正
auto_fix_security_issues() {
    log_check "自動セキュリティ修正を実行中..."
    
    read -p "検出されたセキュリティ問題を自動修正しますか？ (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # ディレクトリ権限の修正
        if [ -d ".git/ai-tracker" ]; then
            chmod 700 .git/ai-tracker
            log_success "ディレクトリ権限を 700 に修正しました"
        fi
        
        # ファイル権限の修正
        for file in .git/ai-tracker/*.jsonl; do
            if [ -f "$file" ]; then
                chmod 600 "$file"
                log_success "ファイル権限を 600 に修正しました: $(basename "$file")"
            fi
        done
        
        # Claude hooks 権限の修正
        local claude_config="$HOME/.claude/hooks-aict.json"
        if [ -f "$claude_config" ]; then
            chmod 600 "$claude_config"
            log_success "Claude hooks 設定ファイルの権限を修正しました"
        fi
        
        # Git hooks 権限の修正
        if [ -f ".git/hooks/post-commit" ]; then
            chmod 755 .git/hooks/post-commit
            log_success "Git hooks の権限を修正しました"
        fi
    else
        log_info "自動修正をスキップしました"
    fi
}

# セキュリティレポートの生成
generate_security_report() {
    local report_file="security-report-$(date +%Y%m%d-%H%M%S).txt"
    
    cat > "$report_file" << EOF
AICT セキュリティスキャンレポート
================================
実行日時: $(date)
実行ディレクトリ: $(pwd)

スキャン結果:
- 総チェック数: $TOTAL_CHECKS
- セキュリティ問題: $SECURITY_ISSUES
- プライバシー問題: $PRIVACY_ISSUES

総合評価: $(if [ $((SECURITY_ISSUES + PRIVACY_ISSUES)) -eq 0 ]; then echo "✅ 安全"; elif [ $SECURITY_ISSUES -eq 0 ]; then echo "⚠️ 注意"; else echo "❌ 要対応"; fi)

詳細は実行ログを参照してください。

推奨アクション:
1. セキュリティ問題がある場合は直ちに修正
2. プライバシー問題は設定で改善可能
3. 定期的なセキュリティスキャンの実行
4. セキュリティアップデートの適用

次回スキャン推奨日: $(date -d "+1 month" 2>/dev/null || date -v+1m 2>/dev/null || echo "1ヶ月後")
EOF

    log_success "セキュリティレポートを生成しました: $report_file"
}

# 使用方法の表示
show_usage() {
    cat << EOF
AI Code Tracker (AICT) セキュリティスキャンスクリプト

使用方法:
  $0 [オプション]

オプション:
  -h, --help       このヘルプを表示
  -v, --verbose    詳細な出力を表示
  --auto-fix       検出された問題を自動修正
  --report-only    レポートのみ生成（修正なし）
  --quick          基本チェックのみ実行

例:
  $0               # 標準スキャン
  $0 --auto-fix    # 自動修正付きスキャン
  $0 --quick       # 高速スキャン

EOF
}

# メイン関数
main() {
    local auto_fix=false
    local report_only=false
    local quick_scan=false
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
            --auto-fix)
                auto_fix=true
                shift
                ;;
            --report-only)
                report_only=true
                shift
                ;;
            --quick)
                quick_scan=true
                shift
                ;;
            *)
                log_error "不明なオプション: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    echo "🔒 AICT セキュリティスキャン"
    echo "==============================="
    echo ""
    
    # セキュリティチェックの実行
    check_file_permissions
    check_data_leakage
    check_hooks_security
    check_claude_hooks
    
    if [ "$quick_scan" = false ]; then
        check_network_security
        check_dependencies
        check_privacy_settings
    fi
    
    # 結果の表示
    echo ""
    echo "📊 スキャン結果"
    echo "==============================="
    echo "総チェック数: $TOTAL_CHECKS"
    echo "セキュリティ問題: $SECURITY_ISSUES"
    echo "プライバシー問題: $PRIVACY_ISSUES"
    echo ""
    
    # 総合評価
    if [ $((SECURITY_ISSUES + PRIVACY_ISSUES)) -eq 0 ]; then
        log_success "✅ セキュリティ評価: 安全"
    elif [ $SECURITY_ISSUES -eq 0 ]; then
        log_warning "⚠️ セキュリティ評価: 注意（プライバシー問題のみ）"
    else
        log_error "❌ セキュリティ評価: 要対応（セキュリティ問題あり）"
    fi
    
    # 自動修正
    if [ "$auto_fix" = true ] && [ $SECURITY_ISSUES -gt 0 ]; then
        auto_fix_security_issues
    fi
    
    # レポート生成
    if [ "$report_only" = true ] || [ $((SECURITY_ISSUES + PRIVACY_ISSUES)) -gt 0 ]; then
        generate_security_report
    fi
    
    # 終了コード
    if [ $SECURITY_ISSUES -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# メイン関数の実行
main "$@"