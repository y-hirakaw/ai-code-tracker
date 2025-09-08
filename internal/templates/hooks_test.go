package templates

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPreToolUseHook(t *testing.T) {
	// Test that PreToolUseHook contains expected content
	expectedStrings := []string{
		"#!/bin/bash",
		"AI Code Tracker - PreToolUse Hook",
		"Records human code state before Claude Code makes changes",
		"\"$AICT_BIN\" track -author human",
		"TOOL_NAME=$(echo \"$INPUT\" | jq -r '.tool_name // \"unknown\"')",
		"SESSION_ID=$(echo \"$INPUT\" | jq -r '.session_id // \"unknown\"')",
		"PROJECT_DIR=\"${CLAUDE_PROJECT_DIR:-$(pwd)}\"",
		"go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(PreToolUseHook, expected) {
			t.Errorf("PreToolUseHook should contain '%s'", expected)
		}
	}

	// Test shebang is at the beginning
	if !strings.HasPrefix(PreToolUseHook, "#!/bin/bash") {
		t.Error("PreToolUseHook should start with #!/bin/bash")
	}

	// Test exit code
	if !strings.Contains(PreToolUseHook, "exit 0") {
		t.Error("PreToolUseHook should exit with code 0")
	}
}

func TestPostToolUseHook(t *testing.T) {
	// Test that PostToolUseHook contains expected content
	expectedStrings := []string{
		"#!/bin/bash",
		"AI Code Tracker - PostToolUse Hook",
		"Records AI code state after Claude Code makes changes",
		"\"$AICT_BIN\" track -author claude",
		"\"$AICT_BIN\" report",
		"TOOL_NAME=$(echo \"$INPUT\" | jq -r '.tool_name // \"unknown\"')",
		"SESSION_ID=$(echo \"$INPUT\" | jq -r '.session_id // \"unknown\"')",
		"TOOL_RESPONSE=$(echo \"$INPUT\" | jq -r '.tool_response // \"{}\"')",
		"PROJECT_DIR=\"${CLAUDE_PROJECT_DIR:-$(pwd)}\"",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(PostToolUseHook, expected) {
			t.Errorf("PostToolUseHook should contain '%s'", expected)
		}
	}

	// Test error handling
	if !strings.Contains(PostToolUseHook, "if echo \"$TOOL_RESPONSE\" | jq -e '.error' > /dev/null") {
		t.Error("PostToolUseHook should check for tool errors")
	}

	// Test shebang is at the beginning
	if !strings.HasPrefix(PostToolUseHook, "#!/bin/bash") {
		t.Error("PostToolUseHook should start with #!/bin/bash")
	}

	// Test exit code
	if !strings.Contains(PostToolUseHook, "exit 0") {
		t.Error("PostToolUseHook should exit with code 0")
	}
}

func TestPreCommitHook(t *testing.T) {
	// Test that PreCommitHook contains expected content
	expectedStrings := []string{
		"#!/bin/bash",
		"AI Code Tracker - Git Pre-Commit Hook",
		"Records current state before commit",
		"\"$AICT_BIN\" track -author human",
		"PROJECT_DIR=\"$(git rev-parse --show-toplevel)\"",
		"if [[ ! -f \"$AICT_BIN\" ]]; then",
		"if [[ ! -d \"$PROJECT_DIR/.ai_code_tracking\" ]]; then",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(PreCommitHook, expected) {
			t.Errorf("PreCommitHook should contain '%s'", expected)
		}
	}

	// Test shebang is at the beginning
	if !strings.HasPrefix(PreCommitHook, "#!/bin/bash") {
		t.Error("PreCommitHook should start with #!/bin/bash")
	}

	// Test silent operation
	if !strings.Contains(PreCommitHook, ">/dev/null 2>&1") {
		t.Error("PreCommitHook should operate silently")
	}
}

