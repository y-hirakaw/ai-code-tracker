#!/bin/bash
# AICT PostToolUse Hook Script
# Claude Codeのファイル編集後の変更を記録する

# 実行確認メッセージ（stdoutに出力してClaude Codeに表示される）
echo "✅ [AICT PostToolUse Hook] 実行開始"

# デバッグ出力
echo "[AICT PostToolUse Hook] Called at $(date)" >&2

# 標準入力からJSONを読み込む
INPUT=$(cat)

# ファイルパスとモデル情報を抽出
FILE=$(echo "$INPUT" | jq -r '.tool_input.path // .tool_input.file_path // empty')
MODEL=$(echo "$INPUT" | jq -r '.metadata.model // "claude-sonnet-4"')
echo "[AICT PostToolUse Hook] FILE=$FILE, MODEL=$MODEL" >&2

MESSAGE="Claude Code automated edit"
SESSION_FILE="/tmp/aict-session-$(date +%Y%m%d).tmp"

# aictコマンドのパスを探す
AICT_CMD=""
if command -v aict >/dev/null 2>&1; then
    AICT_CMD="aict"
elif [ -x "./aict" ]; then
    AICT_CMD="./aict"
elif [ -x "$(dirname "$0")/../aict" ]; then
    AICT_CMD="$(dirname "$0")/../aict"
else
    echo "[AICT PostToolUse Hook] ERROR: aict command not found" >&2
    echo '{"continue": true}'
    exit 0
fi

echo "[AICT PostToolUse Hook] Using aict command: $AICT_CMD" >&2

# ファイルが指定されている場合のみ処理
if [ -n "$FILE" ]; then
    # セッションファイルが存在する場合（pre-editが実行された場合）
    if [ -f "$SESSION_FILE" ]; then
        SESSION_ID=$(cat "$SESSION_FILE" 2>/dev/null)
        echo "[AICT PostToolUse Hook] Found session ID: $SESSION_ID" >&2
        echo "[AICT PostToolUse Hook] Running post-edit tracking..." >&2
        
        # post-editトラッキングを実行
        $AICT_CMD track --post-edit --ai --author "Claude Code" --model "$MODEL" --files "$FILE" --session "$SESSION_ID" --message "$MESSAGE" 2>&1 | sed 's/^/[AICT] /' >&2
        
        # セッションファイルを削除
        rm -f "$SESSION_FILE"
        echo "[AICT PostToolUse Hook] Session file cleaned up" >&2
    else
        # セッションなしの通常トラッキング
        echo "[AICT PostToolUse Hook] No session found, running normal AI tracking..." >&2
        $AICT_CMD track --ai --author "Claude Code" --model "$MODEL" --files "$FILE" --message "$MESSAGE" 2>&1 | sed 's/^/[AICT] /' >&2
    fi
else
    echo "[AICT PostToolUse Hook] No file specified" >&2
fi

# 処理を続行（実行確認メッセージ付き）
MESSAGE=$(cat << 'EOF'
✅ [AICT PostToolUse Hook] AI編集を記録しました
- ファイル: FILE_PLACEHOLDER
- モデル: MODEL_PLACEHOLDER
EOF
)

ESCAPED_MESSAGE=$(echo "$MESSAGE" | sed "s|FILE_PLACEHOLDER|${FILE}|g" | sed "s|MODEL_PLACEHOLDER|${MODEL}|g" | jq -Rs .)
cat << EOF
{
    "continue": true,
    "reason": $ESCAPED_MESSAGE
}
EOF