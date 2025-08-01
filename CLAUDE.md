# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
必ず日本語でチャットに返答してください

README.mdだけは英語で記載すること

## Project Overview

AI Code Tracker (AICT) is a Go-based tool designed to track the proportion of AI-generated versus human-written code in a repository. The project aims to integrate with Claude Code hooks and Git post-commit hooks to automatically monitor code generation metrics and help achieve configurable AI code generation targets.

**Current Status**: This is a newly initialized project with only design documentation. The actual Go implementation has not been started yet.

## Architecture

The planned system follows this structure:

```
ai-code-tracker/
├── cmd/
│   └── aict/              # Main CLI tool
│       └── main.go
├── internal/
│   ├── tracker/           # Core tracking logic
│   │   ├── checkpoint.go  # Checkpoint management
│   │   ├── analyzer.go    # Analysis logic
│   │   └── types.go       # Type definitions
│   ├── storage/           # Data persistence
│   │   ├── json.go        # JSON serialization
│   │   └── metrics.go     # Metrics management
│   ├── git/               # Git integration
│   │   └── diff.go        # Git diff processing
│   └── templates/         # Hook templates
│       └── hooks.go       # Embedded hook scripts
├── .claude-code/
│   └── config.json        # Claude Code configuration
└── .ai_code_tracking/     # AI tracking data directory (created by 'aict init')
    ├── config.json        # Tracking configuration
    ├── hooks/             # Generated hook scripts
    │   ├── pre-tool-use.sh
    │   ├── post-tool-use.sh
    │   └── post-commit
    ├── checkpoints/       # Code snapshots
    └── metrics/           # Tracking metrics
```

## Development Commands

**Note**: These commands are planned but not yet implemented since the Go code doesn't exist.

```bash
# Build the project
go build -o bin/aict ./cmd/aict

# Run tests (when implemented)
go test ./...

# Format code
go fmt ./...

# Run linter (when configured)
golangci-lint run

# Install dependencies
go mod tidy
```

## Key Features to Implement

1. **Checkpoint System**: Record code state before/after Claude Code edits using JSON snapshots
2. **Git Integration**: Analyze git diff to track line changes by author
3. **Metrics Tracking**: Calculate AI vs human code ratios and progress toward targets
4. **Hook Integration**: Automatic tracking via Claude Code hooks and Git post-commit hooks
5. **CLI Interface**: Commands for tracking, analysis, reporting, and configuration

## Configuration

The system will use `.ai_code_tracking/config.json` for configuration:
- Target AI percentage (default: 80%)
- File extensions to track (code files only)
- Exclude patterns (test files, generated files)
- Author mappings (human name from git.user.name, AI tools)

## Data Flow

1. Claude Code Pre-Tool Hook → Record human state checkpoint
2. Claude Code performs edits
3. Claude Code Post-Tool Hook → Record AI state checkpoint  
4. Human makes additional edits
5. Git commit → Post-commit hook analyzes changes and updates metrics

## Implementation Notes

- Use Go's standard library for file operations and JSON handling
- Implement efficient diff parsing to handle large repositories
- Store data in JSON format for simplicity and debuggability
- Focus on code files only (exclude documentation, configuration files)
- Track line additions only (ignore deletions for cleaner metrics)

## Testing Strategy

When implementing:
- Unit tests for core logic (diff parsing, metrics calculation)
- Integration tests for hook workflows
- Performance tests for large repositories
- Test coverage target: 80%+

## AICT使用時の注意事項

**重要**: `aict track`でコード変更を記録するには、設定された拡張子（`.go`, `.py`, `.js`等）のファイルを編集する必要があります。マークダウンファイルやテキストファイルのみの変更では記録されません。

## バージョン更新手順

新しいバージョンをリリースする際は以下の手順に従ってください：

### 1. バージョン番号の更新
```bash
# cmd/aict/main.go の version 定数を更新
# 例: version = "0.3.4" → version = "0.3.5"
```

### 2. README.mdのバージョン更新
```bash
# README.md の先頭タイトルを更新
# 例: # AI Code Tracker (AICT) v0.3.4 → # AI Code Tracker (AICT) v0.3.5
```

### 3. ビルドとテスト
```bash
# プロジェクトをビルド
go build -o bin/aict ./cmd/aict

# バージョン確認
./bin/aict version
```

### 4. 変更のコミットとプッシュ
```bash
# 変更をステージング
git add .

# 詳細なコミットメッセージでコミット
git commit -m "feat: [機能概要] and bump to v[バージョン]

- [変更内容1]
- [変更内容2]
- Bumped version to [バージョン]"

# リモートにプッシュ
git push origin main
```

### 5. タグの作成とプッシュ
```bash
# アノテーション付きタグを作成
git tag -a v[バージョン] -m "Release v[バージョン] - [リリース概要]

- [主要な変更点1]
- [主要な変更点2]"

# タグをリモートにプッシュ
git push origin v[バージョン]
```

### 6. リリース後の確認
```bash
# タグが正しく作成されたことを確認
git tag -l

# go install でインストールできることを確認（新しいターミナルで）
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@v[バージョン]
```

### 注意事項
- バージョン番号はセマンティックバージョニング（major.minor.patch）に従う
- 破壊的変更がある場合はメジャーバージョンを上げる
- 新機能追加の場合はマイナーバージョンを上げる
- バグ修正の場合はパッチバージョンを上げる
- Go Module Proxy のキャッシュ更新には時間がかかる場合がある
