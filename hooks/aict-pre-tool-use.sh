#!/bin/bash
# AICT PreToolUse Hook Script
# Claude Codeのファイル編集前の状態を記録する

# 実行確認メッセージ（stdoutに出力してClaude Codeに表示される）
echo "🔧 [AICT PreToolUse Hook] 実行開始"

# デバッグ出力
echo "[AICT PreToolUse Hook] Called at $(date)" >&2

# 標準入力からJSONを読み込む
INPUT=$(cat)

# ファイルパスを抽出
FILE=$(echo "$INPUT" | jq -r '.tool_input.path // .tool_input.file_path // empty')
echo "[AICT PreToolUse Hook] FILE=$FILE" >&2

# セッションIDを生成
SESSION_ID=$(date +%s%N | cut -b1-13)

# ファイルが存在する場合のみ処理
if [ -n "$FILE" ] && [ -f "$FILE" ]; then
    echo "[AICT PreToolUse Hook] Running pre-edit with session=$SESSION_ID" >&2
    
    # aictコマンドのパスを探す
    AICT_CMD=""
    if command -v aict >/dev/null 2>&1; then
        AICT_CMD="aict"
    elif [ -x "./aict" ]; then
        AICT_CMD="./aict"
    elif [ -x "$(dirname "$0")/../aict" ]; then
        AICT_CMD="$(dirname "$0")/../aict"
    else
        echo "[AICT PreToolUse Hook] ERROR: aict command not found" >&2
        echo '{"decision": "approve"}'
        exit 0
    fi
    
    echo "[AICT PreToolUse Hook] Using aict command: $AICT_CMD" >&2
    
    # pre-editトラッキングを実行
    $AICT_CMD track --pre-edit --files "$FILE" --session "$SESSION_ID" 2>&1 | sed 's/^/[AICT] /' >&2
    
    # セッションIDを一時ファイルに保存
    echo "$SESSION_ID" > "/tmp/aict-session-$(date +%Y%m%d).tmp"
    echo "[AICT PreToolUse Hook] Session ID saved: $SESSION_ID" >&2
else
    echo "[AICT PreToolUse Hook] No file to track or file not found" >&2
fi

# 処理を承認（実行確認メッセージ付き）
MESSAGE=$(cat << 'EOF'
🔧 [AICT PreToolUse Hook] 編集前状態を記録しました
- ファイル: FILE_PLACEHOLDER
- セッション: SESSION_PLACEHOLDER
EOF
)

ESCAPED_MESSAGE=$(echo "$MESSAGE" | sed "s|FILE_PLACEHOLDER|${FILE}|g" | sed "s|SESSION_PLACEHOLDER|${SESSION_ID}|g" | jq -Rs .)
cat << EOF
{
    "decision": "approve",
    "reason": $ESCAPED_MESSAGE
}
EOF