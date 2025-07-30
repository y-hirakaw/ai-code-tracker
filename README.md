# AI Code Tracker (AICT)

Claude CodeとGitと連携してAI生成コードと人間が書いたコードの割合を追跡するGoベースのCLIツール。

## 🎯 特徴

- **自動追跡**: Claude Codeフックとの連携で編集を自動記録
- **正確な分析**: チェックポイント間の差分で正確な行数カウント
- **リアルタイム報告**: 目標達成率と詳細統計の表示
- **設定可能**: 追跡対象ファイル拡張子と除外パターン
- **軽量**: JSON形式での効率的なデータ保存

## 🚀 クイックスタート

### 1. セットアップ

```bash
# リポジトリをクローン
git clone https://github.com/y-hirakaw/ai-code-tracker.git
cd ai-code-tracker

# ビルド
go build -o bin/aict ./cmd/aict

# 初期化
./bin/aict init

# フック設定（Claude Code連携）
./bin/aict setup-hooks
```

### 2. 手動使用

```bash
# 人間のコード状態を記録
./bin/aict track -author human

# AIのコード状態を記録  
./bin/aict track -author claude

# 現在の統計を表示
./bin/aict report
```

### 3. 自動使用（Claude Code連携）

フック設定後、Claude Codeでファイルを編集すると自動的に追跡されます：

1. **PreToolUse**: Claude編集前に人間状態を記録
2. **PostToolUse**: Claude編集後にAI状態を記録
3. **Post-commit**: コミット時にメトリクス保存

## 📊 出力例

```
AI Code Tracking Report
======================
Total Lines: 817
AI Lines: 14 (1.7%)
Human Lines: 803 (98.3%)

Target: 80.0% AI code
Progress: 2.1%

Last Updated: 2025-07-30 16:04:08
```

## ⚙️ 設定

`.ai_code_tracking/config.json`で設定をカスタマイズ：

```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".swift"],
  "exclude_patterns": ["*_test.go", "*.test.js"],
  "author_mappings": {"y-hirakaw": "human"}
}
```

## 🔧 Claude Codeフック

`.claude/settings.json`でフックが設定されます：

```json
{
  "hooks": [
    {
      "event": "PreToolUse",
      "matcher": "Write|Edit|MultiEdit",
      "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/hooks/pre-tool-use.sh"}]
    },
    {
      "event": "PostToolUse", 
      "matcher": "Write|Edit|MultiEdit",
      "hooks": [{"type": "command", "command": "$CLAUDE_PROJECT_DIR/hooks/post-tool-use.sh"}]
    }
  ]
}
```

## 📁 ファイル構造

```
ai-code-tracker/
├── bin/aict                   # CLI実行ファイル
├── cmd/aict/main.go          # CLIエントリーポイント
├── internal/
│   ├── tracker/              # コア追跡ロジック
│   ├── storage/              # データ永続化
│   └── git/                  # Git統合
├── hooks/                    # フックスクリプト
│   ├── pre-tool-use.sh
│   ├── post-tool-use.sh
│   └── post-commit
├── .claude/
│   └── settings.json         # Claude Codeフック設定
└── .ai_code_tracking/        # 追跡データ
    ├── config.json
    ├── checkpoints/
    └── metrics/
```

## 🎯 使用ケース

### 開発目標管理
- AI生成コードの目標割合を設定（例：80%）
- プロジェクト進捗の可視化
- コード品質とAI活用のバランス

### チーム分析
- メンバー別のAI活用度
- プロジェクト間の比較
- 生産性指標の追跡

### 品質管理
- AIコードのレビュー対象特定
- 人間によるコード品質チェック
- バランスの取れた開発促進

## 🔄 ワークフロー

1. **初期化**: `aict init`でプロジェクト設定
2. **フック設定**: `aict setup-hooks`で自動追跡有効化
3. **開発**: Claude Codeで通常通り開発
4. **監視**: `aict report`で進捗確認
5. **調整**: 目標達成に向けた開発戦略調整

## 🛠️ 技術仕様

- **言語**: Go 1.21+
- **依存関係**: 標準ライブラリのみ
- **データ形式**: JSON
- **フック**: Claude Code hooks, Git post-commit
- **対応プラットフォーム**: macOS, Linux, Windows

## 📈 メトリクス

追跡される指標：
- 総行数
- AI生成行数・割合
- 人間作成行数・割合
- 目標達成率
- 最終更新時刻

## 🔒 セキュリティ

- ローカルファイルシステムのみ使用
- 外部通信なし
- 設定可能な追跡対象
- フック実行の透明性

## 🤝 貢献

Issue報告やPull Requestを歓迎します。

## 📄 ライセンス

MIT License

---

🤖 このプロジェクトはClaude Codeとの協力により開発されました。