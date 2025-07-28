
```json
{
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