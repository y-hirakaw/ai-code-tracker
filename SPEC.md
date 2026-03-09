# ai-code-tracker (aict) 仕様書

> **注意**: このドキュメントはプロジェクト開始時の初期設計仕様書です（v0.1.0想定）。
> 現在の実装（v1.5.0-beta.1）とは差異があります。実装状況は `TODO.md` を、
> 現在の機能仕様は `CLAUDE.md` を参照してください。
>
> **主な差異**:
> - ロードマップのチェックボックスは更新されていません（Phase 1-2は実装済み）
> - `--branch`, `--last`, `--by-file`, `--by-author`, `--verbose`, `--format csv` は未実装
> - 実装されたオプション: `--since`（`--last`の代替）, `--range`, `--format json|table`
> - ディレクトリ構造は実装に伴い変更されています

## プロジェクト概要

**ai-code-tracker (aict)** は、AIと人間によるコード生成率を正確に計測するためのGitツールです。

### 目的
- PRやブランチ単位でAI/人間のコード貢献度を可視化
- CI/CDパイプラインでの自動計測
- チーム内でのAI活用状況の把握

### 特徴
- 🎯 **シンプル**: コード生成率の計測機能に特化
- 🚀 **高速**: Go言語による軽量実装
- 🔧 **統合容易**: 単一バイナリで配布、既存ワークフローに簡単に組み込み
- 📊 **CI/CD対応**: JSON出力でパイプライン統合が容易

---

## コア概念

### 1. チェックポイントシステム

開発中に「誰がコードを書いたか」を記録するための境界マーカー。

```bash
# パターン1: AIが作業を開始する前
$ aict checkpoint

# AIがコード生成（例: Claude Code, Cursor, GitHub Copilotなど）

# パターン2: AI作業完了後
$ aict checkpoint --author "Claude Code" --model "claude-sonnet-4"

# 人間が手動修正

# パターン3: 人間の作業完了
$ aict checkpoint
```

**特性:**
- コミット前の一時データ（`.git/aict/checkpoints/`に保存）
- Gitヒストリーには含まれない
- 何度でも作成・修正可能

### 2. Authorship Log

コミット時に生成される、行単位での作成者情報を含むJSON記録。

**保存場所:** Git notes (`refs/aict/authorship/{commit-sha}`)

**フォーマット:**
```json
{
  "version": "1.0",
  "commit": "abc123def456...",
  "timestamp": "2024-11-23T10:30:00Z",
  "files": {
    "src/main.go": {
      "authors": [
        {
          "name": "Yuta",
          "lines": [[1, 20], [71, 75]],
          "type": "human"
        },
        {
          "name": "Claude Code",
          "lines": [[21, 70]],
          "type": "ai",
          "metadata": {
            "model": "claude-sonnet-4"
          }
        }
      ]
    }
  }
}
```

**lines 配列の形式:**
- `[10]`: 単一行（10行目）
- `[10, 20]`: 範囲（10-20行目）
- 複数の範囲を配列で保持

---

## コマンド仕様

### `aict init`

リポジトリでaictを初期化。

```bash
$ aict init
```

**動作:**
- `.git/aict/` ディレクトリ作成
- 設定ファイル `.git/aict/config.json` 生成
- Git hooks設定（オプション）

**設定例:**
```json
{
  "version": "1.0",
  "default_author": "Yuta",
  "ai_agents": ["Claude Code", "Cursor", "GitHub Copilot"]
}
```

### `aict checkpoint`

開発の区切りで作成者を記録。

```bash
# 基本（人間の作業区切り）
$ aict checkpoint

# AI作業完了時
$ aict checkpoint --author "Claude Code" --model "claude-sonnet-4"

# メッセージ付き
$ aict checkpoint --author "Cursor" --message "Implemented authentication logic"
```

**オプション:**
- `--author <name>`: 作成者名（デフォルト: config.default_author）
- `--model <model>`: AIモデル名（AIエージェントの場合）
- `--message <msg>`: メモ（オプション）

**動作:**
1. 前回のチェックポイント以降の変更を検出
2. 変更行数を計算（`git diff`使用）
3. チェックポイントデータを `.git/aict/checkpoints/latest.json` に追記

**チェックポイントデータ形式:**
```json
{
  "timestamp": "2024-11-23T10:30:00Z",
  "author": "Claude Code",
  "type": "ai",
  "metadata": {
    "model": "claude-sonnet-4"
  },
  "changes": {
    "src/main.go": {
      "added": 50,
      "deleted": 5,
      "lines": [[21, 70]]
    }
  }
}
```

