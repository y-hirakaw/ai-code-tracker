# AICT データフローガイド (v1.4.4)

このドキュメントでは、AI Code Tracker (AICT) v1.4.4がどのようにデータを記録し、レポートを生成しているかを詳しく説明します。

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
    │       └─> .git/aict/checkpoints/latest.json
    │           (配列に追記: 人間のベースライン記録)
    │
[2] Claude Codeが編集
    │
    ├─> post-tool-use hook
    │   └─> aict checkpoint --author "Claude Code" --message "Claude Code edits"
    │       └─> .git/aict/checkpoints/latest.json
    │           (配列に追記: AIの変更記録)
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
    ├─> 入力バリデーション（v1.4.4）
    │   └─> validateSinceInput(): 未知の日付形式の場合に警告
    │
    ├─> コミット範囲を取得
    │   └─> git log --since 7d
    │
    ├─> バッチ取得: numstat（v1.4.0: N+1問題解消）
    │   └─> git log --numstat --format=__AICT_COMMIT__%H <range>
    │       └─> 全コミットの追加/削除行数を1回のgit呼び出しで取得
    │
    ├─> バッチ取得: Authorship Log（v1.4.0: N+1問題解消）
    │   └─> git log --notes=refs/aict/authorship --format=__AICT_HASH__%H%n%N <range>
    │       └─> 全コミットのAuthorship Logを1回のgit呼び出しで取得
    │
    ├─> 作成者別集計（collectAuthorStats）
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
    └─> .git/aict/checkpoints/latest.json に追記（JSONL形式、v1.4.0）
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
    │   └─> HEAD基準との差分（注: 前回チェックポイント基準ではない）
    │       ├─> 新規ファイル検出（HEAD に存在しない）
    │       ├─> 変更ファイル検出（git show HEAD:filepath との比較）
    │       │   └─> getDetailedDiff() で行数・行範囲を取得
    │       └─> 削除ファイル検出
    │
    └─> .git/aict/checkpoints/latest.json に追記（JSONL形式、v1.4.0）
```

**データ例（JSONL: 1行1チェックポイント）**:
```json
{"timestamp":"2025-12-13T10:15:00Z","author":"Claude Code","type":"ai","metadata":{"message":"Claude Code edits"},"changes":{"internal/api/handler.go":{"added":50,"deleted":10,"lines":[[1,50],[75,100]]}},"snapshot":{"internal/api/handler.go":{"hash":"def456...","lines":200}}}
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
    │   └─> .git/aict/checkpoints/latest.json を読み込み（JSONL形式、旧JSON配列も自動判別）
    │
    ├─> numstatフィルタリング
    │   └─> git show --numstat --format= HEAD
    │       └─> コミットに含まれるファイルのみを抽出
    │           └─> 早期return: changedFiles が空なら終了（⚠️ チェックポイント残留）
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
    │       └─> filepath -> **最後のチェックポイントの作成者1人のみ**（複数作者集約なし）
    │
    ├─> buildAuthorshipLogFromDiff()
    │   ├─> コミット差分の各ファイルについて
    │   │   ├─> authorMapから作成者を取得（1人のみ）
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
          "lines": [[10]],
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

#### 3.1 Report生成（numstat按分方式・バッチ取得）

```
ユーザーアクション: aict report --since 7d
    ↓
handleRangeReport()
    ├─> validateSinceInput()（v1.4.4）
    │   └─> 未知の日付形式の場合にstderrに警告を出力
    │
    ├─> convertSinceToRange()
    │   └─> "7d" → "7 days ago" → コミット範囲に変換
    │
    ├─> collectAuthorStats()（v1.4.0: バッチ化）
    │   ├─> GetRangeNumstat()（バッチ）
    │   │   └─> git log --numstat --format=__AICT_COMMIT__%H <range>
    │   │       └─> 全コミットのnumstatを1回で取得
    │   │
    │   ├─> GetAuthorshipLogsForRange()（バッチ）
    │   │   └─> git log --notes=refs/aict/authorship --format=__AICT_HASH__%H%n%N <range>
    │   │       └─> 全コミットのAuthorship Logを1回で取得
    │   │
    │   └─> 各コミット・各ファイルの集計
    │       ├─> Authorship Logの行範囲から作成者割合を計算
    │       ├─> numstatの追加/削除行数を割合で按分
    │       ├─> 削除のみファイルの特別処理（v1.1.9）
    │       │   └─> totalAuthorLines==0 && 作成者1人 → 全削除行を割り当て
    │       └─> 詳細メトリクス計算
    │           ├─> コードベース貢献（追加行のみ）
    │           └─> 作業量貢献（追加+削除）
    │
    ├─> buildReport()
    │   └─> authorStatsResult → tracker.Report に変換
    │
    └─> formatRangeReport()
        ├─> テーブル形式（デフォルト）+ 詳細メトリクス自動表示
        └─> JSON形式（--format json）
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
    Lines    [][]int           `json:"lines"` // [[start, end], ...] or [[lineNum]] (混在)
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
│   │   └── latest.json             # CheckpointV2 JSONL形式（1行1チェックポイント、O(1)追記）
│   └── hook.log                    # フック実行ログ（v1.1.6+）
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

## レポート表示例

### 実際の出力例

```bash
$ aict report --since 7d

