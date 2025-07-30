# AI Code Tracker (AICT) v0.3.5

A Go-based CLI tool for tracking the proportion of AI-generated versus human-written code with **ultra-lightweight JSONL storage**, integrated with Claude Code and Git.

## 🎯 Features

- **Ultra-Lightweight**: JSONL format reduces storage by 70%+ (~100 bytes per record)
- **Automatic Tracking**: Integrated with Claude Code hooks for automatic edit recording
- **Simple Architecture**: No baseline concept - pure differential tracking
- **Accurate Analysis**: Git numstat-based precise line counting
- **Real-time Reporting**: Instant AI/Human ratio calculations
- **Scalable**: Handles large codebases (10K+ files) efficiently
- **Configurable**: Customizable tracked file extensions and exclusion patterns
- **Smart Skip**: Automatically skips recording when only non-tracked files are modified

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

# Reset tracking from current state (with confirmation)
aict reset
```

### 4. Automatic Usage (Claude Code Integration)

After running `aict setup-hooks`, editing files with Claude Code will automatically track changes:

1. **PreToolUse**: Records human state before Claude edits
2. **PostToolUse**: Records AI state after Claude edits  
3. **Post-commit**: Saves metrics on Git commit

Hook files are created in `.ai_code_tracking/hooks/` with confirmation prompts for existing configurations.

## 📊 Output Example

```
AI Code Tracking Report
======================
Added Lines: 505
  AI Lines: 236 (46.7%)
  Human Lines: 269 (53.3%)

Target: 80.0% AI code
Progress: 58.4%

Last Updated: 2025-07-30 17:24:59
```

**JSONL Record Format** (ultra-lightweight):
```json
{"timestamp":"2025-07-30T17:24:59+09:00","author":"claude","added":236,"deleted":45}
```

## ⚙️ Configuration

Customize settings in `.ai_code_tracking/config.json`:

```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".swift"],
  "exclude_patterns": ["*_test.go", "*.test.js"],
  "author_mappings": {"y-hirakaw": "human"}
}
```

**Note**: Only files with extensions listed in `tracked_extensions` are monitored. Changes to other files (like `.md`, `.txt`) are automatically skipped for efficiency.

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
├── cmd/aict/main.go          # CLI entry point
├── internal/
│   ├── tracker/              # Core tracking logic
│   ├── storage/              # Data persistence
│   └── git/                  # Git integration
├── .claude/
│   └── settings.json         # Claude Code hook configuration
└── .ai_code_tracking/        # Tracking data
    ├── config.json
    ├── hooks/                # Hook scripts (auto-generated)
    │   ├── pre-tool-use.sh
    │   ├── post-tool-use.sh
    │   └── post-commit
    ├── checkpoints/
    └── metrics/
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
| `aict report` | Display current metrics (baseline excluded) |
| `aict reset` | Reset metrics and create new baseline (with confirmation) |
| `aict version` | Show version information |

## 🔄 Workflow

1. **Initialize**: `aict init` creates baseline from existing code (excluded from metrics)
2. **Setup Hooks**: `aict setup-hooks` enables Claude Code and Git integration
3. **Develop**: Code normally with Claude Code (tracks only changes from baseline)
4. **Monitor**: `aict report` to check progress on added lines only
5. **Reset**: `aict reset` to start fresh tracking from current state (optional)
6. **Adjust**: Modify development strategy to achieve targets

## 🛠️ Technical Specifications

- **Language**: Go 1.21+
- **Dependencies**: Standard library only
- **Data Format**: Ultra-lightweight JSONL (~100 bytes per record)
- **Hooks**: Claude Code hooks, Git post-commit
- **Supported Platforms**: macOS, Linux, Windows
- **Smart Features**: Tracked file extension filtering, Smart skip for efficiency

## 📈 Metrics

Tracked indicators:
- Total line count (including baseline)
- Baseline line count (excluded from metrics)
- AI-generated lines and percentage (of added lines)
- Human-written lines and percentage (of added lines)
- Target achievement rate (based on added lines only)
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