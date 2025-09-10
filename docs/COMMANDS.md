# AICT Commands Reference

Complete command reference for AI Code Tracker (AICT) v0.6.0.

## 📋 Core Commands

### `aict init`
Initialize tracking in the current directory.

```bash
aict init
```

**What it does:**
- Creates `.ai_code_tracking/` directory
- Generates `config.json` with default settings
- Creates hook scripts (but doesn't install them)
- Detects Git user name for human author mapping

---

### `aict setup-hooks`
Setup Claude Code and Git hooks for automatic tracking.

```bash
aict setup-hooks
```

**What it does:**
- Creates `.claude/settings.json` with hook configuration
- Merges safely with existing Claude settings
- Installs Git post-commit hook
- Enables automatic AI/Human code tracking

---

### `aict track -author <name>`
Create a manual checkpoint for the specified author.

```bash
aict track -author human
aict track -author claude
aict track -author "Custom Name"
```

**Options:**
- `-author <name>`: Specify the author name (required)

**What it does:**
- Analyzes current code state using `git diff`
- Records line changes in JSONL format
- Associates changes with specified author
- Skips automatically if no tracked file changes

---

## 📊 Report Commands

### Basic Reports

#### `aict report`
Display current tracking metrics.

```bash
aict report
```

**Output:** Basic table format with AI/Human ratios and progress.

#### `aict report --format <format>`
Specify output format.

```bash
aict report --format table    # Default table format
aict report --format graph    # ASCII graph visualization
aict report --format json     # JSON output for scripts
aict report --format csv      # CSV for spreadsheet analysis
```

### Period-Based Reports

#### `aict report --last <duration>`
Show report for last N days/weeks/months.

```bash
aict report --last 7d         # Last 7 days
aict report --last 2w         # Last 2 weeks  
aict report --last 1m         # Last 1 month
```

#### `aict report --since <date>`
Show report since specific date/time.

```bash
aict report --since "2025-01-01"
aict report --since "2 weeks ago"
aict report --since "yesterday"
```

#### `aict report --from <date> --to <date>`
Show report for specific date range.

```bash
aict report --from 2025-01-01 --to 2025-01-15
aict report --from "1 week ago" --to "yesterday"
```

### Branch-Based Reports

**Note:** Branch options are mutually exclusive - use only one at a time.

#### `aict report --branch <name>`
Show report for specific branch.

```bash
aict report --branch main
aict report --branch feature/ui-improvements
```

#### `aict report --branch-regex <pattern>`
Show report for branches matching regex pattern.

```bash
aict report --branch-regex "^feature/"        # All feature branches
aict report --branch-regex "(hotfix|bugfix)"  # Hotfix or bugfix branches
aict report --branch-regex "^release/v[0-9]"  # Release branches
```

#### `aict report --branch-pattern <pattern>` *(v0.6.0+)*
Show report for branches matching glob pattern.

```bash
aict report --branch-pattern "feature/*"      # All feature branches
aict report --branch-pattern "*/fix-*"        # Any branch with fix- in name
aict report --branch-pattern "release/v*.*"   # Release version branches
```

#### `aict report --all-branches`
Show summary of all branches with statistics.

```bash
aict report --all-branches
```

### Combined Filtering

You can combine period and branch filtering:

```bash
# Feature branches in last week
aict report --branch-pattern "feature/*" --last 7d

# Main branch since project start
aict report --branch main --since "2024-01-01"

# All branches in date range with graph format
aict report --all-branches --from "2025-01-01" --to "2025-01-31" --format graph
```

---

## 🔧 Configuration Commands

### `aict config`
View and edit configuration settings.

```bash
aict config
```

**Configuration file location:** `.ai_code_tracking/config.json`

**Key settings:**
- `target_ai_percentage`: Target AI code percentage (default: 80%)
- `tracked_extensions`: File extensions to track (e.g., [".go", ".py", ".js"])
- `exclude_patterns`: Files to exclude (e.g., ["*_test.go", "*_generated.go"])
- `author_mappings`: Map Git usernames to human authors

---

## 🔄 Maintenance Commands

### `aict reset`
Reset metrics and create new baseline from current codebase state.

```bash
aict reset
```

**Warning:** This command requires confirmation as it clears all tracking history.

**What it does:**
- Archives current `checkpoints.jsonl`
- Creates fresh baseline from current code state
- Preserves configuration settings

---

### `aict version`
Show version information.

```bash
aict version
```

### `aict help`
Show help information with available commands.

```bash
aict help
```

---

## 🚀 Quick Reference

### Daily Workflow
```bash
# Setup (once per project)
aict init
aict setup-hooks

# Development cycle (automatic with hooks)
# - Write code
# - Commit changes (triggers automatic tracking)

# Check progress
aict report                           # Current status
aict report --last 7d --format graph # Weekly progress graph
```

### Analysis Examples
```bash
# Project overview
aict report --all-branches

# Feature development analysis
aict report --branch-pattern "feature/*" --last 1m

# Performance review preparation
aict report --since "1 month ago" --format csv > monthly_report.csv
```

### Troubleshooting
```bash
# Check configuration
aict config

# Manual tracking if hooks fail
aict track -author claude

# Reset if data corrupted
aict reset
```

---

## 💡 Tips

- **Automatic vs Manual**: Use `aict setup-hooks` for automatic tracking, `aict track` for manual control
- **Pattern Matching**: Glob patterns (`feature/*`) are more intuitive than regex (`^feature/`)
- **Output Formats**: Use `--format csv` for data analysis, `--format graph` for visualization
- **Branch Filtering**: Combine with period filtering for detailed analysis
- **Performance**: Reports generate in <1 second for typical project sizes

---

*For more information, visit the [project repository](https://github.com/y-hirakaw/ai-code-tracker).*