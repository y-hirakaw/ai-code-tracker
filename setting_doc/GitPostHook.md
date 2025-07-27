

```bash
#!/bin/bash
# .git/hooks/post-commit
# AI Code Tracker - Gitコミット後の自動トラッキング

# デバッグモード（環境変数で制御）
if [ "$ACT_DEBUG" = "1" ]; then
    set -x
    exec 2>>.git/ai-tracker/hook-debug.log
fi

# AI Code Trackerがインストールされているか確認
if ! command -v aict &> /dev/null; then
    # サイレントに終了（開発フローを妨げない）
    exit 0
fi

# プロジェクトがACTで初期化されているか確認
if [ ! -d ".git/ai-tracker" ]; then
    exit 0
fi

# 重複トラッキング防止機能
check_duplicate() {
    local LOCKFILE=".git/ai-tracker/.commit-lock"
    local CURRENT_TIME=$(date +%s)
    
    # ロックファイルが存在する場合
    if [ -f "$LOCKFILE" ]; then
        local LOCK_TIME=$(cat "$LOCKFILE" 2>/dev/null || echo 0)
        local TIME_DIFF=$((CURRENT_TIME - LOCK_TIME))
        
        # 5秒以内の場合は重複とみなす
        if [ $TIME_DIFF -lt 5 ]; then
            [ "$ACT_DEBUG" = "1" ] && echo "[ACT] Skipping duplicate tracking (${TIME_DIFF}s since last)"
            exit 0
        fi
    fi
    
    # 新しいロックを作成
    echo "$CURRENT_TIME" > "$LOCKFILE"
}

# メイン処理
main() {
    # 重複チェック
    check_duplicate
    
    # コミット情報を取得
    local COMMIT_SHA=$(git rev-parse HEAD)
    local COMMIT_MSG=$(git log -1 --pretty=%B)
    local COMMIT_AUTHOR=$(git log -1 --pretty=%an)
    local COMMIT_EMAIL=$(git log -1 --pretty=%ae)
    
    # Claude Codeのコミットパターン
    local IS_CLAUDE=false
    if [[ "$COMMIT_AUTHOR" =~ ^(Claude|claude) ]] || \
       [[ "$COMMIT_EMAIL" =~ claude ]] || \
       [[ "$COMMIT_MSG" =~ "Claude Code:" ]] || \
       [[ "$COMMIT_MSG" =~ "AI-assisted" ]]; then
        IS_CLAUDE=true
    fi
    
    # 最近のトラッキングイベントを確認
    local LAST_EVENT=$(aict status --json 2>/dev/null | jq -r '.last_event.type // "none"')
    local LAST_TIME=$(aict status --json 2>/dev/null | jq -r '.last_event.timestamp // ""')
    
    # 時間差を計算
    if [ -n "$LAST_TIME" ] && [ "$LAST_TIME" != "null" ]; then
        local LAST_EPOCH=$(date -d "$LAST_TIME" +%s 2>/dev/null || echo 0)
        local CURRENT_EPOCH=$(date +%s)
        local TIME_SINCE_LAST=$((CURRENT_EPOCH - LAST_EPOCH))
    else
        local TIME_SINCE_LAST=999999
    fi
    
    # トラッキング実行の判断
    if [ "$IS_CLAUDE" = true ] && [ "$TIME_SINCE_LAST" -lt 10 ]; then
        # Claude Codeのコミットで、直前にhookが動作していれば、既に記録済みとみなす
        [ "$ACT_DEBUG" = "1" ] && echo "[ACT] Claude commit already tracked by hooks"
    else
        # トラッキング実行
        local TRACK_ARGS=(
            "--commit-ref" "$COMMIT_SHA"
            "--message" "Git commit: ${COMMIT_MSG:0:50}"
            "--quiet"
        )
        
        if [ "$IS_CLAUDE" = true ]; then
            # AIコミットとして記録
            aict track --ai --author "$COMMIT_AUTHOR" --model "claude-3-opus" "${TRACK_ARGS[@]}"
        else
            # 人間のコミットとして記録
            aict track --author "$COMMIT_AUTHOR" "${TRACK_ARGS[@]}"
        fi
        
        # 結果を確認（デバッグ用）
        if [ $? -eq 0 ]; then
            [ "$ACT_DEBUG" = "1" ] && echo "[ACT] Successfully tracked commit $COMMIT_SHA"
        else
            [ "$ACT_DEBUG" = "1" ] && echo "[ACT] Failed to track commit $COMMIT_SHA"
        fi
    fi
    
    # コミット後の統計表示（オプション）
    if [ "$ACT_SHOW_STATS" = "1" ]; then
        echo "───────────────────────────────────────"
        aict stats --brief
        echo "───────────────────────────────────────"
    fi
}

# エラーハンドリング
trap 'rm -f .git/ai-tracker/.commit-lock' EXIT

# メイン処理実行
main

# 正常終了
exit 0
```