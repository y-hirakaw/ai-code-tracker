# AI Code Tracker (AICT) v0.7.0

A Go-based CLI tool for tracking the proportion of AI-generated versus human-written code with **Git notes-based authorship tracking**, integrated with Claude Code and Git workflows.

## 🎯 Features

- **Git Notes Integration**: Authorship logs stored in `refs/aict/authorship` for version control
- **Line-Level Tracking**: Precise line range tracking for each code change
- **Checkpoint System**: Record development checkpoints with author and metadata
- **Commit Range Reports**: Analyze AI/Human ratios for commit ranges (`--range origin/main..HEAD`)
- **Author Type Detection**: Automatic classification of AI vs Human contributors
- **Git Workflow Integration**: Seamless integration with git commit workflows
- **Remote Sync**: Push/fetch authorship logs to/from remote repositories
- **Accurate Analysis**: Git diff-based precise line counting with range tracking
- **Configurable**: Customizable tracked file extensions, exclusion patterns, and AI agent list
- **Multiple Output Formats**: Table and JSON output formats for reports
- **SPEC.md Compliant**: Implements detailed specification for enterprise use cases

## 🆕 What's New

### v0.7.0 (SPEC.md Implementation - Breaking Changes)
- **Git Notes-based Storage**: Authorship logs now stored in `refs/aict/authorship` (replaces `.ai_code_tracking/`)
- **New Checkpoint System**: `aict checkpoint` records development checkpoints with line ranges
- **Commit Integration**: `aict commit` generates Authorship Logs from checkpoints (post-commit hook compatible)
- **Range Reports**: `aict report --range <base>..<head>` analyzes commit ranges
- **Author Classification**: Automatic AI vs Human detection based on configurable agent list
- **Remote Sync**: `aict sync push/fetch` to share authorship logs across team
- **Line Range Tracking**: Precise tracking of which lines were modified by each author
- **⚠️ Breaking Change**: This version uses a new storage format and is not backward compatible with v0.6.x

### v0.6.0
- Branch reporting with `--branch-pattern` CLI option for glob-style pattern matching
- Enhanced branch filtering with four mutually exclusive options
- Combined period and branch filtering support

## 🚀 Quick Start

### 1. Installation

#### Option A: Direct Install (Recommended)
```bash
# Install directly from GitHub repository
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest

# Make sure GOPATH/bin is in your PATH
export PATH=$PATH:$(go env GOPATH)/bin
# Add to your shell profile: echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

#### Option B: Build from Source
```bash
# Clone and build AI Code Tracker
git clone https://github.com/y-hirakaw/ai-code-tracker.git
cd ai-code-tracker
go build -o bin/aict ./cmd/aict

# Optional: Add to PATH for global access
export PATH=$PATH:$(pwd)/bin
# Or copy to system location: sudo cp bin/aict /usr/local/bin/
```

### 2. Setup in Your Project

```bash
# Navigate to your project directory
cd /path/to/your-project

# Initialize AI Code Tracker (creates .git/aict/ directory)
aict init
```

### 3. Basic Workflow

```bash
# 1. Make code changes
vim main.go

# 2. Stage your changes
git add .

# 3. Record a checkpoint (before commit)
aict checkpoint --author "Your Name" --message "Implemented feature X"
# Or for AI-generated code:
aict checkpoint --author "Claude Code" --model "claude-sonnet-4" --message "Generated API handlers"

# 4. Commit your changes
git commit -m "Add feature X"

# 5. Generate Authorship Log (auto-run by post-commit hook, or manual)
aict commit

# 6. View statistics for a commit range
aict report --range origin/main..HEAD

# 7. (Optional) Sync authorship logs with remote
aict sync push
```

### 4. Advanced Usage

```bash
# View report in JSON format
aict report --range HEAD~5..HEAD --format json

# Sync authorship logs from remote
aict sync fetch

# View configuration
cat .git/aict/config.json
```

For complete SPEC.md documentation and advanced usage patterns, see **[SPEC.md](SPEC.md)**.

## 📊 Output Examples

### Basic Report
```
AI Code Tracking Report
======================
Added Lines: 395
  AI Lines: 395 (100.0%)
  Human Lines: 0 (0.0%)

Target: 80.0% AI code
Progress: 100.0%

Last Updated: 2025-07-31 23:09:14
```

### Period Report (Table Format)
```
AI Code Tracking Report (Period)
=================================
Period: 2025-07-24 to 2025-07-31
Total Lines: 395
  AI Lines: 395 (100.0%)
  Human Lines: 0 (0.0%)

Target: 80.0% AI code
Progress: 100.0%

Daily Breakdown:
Date       | AI Lines | Human Lines | AI %
-----------+----------+-------------+------
2025-07-31 |      395 |           0 | 100.0
```

### Period Report (Graph Format)
```
AI vs Human Code Contributions (Period)
========================================
Period: 2025-07-30 to 2025-07-31

Daily AI Percentage Trend:
07-31 [██████████████████████████████████████████████████] 100.0% (395/395)

