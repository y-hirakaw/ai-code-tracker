#!/bin/bash

# AI Code Tracker - PreToolUse Hook
# Records human code state before Claude Code makes changes

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"

# Try to find aict binary (in order of preference)
if command -v aict >/dev/null 2>&1; then
    AICT_BIN="aict"
elif [[ -f "$PROJECT_DIR/bin/aict" ]]; then
    AICT_BIN="$PROJECT_DIR/bin/aict"
else
    echo "Warning: AI Code Tracker (aict) not found in PATH or $PROJECT_DIR/bin/aict" >&2
    echo "Please install aict: go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest" >&2
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