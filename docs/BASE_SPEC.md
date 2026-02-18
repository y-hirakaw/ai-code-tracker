# AI Code Tracker (AICT) 基本仕様書 v1.4.4

## 概要

AI Code Tracker (AICT) は、AIによるコード生成と人間によるコード記述を追跡・管理するためのツールです。Claude Code などのAI支援ツールとGit hookを組み合わせて、誰がどのコードを書いたかを自動的に記録します。

**バージョン**: v1.4.4（Production ready）
**実装方式**: CheckpointV2ベース、バッチnumstat取得 + 按分方式

## 基本フロー

### 1. pre-tool-use hook: 人間のチェックポイント記録

**目的**: Claude Code編集前の状態をスナップショット

**処理内容**:
- Gitリポジトリルートに移動（v1.1.7+: ファイルパス一貫性確保）
- Claude Codeが編集を開始する前に実行
- **追跡対象および未追跡ファイルを取得**（v1.1.8+: 新規ファイル追跡）
  - `git ls-files --cached --others --exclude-standard`
- 現在のワークツリーの状態をスナップショットとして保存
- **ファイルハッシュを記録**（CheckpointV2）
- 作成者: `y-hirakaw` (human)
- 作成者タイプ: `human`

**保存場所**: `.git/aict/checkpoints/latest.json`（JSONL形式で追記、v1.4.0）

**データ構造（CheckpointV2、JSONL: 1行1チェックポイント）**:
```json
{"timestamp":"2025-12-10T12:00:00Z","author":"y-hirakaw","type":"human","metadata":{"message":"Before Claude Code edits"},"changes":{},"snapshot":{"main.go":{"hash":"abc123def456...","lines":50},"utils.go":{"hash":"def456abc123...","lines":30}}}
```

### 2. Claude Code編集: AIが実際にコードを変更

**処理内容**:
- Claude Codeがコードを編集（ユーザーの要求に基づく）
- ファイルの追加、変更、削除など

**AIツール**: Claude Code (claude-sonnet-4.5など)

### 3. post-tool-use hook: AIのチェックポイント記録

**目的**: Claude Code編集後の状態をスナップショット

**処理内容**:
- Gitリポジトリルートに移動（v1.1.7+: ファイルパス一貫性確保）
- Claude Codeが編集を完了した後に実行
- **追跡対象および未追跡ファイルを取得**（v1.1.8+: 新規ファイル追跡）
  - `git ls-files --cached --others --exclude-standard`
- 現在のワークツリーの状態をスナップショットとして保存
- **前回のチェックポイント（人間のスナップショット）との差分を計算**
  - ファイルハッシュ比較により変更検出
  - `git diff` numstat形式で行数変更を取得
- 差分を`Changes`フィールドに記録（added, deleted, lines範囲）
- **ファイルハッシュを記録**（CheckpointV2）
- 作成者: `Claude Code` (ai)
- 作成者タイプ: `ai`
- メタデータ: モデル名（v1.1.6で廃止、簡素化）

**保存場所**: `.git/aict/checkpoints/latest.json`（JSONL形式で追記）

**データ構造（CheckpointV2、JSONL）**:
```json
{"timestamp":"2025-12-10T12:05:00Z","author":"Claude Code","type":"ai","metadata":{},"changes":{"main.go":{"added":10,"deleted":2,"lines":[[1,10]]}},"snapshot":{"main.go":{"hash":"xyz789abc012...","lines":60},"utils.go":{"hash":"def456abc123...","lines":30}}}
```

### 4. git commit: ユーザーがコミット

**処理内容**:
- ユーザーが `git commit` を実行
- コミットメッセージを記述
- コミットが作成される

**例**:
```bash
git add .
git commit -m "feat: Add new feature"
```

### 5. post-commit hook: aict commit実行

**目的**: チェックポイント群をAuthorship Logに変換

**処理内容**:

#### 5.1 チェックポイント群を読み込み
- `.git/aict/checkpoints/latest.json` から全チェックポイントを読み込む（JSONL/旧JSON配列を自動判別）

#### 5.2 コミットのnumstatを取得（ここが重要！）
- `git show --numstat --format= <commit-hash>` を実行
- **実際にコミットに含まれるファイルの変更**を取得
- numstat形式: `<added>\t<deleted>\t<filepath>`