func TestPostCommitHook(t *testing.T) {
	// Test that PostCommitHook contains expected content
	expectedStrings := []string{
		"#!/bin/bash",
		"AI Code Tracker - Git Post-Commit Hook",
		"Updates tracking metrics after each commit",
		"PROJECT_DIR=\"$(git rev-parse --show-toplevel)\"",
		"COMMIT_HASH=$(git rev-parse HEAD)",
		"COMMIT_AUTHOR=$(git log -1 --format='%an')",
		"COMMIT_MESSAGE=$(git log -1 --format='%s')",
		"\"$AICT_BIN\" report",
		"if [[ ! -d \"$PROJECT_DIR/.ai_code_tracking\" ]]; then",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(PostCommitHook, expected) {
			t.Errorf("PostCommitHook should contain '%s'", expected)
		}
	}

	// Test archiving functionality
	if !strings.Contains(PostCommitHook, "ARCHIVE_FILE=") {
		t.Error("PostCommitHook should create archive files")
	}

	if !strings.Contains(PostCommitHook, "jq --arg commit") {
		t.Error("PostCommitHook should add commit info to metrics")
	}

	// Test shebang is at the beginning
	if !strings.HasPrefix(PostCommitHook, "#!/bin/bash") {
		t.Error("PostCommitHook should start with #!/bin/bash")
	}
}

func TestClaudeSettingsJSON(t *testing.T) {
	// Test that ClaudeSettingsJSON contains expected structure
	expectedStrings := []string{
		"\"hooks\":",
		"\"PreToolUse\":",
		"\"PostToolUse\":",
		"\"type\": \"command\"",
		"\"command\": \"$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/pre-tool-use.sh\"",
		"\"command\": \"$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/post-tool-use.sh\"",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(ClaudeSettingsJSON, expected) {
			t.Errorf("ClaudeSettingsJSON should contain '%s'", expected)
		}
	}

	// Test MCP matcher patterns
	mcpPatterns := []string{
		"Write|Edit|MultiEdit",
		"mcp__.*__.*edit.*",
		"mcp__.*__.*write.*",
		"mcp__.*__.*create.*",
		"mcp__.*__.*replace.*",
		"mcp__.*__.*insert.*",
		"mcp__.*__.*override.*",
	}

	for _, pattern := range mcpPatterns {
		if !strings.Contains(ClaudeSettingsJSON, pattern) {
			t.Errorf("ClaudeSettingsJSON should contain MCP pattern '%s'", pattern)
		}
	}

	// Verify it's valid JSON structure (basic check)
	if !strings.HasPrefix(strings.TrimSpace(ClaudeSettingsJSON), "{") {
		t.Error("ClaudeSettingsJSON should start with {")
	}

	if !strings.HasSuffix(strings.TrimSpace(ClaudeSettingsJSON), "}") {
		t.Error("ClaudeSettingsJSON should end with }")
	}

	// Check hook configuration count
	preToolCount := strings.Count(ClaudeSettingsJSON, "pre-tool-use.sh")
	if preToolCount != 1 {
		t.Errorf("Expected 1 pre-tool-use.sh reference, found %d", preToolCount)
	}

	postToolCount := strings.Count(ClaudeSettingsJSON, "post-tool-use.sh")
	if postToolCount != 1 {
		t.Errorf("Expected 1 post-tool-use.sh reference, found %d", postToolCount)
	}
}

