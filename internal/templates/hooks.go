package templates

// PreToolUseHook template - records human checkpoint before Claude Code edits
const PreToolUseHook = `#!/bin/bash

# AI Code Tracker - PreToolUse Hook (SPEC.md)
# Records human checkpoint before Claude Code makes edits

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

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
    exit 0
fi

# Get git user name
GIT_USER=$(git config user.name 2>/dev/null || echo "Developer")

# Record human checkpoint before AI edits
"$AICT_BIN" checkpoint --author "$GIT_USER" --message "Before Claude Code edits" 2>/dev/null || true

exit 0`

// PostToolUseHook template - records AI checkpoint after Claude Code edits
const PostToolUseHook = `#!/bin/bash

# AI Code Tracker - PostToolUse Hook (SPEC.md)
# Records AI checkpoint after Claude Code edits

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

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
    exit 0
fi

# Record AI checkpoint after edits
"$AICT_BIN" checkpoint --author "Claude Code" --message "Claude Code edits" 2>/dev/null || true

exit 0`

// PreCommitHook template - no longer needed for SPEC.md approach
const PreCommitHook = `#!/bin/bash
# AI Code Tracker - Pre-Commit Hook (not used in SPEC.md)
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
const ClaudeSettingsJSON = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit|mcp__.*__.*edit.*|mcp__.*__.*write.*|mcp__.*__.*create.*|mcp__.*__.*replace.*|mcp__.*__.*insert.*|mcp__.*__.*override.*",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.git/aict/hooks/pre-tool-use.sh"
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
            "command": "$CLAUDE_PROJECT_DIR/.git/aict/hooks/post-tool-use.sh"
          }
        ]
      }
    ]
  }
}`