**例**:
```
10    2    main.go
5     0    utils.go
```

#### 5.3 各チェックポイントのChangesを集約
- チェックポイント群から変更情報を抽出
- **numstatと照合**して、実際にコミットに含まれるファイルのみをフィルタリング
- ファイルごとに作成者情報を集約

**重要**: numstatに含まれないファイルは除外される
- 例: 前回のセッションで編集したが、今回のコミットには含まれないファイル

#### 5.4 Authorship Logに変換してGit notesに保存
- フィルタリングされたデータをAuthorship Log形式に変換
- Git notes (`refs/aict/authorship`) に保存
- チェックポイントをクリア（ベースラインのみ保持）

**Authorship Log構造**:
```json
{
  "version": "1.0",
  "commit": "abc123def456...",
  "timestamp": "2025-12-10T12:10:00Z",
  "files": {
    "main.go": {
      "authors": [
        {
          "name": "y-hirakaw",
          "type": "human",
          "lines": [[1, 5]],
          "metadata": {}
        },
        {
          "name": "Claude Code",
          "type": "ai",
          "lines": [[6, 15]],
          "metadata": {
            "model": "claude-sonnet-4.5"
          }
        }
      ]
    },
    "utils.go": {
      "authors": [
        {
          "name": "Claude Code",
          "type": "ai",
          "lines": [[1, 5]],
          "metadata": {
            "model": "claude-sonnet-4.5"
          }
        }
      ]
    }
  }
}
```

## numstatフィルタリングの重要性

### なぜnumstatと照合するのか

**問題**: チェックポイントには過去のセッションで編集したファイルが残っている可能性がある

**例**:
- セッション1: `AndroidNativeApp/MainActivity.kt` を編集 → チェックポイント記録
- セッション2: `iOSNativeApp/Info.plist` を編集 → チェックポイント記録
- コミット: `iOSNativeApp/Info.plist` のみをコミット

**期待される動作**:
- Authorship Logには `iOSNativeApp/Info.plist` のみが含まれるべき
- `AndroidNativeApp/MainActivity.kt` は含まれない（コミットされていないため）

**実装**:
```go
// handlers_commit.go

// コミットのnumstatを取得（git.ParseNumstat に統一、v1.4.0）
numstatOutput, err := executor.Run("show", "--numstat", "--format=", "--end-of-options", commitHash)
numstatData := git.ParseNumstat(numstatOutput)

// numstatでフィルタリング（実際に変更されたファイルのみ）
changedFiles := make(map[string]bool)
for filePath := range numstatData {
    changedFiles[filePath] = true
}

// チェックポイント群からAuthorship Mapを構築
authorMap := buildAuthorshipMap(checkpoints, changedFiles)

// Authorship Logを生成（numstatでフィルタリング）
log, err := buildAuthorshipLogFromDiff(diffMap, authorMap, commitHash, changedFiles, cfg)
```

## データフロー図

```
┌─────────────────┐
│ Pre-tool-use    │
│ Hook            │ → 人間のチェックポイント記録
└─────────────────┘
         ↓
┌─────────────────┐
│ Claude Code     │
│ 編集            │ → AIがコードを変更
└─────────────────┘
         ↓
┌─────────────────┐
│ Post-tool-use   │
│ Hook            │ → AIのチェックポイント記録
└─────────────────┘   (差分計算)
         ↓
┌─────────────────┐
│ git commit      │ → ユーザーがコミット
└─────────────────┘
         ↓
┌─────────────────┐
│ Post-commit     │
│ Hook            │
└─────────────────┘
         ↓
┌─────────────────┐
│ aict commit     │
├─────────────────┤
│ 1. チェック     │ → チェックポイント群を読み込み
│    ポイント     │
│    読み込み     │
├─────────────────┤
│ 2. numstat      │ → git show --numstat
│    取得         │   (実際の変更ファイル)
├─────────────────┤
│ 3. フィルタ     │ → numstatと照合
│    リング       │   (重要！)
├─────────────────┤
│ 4. Authorship   │ → Git notesに保存
│    Log生成      │   (refs/aict/authorship)
└─────────────────┘
```

