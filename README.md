# AI Code Tracker (AICT) v0.5.1

A Go-based CLI tool for tracking the proportion of AI-generated versus human-written code with **ultra-lightweight JSONL storage**, integrated with Claude Code and Git.

## 🎯 Features

- **Ultra-Lightweight**: JSONL format reduces storage by 70%+ (~100 bytes per record)
- **Period-Specified Reports**: Analyze AI/Human ratios for specific time ranges (--since, --last, --from/--to)
- **Multiple Output Formats**: Table, ASCII graph, JSON, and **CSV** output formats
- **Automatic Tracking**: Integrated with Claude Code hooks for automatic edit recording
- **Simple Architecture**: No baseline concept - pure differential tracking
- **Accurate Analysis**: Git numstat-based precise line counting
- **Real-time Reporting**: Instant AI/Human ratio calculations with daily breakdown
- **Scalable**: Handles large codebases (10K+ files) efficiently
- **Configurable**: Customizable tracked file extensions and exclusion patterns
- **Smart Skip**: Automatically skips recording when only non-tracked files are modified
- **Test Code Tracking**: Includes test files as legitimate code contributions

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

### 3. Manual Usage

```bash
# Navigate to your project directory first
cd /path/to/your-project

# Record human code state
aict track -author human

# Record AI code state  
aict track -author claude

# Display current statistics (baseline excluded)
aict report

# Period-specified reports (v0.4.0+ features)
aict report --last 7d                    # Last 7 days
aict report --since "2 weeks ago"        # Since 2 weeks ago
aict report --from 2025-01-01 --to 2025-01-15  # Date range
aict report --last 1w --format graph     # ASCII graph format
aict report --last 1m --format json      # JSON format
aict report --last 1w --format csv       # CSV format (v0.5.0 new!)

# Reset tracking from current state (with confirmation)
aict reset
```

### 4. Automatic Usage (Claude Code Integration)

After running `aict setup-hooks`, editing files with Claude Code will automatically track changes:

1. **PreToolUse**: Records human state before Claude edits
2. **PostToolUse**: Records AI state after Claude edits  
3. **Post-commit**: Saves metrics on Git commit

Hook files are created in `.ai_code_tracking/hooks/` with confirmation prompts for existing configurations.

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

### CSV Output (NEW in v0.5.0!)
```csv
Date,AI_Lines,Human_Lines,Total_Lines,AI_Percentage,Human_Percentage,Target_Percentage,Progress
2025-07-30,1458,1209,2667,54.7,45.3,80.0,68.3
2025-07-31,580,0,580,100.0,0.0,80.0,125.0
2025-08-01,227,0,227,100.0,0.0,80.0,125.0
```

**JSONL Record Format** (ultra-lightweight):
```json
{"timestamp":"2025-07-31T23:09:14+09:00","author":"claude","added":395,"deleted":271}
```

## ⚙️ Configuration

Customize settings in `.ai_code_tracking/config.json`:

```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".swift"],
  "exclude_patterns": ["*_generated.go"],
  "author_mappings": {"y-hirakaw": "human"}
}
```

**Note**: Only files with extensions listed in `tracked_extensions` are monitored. Changes to other files (like `.md`, `.txt`) are automatically skipped for efficiency. As of v0.4.0, test files are included as legitimate code contributions.

## 🔧 Claude Code Hooks

`aict setup-hooks` creates `.claude/settings.json` (with merge confirmation for existing settings):

```json
{
  "hooks": [
    {
      "event": "PreToolUse",
      "matcher": "Write|Edit|MultiEdit",
      "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/pre-tool-use.sh"}]
    },
    {
      "event": "PostToolUse", 
      "matcher": "Write|Edit|MultiEdit",
      "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/.ai_code_tracking/hooks/post-tool-use.sh"}]
    }
  ]
}
```

## 📁 File Structure

```
ai-code-tracker/
├── bin/aict                   # CLI executable
├── cmd/aict/
│   ├── main.go               # CLI entry point
│   └── handlers.go           # Period report handlers (v0.4.0+)
├── internal/
│   ├── tracker/              # Core tracking logic
│   ├── period/               # Period analysis (v0.4.0+)
│   ├── storage/              # Data persistence
│   └── git/                  # Git integration
├── .claude/
│   └── settings.json         # Claude Code hook configuration
└── .ai_code_tracking/        # Tracking data
    ├── config.json
    ├── checkpoints.jsonl     # Ultra-lightweight records
    ├── hooks/                # Hook scripts (auto-generated)
    │   ├── pre-tool-use.sh
    │   ├── post-tool-use.sh
    │   └── post-commit
    └── metrics/              # Legacy compatibility
```

## 🎯 Use Cases

### Development Goal Management
- Set target AI code percentage (e.g., 80%)
- Visualize project progress
- Balance code quality and AI utilization

### Team Analysis
- Track AI utilization by team member
- Compare across projects
- Monitor productivity metrics

### Quality Management
- Identify AI code for review
- Ensure human quality checks
- Promote balanced development

## 🔗 Existing Configuration Integration

When existing Git hooks or Claude settings are detected, `aict setup-hooks` safely merges configurations:

```bash
$ aict setup-hooks
Warning: Git post-commit hook already exists at .git/hooks/post-commit
Do you want to merge AI Code Tracker functionality? (y/N): y
✓ Git post-commit hook merged with existing hook

Warning: Claude settings already exist at .claude/settings.json  
Do you want to merge AI Code Tracker hooks? (y/N): y
✓ Claude Code hooks merged with existing settings
```

- **Git Hook Merge**: Appends AICT functionality to existing hooks
- **Claude Settings Merge**: Adds hooks section to existing configuration

## 📋 Commands

| Command | Description |
|---------|-------------|
| `aict init` | Initialize tracking with baseline from existing codebase |
| `aict setup-hooks` | Setup Claude Code and Git hooks (with merge confirmation) |
| `aict track -author <name>` | Create manual checkpoint |
| `aict report [options]` | Display current metrics (baseline excluded) |
| `aict report --last 7d` | Show report for last 7 days |
| `aict report --since "2 weeks ago"` | Show report since specific time |
| `aict report --from DATE --to DATE` | Show report for date range |
| `aict report --format graph` | Show ASCII graph format |
| `aict report --format json` | Show JSON format |
| `aict report --format csv` | Show CSV format (NEW in v0.5.0!) |
| `aict reset` | Reset metrics and create new baseline (with confirmation) |
| `aict version` | Show version information |
| `aict help` | Show help information |
| `aict config` | Show configuration |

## 🔄 Workflow

1. **Initialize**: `aict init` creates baseline from existing code (excluded from metrics)
2. **Setup Hooks**: `aict setup-hooks` enables Claude Code and Git integration
3. **Develop**: Code normally with Claude Code (tracks only changes from baseline)
4. **Monitor**: `aict report` to check progress on added lines only
5. **Analyze**: Use period reports (`--last 1w`, `--format csv`) for detailed analysis
6. **Export**: Use CSV format for Excel/Google Sheets analysis
7. **Reset**: `aict reset` to start fresh tracking from current state (optional)
8. **Adjust**: Modify development strategy to achieve targets

## 🛠️ Technical Specifications

- **Language**: Go 1.21+
- **Dependencies**: Standard library only
- **Data Format**: Ultra-lightweight JSONL (~100 bytes per record)
- **Period Analysis**: Flexible time range filtering with multiple output formats
- **Export Formats**: Table, Graph, JSON, CSV (v0.5.0+)
- **Hooks**: Claude Code hooks, Git post-commit
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