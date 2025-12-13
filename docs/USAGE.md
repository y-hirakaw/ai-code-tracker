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
- **Post-tool-use hook**: Claude Code編集後にAIチェックポイントを自動記録
- **Post-commit hook**: コミット時に自動的にAuthorship Logを生成

**フックセットアップ後は、手動でチェックポイント記録する必要はありません！**

### 2-a. 手動でチェックポイントを記録する場合

フックを使わない場合、または手動で記録したい場合:

```bash
# 人間が書いた場合
aict checkpoint --author "Your Name"

# AIが生成した場合
aict checkpoint --author "Claude Code"

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

#### 詳細メトリクス（--detailedオプション）

`--detailed` フラグを使用すると、3つの異なる視点でコード貢献度を確認できます：

```bash
# 詳細メトリクス付きレポート
aict report --since 7d --detailed
```

**3つのメトリクス視点**:
1. **コードベース貢献**：純粋な追加行数（最終的なコード量への寄与）
2. **作業量貢献**：追加+削除の合計（実際の作業量）
3. **新規ファイル**：完全新規のコードのみ（新規ファイルがある場合のみ表示）

これにより、リファクタリング作業（削除+書き直し）やコード整理（削除のみ）も適切に評価されます。

標準レポート（`--detailed`なし）では、最終的なコード量への寄与（追加行数）のみが表示されます。

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
| `--detailed` | 詳細メトリクス表示（コードベース貢献、作業量貢献、新規ファイル） | 無効 |

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

### テーブル形式（標準）

```
📊 AI Code Generation Report (since 7d)

Commits: 5
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Lines Changed: 52行
  🤖 AI-generated:     2行 (3.8%)
  👤 Human-written:   50行 (96.2%)

By Author:
  👤 y-hirakaw                50行追加 (96.2%) - 5 commits
  🤖 Claude Code               2行追加 (3.8%) - 1 commits
```

### テーブル形式（--detailed付き）

```
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

## データの削除・リセット

AICTのトラッキングデータを削除したい場合（他ツールへの移行、テストデータのクリア等）は、以下のコマンドを使用します。

### チェックポイントのみ削除

```bash
aict debug clean
```

**削除されるもの**:
- `.git/aict/checkpoints/latest.json` (作業中のチェックポイント)

**残るもの**:
- Git notes (`refs/aict/authorship`) - コミット済みのAuthorship Log
- 設定ファイル (`.git/aict/config.json`)

**用途**:
- 開発中の不要なチェックポイントをクリア
- コミット前に記録をリセット

### Git notesを削除

```bash
aict debug clear-notes
```

**削除されるもの**:
- `refs/aict/authorship` - Authorship Log
- `refs/notes/aict` など、AICT関連のすべてのGit notes

**残るもの**:
- チェックポイント (`.git/aict/checkpoints/`)
- 設定ファイル (`.git/aict/config.json`)

**用途**:
- コミット済みのAuthorship Log履歴を完全削除
- プロジェクトのトラッキング履歴をリセット

### 完全削除（AICTを完全にアンインストール）

```bash
# 1. チェックポイントを削除
aict debug clean

# 2. Git notesを削除
aict debug clear-notes

# 3. 設定ファイルとディレクトリを削除
rm -rf .git/aict

# 4. フックを削除（必要に応じて）
rm .git/hooks/post-commit
rm .claude/hooks/pre-tool-use.sh
rm .claude/hooks/post-tool-use.sh
```

**用途**:
- AICTを完全にアンインストール
- 他のツール（Claude Code公式機能等）への移行

### リモートのGit notesも削除

ローカルで`aict debug clear-notes`を実行後、リモートにもプッシュ済みの場合：

```bash
# リモートのGit notesを削除
git push origin :refs/aict/authorship
git push origin :refs/notes/aict
```

**注意**: チームで使用している場合は、他のメンバーに事前に通知してください。

---

## トラブルシューティング

### 既存コードが誤ってカウントされている

既存のコードベースにaictを導入した際、既存コード全体が「変更」としてカウントされてしまった場合：

```bash
# 1. チェックポイントデータをクリア
aict debug clean

# 2. （必要に応じて）Git notesもクリア
aict debug clear-notes

# 3. 新しいベースラインチェックポイントを作成
aict checkpoint --author "Your Name"
```

**説明**: aictはスナップショットベースのトラッキングを使用しており、初回チェックポイント時に既存コードを除外します。古いデータがある場合は、上記の手順でクリーンアップして再スタートしてください。

### レポート結果がおかしい

計測結果に異常がある場合、デバッグコマンドで状態を確認できます：

```bash
# チェックポイントの詳細を表示
aict debug show

# 問題があれば、データをリセット
aict debug clean          # チェックポイントのみ削除
aict debug clear-notes    # Git notesも削除
```

### チェックポイントが記録されない

- 追跡対象の拡張子（`.go`, `.py`等）のファイルを編集していることを確認
- `git diff` で変更が検出されることを確認
- `aict debug show` でチェックポイントの状態を確認

### Authorship Logが生成されない

- チェックポイントが記録されていることを確認: `ls .git/aict/checkpoints/`
- Git notesを確認: `git notes --ref=refs/aict/authorship show HEAD`
- チェックポイントがない場合: `aict checkpoint` で手動作成

### フックが動作しない

- フックファイルが実行可能か確認: `ls -la .git/hooks/post-commit`
- `.claude/settings.json` が正しく設定されているか確認
- `aict` コマンドがPATHに含まれているか確認
- フックの再セットアップ: `aict setup-hooks`

## 既知の制限事項

### Bashコマンドによるファイル削除

`rm` コマンドなどのBashツールで直接ファイル削除した場合、Claude Code hooksが実行されないため、削除が人間の作業として記録される可能性があります。

**影響範囲**:
- ファイル削除のみに影響
- ファイル編集や追加には影響なし

**推奨事項**:
- プロダクションコードでは `rm` コマンドの使用を禁止する運用が多い
- ファイル削除自体が頻繁に発生するケースは少ない
- 全体的な追跡精度への影響は限定的（99%以上の精度を維持）

## 詳細仕様

完全な仕様については以下のドキュメントを参照してください：
- [BASE_SPEC.md](./BASE_SPEC.md) - 基本仕様
- [DATA_FLOW.md](./DATA_FLOW.md) - データフロー詳細
