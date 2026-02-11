# AI Code Tracker (AICT) v1.4.1

A Git-integrated tool to measure the proportion of AI-generated vs human-written code in your repository.

## Features

- Tracks AI and human code contributions per commit using Git notes
- Checkpoint-based workflow for recording authorship boundaries
- Date-based (`--since`) and range-based (`--range`) report filtering
- Table and JSON output formats with detailed metrics
- Automatic integration with Claude Code hooks and Git post-commit hooks
- Single binary, no external dependencies

## Installation

```bash
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest
```

Or build from source:

```bash
git clone https://github.com/y-hirakaw/ai-code-tracker.git
cd ai-code-tracker
go build -o bin/aict ./cmd/aict
```

## Quick Start

```bash
# Initialize in your repo
aict init

# Set up automatic hooks (Claude Code + Git post-commit)
aict setup-hooks

# After coding, record a checkpoint
aict checkpoint --author "your-name"

# After AI generates code
aict checkpoint --author "claude"

# Commit triggers automatic Authorship Log generation
git add . && git commit -m "Add feature"

# View report
aict report --since 7d
```

## Commands

| Command | Description |
|---------|-------------|
| `aict init` | Initialize tracking in current repo |
| `aict checkpoint --author <name>` | Record authorship checkpoint |
| `aict commit` | Generate Authorship Log from checkpoints |
| `aict report --since <period>` | Report by date (e.g., `7d`, `2w`, `1m`) |
| `aict report --range <range>` | Report by commit range (e.g., `origin/main..HEAD`) |
| `aict report --format json` | JSON output for CI/CD integration |
| `aict sync push/fetch` | Sync authorship data with remote |
| `aict setup-hooks` | Set up Claude Code and Git hooks |
| `aict debug show/clean/clear-notes` | Debug and cleanup utilities |
| `aict version` | Show version |
| `aict help` | Show help |

## Report Output

```
aict report --since 7d

Commits: 10
期間: 2025-12-01 〜 2025-12-08
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total:  500 lines
  🤖 AI:  350 lines (70.0%)
  👤 Human: 150 lines (30.0%)

By Author:
  claude     300 lines (60.0%)
  developer  200 lines (40.0%)
```

## How It Works

1. **Pre-tool-use hook** records a human checkpoint before Claude Code edits
2. **Post-tool-use hook** records an AI checkpoint after Claude Code edits
3. **Git post-commit hook** runs `aict commit` to generate an Authorship Log
4. Authorship data is stored in Git notes (`refs/aict/authorship`)
5. Reports aggregate data across commits with AI/human classification

## Development

```bash
# Run tests
go test ./...

# Run integration tests
./test_since_option.sh     # --since option tests
./test_functional.sh       # Full workflow E2E test

# Build
go build -o bin/aict ./cmd/aict
```

## License

MIT
