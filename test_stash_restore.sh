#!/bin/bash

# AI Code Tracker - Stash/Restore Integration Test (Issue #8)
# stash/pop後のコミットでAI著者情報が保持されることを検証
#
# 使い方:
#   ./test_stash_restore.sh
#
# 前提:
#   go build -o bin/aict ./cmd/aict

set -e

# CI環境ではカラー出力を無効化
if [ "${CI:-}" = "true" ] || [ "${NO_COLOR:-}" != "" ]; then
    GREEN=''
    RED=''
    YELLOW=''
    NC=''
else
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
fi

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

git init -b main
git config user.email "test@test.com"
git config user.name "Test User"

echo -e "${YELLOW}=== Stash/Restore Integration Tests (Issue #8) ===${NC}"

# =============================================
# セットアップ: ベースラインコミット
# =============================================
echo 'package main' > main.go
echo 'func hello() {}' >> main.go
git add main.go
git commit -m "Initial commit"

$AICT init > /dev/null 2>&1

# post-commit hookが新しいバイナリを使うように上書き
# (PATH上の古いaictではなく、テスト対象のビルド済みバイナリを使用)
cat > .git/hooks/post-commit << HOOK
#!/bin/bash
PROJECT_DIR="\$(git rev-parse --show-toplevel)"
if [[ ! -d "\$PROJECT_DIR/.git/aict" ]]; then
    exit 0
fi
"$AICT" commit 2>/dev/null || true
exit 0
HOOK
chmod +x .git/hooks/post-commit

# =============================================
# Test 1: 基本的な stash → pop → commit でAI帰属が保持される
# =============================================
echo "--- Test 1: Basic stash/pop preserves AI attribution ---"

# AI編集をシミュレート（checkpoint 2つ: human baseline + AI変更）
$AICT checkpoint --author "Developer" > /dev/null 2>&1

# AI がファイルを編集
cat > main.go << 'EOF'
package main

import "fmt"

func hello() {
    fmt.Println("Hello from AI")
}

func newFunc() {
    fmt.Println("New function by AI")
}
EOF

$AICT checkpoint --author "Claude Code" > /dev/null 2>&1

# stash して pop
git stash
git stash pop

# コミット
git add main.go
git commit -m "AI changes via stash/pop"

# レポート確認
REPORT=$($AICT report --range HEAD~1..HEAD 2>&1 || true)
assert_contains "$REPORT" "Claude Code" "Test 1: AI attribution preserved after stash/pop"

# =============================================
# Test 2: stash 中に別コミットが介入してもAI帰属が保持される
# =============================================
echo "--- Test 2: Stash with intervening commit ---"

# AI編集
$AICT checkpoint --author "Developer" > /dev/null 2>&1

cat > main.go << 'EOF'
package main

import "fmt"

func hello() {
    fmt.Println("Modified by AI again")
}

func newFunc() {
    fmt.Println("New function by AI")
}

func anotherFunc() {
    fmt.Println("Another AI function")
}
EOF

$AICT checkpoint --author "Claude Code" > /dev/null 2>&1

# stash
git stash

# 別のファイルで介入コミット
echo 'package main' > utils.go
echo 'func util() {}' >> utils.go
git add utils.go
git commit -m "Human intervening commit"

# pop して コミット
git stash pop

git add main.go
git commit -m "AI changes after intervening commit"

REPORT=$($AICT report --range HEAD~1..HEAD 2>&1 || true)
assert_contains "$REPORT" "Claude Code" "Test 2: AI attribution preserved after intervening commit"

# =============================================
# Test 3: チェックポイントなしの stash pop → Human扱いが正しい
# =============================================
echo "--- Test 3: No checkpoint stash pop defaults to human ---"

# チェックポイントなしで直接編集
cat > main.go << 'EOF'
package main

import "fmt"

func hello() {
    fmt.Println("Human edit only")
}
EOF

git stash
git stash pop

git add main.go
git commit -m "Human only changes"

REPORT=$($AICT report --range HEAD~1..HEAD 2>&1 || true)
assert_not_contains "$REPORT" "Claude Code" "Test 3: No AI attribution without checkpoint"

# =============================================
# Test 4: stash apply（削除せず）でもAI帰属が保持される
# =============================================
echo "--- Test 4: Stash apply preserves AI attribution ---"

$AICT checkpoint --author "Developer" > /dev/null 2>&1

cat > main.go << 'EOF'
package main

import "fmt"

func hello() {
    fmt.Println("Stash apply test")
}

func applyFunc() {
    fmt.Println("Applied function")
}
EOF

$AICT checkpoint --author "Claude Code" > /dev/null 2>&1

git stash
git stash apply

git add main.go
git commit -m "AI changes via stash apply"

REPORT=$($AICT report --range HEAD~1..HEAD 2>&1 || true)
assert_contains "$REPORT" "Claude Code" "Test 4: AI attribution preserved with stash apply"

# stash をクリーンアップ
git stash drop 2>/dev/null || true

# =============================================
# Test 5: 選択的チェックポイント削除が機能する
#   (コミット後にstash関連のチェックポイントが残っていないか確認)
# =============================================
echo "--- Test 5: Consumed checkpoints are removed ---"

$AICT checkpoint --author "Developer" > /dev/null 2>&1

cat > main.go << 'EOF'
package main

import "fmt"

func hello() {
    fmt.Println("Final test")
}
EOF

$AICT checkpoint --author "Claude Code" > /dev/null 2>&1

git add main.go
git commit -m "Normal commit to consume checkpoints"

# debug show でチェックポイントが消えていることを確認
DEBUG_OUT=$($AICT debug show 2>&1 || true)
assert_contains "$DEBUG_OUT" "チェックポイントはありません" "Test 5: Checkpoints consumed after commit"

# =============================================
# 結果サマリー
# =============================================
echo ""
echo -e "${YELLOW}=== Results ===${NC}"
echo -e "Total: $TOTAL  Passed: ${GREEN}$PASSED${NC}  Failed: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