### `aict commit` (Git hookとして自動実行)

コミット時に自動的に実行され、Authorship Logを生成。

```bash
# 通常は git commit で自動実行される
$ git commit -m "Add new feature"
→ 内部的に aict commit が実行される

# 手動実行も可能
$ aict commit
```

**動作:**
1. `.git/aict/checkpoints/latest.json` を読み込み
2. チェックポイント群をAuthorship Log形式に変換
3. Git notes として保存 (`refs/aict/authorship/{commit-sha}`)
4. チェックポイントファイルをクリア

### `aict report`

コミット、ブランチ、または期間のコード生成レポートを表示。

#### 基本的な使い方

```bash
# 現在のHEADコミット
$ aict report

# 特定のコミット
$ aict report abc123

# 現在のブランチ全体
$ aict report --branch feature-branch
```

**出力例:**
```
📊 AI Code Generation Report

Branch: feature-branch (15 commits)
Period: 2024-11-15 ~ 2024-11-23
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total Lines:        500
  🤖 AI Generated:    350 (70.0%)
  👤 Human Written:   150 (30.0%)

By Author:
  Claude Code:        300 lines (60.0%)
  Cursor:              50 lines (10.0%)
  Yuta:               150 lines (30.0%)

Top Files:
  src/main.go:        200 lines (75% AI)
  src/utils.go:       150 lines (80% AI)
  src/api.go:         100 lines (50% AI)
  src/helper.go:       50 lines (60% AI)
```

#### 期間指定

```bash
# 過去7日間
$ aict report --last 7d

# 過去30日間
$ aict report --last 30d

# 今週
$ aict report --last week

# 今月
$ aict report --last month

# カスタム期間
$ aict report --since "2024-11-01" --until "2024-11-30"
```

**期間指定フォーマット:**
- `7d`, `30d`: 日数
- `2w`: 週数
- `3m`: 月数
- `week`: 今週（月曜日から）
- `month`: 今月（1日から）
- ISO 8601形式: `2024-11-01T00:00:00Z`

#### ブランチ/PR範囲指定

```bash
# 特定のブランチ
$ aict report --branch feature-branch

# ブランチ範囲（PRに相当）
$ aict report --range origin/main..feature-branch

# 現在のブランチとmainの差分
$ aict report --range origin/main..HEAD
```

#### 詳細表示オプション

```bash
# ファイル別の詳細
$ aict report --branch feature-branch --by-file

# 作成者別の詳細
$ aict report --branch feature-branch --by-author

# すべての詳細
$ aict report --branch feature-branch --verbose
```

#### JSON出力（CI/CD用）

```bash
$ aict report --branch feature-branch --json
```

**JSON出力例:**
```json
{
  "branch": "feature-branch",
  "range": "origin/main..feature-branch",
  "commits": 15,
  "period": {
    "start": "2024-11-15T10:30:00Z",
    "end": "2024-11-23T15:45:00Z"
  },
  "summary": {
    "total_lines": 500,
    "ai_lines": 350,
    "human_lines": 150,
    "ai_percentage": 70.0
  },
  "by_file": [
    {
      "path": "src/main.go",
      "total_lines": 200,
      "ai_lines": 150,
      "human_lines": 50,
      "ai_percentage": 75.0
    },
    {
      "path": "src/utils.go",
      "total_lines": 150,
      "ai_lines": 120,
      "human_lines": 30,
      "ai_percentage": 80.0
    }
  ],
  "by_author": [
    {
      "name": "Claude Code",
      "type": "ai",
      "lines": 300,
      "percentage": 60.0,
      "commits": 8
    },
    {
      "name": "Cursor",
      "type": "ai",
      "lines": 50,
      "percentage": 10.0,
      "commits": 2
    },
    {
      "name": "Yuta",
      "type": "human",
      "lines": 150,
      "percentage": 30.0,
      "commits": 5
    }
  ]
}
```

**オプション一覧:**
- `--branch <name>`: ブランチ名を指定
- `--range <base>..<head>`: コミット範囲指定
- `--last <period>`: 相対期間指定（7d, 30d, week, monthなど）
- `--since <date>`: 開始日時（ISO 8601形式）
- `--until <date>`: 終了日時
- `--by-file`: ファイル別の詳細表示
- `--by-author`: 作成者別の詳細表示
- `--verbose`: すべての詳細表示
- `--json`: JSON形式で出力
- `--format <format>`: 出力形式（text, json, csv）

