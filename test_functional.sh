#!/bin/bash

# AI Code Tracker - Functional Integration Test
# 仮リポジトリで複数コミットの全コマンドを動作確認する
#
# 使い方:
#   ./test_functional.sh
#
# 前提:
#   go build -o bin/aict ./cmd/aict

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
TOTAL=0

pass() {
    PASSED=$((PASSED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    FAILED=$((FAILED + 1))
    TOTAL=$((TOTAL + 1))
    echo -e "${RED}✗${NC} $1"
    if [ ! -z "$2" ]; then
        echo -e "  ${RED}→ $2${NC}"
    fi
}

assert_contains() {
    if echo "$1" | grep -q "$2"; then
        pass "$3"
    else
        fail "$3" "Expected to contain: $2"
    fi
}

assert_not_contains() {
    if echo "$1" | grep -q "$2"; then
        fail "$3" "Should not contain: $2"
    else
        pass "$3"
    fi
}

# ビルド確認
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
AICT="$SCRIPT_DIR/bin/aict"

if [ ! -f "$AICT" ]; then
    echo -e "${YELLOW}Building aict...${NC}"
    cd "$SCRIPT_DIR"
    go build -o bin/aict ./cmd/aict
fi

# テスト用リポジトリ作成
TMPDIR=$(mktemp -d)
ORIGINAL_DIR=$(pwd)
cd "$TMPDIR"

cleanup() {
    cd "$ORIGINAL_DIR"
    rm -rf "$TMPDIR"
}
trap cleanup EXIT

git init -q
git config user.name "TestUser"
git config user.email "test@example.com"

echo ""
echo "=== AI Code Tracker Functional Test ==="
echo "Temp repo: $TMPDIR"
echo ""

# --- Test 1: init ---
echo "--- init ---"
OUTPUT=$($AICT init 2>&1)
assert_contains "$OUTPUT" "initialized successfully" "init: 正常初期化"
assert_contains "$OUTPUT" "TestUser" "init: デフォルト作成者"

# --- Test 2: version ---
echo "--- version ---"
OUTPUT=$($AICT version 2>&1)
assert_contains "$OUTPUT" "1.4.1" "version: バージョン表示"

# --- Test 3: コミット1 (human) ---
echo "--- commit 1: human initial ---"
$AICT checkpoint --author "TestUser" > /dev/null 2>&1

cat > main.go <<'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    fmt.Println("Version: 1.0")
}
EOF

OUTPUT=$($AICT checkpoint --author "TestUser" 2>&1)
assert_contains "$OUTPUT" "1 files" "checkpoint: human変更検出"

git add main.go
git commit -q -m "Initial implementation"
OUTPUT=$($AICT commit 2>&1)
assert_contains "$OUTPUT" "Authorship log created" "commit: Authorship Log生成"

# --- Test 4: コミット2 (AI) ---
echo "--- commit 2: AI utilities ---"
$AICT checkpoint --author "TestUser" > /dev/null 2>&1

cat > utils.go <<'EOF'
package main

import "strings"

func ToUpper(s string) string {
    return strings.ToUpper(s)
}

func ToLower(s string) string {
    return strings.ToLower(s)
}
EOF

OUTPUT=$($AICT checkpoint --author "claude" 2>&1)
assert_contains "$OUTPUT" "claude" "checkpoint: AI作成者記録"

git add utils.go
git commit -q -m "Add string utilities"
$AICT commit > /dev/null 2>&1

# --- Test 5: コミット3 (human) ---
echo "--- commit 3: human addition ---"
$AICT checkpoint --author "TestUser" > /dev/null 2>&1

cat >> main.go <<'EOF'

func greet(name string) string {
    return "Hello, " + name + "!"
}
EOF

$AICT checkpoint --author "TestUser" > /dev/null 2>&1
git add main.go
git commit -q -m "Add greet function"
$AICT commit > /dev/null 2>&1

# --- Test 6: コミット4 (AI) ---
echo "--- commit 4: AI more code ---"
$AICT checkpoint --author "TestUser" > /dev/null 2>&1

cat > helper.go <<'EOF'
package main

import "fmt"

func formatName(first, last string) string {
    return fmt.Sprintf("%s %s", first, last)
}

func repeat(s string, n int) string {
    result := ""
    for i := 0; i < n; i++ {
        result += s
    }
    return result
}
EOF

$AICT checkpoint --author "claude" > /dev/null 2>&1
git add helper.go
git commit -q -m "Add helper functions"
$AICT commit > /dev/null 2>&1

# --- Test 7: コミット5 (human リファクタリング) ---
echo "--- commit 5: human refactor ---"
$AICT checkpoint --author "TestUser" > /dev/null 2>&1

cat > main.go <<'EOF'
package main

import "fmt"

const version = "2.0"

func main() {
    fmt.Printf("App v%s\n", version)
}

func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
EOF

$AICT checkpoint --author "TestUser" > /dev/null 2>&1
git add main.go
git commit -q -m "Refactor main"
$AICT commit > /dev/null 2>&1

# --- Test 8: report (table) ---
echo "--- report: table ---"
OUTPUT=$($AICT report --since 7d 2>&1)
assert_contains "$OUTPUT" "claude" "report table: AI作成者表示"
assert_contains "$OUTPUT" "TestUser" "report table: human作成者表示"
assert_contains "$OUTPUT" "AI生成" "report table: AI貢献セクション"
assert_contains "$OUTPUT" "開発者" "report table: 開発者セクション"

# --- Test 9: report (json) ---
echo "--- report: json ---"
OUTPUT=$($AICT report --since 7d --format json 2>&1)
assert_contains "$OUTPUT" '"name": "claude"' "report json: AI作成者"
assert_contains "$OUTPUT" '"type": "ai"' "report json: AI種別"
assert_contains "$OUTPUT" '"name": "TestUser"' "report json: human作成者"
assert_contains "$OUTPUT" '"type": "human"' "report json: human種別"

# --- Test 10: report (range) ---
echo "--- report: range ---"
OUTPUT=$($AICT report --range HEAD~2..HEAD 2>&1)
assert_contains "$OUTPUT" "Commits: 2" "report range: コミット数"

# --- Test 11: debug commands ---
echo "--- debug ---"
OUTPUT=$($AICT debug show 2>&1)
assert_contains "$OUTPUT" "チェックポイントはありません" "debug show: コミット後は空"

# チェックポイントを手動作成して表示テスト
$AICT checkpoint --author "TestUser" > /dev/null 2>&1
cat >> main.go <<'EOF'
// test line
EOF
$AICT checkpoint --author "TestUser" --message "test msg" > /dev/null 2>&1
OUTPUT=$($AICT debug show 2>&1)
assert_contains "$OUTPUT" "チェックポイント情報" "debug show: チェックポイント表示"

OUTPUT=$($AICT debug clean 2>&1)
assert_contains "$OUTPUT" "削除しました" "debug clean: 削除成功"

OUTPUT=$($AICT debug show 2>&1)
assert_contains "$OUTPUT" "チェックポイントはありません" "debug show: クリーン後は空"

# --- Test 12: help ---
echo "--- help ---"
OUTPUT=$($AICT help 2>&1)
assert_contains "$OUTPUT" "aict init" "help: initコマンド記載"
assert_contains "$OUTPUT" "aict report" "help: reportコマンド記載"
assert_contains "$OUTPUT" "aict debug" "help: debugコマンド記載"

# --- Test 13: error cases ---
echo "--- error cases ---"
OUTPUT=$($AICT sync 2>&1) || true
assert_contains "$OUTPUT" "subcommand required" "error: sync引数なし"

OUTPUT=$($AICT debug 2>&1) || true
assert_contains "$OUTPUT" "subcommand required" "error: debug引数なし"

OUTPUT=$($AICT report --since 7d --range HEAD~1..HEAD 2>&1) || true
assert_contains "$OUTPUT" "mutually exclusive" "error: since+range排他"

# --- 結果表示 ---
echo ""
echo "========================================="
echo -e "Results: ${GREEN}${PASSED} passed${NC}, ${RED}${FAILED} failed${NC}, ${TOTAL} total"
echo "========================================="

if [ $FAILED -gt 0 ]; then
    exit 1
fi
