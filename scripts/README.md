# AICT インストールスクリプト

このディレクトリには AI Code Tracker (AICT) のインストールと設定を自動化するスクリプトが含まれています。

## スクリプト一覧

### 📦 install.sh
AICT のシステムインストールを自動化するスクリプトです。

**主な機能:**
- プラットフォーム自動検出 (macOS/Linux, amd64/arm64)
- 依存関係チェック (Git, Go, Claude Code)
- バイナリのビルドまたはインストール
- システムパスへの追加
- 設定ディレクトリの作成
- インストール後のテスト

**使用例:**
```bash
# 標準インストール
bash scripts/install.sh

# カスタムディレクトリにインストール
bash scripts/install.sh --install-dir ~/bin

# hooks 設定をスキップしてインストール
bash scripts/install.sh --no-hooks

# 既存インストールを強制上書き
bash scripts/install.sh --force
```

### 🔗 setup-hooks.sh
Git hooks と Claude Code hooks の設定を自動化するスクリプトです。

**主な機能:**
- Git リポジトリの確認
- AICT の初期化
- Git post-commit hook の設定
- Claude Code hooks の設定
- 設定状況の確認
- テスト実行

**使用例:**
```bash
# 全ての hooks を設定
bash scripts/setup-hooks.sh

# Git hooks のみ設定
bash scripts/setup-hooks.sh --git-only

# Claude Code hooks のみ設定
bash scripts/setup-hooks.sh --claude-only

# 現在の設定状況を確認
bash scripts/setup-hooks.sh --status-only

# hooks を削除
bash scripts/setup-hooks.sh --remove
```

## インストール手順

### 1. 新規インストール

```bash
# 1. リポジトリをクローンまたはダウンロード
git clone <repository-url>
cd ai-code-tracker

# 2. ビルド (オプション、インストールスクリプトが自動実行)
make build

# 3. インストール
bash scripts/install.sh

# 4. Gitリポジトリで hooks を設定
cd /path/to/your/project
bash /path/to/ai-code-tracker/scripts/setup-hooks.sh
```

### 2. 手動設定

AICT がインストール済みの場合：

```bash
# 1. プロジェクト初期化
aict init

# 2. hooks 設定
aict setup

# 3. 設定確認
aict setup --status
```

### 3. 既存プロジェクトへの追加

```bash
# Git リポジトリのルートで実行
cd /path/to/your/project

# AICT を初期化
aict init

# hooks を設定
bash /path/to/ai-code-tracker/scripts/setup-hooks.sh
```

## トラブルシューティング

### 権限エラー
```bash
# ユーザーディレクトリにインストール
bash scripts/install.sh --install-dir ~/bin
```

### Git hooks が動作しない
```bash
# hooks の再設定
bash scripts/setup-hooks.sh --remove
bash scripts/setup-hooks.sh
```

### Claude Code hooks が動作しない
```bash
# Claude Code の確認
which claude

# hooks の状況確認
aict setup --status

# Claude Code hooks のみ再設定
bash scripts/setup-hooks.sh --claude-only
```

### 設定ファイルの場所
- **AICT 設定**: `~/.aict/config.json`
- **Claude Code hooks**: `~/.claude/hooks-aict.json`
- **プロジェクトデータ**: `.git/ai-tracker/`

## 環境変数

以下の環境変数で動作をカスタマイズできます：

```bash
# デバッグモードを有効化
export AICT_DEBUG=1

# 統計表示を無効化
export AICT_SHOW_STATS=0

# カスタム設定ディレクトリ
export AICT_CONFIG_DIR=~/.config/aict
```

## アンインストール

```bash
# hooks を削除
aict setup --remove

# バイナリを削除
sudo rm /usr/local/bin/aict  # または該当するパス

# 設定ファイルを削除
rm -rf ~/.aict
rm -f ~/.claude/hooks-aict.json

# プロジェクトデータを削除 (任意)
rm -rf .git/ai-tracker
```

## サポート

問題が発生した場合：

1. **設定確認**: `aict setup --status`
2. **ログ確認**: `aict --help` でコマンド確認
3. **再インストール**: `bash scripts/install.sh --force`
4. **手動設定**: `aict setup` コマンドを使用

詳細な使用方法については、プロジェクトのメインドキュメントを参照してください。