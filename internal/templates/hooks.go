package templates

// PreToolUseHook template - records human checkpoint before Claude Code edits
const PreToolUseHook = `#!/bin/bash

# AI Code Tracker - PreToolUse Hook (SPEC.md)
# Records human checkpoint before Claude Code makes edits

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Log file
LOG_FILE="$PROJECT_DIR/.git/aict/hook.log"

# Check if AI Code Tracker is initialized
if [[ ! -d "$PROJECT_DIR/.git/aict" ]]; then
    exit 0
fi

# Try to find aict binary
if command -v aict >/dev/null 2>&1; then
    AICT_BIN="aict"
elif [[ -f "$PROJECT_DIR/bin/aict" ]]; then
    AICT_BIN="$PROJECT_DIR/bin/aict"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] pre-tool-use: aict binary not found" >> "$LOG_FILE"
    exit 0
fi

# Get git user name
GIT_USER=$(git config user.name 2>/dev/null || echo "Developer")

# Record human checkpoint before AI edits
echo "[$(date '+%Y-%m-%d %H:%M:%S')] pre-tool-use: Recording checkpoint for $GIT_USER" >> "$LOG_FILE"
if "$AICT_BIN" checkpoint --author "$GIT_USER" --message "Before Claude Code edits" 2>> "$LOG_FILE"; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] pre-tool-use: Checkpoint recorded successfully" >> "$LOG_FILE"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] pre-tool-use: Failed to record checkpoint (exit code: $?)" >> "$LOG_FILE"
fi

exit 0`

// PostToolUseHook template - records AI checkpoint after Claude Code edits
const PostToolUseHook = `#!/bin/bash

# AI Code Tracker - PostToolUse Hook (SPEC.md)
# Records AI checkpoint after Claude Code edits

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Log file
LOG_FILE="$PROJECT_DIR/.git/aict/hook.log"

# Check if AI Code Tracker is initialized
if [[ ! -d "$PROJECT_DIR/.git/aict" ]]; then
    exit 0
fi

# Try to find aict binary
if command -v aict >/dev/null 2>&1; then
    AICT_BIN="aict"
elif [[ -f "$PROJECT_DIR/bin/aict" ]]; then
    AICT_BIN="$PROJECT_DIR/bin/aict"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] post-tool-use: aict binary not found" >> "$LOG_FILE"
    exit 0
fi

# Record AI checkpoint after edits
echo "[$(date '+%Y-%m-%d %H:%M:%S')] post-tool-use: Recording checkpoint for Claude Code" >> "$LOG_FILE"
if "$AICT_BIN" checkpoint --author "Claude Code" --message "Claude Code edits" 2>> "$LOG_FILE"; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] post-tool-use: Checkpoint recorded successfully" >> "$LOG_FILE"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] post-tool-use: Failed to record checkpoint (exit code: $?)" >> "$LOG_FILE"
fi

exit 0`

// PostCommitHook template - generates Authorship Log after commit
const PostCommitHook = `#!/bin/bash

# AI Code Tracker - Git Post-Commit Hook (SPEC.md)
# Generates Authorship Log from checkpoints

set -e

# Get project directory
PROJECT_DIR="$(git rev-parse --show-toplevel)"

# Try to find aict binary
if command -v aict >/dev/null 2>&1; then
    AICT_BIN="aict"
elif [[ -f "$PROJECT_DIR/bin/aict" ]]; then
    AICT_BIN="$PROJECT_DIR/bin/aict"
else
    exit 0
fi

# Check if AI Code Tracker is initialized
if [[ ! -d "$PROJECT_DIR/.git/aict" ]]; then
    exit 0
fi

# Generate Authorship Log from checkpoints
"$AICT_BIN" commit 2>/dev/null || true

exit 0`

// ClaudeSettingsJSON template for Claude Code hook configuration
// hookスクリプトが存在しない場合でもエラーにならないよう test -x でガード (#5)
const ClaudeSettingsJSON = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit|mcp__.*__.*edit.*|mcp__.*__.*write.*|mcp__.*__.*create.*|mcp__.*__.*replace.*|mcp__.*__.*insert.*|mcp__.*__.*override.*",
        "hooks": [
          {
            "type": "command",
            "command": "test -x \"$CLAUDE_PROJECT_DIR/.git/aict/hooks/pre-tool-use.sh\" && \"$CLAUDE_PROJECT_DIR/.git/aict/hooks/pre-tool-use.sh\" || true"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit|mcp__.*__.*edit.*|mcp__.*__.*write.*|mcp__.*__.*create.*|mcp__.*__.*replace.*|mcp__.*__.*insert.*|mcp__.*__.*override.*",
        "hooks": [
          {
            "type": "command",
            "command": "test -x \"$CLAUDE_PROJECT_DIR/.git/aict/hooks/post-tool-use.sh\" && \"$CLAUDE_PROJECT_DIR/.git/aict/hooks/post-tool-use.sh\" || true"
          }
        ]
      }
    ]
  }
}`
