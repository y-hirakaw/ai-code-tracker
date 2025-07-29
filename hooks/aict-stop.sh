#!/bin/bash
# AICT Stop Hook Script
# セッション終了時に統計情報を表示する

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

# 処理を続行し、ユーザーメッセージに統計情報を含める
echo "{\"continue\": true, \"userMessage\": \"📊 AICT Session: $STATS\"}"