### `aict sync`

Authorship LogをリモートリポジトリとSync。

```bash
# Push
$ aict sync push
$ git push  # Git notes も自動的にpushされる（hook設定時）

# Fetch
$ aict sync fetch
$ git fetch  # Git notes も自動的にfetchされる（hook設定時）
```

**動作:**
- Git notesの `refs/aict/authorship/*` をpush/fetch
- 内部的に `git push/fetch refs/notes/aict/*` を実行

---

## データフロー

### 開発フロー全体

```
1. 開発開始
   $ aict checkpoint
   ↓
2. AIがコード生成
   ↓
3. AI作業完了
   $ aict checkpoint --author "Claude Code"
   ↓
4. 人間が修正
   ↓
5. 修正完了
   $ aict checkpoint
   ↓
6. コミット
   $ git commit -m "Add feature"
   → aict commit (自動実行)
   → Authorship Log生成
   ↓
7. Push
   $ git push
   → Authorship Logもpush
   ↓
8. 統計確認
   $ aict stats --range origin/main..feature-branch
```

### チェックポイント → Authorship Log変換

```
チェックポイント群:
[
  {author: "Yuta", lines: [[1,20]]},
  {author: "Claude Code", lines: [[21,70]], model: "claude-sonnet-4"},
  {author: "Yuta", lines: [[71,75]]}
]
        ↓ git commit時に集約
Authorship Log:
{
  "files": {
    "src/main.go": {
      "authors": [
        {"name": "Yuta", "lines": [[1,20], [71,75]], "type": "human"},
        {"name": "Claude Code", "lines": [[21,70]], "type": "ai", ...}
      ]
    }
  }
}
        ↓ Git notesに保存
refs/aict/authorship/abc123def...
```

---

## レポート生成ロジック

### 方法1: 単純集計（デフォルト）

**用途:** PR/ブランチの開発活動総量を測定

```
1. git log <range> でコミット一覧取得
2. 各コミットのAuthorship Logを読み込み
3. すべてのAI行数、人間行数を合算
4. 割合を計算
```

**特徴:**
- 高速
- 削除された行もカウント（開発活動の総量）
- 「このPRでどれだけコードを書いたか」を表す

**実装疑似コード:**
```go
func GenerateReport(commitRange string) *Report {
    commits := getCommits(commitRange)
    totalAI := 0
    totalHuman := 0
    
    for _, commit := range commits {
        log := readAuthorshipLog(commit.SHA)
        for _, file := range log.Files {
            for _, author := range file.Authors {
                lineCount := countLines(author.Lines)
                if author.Type == "ai" {
                    totalAI += lineCount
                } else {
                    totalHuman += lineCount
                }
            }
        }
    }
    
    return &Report{
        Summary: SummaryStats{
            AILines: totalAI,
            HumanLines: totalHuman,
            AIPercentage: float64(totalAI) / float64(totalAI + totalHuman) * 100,
        },
    }
}
```

### 方法2: blame方式（将来実装）

**用途:** 最終成果物の正確な割合を測定

```
1. git diff <range> でファイル一覧と追加行取得
2. 各追加行に git blame で作成コミット特定
3. Authorship Logから実際の作成者を取得
4. 集計
```

**特徴:**
- 正確
- 削除された行は除外
- 「最終的に残ったコードの何%がAI製か」を表す

---

## Git Hooks統合

### post-commit hook

```bash
#!/bin/sh
# .git/hooks/post-commit

# aictでAuthorship Logを生成
aict commit

exit 0
```

### pre-push hook

```bash
#!/bin/sh
# .git/hooks/pre-push

# Authorship Logもpush
git push origin "refs/notes/aict/*"

exit 0
```

### post-merge / post-rebase hook

```bash
#!/bin/sh
# .git/hooks/post-merge

# Authorship Logを同期
aict sync fetch

exit 0
```

---

## CI/CD統合例

### GitHub Actions