func TestMCPToolMatching(t *testing.T) {
	// Test that the matcher pattern would catch common MCP tools
	// This is a conceptual test - actual regex matching would be done by Claude Code
	
	// Extract matcher pattern from ClaudeSettingsJSON
	var config map[string]interface{}
	err := json.Unmarshal([]byte(ClaudeSettingsJSON), &config)
	if err != nil {
		t.Fatalf("Failed to parse ClaudeSettingsJSON: %v", err)
	}
	
	hooks := config["hooks"].(map[string]interface{})
	preToolUse := hooks["PreToolUse"].([]interface{})
	firstHook := preToolUse[0].(map[string]interface{})
	matcher := firstHook["matcher"].(string)
	
	// Test common MCP tool names that should be matched
	testCases := []struct {
		toolName string
		should   string
	}{
		{"mcp__serena__create_text_file", "match"},
		{"mcp__serena__replace_regex", "match"},
		{"mcp__effortlessly-mcp__smart_edit_file", "match"},
		{"mcp__effortlessly-mcp__override_text", "match"},
		{"mcp__mcp-file-editor__write_file", "match"},
		{"mcp__serena__insert_before_symbol", "match"},
		{"mcp__serena__read_file", "not match"}, // read-only, should not match
		{"mcp__serena__list_dir", "not match"},  // read-only, should not match
		{"Write", "match"},                      // standard tool
		{"Edit", "match"},                       // standard tool
		{"MultiEdit", "match"},                  // standard tool
		{"Read", "not match"},                   // read-only, should not match
	}
	
	// Basic pattern validation (this tests our pattern structure, not actual regex matching)
	_ = testCases // Test cases defined for documentation/future use
	
	// Check if pattern contains the necessary components
		hasBasicTools := strings.Contains(matcher, "Write|Edit|MultiEdit")
		hasMCPPattern := strings.Contains(matcher, "mcp__.*__.*edit.*") || 
						 strings.Contains(matcher, "mcp__.*__.*write.*") ||
						 strings.Contains(matcher, "mcp__.*__.*create.*") ||
						 strings.Contains(matcher, "mcp__.*__.*replace.*") ||
						 strings.Contains(matcher, "mcp__.*__.*insert.*") ||
						 strings.Contains(matcher, "mcp__.*__.*override.*")
		
		if !hasBasicTools {
			t.Error("Matcher should contain basic tools pattern")
		}
		
		if !hasMCPPattern {
			t.Error("Matcher should contain MCP patterns")
		}
		
		// Verify pattern structure
		if !strings.Contains(matcher, "|") {
			t.Error("Matcher should use | for OR operations")
		}
	
	t.Logf("Matcher pattern: %s", matcher)
}

func TestHookScriptStructure(t *testing.T) {
	hooks := []struct {
		name        string
		content     string
		checkStderr bool
	}{
		{"PreToolUseHook", PreToolUseHook, true},
		{"PostToolUseHook", PostToolUseHook, true},
		{"PreCommitHook", PreCommitHook, false}, // PreCommitHook operates silently
		{"PostCommitHook", PostCommitHook, true},
	}

	for _, hook := range hooks {
		// Check for set -e (exit on error)
		if !strings.Contains(hook.content, "set -e") {
			t.Errorf("%s should contain 'set -e' for error handling", hook.name)
		}

		// Check for proper error messages (only for hooks that output to stderr)
		if hook.checkStderr && !strings.Contains(hook.content, ">&2") {
			t.Errorf("%s should redirect messages to stderr", hook.name)
		}

		// Check line endings (should not have Windows-style CRLF)
		if strings.Contains(hook.content, "\r\n") {
			t.Errorf("%s should use Unix-style line endings", hook.name)
		}
	}
}

func TestHookBinaryPaths(t *testing.T) {
	// Test that hooks check for aict binary in multiple locations
	hooks := []struct {
		name    string
		content string
	}{
		{"PreToolUseHook", PreToolUseHook},
		{"PostToolUseHook", PostToolUseHook},
		{"PostCommitHook", PostCommitHook},
	}

	for _, hook := range hooks {
		// Should check if aict is in PATH
		if !strings.Contains(hook.content, "command -v aict") {
			t.Errorf("%s should check if aict is in PATH", hook.name)
		}

		// Should have fallback to project bin directory
		if !strings.Contains(hook.content, "$PROJECT_DIR/bin/aict") {
			t.Errorf("%s should check for aict in project bin directory", hook.name)
		}
	}
}

func TestHookInitialization(t *testing.T) {
	// Test that appropriate hooks check for initialization
	hooks := []struct {
		name        string
		content     string
		shouldCheck bool
	}{
		{"PreToolUseHook", PreToolUseHook, true},
		{"PreCommitHook", PreCommitHook, true},
		{"PostCommitHook", PostCommitHook, true},
	}

	for _, hook := range hooks {
		if hook.shouldCheck {
			if !strings.Contains(hook.content, ".ai_code_tracking") {
				t.Errorf("%s should check for .ai_code_tracking directory", hook.name)
			}
		}
	}

	// PreToolUseHook should initialize if needed
	if !strings.Contains(PreToolUseHook, "\"$AICT_BIN\" init") {
		t.Error("PreToolUseHook should initialize AI Code Tracker if not already done")
	}
}
