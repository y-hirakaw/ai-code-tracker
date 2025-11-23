package main

import (
	"testing"
)

func TestGetLatestCommitHash(t *testing.T) {
	// getLatestCommitHash は git rev-parse を実行するため、実際のGitリポジトリが必要
	// このテストはスキップ（統合テストで検証済み）
	t.Skip("getLatestCommitHash requires actual git repository - covered by integration tests")
}

func TestHandleCommit(t *testing.T) {
	// handleCommit は複数のシステムコンポーネント（storage, git notes）を使用するため
	// 統合テストで検証済み
	t.Skip("handleCommit is a complex integration - covered by integration tests")
}