Target [████████████████████████████████████████          ] 80.0%
```

### CSV Output
```csv
Date,AI_Lines,Human_Lines,Total_Lines,AI_Percentage,Human_Percentage,Target_Percentage,Progress
2025-07-30,1458,1209,2667,54.7,45.3,80.0,68.3
2025-07-31,580,0,580,100.0,0.0,80.0,125.0
2025-08-01,227,0,227,100.0,0.0,80.0,125.0
```

**JSONL Record Format** (ultra-lightweight):
```json
{"timestamp":"2025-07-31T23:09:14+09:00","author":"claude","branch":"feature/xyz","commit":"def456","added":395,"deleted":271}
```
Note: `branch` and `commit` may be omitted if unavailable (backward compatible via `omitempty`).

### Branch Reports (Examples)
```
All Branches Report
===================

Overall Statistics:
  Total Records: 128
  Unique Branches: 6
  Records with Branch Info: 120
  Records without Branch Info: 8 (shown as 'main (inferred)')

Group Summary:
  Total Records: 128
  Total Added Lines: 5421
  Group AI Ratio: 78.5% (target: 80.0%)
  Progress: 📊 98.1% of target

Per-Branch Breakdown:
  main: AI 80.2% (812/1012 lines) [23 records]
  feature/abc: AI 76.4% (1350/1767 lines) [41 records]
  hotfix/x: AI 81.0% (210/259 lines) [7 records]
```

```
Branch Pattern Report: "^feature/"
==================================
Matching Branches: feature/abc, feature/xyz
Total Records: 62
Added Lines: 3117 (AI: 2456, Human: 661)
Group AI Ratio: 78.8%
Progress: 📊 98.5% (target: 80.0%)

Per-Branch Breakdown:
  feature/abc: AI 76.4% (1350/1767 lines) [41 records]
  feature/xyz: AI 81.9% (1106/1350 lines) [21 records]
```

## ⚙️ Configuration

Customize settings in `.ai_code_tracking/config.json`:

```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".swift"],
  "exclude_patterns": ["*_test.go", "*_generated.go"],
  "author_mappings": {"y-hirakaw": "human"}
}
```

**Note**: Only files with extensions listed in `tracked_extensions` are monitored. Changes to other files (like `.md`, `.txt`) are automatically skipped for efficiency. By default in v0.5.x, `*_test.go` is excluded; remove it from `exclude_patterns` if you want to include test files.

## 🔧 Claude Code Hooks

`aict setup-hooks` creates `.claude/settings.json` (with merge confirmation for existing settings):

```json
{
  "hooks": [
    {
      "event": "PreToolUse",
      "matcher": "Write|Edit|MultiEdit|mcp__.*__.*edit.*|mcp__.*__.*write.*|mcp__.*__.*create.*|mcp__.*__.*replace.*|mcp__.*__.*insert.*|mcp__.*__.*override.*",
      "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/pre-tool-use.sh"}]
    },
    {
      "event": "PostToolUse", 
      "matcher": "Write|Edit|MultiEdit|mcp__.*__.*edit.*|mcp__.*__.*write.*|mcp__.*__.*create.*|mcp__.*__.*replace.*|mcp__.*__.*insert.*|mcp__.*__.*override.*",
      "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/post-tool-use.sh"}]
    }
  ]
}
```

## 🔗 Existing Configuration Integration

`aict setup-hooks` safely merges with existing Git hooks and Claude settings, preserving existing functionality while adding AICT tracking. When conflicts occur, user confirmation is required before merging, with automatic backup creation when necessary.

## 📋 Commands

For complete command reference, options, and advanced usage examples, see **[COMMANDS.md](docs/COMMANDS.md)**.

## 🔄 Workflow

Initialize with `aict init`, enable hooks via `aict setup-hooks`, then develop and monitor with `aict report`. For advanced usage, see [COMMANDS.md](docs/COMMANDS.md).

## 🛠️ Technical Specifications

- **Language**: Go 1.21+
- **Dependencies**: Standard library only
- **Data Format**: Ultra-lightweight JSONL (~100 bytes per record)
- **Period Analysis**: Flexible time range filtering with multiple output formats
- **Export Formats**: Table, Graph, JSON, CSV (v0.5.0+)
 
- **Supported Platforms**: macOS, Linux, Windows
- **Smart Features**: Tracked file extension filtering, Smart skip for efficiency
- **Test Coverage**: 89.3% for period analysis package

## 📈 Metrics

Tracked indicators:
- Total line count (including baseline)
- Baseline line count (excluded from metrics)
- AI-generated lines and percentage (of added lines)
- Human-written lines and percentage (of added lines)
- Target achievement rate (based on added lines only)
- Daily breakdown with trend analysis (v0.4.0+)
- Period-specific statistics with multiple time ranges
- CSV export for external analysis (v0.5.0+)
- Last update timestamp

## 🔒 Security

- Local filesystem only
- No external communication
- Configurable tracking scope
- Transparent hook execution

## 🤝 Contributing

Issue reports and Pull Requests are welcome.

## 📄 License

MIT License

---

🤖 This project was developed in collaboration with Claude Code.
# Test comment
// Final test
