# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
必ず日本語でチャットに返答してください

README.mdだけは英語で記載すること

## Project Overview

AI Code Tracker (AICT) is a Go-based tool designed to track the proportion of AI-generated versus human-written code in a repository. The project integrates with Claude Code hooks and Git post-commit hooks to automatically monitor code generation metrics.

**Current Version**: v1.0.6 (Production ready)

**Key Features**:
- Git notes-based authorship tracking (`refs/aict/authorship`)
- Snapshot-based checkpoint system for human/AI code tracking
- Baseline preservation to exclude existing code from tracking
- Date-based report filtering with `--since` option
- Automatic Claude Code hooks integration
- Table and JSON output formats
- Debug commands for development and testing

## Architecture

Current implementation structure:

```
ai-code-tracker/
├── cmd/aict/              # Main CLI entry point
│   ├── main.go            # CLI commands (init, checkpoint, commit, report, sync, setup-hooks, debug)
│   ├── handlers_*.go      # Command handlers
│   ├── handlers_debug.go  # Debug command handlers
│   └── *_test.go          # Unit tests
├── internal/
│   ├── authorship/        # Authorship line tracking
│   ├── checkpoint/        # Checkpoint management
│   ├── config/            # Configuration handling
│   ├── gitexec/           # Git command execution abstraction (Phase 2)
│   ├── gitnotes/          # Git notes integration (refs/aict/authorship)
│   ├── hooks/             # Hook template generation
│   ├── storage/           # .git/aict/ storage management (Phase 3)
│   └── tracker/           # Core tracking types and analysis (Phase 4 refactored)
├── .git/aict/             # Created by 'aict init'
│   ├── config.json        # Project configuration
│   └── checkpoints/       # Checkpoint snapshots
├── .claude/
│   └── settings.json      # Claude Code hooks configuration
└── test_since_option.sh   # Integration test suite
```

## Development Commands

```bash
# Build the project
go build -o bin/aict ./cmd/aict

# Run unit tests
go test ./...

# Run integration tests
./test_since_option.sh

# Format code
go fmt ./...

# Install dependencies
go mod tidy

# Verify version
./bin/aict version
```

## Core Features (Implemented)

### 1. Checkpoint System
- Records code state before/after Claude Code edits
- Stores in `.git/aict/checkpoints/` as JSON
- Tracks author (human/AI), timestamp, and git diff

### 2. Authorship Tracking
- Uses Git notes (`refs/aict/authorship`) for persistence
- Analyzes git diff to track line changes by author
- Calculates AI vs human code ratios per commit

### 3. Report Generation
- `--range`: Commit range filtering (e.g., `origin/main..HEAD`)
- `--since`: Date-based filtering with shorthand (7d, 2w, 1m, 1y)
- `--format`: Output as table or JSON
- Displays: total lines, AI%, human%, per-author stats, per-file breakdown

### 4. Hook Integration
- Claude Code hooks (pre-tool-use, post-tool-use)
- Git post-commit hook for automatic Authorship Log generation
- Setup via `aict setup-hooks`

### 5. CLI Commands
- `aict init` - Initialize project tracking
- `aict checkpoint --author <name> [--model <name>] [--message <msg>]` - Manual checkpoint
  - `--model` is optional and no longer included in auto-generated hooks
- `aict commit` - Generate Authorship Log from checkpoints
- `aict report --range/--since` - Show statistics
- `aict sync push/fetch` - Sync with remote
- `aict setup-hooks` - Setup automatic tracking
- `aict debug [show|clean|clear-notes]` - Debug and cleanup commands

### 6. Debug Commands (v1.0.3+)
- `aict debug show` - Display checkpoint details (timestamp, author, changes)
- `aict debug clean` - Remove all checkpoint data from `.git/aict/checkpoints/`
- `aict debug clear-notes` - Remove all AICT-related Git notes (refs/notes/aict, refs/aict/authorship, etc.)
- **Use Case**: Clean up test data during development, reset tracking state

## Configuration

`.git/aict/config.json` settings:
- `target_ai_percentage`: Target AI generation rate (default: 80%)
- `tracked_extensions`: File extensions to track (`.go`, `.py`, `.js`, `.ts`, etc.)
- `exclude_patterns`: Patterns to exclude (`*_test.go`, `vendor/*`, etc.)
- `default_author`: Default author name
- `ai_agents`: List of AI agent names (auto-classified as AI)

## Data Flow (with hooks enabled)

1. **Pre-tool-use hook**: Records human checkpoint before Claude Code edits
2. **Claude Code edits**: AI modifications to code
3. **Post-tool-use hook**: Records AI checkpoint after Claude Code edits
4. **Human edits**: Additional manual changes
5. **Git commit**: Triggers post-commit hook
6. **Post-commit hook**: Runs `aict commit` to generate Authorship Log
7. **Authorship Log**: Stored in Git notes (`refs/aict/authorship`)

## Testing

### Integration Tests
Run comprehensive test suite:
```bash
./test_since_option.sh
```

**Test Coverage** (16 tests, 100% pass rate):
- Shorthand notation (7d, 2w, 1m, 1y)
- Relative dates (yesterday, N days ago)
- Absolute dates (2025-01-01)
- Error handling (mutual exclusivity, invalid input)
- Output formats (table, JSON)
- Edge cases (initial commits, very old dates)
- Real-world scenarios (sprint review, daily standup, monthly release)

### Unit Tests
```bash
go test ./...
```

