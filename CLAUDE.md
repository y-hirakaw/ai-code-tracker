# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.
必ず日本語でチャットに返答してください

## Project Overview

AI Code Tracker (AICT) is a Go-based tool designed to track the proportion of AI-generated versus human-written code in a repository. The project integrates with Claude Code hooks and Git post-commit hooks to automatically monitor code generation metrics.

**Current Version**: v1.5.1-beta.1

**Key Features**:
- Git notes-based authorship tracking (`refs/aict/authorship`)
- Snapshot-based checkpoint system for human/AI code tracking
- Baseline preservation to exclude existing code from tracking
- Stash/restore時のAI帰属情報保全
- Date-based report filtering with `--since` option
- Automatic Claude Code hooks integration
- Table and JSON output formats
- Debug commands for development and testing
- チェックポイントファイルのアドバイザリロック（TOCTOU防止）

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
│   ├── authorship/        # Authorship Log構築・パース
│   ├── git/               # numstat解析ユーティリティ
│   ├── gitexec/           # Git実行抽象化・モックサポート
│   ├── gitnotes/          # Git notes操作 (refs/aict/authorship)
│   ├── storage/           # .git/aict/ ストレージ管理
│   ├── templates/         # Hook/設定テンプレート定数
│   ├── testutil/          # テスト共通ユーティリティ
│   └── tracker/           # 追跡型定義・分析エンジン
├── .git/aict/             # Created by 'aict init'
│   ├── config.json        # Project configuration
│   └── checkpoints/       # Checkpoint snapshots
├── .claude/
│   └── settings.json      # Claude Code hooks configuration
├── test_since_option.sh   # --since option integration tests
└── test_functional.sh     # Full functional test (multi-commit workflow)
```

## Development Commands

```bash
# Build the project
go build -o bin/aict ./cmd/aict

# Run unit tests
go test ./...

# Run integration tests
./test_since_option.sh     # --since option tests (16 tests)
./test_functional.sh       # Full workflow test (25 tests)

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
- Stores in `.git/aict/checkpoints/` as JSONL (旧JSON array形式も自動マイグレーション)
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

```bash
# ユニットテスト
go test ./...

# 統合テスト
./test_since_option.sh     # --since option tests (16 tests)
./test_functional.sh       # Full workflow test (25 tests) ※リファクタ後に実行推奨
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

### 更新対象ファイル
以下の全ファイルでバージョン番号を更新（`replace_all`推奨）:
- `cmd/aict/main.go` (version定数)
- `README.md`, `CLAUDE.md`, `SPEC.md`, `IMPLEMENTATION_STATUS.md`
- `test_functional.sh`
- `docs/USAGE.md`, `docs/BASE_SPEC.md`, `docs/DATA_FLOW.md`

### リリース手順
```bash
# 1. 全ファイルのバージョン更新後
go build -o bin/aict ./cmd/aict && ./bin/aict version

# 2. テスト
go test ./... && ./test_functional.sh

# 3. コミット・タグ・プッシュ
git add <files> && git commit -m "chore: bump version to v[VERSION]"
git tag -a v[VERSION] -m "Release v[VERSION] - [概要]"
git push origin main && git push origin v[VERSION]
```

### 注意事項
- セマンティックバージョニング（major.minor.patch）に従う
- プレリリース版（例: `v1.5.1-beta.1`）は `@latest` に含まれない
- Go Module Proxy のキャッシュ更新には時間がかかる場合がある

## テスト開発ガイドライン

### テストパターン
- **純粋関数**: テーブル駆動テスト（`t.Run` + サブテスト）
- **Git操作**: `testutil.TempGitRepo(t)` でテンポラリリポジトリ作成
- **Git notes**: `gitexec.NewMockExecutor()` でモック化
- **統合テスト**: `testutil.InitAICT(t, tmpDir)` でAICT環境構築
- **cmd/aict/**: `newExecutor` のDIパターンでモック注入可能（カバレッジ56.7%）

### 注意点
- テスト内で `os.Chdir` する場合は必ず `defer os.Chdir(originalDir)` でリストア
- `expandShorthandDate` は未知の形式をそのままgitに渡す（エラーにならない）

## リファクタリング進捗状況

プロジェクトは継続的な品質改善を行っています。詳細は `TODO.md` を参照してください。

### 完了済みフェーズサマリー

| Phase | 内容 | 成果 |
|-------|------|------|
| 1 | テストインフラ基盤整備 | `internal/testutil/` 作成、テストセットアップ60%削減 |
| 2 | Gitコマンド抽象化 | `internal/gitexec/` 作成、os/exec依存削除 |
| 3 | ストレージ抽象化評価 | 既に完了済みと確認 |
| 4.1 | 高複雑度関数リファクタリング | CC=11→3（86%削減） |
| 4.2 | Numstat解析集約化 | `internal/git/` に統合、82-93%削減 |
| 4.5 | 複数メトリクス対応 | 3視点測定（コードベース貢献・作業量・新規ファイル） |
| 5 | パフォーマンス改善 | N+1問題解消、JSONL化 |
| 6 | セキュリティ強化 | Git引数注入防止、ValidateRevisionArg |
| 7 | テスト品質向上 | gitnotes/templates/handlers_range テスト追加 |
| 8 | その他 | コメント言語統一、変数シャドウイング解消 |
| 9 | データ整合性・バグリスク | ファイルフィルタ統一、ゼロ除算ガード |
| 10 | コード品質・アーキテクチャ | collectAuthorStats分割、DIパターン化 |
| 11 | テスト品質向上 | cmd/aictカバレッジ27.3%→56.7% |
| 12 | Stash/Restore対応 | AI帰属保全、チェックポイント再生成 |
| 13 | TOCTOU修正・API改善 | ロック保護、PurgeExpiredCheckpoints API簡素化、テスト+502行 |
