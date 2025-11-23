# AI Code Tracker (AICT) v0.6.1

A Go-based CLI tool for tracking the proportion of AI-generated versus human-written code with **ultra-lightweight JSONL storage**, integrated with Claude Code and Git.

## 🎯 Features

- **Ultra-Lightweight**: JSONL format reduces storage by 70%+ (~150 bytes per record)
- **Branch Statistics**: Analyze AI/Human ratios by branch with regex and glob pattern matching
- **Period-Specified Reports**: Analyze AI/Human ratios for specific time ranges (--since, --last, --from/--to)
- **Multiple Output Formats**: Table, ASCII graph, JSON, and **CSV** output formats
- **Automatic Tracking**: Integrated with Claude Code hooks for automatic edit recording
- **Simple Architecture**: No baseline concept - pure differential tracking
- **Accurate Analysis**: Git numstat-based precise line counting
- **Fast Reporting**: Sub-second AI/Human ratio calculations with daily breakdown
- **Scalable**: Handles large codebases efficiently with optimized JSONL storage
- **Configurable**: Customizable tracked file extensions and exclusion patterns
- **Smart Skip**: Automatically skips recording when only non-tracked files are modified
- **Test File Handling**: `*_test.go` is excluded by default (configurable)

## 🆕 What's New

### v0.6.0
- Complete branch reporting Phase 5: `--branch-pattern` CLI option for glob-style pattern matching.
- Intelligent pattern detection automatically chooses between exact, regex, and glob matching.
- Enhanced branch filtering with four mutually exclusive options for maximum flexibility.
- Combined period and branch filtering support for complex reporting scenarios.
- Comprehensive test coverage with 133+ test cases and 95%+ coverage.

### v0.5.4
- Complete branch reporting Phase 2: `--branch`, `--branch-regex`, `--all-branches` CLI options.
- Regex-based branch grouping with per-branch breakdown and group summary.
- Overall record stats now include counts for records without branch info (shown as `main (inferred)`).
- Improved Git branch detection and normalization (handles detached HEAD and remotes).
- Validation for mutually exclusive branch flags with clear error messages.

### v0.5.3
- Introduced branch-aware JSONL records (`branch` field) and foundational analysis APIs.
- Internal plumbing for future CLI branch reporting.

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

# Initialize AI Code Tracker (creates .ai_code_tracking/ directory)
aict init

# Setup hooks for automatic tracking with Claude Code and Git
aict setup-hooks
```

### 3. Basic Usage

```bash
# Navigate to your project directory first
cd /path/to/your-project

# Record code changes manually (if hooks not used)
aict track -author human   # After human coding
aict track -author claude  # After AI coding

# View current statistics
aict report

# View progress over time
aict report --last 7d
```

For complete command reference and advanced usage, see **[COMMANDS.md](docs/COMMANDS.md)**.

### 4. Automatic Usage

Automatic tracking is enabled by `aict setup-hooks`. For hook details and MCP matchers, see the "Claude Code Hooks" section below.

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
