# AICT データフローガイド (v1.2.0)

このドキュメントでは、AI Code Tracker (AICT) v1.2.0がどのようにデータを記録し、レポートを生成しているかを詳しく説明します。

## 目次

1. [概要](#概要)
2. [データフロー全体図](#データフロー全体図)
3. [詳細フロー](#詳細フロー)
4. [データ構造](#データ構造)
5. [ストレージ構成](#ストレージ構成)
6. [既知の制限](#既知の制限)

---

## 概要

AICTは以下の3つの主要ステップでコードの作成者情報を追跡します：

1. **チェックポイント記録** - コード変更前後の状態をスナップショット
2. **Authorship Log生成** - コミット差分とチェックポイントから作成者情報を抽出
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
    │   └─> aict checkpoint --author y-hirakaw --message "Before Claude Code edits"
    │       └─> .git/aict/checkpoints/{timestamp}.json
    │           (人間のベースライン記録)
    │
[2] Claude Codeが編集
    │
    ├─> post-tool-use hook
    │   └─> aict checkpoint --author "Claude Code" --message "Claude Code edits"
    │       └─> .git/aict/checkpoints/{timestamp}.json
    │           (AIの変更記録)
    │
[3] 人間が追加編集（任意）
    │
    └─> git commit
        │
        ├─> post-commit hook
        │   └─> aict commit
        │       ├─> Checkpointsを読み込み
        │       ├─> git diff HEAD~1 HEAD --numstat で完全な差分を取得
        │       ├─> Authorship Logに変換（差分 + チェックポイント作成者マッピング）
        │       └─> Git notesに保存
        │           └─> refs/aict/authorship
        │               (コミット単位の作成者情報)
        │
        └─> Checkpointsクリア

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
    ├─> 各コミットのnumstatを取得
    │   └─> git show --numstat <commit>
    │       └─> 実際の追加/削除行数を取得
    │
    ├─> 作成者別集計
    │   ├─> Authorship Logの行範囲から作成者割合を計算
    │   ├─> numstatの追加/削除行数を作成者割合で按分
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
aict checkpoint --author y-hirakaw --message "Before Claude Code edits"
    ↓
handleCheckpoint()
    ├─> Gitリポジトリルートに移動（v1.1.7）
    │   └─> git rev-parse --show-toplevel
    │
    ├─> captureSnapshot()
    │   └─> git ls-files --cached --others --exclude-standard (v1.1.8)
    │       ├─> 追跡済みファイルを取得
    │       ├─> 未追跡の新規ファイルも取得
    │       └─> 各ファイルのハッシュと行数を記録
    │
    ├─> detectChangesFromSnapshot()
    │   └─> 前回チェックポイントとの差分を検出
    │
    └─> .git/aict/checkpoints/{timestamp}.json に保存
```

**データ例**:
```json
{
  "timestamp": "2025-12-13T10:00:00Z",
  "author": "y-hirakaw",
  "type": "human",
  "metadata": {
    "message": "Before Claude Code edits"
  },
  "changes": {},
  "snapshot": {
    "internal/api/handler.go": {
      "hash": "abc123...",
      "lines": 150
    }
  }
}
```

#### 1.2 Post-tool-use Hook（編集後）

```
Claude Codeが編集完了
    ↓
.claude/hooks/post-tool-use.sh
    ↓
aict checkpoint --author "Claude Code" --message "Claude Code edits"
    ↓
handleCheckpoint()
    ├─> Gitリポジトリルートに移動
    ├─> captureSnapshot()
    │   └─> 全ファイル（追跡済み + 新規）のスナップショット
    │
    ├─> detectChangesFromSnapshot()
    │   └─> 前回チェックポイント（pre-tool-use）との差分
    │       ├─> 新規ファイル検出
    │       ├─> 変更ファイル検出（git diffで詳細取得）
    │       └─> 削除ファイル検出
    │
    └─> .git/aict/checkpoints/{timestamp}.json に保存
```

**データ例**:
```json
{
  "timestamp": "2025-12-13T10:15:00Z",
  "author": "Claude Code",
  "type": "ai",
  "metadata": {
    "message": "Claude Code edits"
  },
  "changes": {
    "internal/api/handler.go": {
      "added": 50,
      "deleted": 10,
      "lines": [[1, 50], [75, 100]]
    }
  },
  "snapshot": {
    "internal/api/handler.go": {
      "hash": "def456...",
      "lines": 190
    }
  }
}
```

### 2. Authorship Log生成フェーズ

#### 2.1 Post-commit Hook（コミットベース追跡）

```
ユーザーアクション: git commit
    ↓
.git/hooks/post-commit
    ↓
aict commit
    ↓
handleCommit()
    ├─> LoadCheckpoints()
    │   └─> .git/aict/checkpoints/*.json を読み込み
    │
    ├─> getCommitDiff()
    │   └─> git diff HEAD~1 HEAD --numstat
    │       └─> 完全なコミット差分を取得（全ファイル）
    │           ├─> 追加行数
    │           ├─> 削除行数
    │           └─> ファイルパス
    │
    ├─> buildAuthorshipMap()
    │   └─> チェックポイントから作成者マッピングを構築
    │       └─> filepath -> 最後に変更したチェックポイント
    │
    ├─> buildAuthorshipLogFromDiff()
    │   ├─> コミット差分の各ファイルについて
    │   │   ├─> authorMapから作成者を取得
    │   │   ├─> チェックポイントがない場合はデフォルト作成者
    │   │   └─> 行範囲情報を付与
    │   │
    │   └─> AuthorshipLog を生成
    │
    ├─> ValidateAuthorshipLog()
    │
    ├─> NotesManager.AddAuthorshipLog()
    │   └─> git notes --ref=refs/aict/authorship add <commit>
    │
    └─> ClearCheckpoints()
```

**AuthorshipLog データ例**:
```json
{
  "version": "1.0",
  "commit": "a1b2c3d4e5f6...",
  "timestamp": "2025-12-13T10:30:00Z",
  "files": {
    "internal/api/handler.go": {
      "authors": [
        {
          "name": "Claude Code",
          "type": "ai",
          "lines": [[1, 50], [75, 100]],
          "metadata": {
            "message": "Claude Code edits"
          }
        }
      ]
    },
    "internal/api/routes.go": {
      "authors": [
        {
          "name": "y-hirakaw",
          "type": "human",
          "lines": [],
          "metadata": {
            "message": "No checkpoint found, assigned to default author"
          }
        }
      ]
    }
  }
}
```

### 3. レポート生成フェーズ

#### 3.1 Report生成（numstat按分方式）

```
ユーザーアクション: aict report --since 7d
    ↓
handleRange()
    ├─> parseTimeFilter()
    │   └─> "7d" → 7日前の日時に変換
    │
    ├─> getCommitRange()
    │   └─> git log --since "7 days ago" --format=%H
    │
    ├─> 各コミットについて
    │   ├─> NotesManager.GetAuthorshipLog()
    │   │   └─> git notes --ref=refs/aict/authorship show <commit>
    │   │
    │   ├─> git show --numstat <commit>
    │   │   └─> 実際の追加/削除行数を取得
    │   │
    │   └─> 作成者別集計
    │       ├─> Authorship Logの行範囲から作成者割合を計算
    │       ├─> numstatの追加/削除行数を割合で按分
    │       ├─> 削除のみファイルの特別処理（v1.1.9）
    │       │   └─> totalAuthorLines==0 && 作成者1人 → 全削除行を割り当て
    │       └─> 詳細メトリクス計算
    │           ├─> コードベース貢献（追加行のみ）
    │           └─> 作業量貢献（追加+削除）
    │
    └─> printTableReport() / printJSONReport()
```

---

## データ構造

### CheckpointV2（ファイル保存形式）

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

### AuthorshipLog（Git notes保存形式）

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

### DetailedMetrics（レポート表示）

```go
type DetailedMetrics struct {
    Contributions ContributionMetrics `json:"contributions"` // コードベース貢献
    WorkVolume    WorkVolumeMetrics   `json:"work_volume"`   // 作業量貢献
    NewFiles      NewFileMetrics      `json:"new_files"`     // 新規ファイル
}

type ContributionMetrics struct {
    AIAdditions    int `json:"ai_additions"`
    HumanAdditions int `json:"human_additions"`
}

type WorkVolumeMetrics struct {
    AIAdded      int `json:"ai_added"`
    AIDeleted    int `json:"ai_deleted"`
    AIChanges    int `json:"ai_changes"`
    HumanAdded   int `json:"human_added"`
    HumanDeleted int `json:"human_deleted"`
    HumanChanges int `json:"human_changes"`
}
```

---

## ストレージ構成

### ディレクトリ構造

```
.git/
├── aict/                           # AICT専用ディレクトリ
│   ├── config.json                 # プロジェクト設定
│   ├── checkpoints/
│   │   ├── {timestamp1}.json       # CheckpointV2形式
│   │   └── {timestamp2}.json
│   └── hook.log                    # フック実行ログ（v1.1.6）
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

### フック実行ログ（v1.1.5+）

`.git/aict/hook.log`:
```
[2025-12-13 10:00:00] pre-tool-use: Recording checkpoint for y-hirakaw
[DEBUG] Checkpoint: author=y-hirakaw, files=0 (no changes)
[2025-12-13 10:00:00] pre-tool-use: Checkpoint recorded successfully
[2025-12-13 10:15:00] post-tool-use: Recording checkpoint for Claude Code
[DEBUG] Checkpoint: author=Claude Code, files=5, changes=[file1.go file2.go ...]
[2025-12-13 10:15:00] post-tool-use: Checkpoint recorded successfully
```

---

## 既知の制限

### 1. Bash rmでのファイル削除（v1.2.0で対応見送り）

**問題**:
- `rm`コマンドや直接的なファイル削除はClaude Codeフックをトリガーしない
- post-tool-useフックでファイルが存在しないため、チェックポイントに記録されない
- 結果: 削除がデフォルト作成者（人間）に割り当てられる可能性

**影響**:
- ファイル削除の作成者が不正確になる場合がある
- 削除行数が誤って人間に計上される可能性

**軽減策**:
- プロダクションコードではrmコマンド禁止が一般的
- ファイル削除操作は稀
- 全体の統計への影響は小さい

**判断**:
- コードの複雑性増加を避けるため、v1.2.0では対応見送り
- 実用上許容可能な誤差レベル

### 2. 行範囲の精度

**問題**:
- `git diff --numstat`では行範囲は概算
- 削除のみのファイルは行範囲が空（`lines: []`）

**対応**:
- v1.1.9で削除のみファイルの特別処理を実装
- `totalAuthorLines==0 && 作成者1人`の場合、全削除行を割り当て

---

## レポート表示例

### 実際の出力例

```bash
$ aict report --since 7d

📊 AI Code Generation Report (since 7d)

Commits: 5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【コードベース貢献】（最終的なコード量への寄与）
  総変更行数: 290行
    🤖 AI追加:      290行 (100.0%)
    👤 人間追加:      0行 (0.0%)

【作業量貢献】（実際の作業量）
  総作業量: 317行
    🤖 AI作業:      303行 (95.6%)
       └ 追加: 290行, 削除: 13行
    👤 人間作業:     14行 (4.4%)
       └ 追加: 0行, 削除: 14行

By Author:
  🤖 Claude Code             290行追加 (100.0%) - 1 commits
  👤 y-hirakaw                 0行追加 (0.0%) - 1 commits
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

AICT v1.2.0のデータフローは以下の特徴があります：

### アーキテクチャ
1. **記録** - フック経由でCheckpointV2形式で詳細記録
2. **変換** - コミット時にgit diff + チェックポイントマッピングでAuthorship Log生成
3. **集計** - レポート生成時にnumstat按分方式で正確な統計計算

### 主要改善（v1.1.x → v1.2.0）
- ✅ **v1.1.7**: Gitリポジトリルートからの一貫したパス処理
- ✅ **v1.1.8**: 未追跡ファイル（新規ファイル）の完全サポート
- ✅ **v1.1.9**: 削除のみファイルの正確な按分
- ✅ **v1.2.0**: 不完全な機能を削除、シンプルで安定した実装

### 利点
- ✅ **正確性**: コミットベースの完全な差分追跡
- ✅ **永続性**: Git notesによるコミット単位の履歴管理
- ✅ **柔軟性**: 複数のメトリクスによる多角的分析
- ✅ **同期性**: Git notesを使ったリモート同期
- ✅ **直感性**: `--since`は期間内の変更のみを集計（重複なし）
- ✅ **保守性**: シンプルで理解しやすいコードベース
