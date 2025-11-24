# AI Code Tracker (AICT) 使い方

## インストール

```bash
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest
```

ビルドする場合:
```bash
git clone https://github.com/y-hirakaw/ai-code-tracker.git
cd ai-code-tracker
go build -o bin/aict ./cmd/aict
```

## 基本的な使い方

### 1. 初期化

```bash
cd /path/to/your-project
aict init
```

`.git/aict/` ディレクトリが作成され、設定ファイル `config.json` が生成されます。

### 2. フックのセットアップ（推奨）

Claude Codeとの統合による自動トラッキングを有効にします:

```bash
aict setup-hooks
```

これにより以下がセットアップされます:
- **Pre-tool-use hook**: Claude Code編集前に人間のチェックポイントを自動記録
- **Post-tool-use hook**: Claude Code編集後にAIチェックポイントを自動記録（モデル: claude-sonnet-4.5）
- **Post-commit hook**: コミット時に自動的にAuthorship Logを生成

**フックセットアップ後は、手動でチェックポイント記録する必要はありません！**

### 2-a. 手動でチェックポイントを記録する場合

フックを使わない場合、または手動で記録したい場合:

```bash
# 人間が書いた場合
aict checkpoint --author "Your Name"

# AIが生成した場合
aict checkpoint --author "Claude Code" --model "claude-sonnet-4.5"

# メッセージ付き
aict checkpoint --author "Your Name" --message "Implemented feature X"
```

### 3. コミット

#### フックセットアップ済みの場合

通常通りgitコミットするだけで自動的にAuthorship Logが生成されます:

```bash
git add .
git commit -m "Your commit message"
# → post-commit hookが自動的に aict commit を実行
```

#### 手動の場合

コミット後に明示的に `aict commit` を実行します:

```bash
git add .
git commit -m "Your commit message"
aict commit
```

`aict commit` により、チェックポイントがAuthorship Logに変換され、Git notes (`refs/aict/authorship`) に保存されます。

### 4. レポート表示

コミット範囲のAI/人間のコード生成率を表示します。

#### 日付ベースのフィルタリング（--since）

```bash
# 過去7日間のレポート（簡潔表記）
aict report --since 7d

# 過去2週間のレポート
aict report --since 2w

# 過去1ヶ月のレポート
aict report --since 1m

# 相対日付指定
aict report --since yesterday
aict report --since '7 days ago'

# 絶対日付指定
aict report --since '2025-01-15'
```

#### コミット範囲指定（--range）

```bash
# 最近5コミットのレポート
aict report --range HEAD~5..HEAD

# 特定のブランチとの差分
aict report --range origin/main..HEAD
```

#### 出力フォーマット

```bash
# テーブル形式（デフォルト）
aict report --since 7d

# JSON形式
aict report --since 7d --format json

# JSON出力をファイルに保存
aict report --since 2w --format json > report.json
```

### 5. リモートとの同期

Authorship Logをリモートリポジトリと同期できます:

