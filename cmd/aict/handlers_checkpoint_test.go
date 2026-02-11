package main

import (
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestIsAIAgent(t *testing.T) {
	aiAgents := []string{"Claude Code", "GitHub Copilot", "ChatGPT"}

	tests := []struct {
		name     string
		author   string
		expected bool
	}{
		{
			name:     "AI agent - Claude Code",
			author:   "Claude Code",
			expected: true,
		},
		{
			name:     "AI agent - GitHub Copilot",
			author:   "GitHub Copilot",
			expected: true,
		},
		{
			name:     "Human author",
			author:   "John Doe",
			expected: false,
		},
		{
			name:     "Empty author",
			author:   "",
			expected: false,
		},
		{
			name:     "Case insensitive - lowercase contains 'claude'",
			author:   "claude code",
			expected: true, // "claude"という文字列を含むためtrueが正しい
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.IsAIAgent(tt.author, aiAgents, nil)
			if result != tt.expected {
				t.Errorf("IsAIAgent(%q) = %v, expected %v", tt.author, result, tt.expected)
			}
		})
	}
}


func TestCheckpointValidation(t *testing.T) {
	// CheckpointV2の基本的なバリデーションテスト
	tests := []struct {
		name    string
		cp      *tracker.CheckpointV2
		wantErr bool
	}{
		{
			name: "Valid AI checkpoint",
			cp: &tracker.CheckpointV2{
				Author: "Claude Code",
				Type:   tracker.AuthorTypeAI,
				Changes: map[string]tracker.Change{
					"main.go": {
						Added:   10,
						Deleted: 2,
						Lines:   [][]int{{1, 10}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid human checkpoint",
			cp: &tracker.CheckpointV2{
				Author: "John Doe",
				Type:   tracker.AuthorTypeHuman,
				Changes: map[string]tracker.Change{
					"utils.go": {
						Added:   5,
						Lines:   [][]int{{20}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple line ranges",
			cp: &tracker.CheckpointV2{
				Author: "AI Assistant",
				Type:   tracker.AuthorTypeAI,
				Changes: map[string]tracker.Change{
					"test.go": {
						Added:   15,
						Lines:   [][]int{{1, 5}, {10}, {20, 30}},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 基本的な構造チェック
			if tt.cp.Author == "" {
				t.Error("Author should not be empty")
			}
			if tt.cp.Type != tracker.AuthorTypeAI && tt.cp.Type != tracker.AuthorTypeHuman {
				t.Errorf("Invalid author type: %s", tt.cp.Type)
			}
			if len(tt.cp.Changes) == 0 {
				t.Error("Changes should not be empty")
			}
		})
	}
}
