
# Claude Code Hooks 設定

## 設定ファイルの場所

Claude Code hooksは以下のファイルに設定します：
- **ユーザーグローバル設定**: `~/.claude/settings.json`
- **プロジェクトローカル設定**: `./.claude/settings.json`

## 設定例

`~/.claude/settings.json` に以下の内容を追加：

```json
{
  "model": "sonnet",
  "hooks": {
    "preToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'echo \"{\\\"decision\\\": \\\"approve\\\"}\"'"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'echo \"{\\\"decision\\\": \\\"approve\\\"}\"'"
          }
        ]
      }
    ],
    "postToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'INPUT=$(cat); FILE=$(echo \"$INPUT\" | jq -r \".tool_input.path // .tool_input.file_path // empty\"); if [ -n \"$FILE\" ]; then aict track --ai --author \"Claude Code\" --model \"claude-sonnet-4\" --files \"$FILE\" --message \"Claude Code automated edit\" 2>/dev/null || true; fi; echo \"{\\\"continue\\\": true}\"'"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'echo \"{\\\"continue\\\": true}\"'"
          }
        ]
      }
    ],
    "stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'STATS=$(aict stats 2>/dev/null | head -10 || echo \"No stats available\"); echo \"{\\\"continue\\\": true, \\\"userMessage\\\": \\\"📊 AICT Session Stats:\\n$STATS\\\"}\" 2>/dev/null || echo \"{\\\"continue\\\": true}\"'"
          }
        ]
      }
    ],
    "notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'exit 0'"
          }
        ]
      }
    ]
  }
}
```

## 設定手順

1. **Claude Codeを終了**
2. **設定ファイルを編集**:
   ```bash
   # ユーザーグローバル設定
   vim ~/.claude/settings.json
   
   # または、プロジェクトローカル設定
   vim ./.claude/settings.json
   ```
3. **上記のJSONを追加・保存**
4. **Claude Codeを再起動**

## 動作確認

Claude Codeでファイルを編集すると、以下のような動作をします：

1. **Edit/Write/MultiEdit時**:
   - PreToolUse: 編集を承認
   - PostToolUse: AI編集として自動記録

2. **セッション終了時**:
   - Stop: 簡単な統計を表示

3. **通知時**:
   - Notification: 正常終了

## 重要な注意事項

- **実装済みオプションのみ使用**: `--quiet`、`--pre-edit`などの未実装オプションは使用しない
- **パスの確認**: `aict`コマンドがPATHに含まれていることを確認
- **セキュリティ**: Hooksは現在の環境の権限で実行されるため、設定前に内容を確認