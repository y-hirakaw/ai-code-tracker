package templates

// PreToolUseHook template for recording human state before Claude edits
const PreToolUseHook = `#!/bin/bash

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
exit 0`

// PostToolUseHook template for recording AI state after Claude edits
const PostToolUseHook = `#!/bin/bash

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
exit 0`

// PreCommitHook template for Git pre-commit hook
const PreCommitHook = `#!/bin/bash

# AI Code Tracker - Git Pre-Commit Hook
# Records current state before commit

set -e

# Get project directory
PROJECT_DIR="$(git rev-parse --show-toplevel)"
AICT_BIN="$PROJECT_DIR/bin/aict"

# Check if aict binary exists
if [[ ! -f "$AICT_BIN" ]]; then
    exit 0
fi

# Check if AI Code Tracker is initialized
if [[ ! -d "$PROJECT_DIR/.ai_code_tracking" ]]; then
    exit 0
fi

# Record current state as human contribution
# This captures any uncommitted changes as human work
"$AICT_BIN" track -author human >/dev/null 2>&1

exit 0`

// PostCommitHook template for Git post-commit hook
const PostCommitHook = `#!/bin/bash

# AI Code Tracker - Git Post-Commit Hook
# Updates tracking metrics after each commit

set -e

# Get project directory (Git hook doesn't have CLAUDE_PROJECT_DIR)
PROJECT_DIR="$(git rev-parse --show-toplevel)"

# Try to find aict binary (in order of preference)
if command -v aict >/dev/null 2>&1; then
    AICT_BIN="aict"
elif [[ -f "$PROJECT_DIR/bin/aict" ]]; then
    AICT_BIN="$PROJECT_DIR/bin/aict"
else
    # Silently exit if aict not found (post-commit hook should not be noisy)
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

echo "AI Code Tracker: Post-commit analysis for $COMMIT_HASH" >&2
echo "Author: $COMMIT_AUTHOR" >&2
echo "Message: $COMMIT_MESSAGE" >&2

# Display current tracking status
"$AICT_BIN" report >&2

# Archive current metrics with commit info
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
METRICS_FILE="$PROJECT_DIR/.ai_code_tracking/metrics/current.json"
ARCHIVE_FILE="$PROJECT_DIR/.ai_code_tracking/metrics/archive/commit_${COMMIT_HASH:0:8}_${TIMESTAMP}.json"

if [[ -f "$METRICS_FILE" ]]; then
    mkdir -p "$(dirname "$ARCHIVE_FILE")"
    # Add commit info to archived metrics
    jq --arg commit "$COMMIT_HASH" --arg author "$COMMIT_AUTHOR" --arg message "$COMMIT_MESSAGE" \
       '. + {"commit_hash": $commit, "commit_author": $author, "commit_message": $message}' \
       "$METRICS_FILE" > "$ARCHIVE_FILE"
    echo "AI Code Tracker: Metrics archived to $ARCHIVE_FILE" >&2
fi

exit 0`

// ClaudeSettingsJSON template for Claude Code hook configuration
const ClaudeSettingsJSON = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/pre-tool-use.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
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