# AI Code Tracker (AICT) v0.7.0

AIによるコード生成率を管理してレポート表示するツールです。学習用に開発中のプロジェクトです。

## インストール

```bash
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest
```

## 基本的な使い方

```bash
# 初期化
aict init

# チェックポイントを記録
aict checkpoint --author "Your Name"

# コミット
git commit -m "Your changes"
aict commit

# レポート表示
aict report --range HEAD~5..HEAD
```

## 主な機能

- コード変更のチェックポイント記録
- AI/人間によるコード生成の分類
- コミット範囲のレポート生成
- Git notesを使用した履歴管理
- リモートとの同期機能

詳細は [SPEC.md](SPEC.md) を参照してください。
