# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
必ず日本語でチャットしてください。
日本語が含まれるファイルはUTF-8で保存してください。
コードのコメントは日本語で記載すること
作業完了時に`.claude/workspace/task.md`を更新すること

## Project Overview

AI Code Tracker (AICT) is a system that automatically tracks and distinguishes between AI-generated code and human-written code using Claude Code hooks and Git hooks. The project is designed to integrate seamlessly with existing Git workflows without disrupting development flow.

## Architecture

This is currently a planning phase project with requirements defined in `RDD.md`. The intended architecture will be:

- **Core Language**: Go (planned)
- **CLI Tool**: `aict` command for tracking and reporting
- **Integration Points**: 
  - Claude Code hooks (pre/post tool use)
  - Git post-commit hooks
- **Data Storage**: JSONL format in `.git/ai-tracker/`
- **Hook Configuration**: JSON configuration files for Claude Code integration

## Key Components (Planned)

### Directory Structure
```
cmd/aict/           # CLI entry point
internal/
  ├── tracker/      # Core tracking logic
  ├── hooks/        # Hook processing
  ├── blame/        # Enhanced blame functionality
  ├── stats/        # Statistics and reporting
  └── storage/      # JSONL and index management
pkg/types/          # Common type definitions
scripts/            # Installation and setup scripts
```

### Data Model
- **TrackEvent**: Core tracking event with timestamp, author, event type, and file changes
- **FileInfo**: Per-file change information (lines added/modified/deleted)
- **Storage**: JSONL format with index for fast queries

## Integration Setup

### Claude Code Hooks
The project includes pre-configured Claude Code hooks in `setting_doc/ClaudeCodeHooks.md`:
- **preToolUse**: Tracks file state before AI edits
- **postToolUse**: Records AI changes after edits
- **stop**: Shows session statistics
- **notification**: Handles idle/permission events

### Git Hooks
Post-commit hook configuration in `setting_doc/GitPostHook.md`:
- Automatic tracking of commits
- Duplicate prevention (5-second window)
- Claude Code commit detection
- Optional statistics display

## Development Commands

```bash
go build ./cmd/aict          # Build the CLI tool
go test ./...               # Run tests
go mod tidy                 # Manage dependencies
```

## Release Process

**重要**: リリース時は以下の順序を厳守すること：

1. **バージョン定数を更新**:
   ```bash
   # internal/cli/app.go の Version 定数を更新
   Version = "X.Y.Z"
   
   # internal/web/handlers/simple_api.go のバージョンも同期更新
   "version": "X.Y.Z"
   ```

2. **コミットしてからタグ作成**:
   ```bash
   git add -A
   git commit -m "🔖 バージョンをX.Y.Zに更新"
   git tag vX.Y.Z
   git push origin main
   git push origin vX.Y.Z
   ```

3. **確認**:
   ```bash
   go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@vX.Y.Z
   $(go env GOPATH)/bin/aict version  # "aict version X.Y.Z" が表示されることを確認
   ```

**注意**: タグを作成してからバージョン定数を変更すると、タグとバージョン表示に不整合が生じる。

## Key Features (Planned)

### Automatic Tracking
- Pre/post edit hooks capture AI changes
- Git integration tracks final commit state
- Non-invasive operation (doesn't disrupt workflow)

### CLI Interface
- `aict track` - Manual tracking (usually automated)
- `aict blame <file>` - Enhanced blame with AI/human attribution
- `aict stats` - Statistics and reporting
- `aict init` - Project initialization
- `aict setup` - Hook configuration

### Data Management
- JSONL format for append-only tracking
- Index files for fast queries
- Statistics caching for performance
- Automatic cleanup of old data

## Configuration

The system will use:
- `.git/ai-tracker/` for data storage
- JSON configuration for Claude Code hooks
- Environment variables for debugging (`ACT_DEBUG=1`)
- Optional statistics display (`ACT_SHOW_STATS=1`)

## Important Notes

- This project is in the requirements/planning phase
- No actual Go code exists yet - only specifications in `RDD.md`
- Hook configurations are ready for implementation
- Focus on non-invasive, automatic operation
- Designed specifically for Claude Code integration