```bash
# リモートにプッシュ
aict sync push

# リモートから取得
aict sync fetch
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `aict init` | プロジェクトの初期化（`.git/aict/` ディレクトリ作成） |
| `aict setup-hooks` | Claude Code & Git hooks のセットアップ（推奨） |
| `aict checkpoint [options]` | チェックポイントの記録（手動の場合） |
| `aict commit` | Authorship Logの生成（自動 or 手動） |
| `aict report [options]` | コード生成統計レポート表示 |
| `aict sync push` | Authorship Logをリモートにプッシュ |
| `aict sync fetch` | Authorship Logをリモートから取得 |
| `aict version` | バージョン表示 |

## レポートコマンドのオプション

### 必須オプション（いずれか1つ）

| オプション | 説明 | 例 |
|----------|------|-----|
| `--range <range>` | コミット範囲を指定 | `origin/main..HEAD`, `HEAD~5..HEAD` |
| `--since <date>` | 指定日時以降のコミット | `7d`, `2w`, `1m`, `yesterday`, `2025-01-15` |

**注意**: `--range` と `--since` は同時に指定できません（排他的）。

### オプション

| オプション | 説明 | デフォルト |
|----------|------|-----------|
| `--format <format>` | 出力フォーマット（`table` または `json`） | `table` |

### --since の日付指定形式

| 形式 | 説明 | 例 |
|------|------|-----|
| 簡潔表記 | `<数値><単位>` 形式 | `7d` (7日), `2w` (2週間), `1m` (1ヶ月), `1y` (1年) |
| 相対日付 | Git互換の相対日付 | `yesterday`, `7 days ago`, `2 weeks ago` |
| 絶対日付 | ISO形式の日付 | `2025-01-15`, `2025-01-01` |

## チェックポイントのオプション

| オプション | 説明 | 必須 |
|----------|------|------|
| `--author <name>` | 作成者名 | ✅ 必須 |
| `--model <model>` | AIモデル名（AIエージェントの場合のみ） | AIの場合推奨 |
| `--message <msg>` | メモ・説明 | オプション |

**自動判定**: `--author` が `ai_agents` リストに含まれる場合、自動的にAIとして分類されます。

## 設定ファイル

`.git/aict/config.json` で設定をカスタマイズできます:

```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".java"],
  "exclude_patterns": ["*_test.go", "vendor/*", "node_modules/*"],
  "default_author": "Your Name",
  "ai_agents": ["Claude Code", "GitHub Copilot", "ChatGPT"]
}
```

### 設定項目の説明

| 設定項目 | 説明 | デフォルト値 |
|---------|------|-------------|
| `target_ai_percentage` | 目標AI生成率 (%) | 80.0 |
| `tracked_extensions` | トラッキング対象の拡張子 | `.go`, `.py`, `.js`, `.ts`, `.java` |
| `exclude_patterns` | 除外パターン (glob形式) | `*_test.go`, `vendor/*`, `node_modules/*` |
| `default_author` | デフォルト作成者名 | `git config user.name` の値 |
| `ai_agents` | AIエージェント名のリスト | `Claude Code`, `GitHub Copilot`, `ChatGPT` |

**重要**:
- `tracked_extensions`: この拡張子のファイルのみが追跡対象になります
- `ai_agents`: ここに含まれる名前は自動的にAIとして分類されます

## レポート出力例

### テーブル形式

```
📊 AI Code Generation Report

Range: origin/main..HEAD (5 commits)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total Lines:        150
  🤖 AI Generated:    90 (60.0%)
  👤 Human Written:   60 (40.0%)

By Author:
  🤖 Claude Code       90 lines (60.0%) - 3 commits
  👤 Your Name         60 lines (40.0%) - 2 commits

Top Files:
  main.go                                  50 lines (70% AI)
  handler.go                               40 lines (50% AI)
  utils.go                                 60 lines (100% AI)
```

### JSON形式

```json
{
  "range": "origin/main..HEAD",
  "commits": 5,
  "summary": {
    "total_lines": 150,
    "ai_lines": 90,
    "human_lines": 60,
    "ai_percentage": 60.0
  },
  "by_author": [
    {
      "name": "Claude Code",
      "type": "ai",
      "lines": 90,
      "percentage": 60.0,
      "commits": 3
    },
    {
      "name": "Your Name",
      "type": "human",
      "lines": 60,
      "percentage": 40.0,
      "commits": 2
    }
  ],
  "by_file": [
    {
      "path": "main.go",
      "total_lines": 50,
      "ai_lines": 35,
      "human_lines": 15
    }
  ]
}
```

## 推奨ワークフロー

1. **初回セットアップ**
   ```bash
   cd your-project
   aict init
   aict setup-hooks
   ```

2. **開発サイクル**（フック有効時）
   ```bash
   # Claude Codeで編集
   # → pre-tool-use hookが人間のチェックポイント記録
   # → post-tool-use hookがAIチェックポイント記録

   git add .
   git commit -m "Feature implementation"
   # → post-commit hookが自動的に aict commit 実行
   ```

3. **レポート確認**
   ```bash
   # 日次レビュー（過去24時間）
   aict report --since 1d

   # スプリント振り返り（過去2週間）
   aict report --since 2w

   # PR作成前に確認（ブランチ差分）
   aict report --range origin/main..HEAD

   # JSON形式でエクスポート
   aict report --since 7d --format json > weekly-report.json
   ```

4. **チーム共有**
   ```bash
   # Authorship LogをリモートにPush
   aict sync push

   # チームメンバーがFetch
   aict sync fetch
   ```

## トラブルシューティング

### チェックポイントが記録されない

- 追跡対象の拡張子（`.go`, `.py`等）のファイルを編集していることを確認
- `git diff` で変更が検出されることを確認

### Authorship Logが生成されない

- チェックポイントが記録されていることを確認: `ls .git/aict/checkpoints/`
- Git notesを確認: `git notes --ref=refs/aict/authorship show HEAD`

### フックが動作しない

- フックファイルが実行可能か確認: `ls -la .git/hooks/post-commit`
- `.claude-code/settings.json` が正しく設定されているか確認
- `aict` コマンドがPATHに含まれているか確認

## 詳細仕様

完全な仕様については [SPEC.md](../SPEC.md) を参照してください。
