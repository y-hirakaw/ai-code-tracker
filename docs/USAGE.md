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

コミット範囲のAI/人間のコード生成率を表示します:

```bash
# 最近5コミットのレポート
aict report --range HEAD~5..HEAD

# 特定のブランチとの差分
aict report --range origin/main..HEAD

# JSON形式で出力
aict report --range HEAD~10..HEAD --format json
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
| `aict report --range <range>` | コミット範囲のレポート表示 |
| `aict sync push` | Authorship Logをリモートにプッシュ |
| `aict sync fetch` | Authorship Logをリモートから取得 |
| `aict version` | バージョン表示 |

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

- `target_ai_percentage`: 目標AI生成率 (%)
- `tracked_extensions`: 追跡するファイル拡張子
- `exclude_patterns`: 除外パターン
- `default_author`: デフォルト作成者名
- `ai_agents`: AIエージェントのリスト (ここに含まれる名前は自動的にAIとして分類)

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
   # PR作成前に確認
   aict report --range origin/main..HEAD

   # 最近の開発状況確認
   aict report --range HEAD~10..HEAD
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