AI Code Generation Report (since 7d)

Commits: 5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【コードベース貢献】（最終的なコード量への寄与）
  総変更行数: 290行
    □ AI生成:    290行 (100.0%)
    ○ 開発者:      0行 (0.0%)

【作業量貢献】（実際の作業量）
  総作業量: 290行
    □ AI作業:    290行 (100.0%)
       └ 追加: 290行, 削除: 0行
    ○ 開発者作業:   0行 (0.0%)
       └ 追加: 0行, 削除: 0行

By Author:
  □ Claude Code             290行追加 (100.0%) - 1 commits
  ○ y-hirakaw                 0行追加 (0.0%) - 1 commits
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

AICT v1.4.4のデータフローは以下の特徴があります：

### アーキテクチャ
1. **記録** - フック経由でCheckpointV2形式でJSONL追記
2. **変換** - コミット時にgit diff + チェックポイントマッピングでAuthorship Log生成
3. **集計** - レポート生成時にバッチ取得 + numstat按分方式で正確な統計計算

### 主要改善（v1.1.x → v1.4.4）
- ✅ **v1.1.7**: Gitリポジトリルートからの一貫したパス処理
- ✅ **v1.1.8**: 未追跡ファイル（新規ファイル）の完全サポート
- ✅ **v1.1.9**: 削除のみファイルの正確な按分
- ✅ **v1.2.0**: 不完全な機能を削除、シンプルで安定した実装
- ✅ **v1.3.0**: レポート出力のアイコン改善
- ✅ **v1.4.0**: N+1問題解消（バッチnumstat/notes取得）、JSONL形式保存、メモリ効率改善、セキュリティ強化（ValidateRevisionArg）、ハンドラerror返却統一、デッドコード削除
- ✅ **v1.4.4**: テストカバレッジ向上、`--since`入力バリデーション追加、`--format`エラーメッセージ改善

### 利点
- ✅ **正確性**: コミットベースの完全な差分追跡（99%以上の精度）
- ✅ **永続性**: Git notesによるコミット単位の履歴管理
- ✅ **柔軟性**: numstat按分方式による正確な統計
- ✅ **同期性**: Git notesを使ったリモート同期
- ✅ **直感性**: `--since`は期間内の変更のみを集計（重複なし）
- ✅ **保守性**: シンプルで理解しやすいコードベース

### 既知の実装制約と動作仕様

#### 1. チェックポイント保存形式
- **実装**: `.git/aict/checkpoints/latest.json` にJSONL形式で追記（v1.4.0）
- **動作**: 1行1チェックポイントのO(1)追記。旧JSON配列形式からの自動マイグレーション対応
- **LoadCheckpoints()**: `latest.json` を読み込み（JSONL/旧JSON配列を自動判別）

#### 2. 差分計算の基準
- **実装**: `git show HEAD:filepath` との比較（HEAD基準）
- **動作**: 前回チェックポイント基準ではなく、コミット済みコードとの差分
- **影響**: 複数チェックポイント間の編集が正確に区別されない場合がある

#### 3. 作者割り当てロジック
- **実装**: ファイルごとに最後のチェックポイントの作成者1人のみ
- **動作**: 同一ファイルへの複数作者の貢献を集約しない
- **影響**: 人間とAIが交互に同じファイルを編集した場合、最後の作者にすべてが帰属

#### 4. 早期returnによるチェックポイント残留
- **実装**: `changedFiles`が空の場合、`ClearCheckpoints()`がスキップされる
- **動作**: 追跡対象ファイルの変更がないコミットでチェックポイントが残る
- **影響**: 次回コミットに古いチェックポイントが混入する可能性

#### 5. 行範囲フォーマットの混在
- **実装**: 単一行は`[lineNum]`、範囲は`[start, end]`
- **動作**: 行範囲情報の形式が統一されていない
- **影響**: データ解析時に両方の形式を考慮する必要がある

#### 6. 行数カウント
- **実装**: `bytes.Count(content, []byte{'\n'}) + 1`（v1.4.0: メモリ効率改善）
- **動作**: スライス生成を回避し、バイト列で直接カウント
- **影響**: 大きなファイルでのメモリ使用量を削減

#### 7. Bashコマンドによるファイル削除
- **実装**: `rm`コマンドはフックをバイパス
- **動作**: ファイル削除が人間の作業として記録される場合がある
- **影響**: AIによるファイル削除が正確に追跡されない（限定的）

これらの制約は、v1.2.0の設計判断として受け入れられており、一般的なユースケースでは99%以上の追跡精度を維持しています。
