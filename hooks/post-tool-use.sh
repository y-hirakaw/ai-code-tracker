#!/bin/bash

# AI Code Tracker - PostToolUse Hook  
# Records AI code state after Claude Code makes changes

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

# Read hook input (JSON) from stdin
INPUT=$(cat)

# Extract tool information
TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // "unknown"')
SESSION_ID=$(echo "$INPUT" | jq -r '.session_id // "unknown"')
TOOL_RESPONSE=$(echo "$INPUT" | jq -r '.tool_response // "{}"')

# Check if tool was successful
if echo "$TOOL_RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    echo "AI Code Tracker: Skipping tracking due to tool error" >&2
    exit 0
fi

# Log the post-tool-use event
echo "AI Code Tracker: Recording AI state after $TOOL_NAME (session: ${SESSION_ID:0:8}...)" >&2

# Record AI checkpoint after Claude makes changes
"$AICT_BIN" track -author claude

# Generate and display current report
echo "AI Code Tracker: Current status:" >&2
"$AICT_BIN" report >&2

# Exit successfully
exit 0