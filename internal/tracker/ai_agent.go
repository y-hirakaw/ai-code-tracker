package tracker

import "strings"

// DefaultAINames is the list of common AI agent name patterns (case-insensitive substring match)
var DefaultAINames = []string{"claude", "ai", "assistant", "bot", "copilot", "chatgpt"}

// IsAIAgent checks if the author is an AI agent.
// It checks in the following order:
//  1. Exact match against configuredAgents (e.g., Config.AIAgents)
//  2. AuthorMappings alias resolution, then common AI name check
//  3. Common AI name patterns (case-insensitive substring match)
func IsAIAgent(author string, configuredAgents []string, authorMappings map[string]string) bool {
	// 1. 設定ファイルのAIエージェントリストと完全一致でチェック
	for _, agent := range configuredAgents {
		if author == agent {
			return true
		}
	}

	// 2. AuthorMappingsでエイリアス解決
	resolved := author
	if authorMappings != nil {
		if mapping, exists := authorMappings[author]; exists {
			resolved = mapping
		}
	}

	// 3. 一般的なAI名が含まれているかチェック（大文字小文字を区別しない）
	resolvedLower := strings.ToLower(resolved)
	for _, aiName := range DefaultAINames {
		if strings.Contains(resolvedLower, aiName) {
			return true
		}
	}

	return false
}
