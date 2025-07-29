#!/bin/bash
# AICT Stop Hook Script
# セッション終了時に統計情報を表示する

# 実行確認メッセージ（stdoutに出力してClaude Codeに表示される）
echo "📊 [AICT Stop Hook] セッション終了時の統計表示"

# デバッグ出力
echo "[AICT Stop Hook] Called at $(date)" >&2

# aictコマンドのパスを探す
AICT_CMD=""
if command -v aict >/dev/null 2>&1; then
    AICT_CMD="aict"
elif [ -x "./aict" ]; then
    AICT_CMD="./aict"
elif [ -x "$(dirname "$0")/../aict" ]; then
    AICT_CMD="$(dirname "$0")/../aict"
else
    echo "[AICT Stop Hook] ERROR: aict command not found" >&2
    echo '{"continue": true}'
    exit 0
fi

echo "[AICT Stop Hook] Using aict command: $AICT_CMD" >&2

# 統計情報を取得
STATS=$($AICT_CMD stats --format summary 2>&1 | tr '\n' ' ' || echo "No stats available")
echo "[AICT Stop Hook] Stats: $STATS" >&2

# 処理を続行し、統計情報を詳細表示
MESSAGE=$(cat << EOF
📊 [AICT Stop Hook] セッション統計
$STATS
EOF
)

ESCAPED_MESSAGE=$(echo "$MESSAGE" | jq -Rs .)
cat << EOF
{
    "continue": true,
    "userMessage": $ESCAPED_MESSAGE
}
EOF