```yaml
name: AI Code Report

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0  # 全履歴取得
          
      - name: Install aict
        run: |
          curl -L https://github.com/y-hirakaw/ai-code-tracker/releases/latest/download/aict-linux-amd64 -o /usr/local/bin/aict
          chmod +x /usr/local/bin/aict
          
      - name: Fetch authorship logs
        run: |
          git fetch origin refs/notes/aict/*:refs/notes/aict/*
          
      - name: Generate report
        id: report
        run: |
          aict report --range origin/${{ github.base_ref }}..HEAD --json > report.json
          echo "ai_percentage=$(jq -r '.summary.ai_percentage' report.json)" >> $GITHUB_OUTPUT
          
      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            const report = require('./report.json');
            const body = `## 🤖 AI Code Generation Report
            
            **Total Lines:** ${report.summary.total_lines}
            - 🤖 AI: ${report.summary.ai_lines} (${report.summary.ai_percentage.toFixed(1)}%)
            - 👤 Human: ${report.summary.human_lines} (${(100 - report.summary.ai_percentage).toFixed(1)}%)
            
            ### By Author
            ${report.by_author.map(a => `- ${a.name}: ${a.lines} lines (${a.percentage.toFixed(1)}%)`).join('\n')}
            `;
            
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: body
            });
```

---

## 実装アーキテクチャ

### ディレクトリ構造

```
ai-code-tracker/
├── cmd/
│   └── aict/
│       └── main.go              # エントリーポイント
├── internal/
│   ├── checkpoint/
│   │   ├── checkpoint.go        # チェックポイント管理
│   │   └── storage.go           # ローカルストレージ
│   ├── authorship/
│   │   ├── log.go               # Authorship Log構造体
│   │   ├── builder.go           # チェックポイント→ログ変換
│   │   └── parser.go            # JSON解析
│   ├── git/
│   │   ├── notes.go             # Git notes操作
│   │   ├── diff.go              # git diff解析
│   │   ├── blame.go             # git blame解析
│   │   └── log.go               # git log解析
│   ├── report/
│   │   ├── generator.go         # レポート生成
│   │   ├── aggregator.go        # データ集約
│   │   └── formatter.go         # 出力フォーマット
│   └── config/
│       └── config.go            # 設定管理
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

### 主要な型定義

```go
// Checkpoint represents a development checkpoint
type Checkpoint struct {
    Timestamp time.Time          `json:"timestamp"`
    Author    string             `json:"author"`
    Type      AuthorType         `json:"type"` // "human" or "ai"
    Metadata  map[string]string  `json:"metadata,omitempty"`
    Changes   map[string]Change  `json:"changes"`
}

type Change struct {
    Added   int        `json:"added"`
    Deleted int        `json:"deleted"`
    Lines   [][]int    `json:"lines"` // [[start, end], [single], ...]
}

// AuthorshipLog represents commit-level authorship information
type AuthorshipLog struct {
    Version   string                `json:"version"`
    Commit    string                `json:"commit"`
    Timestamp time.Time             `json:"timestamp"`
    Files     map[string]FileInfo   `json:"files"`
}

type FileInfo struct {
    Authors []AuthorInfo `json:"authors"`
}

