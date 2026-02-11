package templates

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHooksContent(t *testing.T) {
	// Verify hooks start with shebang
	hooks := []string{PreToolUseHook, PostToolUseHook, PostCommitHook}

	for i, hook := range hooks {
		if !strings.HasPrefix(hook, "#!/bin/bash") {
			t.Errorf("Hook %d should start with #!/bin/bash", i)
		}
	}
}

func TestSettingsJSON(t *testing.T) {
	// Verify ClaudeSettingsJSON is valid JSON
	var js map[string]interface{}
	err := json.Unmarshal([]byte(ClaudeSettingsJSON), &js)
	if err != nil {
		t.Errorf("ClaudeSettingsJSON is not valid JSON: %v", err)
	}

	// Verify required fields exist
	hooks, ok := js["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks field missing or invalid")
	}

	if _, ok := hooks["PreToolUse"]; !ok {
		t.Error("PreToolUse hook missing")
	}
	if _, ok := hooks["PostToolUse"]; !ok {
		t.Error("PostToolUse hook missing")
	}
}

func TestHooksExitCleanly(t *testing.T) {
	hooks := map[string]string{
		"PreToolUseHook":  PreToolUseHook,
		"PostToolUseHook": PostToolUseHook,
		"PostCommitHook":  PostCommitHook,
	}

	for name, hook := range hooks {
		if !strings.HasSuffix(strings.TrimSpace(hook), "exit 0") {
			t.Errorf("%s should end with 'exit 0'", name)
		}
	}
}

func TestHooksContainAICTBinaryDetection(t *testing.T) {
	hooks := map[string]string{
		"PreToolUseHook":  PreToolUseHook,
		"PostToolUseHook": PostToolUseHook,
		"PostCommitHook":  PostCommitHook,
	}

	for name, hook := range hooks {
		if !strings.Contains(hook, "command -v aict") {
			t.Errorf("%s should contain 'command -v aict' for binary detection", name)
		}
		if !strings.Contains(hook, "bin/aict") {
			t.Errorf("%s should contain 'bin/aict' fallback path", name)
		}
	}
}

func TestPostCommitHookUsesGitRevParse(t *testing.T) {
	if !strings.Contains(PostCommitHook, "git rev-parse --show-toplevel") {
		t.Error("PostCommitHook should use 'git rev-parse --show-toplevel' for repo root detection")
	}
}

func TestHooksCheckAICTInitialized(t *testing.T) {
	hooks := map[string]string{
		"PreToolUseHook":  PreToolUseHook,
		"PostToolUseHook": PostToolUseHook,
		"PostCommitHook":  PostCommitHook,
	}

	for name, hook := range hooks {
		if !strings.Contains(hook, ".git/aict") {
			t.Errorf("%s should check for .git/aict directory", name)
		}
	}
}
