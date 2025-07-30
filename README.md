# AI Code Tracker (AICT)

A Go-based CLI tool for tracking the proportion of AI-generated versus human-written code, integrated with Claude Code and Git.

## 🎯 Features

- **Automatic Tracking**: Integrated with Claude Code hooks for automatic edit recording
- **Accurate Analysis**: Precise line counting through checkpoint-based differential analysis  
- **Real-time Reporting**: Target achievement rate and detailed statistics display
- **Configurable**: Customizable tracked file extensions and exclusion patterns
- **Lightweight**: Efficient data storage in JSON format

## 🚀 Quick Start

### 1. Setup

```bash
# Clone repository
git clone https://github.com/y-hirakaw/ai-code-tracker.git
cd ai-code-tracker

# Build
go build -o bin/aict ./cmd/aict

# Initialize (creates configuration and hook files)
./bin/aict init

# Setup hooks (enables Claude Code and Git integration)
./bin/aict setup-hooks
```

### 2. Manual Usage

```bash
# Record human code state
./bin/aict track -author human

# Record AI code state  
./bin/aict track -author claude

# Display current statistics
./bin/aict report
```

### 3. Automatic Usage (Claude Code Integration)

After running `aict setup-hooks`, editing files with Claude Code will automatically track changes:

1. **PreToolUse**: Records human state before Claude edits
2. **PostToolUse**: Records AI state after Claude edits  
3. **Post-commit**: Saves metrics on Git commit

Hook files are created in `.ai_code_tracking/hooks/` with confirmation prompts for existing configurations.

## 📊 Output Example

```
AI Code Tracking Report
======================
Total Lines: 817
AI Lines: 14 (1.7%)
Human Lines: 803 (98.3%)

Target: 80.0% AI code
Progress: 2.1%

Last Updated: 2025-07-30 16:04:08
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

## 🔄 Workflow

1. **Initialize**: `aict init` creates project configuration and files
2. **Setup Hooks**: `aict setup-hooks` enables Claude Code and Git integration
3. **Develop**: Code normally with Claude Code (automatic tracking)
4. **Monitor**: `aict report` to check progress
5. **Adjust**: Modify development strategy to achieve targets

## 🛠️ Technical Specifications

- **Language**: Go 1.21+
- **Dependencies**: Standard library only
- **Data Format**: JSON
- **Hooks**: Claude Code hooks, Git post-commit
- **Supported Platforms**: macOS, Linux, Windows

## 📈 Metrics

Tracked indicators:
- Total line count
- AI-generated lines and percentage
- Human-written lines and percentage
- Target achievement rate
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