type AuthorInfo struct {
    Name     string            `json:"name"`
    Type     AuthorType        `json:"type"`
    Lines    [][]int           `json:"lines"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

type AuthorType string

const (
    AuthorTypeHuman AuthorType = "human"
    AuthorTypeAI    AuthorType = "ai"
)

// Report represents generated code generation report
type Report struct {
    Range        string             `json:"range,omitempty"`
    Branch       string             `json:"branch,omitempty"`
    Commits      int                `json:"commits,omitempty"`
    Period       *Period            `json:"period,omitempty"`
    Summary      SummaryStats       `json:"summary"`
    ByFile       []FileStats        `json:"by_file,omitempty"`
    ByAuthor     []AuthorStats      `json:"by_author,omitempty"`
}

type Period struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type SummaryStats struct {
    TotalLines    int     `json:"total_lines"`
    AILines       int     `json:"ai_lines"`
    HumanLines    int     `json:"human_lines"`
    AIPercentage  float64 `json:"ai_percentage"`
}

type FileStats struct {
    Path         string  `json:"path"`
    TotalLines   int     `json:"total_lines"`
    AILines      int     `json:"ai_lines"`
    HumanLines   int     `json:"human_lines"`
    AIPercentage float64 `json:"ai_percentage"`
}

type AuthorStats struct {
    Name       string     `json:"name"`
    Type       AuthorType `json:"type"`
    Lines      int        `json:"lines"`
    Percentage float64    `json:"percentage"`
    Commits    int        `json:"commits,omitempty"`
}
```

---

## 開発ロードマップ

### Phase 1: MVP (v0.1.0)
- [x] 仕様策定
- [ ] `aict init` 実装
- [ ] `aict checkpoint` 実装
- [ ] `aict commit` 実装（Authorship Log生成）
- [ ] `aict report` 実装（単一コミット）
- [ ] `aict report --range` 実装（単純集計）
- [ ] `aict report --last` 実装（相対期間指定）
- [ ] JSON出力対応
- [ ] 基本的なテスト

### Phase 2: 安定版 (v0.2.0)
- [ ] Git hooks自動設定
- [ ] `aict sync` 実装
- [ ] エラーハンドリング強化
- [ ] 設定ファイル対応
- [ ] ドキュメント整備

### Phase 3: 拡張機能 (v0.3.0)
- [ ] `--by-file`, `--by-author` 詳細表示
- [ ] `--format csv` CSV出力
- [ ] GitHub Actions統合例
- [ ] Webダッシュボード（オプション）

### Phase 4: 高度な機能 (v1.0.0)
- [ ] blame方式のレポート生成
- [ ] トレンド分析
- [ ] 複数リポジトリ対応
- [ ] パフォーマンス最適化

---

## インストール方法

### バイナリインストール（推奨）

```bash
# Linux/macOS
curl -L https://github.com/y-hirakaw/ai-code-tracker/releases/latest/download/aict-$(uname -s)-$(uname -m) -o /usr/local/bin/aict
chmod +x /usr/local/bin/aict

# Windowsはリリースページからダウンロード
```

### Go install

```bash
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest
```

### ソースからビルド

```bash
git clone https://github.com/y-hirakaw/ai-code-tracker.git
cd ai-code-tracker
go build -o aict ./cmd/aict
```

---

## 使用例

### 基本的な使い方

```bash
# 1. 初期化
$ cd your-project
$ aict init
✓ Initialized aict in /path/to/your-project

# 2. 開発開始
$ aict checkpoint
✓ Checkpoint created

# 3. AIでコード生成
# (Claude Codeなどで開発)

# 4. AI作業完了を記録
$ aict checkpoint --author "Claude Code" --model "claude-sonnet-4"
✓ Checkpoint created (Claude Code, 50 lines added)

# 5. コミット（自動的にAuthorship Log生成）
$ git add .
$ git commit -m "Add authentication feature"
✓ Authorship log created

# 6. 統計確認
$ aict report
📊 AI Code Generation Report

Commit: HEAD
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total Lines:        50
  🤖 AI Generated:    50 (100%)
  👤 Human Written:    0 (0%)
```

### PR開発での使用例

```bash
# フィーチャーブランチ作成
$ git checkout -b feature/new-api
$ aict checkpoint

# 基礎実装（人間）
$ vim src/api.go
$ aict checkpoint

# AIで詳細実装
# (AIツールでコード生成)
$ aict checkpoint --author "Cursor"

# レビュー修正（人間）
$ vim src/api.go
$ aict checkpoint

# コミット
$ git add .
$ git commit -m "Implement new API endpoint"

# PR作成前にレポート確認
$ aict report --range origin/main..HEAD
📊 AI Code Generation Report

Range: origin/main..HEAD
Period: 2024-11-23 10:00 ~ 2024-11-23 15:30
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total Lines:        200
  🤖 AI Generated:    150 (75.0%)
  👤 Human Written:    50 (25.0%)

By Author:
  Cursor:             150 lines (75.0%)
  Yuta:                50 lines (25.0%)
```

---

## トラブルシューティング

### Q: Authorship Logが見つからない

```bash
# Git notesを確認
$ git notes --ref=refs/aict/authorship list

# リモートから取得
$ git fetch origin refs/notes/aict/*:refs/notes/aict/*
```

### Q: チェックポイントが記録されない

```bash
# チェックポイントファイルを確認
$ cat .git/aict/checkpoints/latest.json

# 権限を確認
$ ls -la .git/aict/
```

### Q: レポートが0%になる

Authorship Logがない古いコミットの可能性があります。aict導入後のコミットのみが計測対象です。

```bash
# Authorship Logがあるコミットを確認
$ git notes --ref=refs/aict/authorship list

# 特定のコミット範囲で確認
$ aict report --range <first-commit-with-aict>..HEAD
```

---

## ライセンス

MIT License

---

## 貢献

Issues、Pull Requestsを歓迎します！

---

## 参考

- Git Notes: https://git-scm.com/docs/git-notes
