# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
必ず日本語でチャットに返答してください

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
│   └── git/               # Git integration
│       └── diff.go        # Git diff processing
├── hooks/
│   ├── aict-pre-tool-use.sh   # Claude Code Pre hook
│   ├── aict-post-tool-use.sh  # Claude Code Post hook
│   └── post-commit             # Git post-commit hook
└── .claude-code/
    └── config.json        # Claude Code configuration
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

The system will use `.ai_tracking/config.json` for configuration:
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