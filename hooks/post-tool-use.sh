#!/bin/bash

# AI Code Tracker - PostToolUse Hook  
# Records AI code state after Claude Code makes changes

set -e

# Get project directory
PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(pwd)}"
AICT_BIN="$PROJECT_DIR/bin/aict"

# Check if aict binary exists
if [[ ! -f "$AICT_BIN" ]]; then
    echo "Warning: AI Code Tracker not found at $AICT_BIN" >&2
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