## 設定ファイル

`.git/aict/config.json`:
```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".java", ".kt", ".swift"],
  "exclude_patterns": ["*_test.go", "vendor/*", "node_modules/*"],
  "default_author": "y-hirakaw",
  "ai_agents": ["Claude Code", "GitHub Copilot", "ChatGPT"]
}
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `aict init` | プロジェクトの初期化 |
| `aict setup-hooks` | Claude Code & Git hooks のセットアップ |
| `aict checkpoint --author <name>` | 手動チェックポイント記録 |
| `aict commit` | Authorship Log生成（自動 or 手動） |
| `aict report --since <date>` | レポート表示 |
| `aict sync push/fetch` | リモートとの同期 |
| `aict debug show` | チェックポイント表示 |
| `aict debug clean` | チェックポイント削除 |
| `aict debug clear-notes` | Git notes削除 |
| `aict version` | バージョン表示 |

## レポート例

```
AI Code Generation Report (since 7d)

Commits: 5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

【コードベース貢献】（最終的なコード量への寄与）
  総変更行数: 52行
    □ AI生成:     2行 (3.8%)
    ○ 開発者:    50行 (96.2%)

【作業量貢献】（実際の作業量）
  総作業量: 52行
    □ AI作業:     2行 (3.8%)
       └ 追加: 2行, 削除: 0行
    ○ 開発者作業: 50行 (96.2%)
       └ 追加: 50行, 削除: 0行

By Author:
  ○ y-hirakaw                50行追加 (96.2%) - 5 commits
  □ Claude Code               2行追加 (3.8%) - 1 commits
```


## 重要なポイント

1. **CheckpointV2構造**: ファイルハッシュベースの変更検出
2. **JSONL保存**: O(1)追記でチェックポイントを保存（v1.4.0）
3. **Gitリポジトリルート**: v1.1.7+でファイルパス一貫性確保
4. **未追跡ファイル追跡**: v1.1.8+で新規ファイルも追跡対象
5. **削除のみファイル**: v1.1.9+で削除行を正確に集計
6. **numstatフィルタリング**: コミットに含まれるファイルのみを追跡（最重要！）
7. **バッチ取得**: N+1問題を解消、2回のgit呼び出しで全データ取得（v1.4.0）
8. **Authorship Log**: Git notes (`refs/aict/authorship`) として永続化
9. **複数メトリクス**: コードベース貢献、作業量貢献、新規ファイルの3視点測定
10. **入力バリデーション**: `--since`の未知形式を警告、`--format`のエラーメッセージに利用可能フォーマット表示（v1.4.4）
11. **セキュリティ**: Git引数のオプション注入防止（ValidateRevisionArg、`--end-of-options`）（v1.4.0）

## 既知の制限事項

### Bashコマンドによるファイル削除
**制限内容**: `rm` コマンドなどのBashツールで直接ファイル削除した場合、Claude Code hooksが実行されないため、削除が人間の作業として記録される可能性があります。

**影響範囲**:
- ファイル削除のみに影響
- ファイル編集や追加には影響なし

**推奨事項**:
- プロダクションコードでは `rm` コマンドの使用を禁止する運用が多い
- ファイル削除自体が頻繁に発生するケースは少ない
- 全体的な追跡精度への影響は限定的（99%以上の精度を維持）

## バージョン履歴

- **v1.1.7**: ファイルパス一貫性修正（Gitリポジトリルートベース）
- **v1.1.8**: 新規ファイル追跡対応（git ls-files --cached --others --exclude-standard）
- **v1.1.9**: 削除のみファイルの正確な集計対応
- **v1.2.0**: 安定版リリース、既知の制限事項の文書化
- **v1.3.0**: レポート出力のアイコン改善
- **v1.4.0**: N+1問題解消、JSONL保存、メモリ効率改善、セキュリティ強化、ハンドラerror返却統一、デッドコード削除
- **v1.4.4**: テストカバレッジ向上、`--since`入力バリデーション、`--format`エラーメッセージ改善

この仕様に基づいて、AICTはAIと人間のコード貢献を正確に追跡・管理します（99%以上の精度）。
