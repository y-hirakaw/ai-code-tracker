
```json
{
  "hooks": {
    "preToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'INPUT=$(cat); FILE=$(echo \"$INPUT\" | jq -r \".tool_input.path // .tool_input.file_path // empty\"); if [ -n \"$FILE\" ] && [ -f \"$FILE\" ]; then aict track --quiet --pre-edit --files \"$FILE\" 2>/dev/null && echo \"{\\\"decision\\\": \\\"approve\\\"}\" || echo \"{\\\"decision\\\": \\\"approve\\\"}\"; else echo \"{\\\"decision\\\": \\\"approve\\\"}\"; fi'"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'aict track --quiet --pre-command 2>/dev/null; echo \"{\\\"decision\\\": \\\"approve\\\"}\"'"
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
            "command": "bash -c 'INPUT=$(cat); FILE=$(echo \"$INPUT\" | jq -r \".tool_input.path // .tool_input.file_path // empty\"); if [ -n \"$FILE\" ]; then aict track --quiet --ai --author \"Claude Code\" --model \"claude-sonnet-4\" --files \"$FILE\" 2>/dev/null || true; fi; echo \"{\\\"continue\\\": true}\"'"
          }
        ]
      },
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'INPUT=$(cat); CMD=$(echo \"$INPUT\" | jq -r \".tool_input.command // empty\"); aict track --quiet --ai --author \"Claude Code\" --model \"claude-sonnet-4\" --command \"$CMD\" 2>/dev/null || true; echo \"{\\\"continue\\\": true}\"'"
          }
        ]
      }
    ],
    "stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'if [ \"$(echo \"$1\" | jq -r \".stop_hook_active // false\")\" = \"false\" ]; then STATS=$(aict stats --format json --session 2>/dev/null || echo \"{}\"); if [ \"$STATS\" != \"{}\" ]; then AI_LINES=$(echo \"$STATS\" | jq -r \".session.ai_lines // 0\"); HUMAN_LINES=$(echo \"$STATS\" | jq -r \".session.human_lines // 0\"); TOTAL=$((AI_LINES + HUMAN_LINES)); if [ $TOTAL -gt 0 ]; then PERCENT=$((AI_LINES * 100 / TOTAL)); echo \"{\\\"continue\\\": true, \\\"userMessage\\\": \\\"📊 Session: AI: ${AI_LINES} lines (${PERCENT}%), Human: ${HUMAN_LINES} lines\\\"}\"; else echo \"{\\\"continue\\\": true}\"; fi; else echo \"{\\\"continue\\\": true}\"; fi; else echo \"{\\\"continue\\\": true}\"; fi'"
          }
        ]
      }
    ],
    "notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "bash -c 'INPUT=$(cat); MSG=$(echo \"$INPUT\" | jq -r \".message // empty\"); if [[ \"$MSG\" == *\"idle\"* ]] || [[ \"$MSG\" == *\"permission\"* ]]; then aict track --quiet --checkpoint \"idle\" 2>/dev/null || true; fi; exit 0'"
          }
        ]
      }
    ]
  }
}
```