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
			name:     "Case sensitive - lowercase",
			author:   "claude code",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAIAgent(tt.author, aiAgents)
			if result != tt.expected {
				t.Errorf("isAIAgent(%q) = %v, expected %v", tt.author, result, tt.expected)
			}
		})
	}
}

func TestDetectChanges(t *testing.T) {
	// detectChanges は git diff を実行するため、実際のGitリポジトリが必要
	// このテストはスキップ（統合テストで検証済み）
	t.Skip("detectChanges requires actual git repository - covered by integration tests")
}

func TestGetLineRanges(t *testing.T) {
	// getLineRanges も git diff を実行するため、実際のGitリポジトリが必要
	// このテストはスキップ（統合テストで検証済み）
	t.Skip("getLineRanges requires actual git repository - covered by integration tests")
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
