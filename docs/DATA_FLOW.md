# AICT データフローガイド

このドキュメントでは、AI Code Tracker (AICT) がどのようにデータを記録し、レポートを生成しているかを詳しく説明します。

## 目次

1. [概要](#概要)
2. [データフロー全体図](#データフロー全体図)
3. [詳細フロー](#詳細フロー)
4. [データ構造](#データ構造)
5. [ストレージ構成](#ストレージ構成)

---

## 概要

AICTは以下の3つの主要ステップでコードの作成者情報を追跡します：

1. **チェックポイント記録** - コード変更前後の状態をスナップショット
2. **Authorship Log生成** - チェックポイントからコミット単位の作成者情報を抽出
3. **レポート生成** - Git notesから統計情報を集計・表示

---

## データフロー全体図

```
┌─────────────────────────────────────────────────────────────────────┐
│ 開発フロー                                                           │
└─────────────────────────────────────────────────────────────────────┘

[1] Claude Code開始前
    │
    ├─> pre-tool-use hook
    │   └─> aict checkpoint --author human
    │       └─> .git/aict/checkpoints/records.jsonl
    │           (人間のベースライン記録)
    │
[2] Claude Codeが編集
    │
    ├─> post-tool-use hook
    │   └─> aict checkpoint --author "Claude Sonnet 4.5"
    │       └─> .git/aict/checkpoints/records.jsonl
    │           (AIの変更記録)
    │
[3] 人間が追加編集（任意）
    │
    └─> git commit
        │
        ├─> post-commit hook
        │   └─> aict commit
        │       ├─> Checkpointを読み込み
        │       ├─> Authorship Logに変換
        │       └─> Git notesに保存
        │           └─> refs/aict/authorship
        │               (コミット単位の作成者情報)
        │
        └─> Checkpointクリア（最新スナップショットは保持）

┌─────────────────────────────────────────────────────────────────────┐
│ レポート生成フロー                                                   │
└─────────────────────────────────────────────────────────────────────┘

aict report --since 7d
    │
    ├─> コミット範囲を取得
    │   └─> git log --since 7d
    │
    ├─> 各コミットのAuthorship Logを取得
    │   └─> git notes --ref=refs/aict/authorship show <commit>
    │
    ├─> 作成者別集計
    │   ├─> AI行数カウント
    │   ├─> 人間行数カウント
    │   └─> パーセンテージ計算
    │
    └─> レポート出力
        ├─> テーブル形式（デフォルト）
        └─> JSON形式（--format json）
```

---

## 詳細フロー

### 1. チェックポイント記録フェーズ

#### 1.1 Pre-tool-use Hook（編集前）

```
ユーザーアクション: Claude Codeを起動
    ↓
.claude/hooks/pre-tool-use.sh
    ↓
aict checkpoint --author human
    ↓
CheckpointManager.CreateCheckpoint()
    ├─> git diff --numstat HEAD を実行
    │   └─> 変更ファイルと行数を取得
    ├─> CheckpointRecord を作成
    │   ├─> Timestamp: 現在時刻
    │   ├─> Author: "human"
    │   ├─> Branch: 現在のブランチ名
    │   ├─> Added: 追加行数
    │   └─> Deleted: 削除行数
    └─> .git/aict/checkpoints/records.jsonl に追記
```

**データ例**:
```json
{
  "timestamp": "2025-12-11T10:00:00Z",
  "author": "human",
  "branch": "feature/new-api",
  "added": 0,
  "deleted": 0
}
```

#### 1.2 Post-tool-use Hook（編集後）

```
Claude Codeが編集完了
    ↓
.claude/hooks/post-tool-use.sh
    ↓
aict checkpoint --author "Claude Sonnet 4.5"
    ↓
CheckpointManager.CreateCheckpoint()
    ├─> git diff --numstat HEAD を実行
    │   └─> 変更ファイルと行数を取得（AIによる変更を含む）
    ├─> CheckpointRecord を作成
    │   ├─> Timestamp: 現在時刻
    │   ├─> Author: "Claude Sonnet 4.5"
    │   ├─> Branch: 現在のブランチ名
    │   ├─> Added: 追加行数（AIによる追加）
    │   └─> Deleted: 削除行数（AIによる削除）
    └─> .git/aict/checkpoints/records.jsonl に追記
```

**データ例**:
```json
{
  "timestamp": "2025-12-11T10:15:00Z",
  "author": "Claude Sonnet 4.5",
  "branch": "feature/new-api",
  "added": 150,
  "deleted": 20
}
```

### 2. Authorship Log生成フェーズ

#### 2.1 Post-commit Hook

```
ユーザーアクション: git commit
    ↓
.git/hooks/post-commit
    ↓
aict commit
    ↓
handleCommit()
    ├─> LoadCheckpoints()
    │   └─> .git/aict/checkpoints/records.jsonl を読み込み
    │       └─> CheckpointRecord の配列を取得
    │
    ├─> getLatestCommitHash()
    │   └─> git rev-parse HEAD
    │       └─> 最新のコミットハッシュを取得
    │
    ├─> BuildAuthorshipLog()
    │   ├─> CheckpointRecord → CheckpointV2 に変換
    │   │   ├─> 前後のCheckpointを比較
    │   │   ├─> git diff で変更箇所を特定
    │   │   └─> 変更をAuthorに紐付け
    │   │
    │   └─> AuthorshipLog を生成
    │       ├─> Version: "1.0"
    │       ├─> Commit: コミットハッシュ
    │       ├─> Timestamp: 現在時刻
    │       └─> Files: ファイルごとの作成者情報
    │           └─> Authors: 作成者と行範囲のマッピング
    │
    ├─> ValidateAuthorshipLog()
    │   └─> データ整合性チェック
    │
    ├─> NotesManager.AddAuthorshipLog()
    │   └─> git notes --ref=refs/aict/authorship add <commit>
    │       └─> AuthorshipLog を JSON として保存
    │
    └─> ClearCheckpoints()
        └─> チェックポイントをクリア
            └─> 最新スナップショットのみベースラインとして保持
```

**AuthorshipLog データ例**:
```json
{
  "version": "1.0",
  "commit": "a1b2c3d4e5f6...",
  "timestamp": "2025-12-11T10:30:00Z",
  "files": {
    "internal/api/handler.go": {
      "authors": [
        {
          "name": "Claude Sonnet 4.5",
          "type": "ai",
          "lines": [[1, 50], [75, 100]],
          "metadata": {}
        },
        {
          "name": "human",
          "type": "human",
          "lines": [[51, 74]],
          "metadata": {}
        }
      ]
    }
  }
}
```

### 3. レポート生成フェーズ

#### 3.1 Report生成

```
ユーザーアクション: aict report --since 7d
    ↓
handleRange()
    ├─> parseTimeFilter()
    │   └─> "7d" → 7日前の日時に変換
    │
    ├─> getCommitRange()
    │   └─> git log --since "7 days ago" --format=%H
    │       └─> コミットハッシュのリストを取得
    │
    ├─> NotesManager.GetAuthorshipLog() (各コミット)
    │   └─> git notes --ref=refs/aict/authorship show <commit>
    │       └─> AuthorshipLog を JSON パース
    │
    ├─> Analyzer.AnalyzeCheckpoints()
    │   ├─> 全AuthorshipLogを集計
    │   ├─> 作成者別行数カウント
    │   │   ├─> AI作成者判定（config.AIAgents）
    │   │   ├─> 行範囲から総行数を計算
    │   │   └─> ファイル別・作成者別に集計
    │   │
    │   └─> AnalysisResult を生成
    │       ├─> TotalLines: 総行数
    │       ├─> AILines: AI作成行数
    │       ├─> HumanLines: 人間作成行数
    │       ├─> Percentage: AI比率
    │       └─> Metrics: 詳細メトリクス（オプション）
    │           ├─> Contributions: コードベース貢献
    │           ├─> WorkVolume: 作業量貢献
    │           └─> NewFiles: 新規ファイル
    │
    └─> printTableReport() / printJSONReport()
        └─> 集計結果を出力
```

#### 3.2 メトリクス表示（常時表示）

```
aict report --since 7d
    ↓
printDetailedMetrics()
    ├─> コードベース貢献
    │   └─> 純粋な追加行数（最終コード量への寄与）
    │       ├─> AI追加: 2行 (3.8%)
    │       └─> 人間追加: 50行 (96.2%)
    │
    ├─> 作業量貢献
    │   └─> 追加+削除の合計（実際の作業量）
    │       ├─> AI作業: 29行 (19.9%)
    │       │   ├─> 追加: 2行
    │       │   └─> 削除: 27行
    │       └─> 人間作業: 117行 (80.1%)
    │           ├─> 追加: 50行
    │           └─> 削除: 67行
    │
    └─> By Author
        └─> 追加行数ベース
            ├─> AI: 2行追加 (3.8%)
            └─> 人間: 50行追加 (96.2%)
```

---

## データ構造

### CheckpointRecord (軽量記録形式)

```go
type CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Branch    string    `json:"branch,omitempty"`
    Commit    string    `json:"commit,omitempty"`
    Added     int       `json:"added"`   // 総追加行数
    Deleted   int       `json:"deleted"` // 総削除行数
}
```

**用途**: `.git/aict/checkpoints/records.jsonl` にJSONL形式で保存

### CheckpointV2 (SPEC準拠の完全形式)

```go
type CheckpointV2 struct {
    Timestamp time.Time             `json:"timestamp"`
    Author    string                `json:"author"`
    Type      AuthorType            `json:"type"` // "human" or "ai"
    Metadata  map[string]string     `json:"metadata,omitempty"`
    Changes   map[string]Change     `json:"changes"`  // filepath -> Change
    Snapshot  map[string]FileSnapshot `json:"snapshot"` // filepath -> FileSnapshot
}

type Change struct {
    Added   int     `json:"added"`
    Deleted int     `json:"deleted"`
    Lines   [][]int `json:"lines"` // [[start, end], ...]
}

type FileSnapshot struct {
    Hash  string `json:"hash"`  // SHA-256 hash
    Lines int    `json:"lines"` // 総行数
}
```

**用途**: Authorship Log生成時の中間形式

### AuthorshipLog (Git notes保存形式)

```go
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
    Type     AuthorType        `json:"type"` // "human" or "ai"
    Lines    [][]int           `json:"lines"` // [[start, end], ...]
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

**用途**: `refs/aict/authorship` Git notesに保存

### AnalysisResult (レポート形式)

```go
type AnalysisResult struct {
    TotalLines  int       `json:"total_lines"`
    AILines     int       `json:"ai_lines"`
    HumanLines  int       `json:"human_lines"`
    Percentage  float64   `json:"percentage"`
    LastUpdated time.Time `json:"last_updated"`

    // 詳細メトリクス（--detailed フラグ時）
    Metrics DetailedMetrics `json:"metrics,omitempty"`
}

type DetailedMetrics struct {
    Contributions ContributionMetrics `json:"contributions"` // コードベース貢献
    WorkVolume    WorkVolumeMetrics   `json:"work_volume"`   // 作業量貢献
    NewFiles      NewFileMetrics      `json:"new_files"`     // 新規ファイル
}
```

**用途**: レポート出力

---

## ストレージ構成

### ディレクトリ構造

```
.git/
├── aict/                           # AICT専用ディレクトリ
│   ├── config.json                 # プロジェクト設定
│   └── checkpoints/
│       └── records.jsonl           # チェックポイント記録（JSONL形式）
│
├── hooks/
│   └── post-commit                 # Git post-commitフック
│
└── refs/
    └── aict/
        └── authorship              # Git notes（コミット単位の作成者情報）

.claude/
└── hooks/
    ├── pre-tool-use.sh             # Claude Code開始前フック
    └── post-tool-use.sh            # Claude Code終了後フック
```

### ストレージ詳細

#### 1. Checkpoints (`records.jsonl`)

- **形式**: JSONL（1行1チェックポイント）
- **サイズ**: 軽量（行数統計のみ）
- **ライフサイクル**: コミット後にクリア（最新スナップショット除く）

**例**:
```jsonl
{"timestamp":"2025-12-11T10:00:00Z","author":"human","branch":"main","added":0,"deleted":0}
{"timestamp":"2025-12-11T10:15:00Z","author":"Claude Sonnet 4.5","branch":"main","added":150,"deleted":20}
```

#### 2. Git Notes (`refs/aict/authorship`)

- **形式**: JSON
- **スコープ**: コミット単位
- **永続性**: Gitリポジトリと同期可能
- **同期**: `aict sync push/fetch`

**コマンド例**:
```bash
# 特定コミットのAuthorship Logを表示
git notes --ref=refs/aict/authorship show <commit-hash>

# リモートにプッシュ
aict sync push

# リモートから取得
aict sync fetch
```

#### 3. Config (`config.json`)

```json
{
  "target_ai_percentage": 80,
  "tracked_extensions": [".go", ".py", ".js", ".ts"],
  "exclude_patterns": ["*_test.go", "vendor/*"],
  "default_author": "human",
  "ai_agents": [
    "Claude Sonnet 4.5",
    "GPT-4",
    "Copilot"
  ]
}
```

---

## レポート表示例

### 実際の出力例

```bash
$ aict report --since 7d

📊 AI Code Generation Report (since 7d)

Commits: 5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【コードベース貢献】（最終的なコード量への寄与）
  総変更行数: 52行
    🤖 AI追加:        2行 (3.8%)
    👤 人間追加:     50行 (96.2%)

【作業量貢献】（実際の作業量）
  総作業量: 146行
    🤖 AI作業:       29行 (19.9%)
       └ 追加: 2行, 削除: 27行
    👤 人間作業:    117行 (80.1%)
       └ 追加: 50行, 削除: 67行

By Author:
  👤 y-hirakaw                50行追加 (96.2%) - 5 commits
  🤖 Claude Code               2行追加 (3.8%) - 1 commits
```

### レポートの読み方

#### コードベース貢献
- **意味**: 期間内に最終的なコードベースに追加された行数
- **計算**: `git diff --numstat`の追加行数のみ
- **用途**: 「実際に増えたコード量」の把握

#### 作業量貢献
- **意味**: 期間内の実際の作業量（追加+削除）
- **計算**: `git diff --numstat`の追加行数+削除行数
- **用途**: 「実際の作業量」の把握（リファクタリング作業も評価）

#### By Author
- **意味**: 作成者別の追加行数
- **計算**: コードベース貢献と同じ（追加行数のみ）
- **用途**: 「誰がどれだけコードを追加したか」の把握

---

## まとめ

AICTのデータフローは以下の3段階で構成されています：

1. **記録** - フック経由でチェックポイントを軽量記録（JSONL）
2. **変換** - コミット時にAuthorship Logに変換してGit notesに保存（JSON）
3. **集計** - レポート生成時にGit notesから統計情報を集計・表示（**差分追跡方式**）

この設計により、以下のメリットがあります：

- ✅ **軽量性**: チェックポイントは行数のみ記録
- ✅ **永続性**: Git notesによるコミット単位の履歴管理
- ✅ **柔軟性**: 複数のメトリクスによる多角的分析
- ✅ **同期性**: Git notesを使ったリモート同期
- ✅ **正確性**: git diffベースの変更追跡
- ✅ **直感性**: `--since`は期間内の変更のみを集計（重複なし）
