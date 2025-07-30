#!/bin/bash

# AI Code Tracker - PreToolUse Hook
# Records human code state before Claude Code makes changes

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
AICT_BIN="$PROJECT_DIR/bin/aict"

# Check if aict binary exists
if [[ ! -f "$AICT_BIN" ]]; then
    echo "Warning: AI Code Tracker not found at $AICT_BIN" >&2
    exit 0
fi

# Initialize if not already done
if [[ ! -d "$PROJECT_DIR/.ai_code_tracking" ]]; then
    echo "Initializing AI Code Tracker..." >&2
    "$AICT_BIN" init
fi

# Read hook input (JSON) from stdin
INPUT=$(cat)

# Extract tool information
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // "unknown"')
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // "unknown"')

# Log the pre-tool-use event
echo "AI Code Tracker: Recording human state before $TOOL_NAME (session: ${SESSION_ID:0:8}...)" >&2

# Record human checkpoint before Claude makes changes
"$AICT_BIN" track -author human

# Exit successfully to allow tool execution
exit 0