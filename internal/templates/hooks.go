package templates

// PreToolUseHook template - no longer needed for git notes approach
const PreToolUseHook = `#!/bin/bash
# AI Code Tracker - PreToolUse Hook (placeholder for compatibility)
exit 0`

// PostToolUseHook template for marking AI edits after Claude Code
const PostToolUseHook = `#!/bin/bash

# AI Code Tracker - PostToolUse Hook
# Marks that AI (Claude Code) made edits, to be recorded on next commit

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Check if AI Code Tracker is initialized
if [[ ! -d "$PROJECT_DIR/.ai_code_tracking" ]]; then
    exit 0
fi

# Read hook input (JSON) from stdin
INPUT=$(cat)

# Extract tool information
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // "unknown"')
TOOL_RESPONSE=$(echo "$INPUT" | jq -r '.tool_response // "{}"')

# Check if tool was successful
if echo "$TOOL_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    exit 0
fi

# Create a marker file indicating AI made edits
# This will be picked up by the post-commit hook
echo "claude" > "$PROJECT_DIR/.ai_code_tracking/.pending_ai_edit"

exit 0`

// PreCommitHook template - no longer needed for git notes approach
const PreCommitHook = `#!/bin/bash
# AI Code Tracker - Pre-Commit Hook (placeholder for compatibility)
exit 0`

// PostCommitHook template for Git post-commit hook
const PostCommitHook = `#!/bin/bash

# AI Code Tracker - Git Post-Commit Hook
# Marks AI-generated code and shows tracking report after commit

set -e

# Get project directory
PROJECT_DIR="$(git rev-parse --show-toplevel)"

# Try to find aict binary
if command -v aict >/dev/null 2>&1; then
    AICT_BIN="aict"
elif [[ -f "$PROJECT_DIR/bin/aict" ]]; then
    AICT_BIN="$PROJECT_DIR/bin/aict"
else
    # Silently exit if aict not found
    exit 0
fi

# Check if AI Code Tracker is initialized
if [[ ! -d "$PROJECT_DIR/.ai_code_tracking" ]]; then
    exit 0
fi

# Get commit information
COMMIT_HASH=$(git rev-parse HEAD)
COMMIT_AUTHOR=$(git log -1 --format='%an')
COMMIT_MESSAGE=$(git log -1 --format='%s')

# Check if post-tool-use hook marked this commit as AI-edited
PENDING_MARKER="$PROJECT_DIR/.ai_code_tracking/.pending_ai_edit"

if [[ -f "$PENDING_MARKER" ]]; then
    # Read the AI tool name from the marker file
    AI_TOOL=$(cat "$PENDING_MARKER" 2>/dev/null || echo "claude")

    # Mark this commit as AI-generated
    "$AICT_BIN" mark-ai-edit --tool "$AI_TOOL" --post-commit 2>/dev/null || true

    # Remove the marker file
    rm -f "$PENDING_MARKER"
fi

echo "AI Code Tracker: Post-commit analysis for $COMMIT_HASH" >&2
echo "Author: $COMMIT_AUTHOR" >&2
echo "Message: $COMMIT_MESSAGE" >&2

# Display current tracking status
"$AICT_BIN" report >&2

exit 0`

// ClaudeSettingsJSON template for Claude Code hook configuration
const ClaudeSettingsJSON = `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit|mcp__.*__.*edit.*|mcp__.*__.*write.*|mcp__.*__.*create.*|mcp__.*__.*replace.*|mcp__.*__.*insert.*|mcp__.*__.*override.*",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/post-tool-use.sh"
          }
        ]
      }
    ]
  }
}`