### Quick Functional Test
```bash
# Build and test basic functionality
go build -o bin/aict ./cmd/aict
./bin/aict version                    # v1.0.3
./bin/aict report --since 7d          # Show last 7 days
./bin/aict report --since 2w          # Show last 2 weeks
./bin/aict report --range HEAD~5..HEAD  # Show last 5 commits

# Debug commands
./bin/aict debug show                 # Show checkpoint details
./bin/aict debug clean                # Clean checkpoints
./bin/aict debug clear-notes          # Clear Git notes
```

## Common Use Cases

### Daily Development Review
```bash
aict report --since 1d
```

### Sprint Retrospective (2 weeks)
```bash
aict report --since 2w
```

### PR Review
```bash
aict report --range origin/main..HEAD
```

### Monthly Release Review
```bash
aict report --since 1m
```

### JSON Export
```bash
aict report --since 7d --format json > report.json
```

## 注意事項・制約

### ファイル追跡制約
- **追跡対象**: `.git/aict/config.json`の`tracked_extensions`で設定
- **デフォルト**: `.go`, `.py`, `.js`, `.ts`, `.java`, `.cpp`, `.c`, `.h`, `.rs`
- **除外対象**: `*_test.go`, `vendor/*`, `node_modules/*`など

### チェックポイント記録条件
以下の場合のみチェックポイントが作成されます：
- 追跡対象拡張子のファイルに変更がある
- `git diff --numstat`で変更が検出される
- 前回と異なる変更量（Added/Deleted）

### Git Notes同期と管理
Authorship Logは`refs/aict/authorship`に保存されます:
```bash
# リモートにプッシュ
aict sync push

# リモートから取得
aict sync fetch

# 手動確認
git notes --ref=refs/aict/authorship show HEAD

# Git notesのクリーンアップ（デバッグ用）
aict debug clear-notes  # すべてのaict関連notesを削除
```

**重要**: Git notesは複数のrefに保存される可能性があります:
- `refs/notes/aict`
- `refs/notes/refs/aict/authorship`
- その他"aict"を含むref

`aict debug clear-notes`コマンドはこれらすべてを自動検出して削除します。

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

## リファクタリング進捗状況

プロジェクトは継続的な品質改善を行っています。詳細は `.claude/plans/recursive-churning-corbato.md` を参照してください。

### 完了済みフェーズ

#### フェーズ1: テストインフラの基盤整備 ✅ (2025-12-09完了)
- 共有テストユーティリティ作成 (`internal/testutil/`)
- テストカバレッジ: 73.8%
- スキップされたテスト有効化（`handlers_commit_test.go`, `handlers_range_test.go`）
- 成果: テストセットアップコード60%削減、全テスト通過

#### フェーズ2: Gitコマンド抽象化 ✅ (2025-12-10完了)
- Git実行インターフェース作成 (`internal/gitexec/`)
- 9ファイルで20+箇所のgitコマンド移行
- os/exec依存削除、テスト容易性向上
- 成果: 40行以上の重複削除、全テスト通過

#### フェーズ3: ストレージ抽象化評価 ✅ (2025-12-10評価完了)
- `internal/storage/aict_storage.go` が既に完全実装済み
- 3ファイルのみで使用、適切に抽象化済み
- 結論: 追加作業不要、実質完了

#### フェーズ4.1: 高複雑度関数のリファクタリング ✅ (2025-12-10完了)
**対象**: `internal/tracker/analyzer.go:AnalyzeCheckpoints()`
- **改善前**: CC=11, 145行
- **改善後**: CC=3, 20行（**86%削減、73%超過達成**）

**抽出されたメソッド**:
- `calculatePercentage()` - パーセンテージ計算ヘルパー (6行)
- `aggregateLinesByAuthor()` - 作成者別集計ヘルパー (7行)
- `analyzeFromNumstat()` - チェックポイントNumstatデータ処理 (42行)
- `analyzeFromCommits()` - コミット間git diff処理 (29行)
- `analyzeFromFiles()` - ファイル行比較フォールバック (33行)

**成果**:
- 循環的複雑度: CC=11 → CC=3（目標CC≤7を大幅達成）
- コード行数: 145行 → 20行（86%削減）
- 可読性: 3つの明確な処理パスに分離
- テスト容易性: 各メソッドが独立してテスト可能
- 全テスト通過: 20テスト、0.4秒

#### フェーズ4.2: Numstat解析の集約化 ✅ (2025-12-10完了)
**対象**: Numstat解析ロジックの重複削除

**新規ファイル**:
- `internal/git/numstat.go` (75行) - 統合numstat解析
- `internal/git/numstat_test.go` (174行) - 包括的テスト

**リファクタリング**:
- `checkpoint.go`: `collectNumstatData()` 簡素化（39行 → 7行、82%削減）
- `analyzer.go`: `getGitNumstat()` 簡素化（42行 → 3行、93%削減）
- `diff_test.go`: testutil依存削除、インポートサイクル解決

**成果**:
- コード重複削減: 2ファイルから25行の重複削除
- 保守性向上: 単一箇所への集約
- テストカバレッジ: 12テスト追加、全テスト通過
- アーキテクチャ: 依存関係整理完了

#### フェーズ4.3: 作成者分類インターフェース ⏭️ (スキップ - 循環依存)
**スキップ理由**: `authorship` ↔ `tracker` 間の循環依存により実装不可能

**代替案**: 現在の `IsAIAuthor()` 実装で十分（テスト済み）

#### フェーズ4.4: GetCurrentBranch()簡素化 ⏭️ (スキップ - 既に実施済)
**確認結果**: フェーズ2で `handleDetachedHead()` と `normalizeBranchName()` が既に抽出済み

### 今後の予定フェーズ

#### フェーズ4.5: 複数メトリクス対応 (計画中)
- コードベース貢献、作業量貢献、新規ファイルの3視点測定
- レポート表示の拡張
- 別セッションで実施予定
