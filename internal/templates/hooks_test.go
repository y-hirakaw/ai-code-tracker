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
