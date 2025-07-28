# AI Code Tracker (AICT)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)
![Security](https://img.shields.io/badge/security-AES256-green.svg)
![i18n](https://img.shields.io/badge/i18n-ja%2Fen-brightgreen.svg)
![Version](https://img.shields.io/badge/version-v0.2.0-blue.svg)
![DuckDB](https://img.shields.io/badge/database-DuckDB-orange.svg)
![Performance](https://img.shields.io/badge/performance-500ns%2Fop-brightgreen.svg)

**AIが生成したコードと人間が書いたコードを自動的に区別・追跡するシステム**

Claude Codeとの完全統合により、透明性のある開発プロセスを実現します。

# このシステムは開発途中のものであり動作は保証できません。このリポジトリの実装や記載内容は動作検証中のものになります。

## ✨ 主要機能

- 🤖 **完全自動化**: Claude Code hooks による非侵襲的な自動追跡
- 🚀 **高速分析**: DuckDB搭載の高性能SQLベース分析基盤（500ns/op）
- 📊 **期間別分析**: 自然言語期間表現対応（"Q1 2024", "this year"等）
- 🔍 **詳細追跡**: AIと人間のコード貢献度を行レベルで記録
- 🌐 **Webダッシュボード**: リアルタイム統計表示とAPI
- 🛡️ **セキュリティ**: AES-256暗号化と監査ログ
- 🌏 **多言語対応**: 日本語・英語完全対応

## 🚀 クイックスタート

### インストール
```bash
# Go 1.19以上が必要
go install github.com/ai-code-tracker/aict/cmd/aict@latest
```

### 初期設定
```bash
cd your-project
aict init       # プロジェクト初期化
aict setup      # Claude Code hooks自動設定
```

### 基本的な使用法
```bash
# AI/人間のコード貢献度を表示
aict blame src/main.go

# 統計表示
aict stats

# 期間別分析
aict period "Q1 2024"
aict period "this year" 
aict period "last 3 months"

# Webダッシュボード起動
aict web
```

## 📊 出力例

### Blame表示（`aict blame`）
```
📁 src/main.go
   1  👤 John Doe       2024-01-15  package main
   2  
   3  🤖 Claude Sonnet  2024-01-15  import (
   4  🤖 Claude Sonnet  2024-01-15      "fmt"
   5  🤖 Claude Sonnet  2024-01-15      "log"
   6  🤖 Claude Sonnet  2024-01-15  )
   7  
   8  👤 John Doe       2024-01-15  func main() {
   9  🤖 Claude Sonnet  2024-01-15      if err := run(); err != nil {
  10  🤖 Claude Sonnet  2024-01-15          log.Fatal(err)
  11  🤖 Claude Sonnet  2024-01-15      }
  12  👤 John Doe       2024-01-16      fmt.Println("Done!") // 手動追加
  13  👤 John Doe       2024-01-15  }

📈 統計: 全13行中 AI: 6行 (46%) / 人間: 7行 (54%)
```

### 統計表示（`aict stats`）
```
📊 AI Code Tracker - プロジェクト統計

🎯 全体サマリー
┌─────────────────┬──────────┬─────────┐
│     項目        │   行数   │   割合  │
├─────────────────┼──────────┼─────────┤
│ 🤖 AI生成       │    892   │   59%   │
│ 👤 人間作成     │    608   │   41%   │
│ 📄 総行数       │  1,500   │  100%   │
├─────────────────┼──────────┼─────────┤
│ 📁 ファイル数   │     25   │         │
│ 👥 貢献者数     │      3   │         │
└─────────────────┴──────────┴─────────┘

🏆 主な貢献者
• 🤖 Claude Sonnet 4: 892行 (59%)
• 👤 開発者A: 380行 (25%)
• 👤 開発者B: 228行 (16%)

📈 最近の活動 (7日間)
• AI編集: 156行追加
• 人間編集: 84行追加
• 編集セッション: 12回
```

### 期間別分析（`aict period`）
```
📊 期間別分析: 2024年第1四半期 (Q1 2024)
════════════════════════════════════════════════

📈 全体サマリー
┌─────────────────┬──────────┬─────────┐
│     指標        │   数値   │   比率  │
├─────────────────┼──────────┼─────────┤
│ 🤖 AI行数       │  1,245   │   68%   │
│ 👤 人間行数     │    587   │   32%   │
│ 📄 総行数       │  1,832   │  100%   │
│ 📁 編集ファイル │     42   │         │
│ 🎯 活動日数     │     58   │         │
└─────────────────┴──────────┴─────────┘

🏆 言語別統計
• Go: 892行 (49%) - AI: 612行, 人間: 280行
• TypeScript: 541行 (30%) - AI: 398行, 人間: 143行
• Python: 284行 (15%) - AI: 156行, 人間: 128行
• Markdown: 115行 (6%) - AI: 79行, 人間: 36行

📅 日別トレンド（最新5日）
• 2024-03-28: AI +45行, 人間 +12行
• 2024-03-27: AI +23行, 人間 +31行
• 2024-03-26: AI +67行, 人間 +8行
• 2024-03-25: AI +34行, 人間 +19行
• 2024-03-24: AI +12行, 人間 +28行

🎯 生産性指標
• AIセッション効率: 28.5行/セッション
• 人間セッション効率: 15.3行/セッション
• 協働ファイル数: 18ファイル (43%)
```

### Webダッシュボード（`aict web`）
```
🌐 AI Code Tracker Web Dashboard starting on port 8080
📁 Data directory: /your-project/.git/ai-tracker
🗣️  Language: ja
🚀 Opening http://localhost:8080/dashboard in browser...
```

## 🌐 Webダッシュボード機能

### 起動方法
```bash
aict web               # デフォルト起動（ポート8080）
aict web -p 3000       # カスタムポート
aict web -l en         # 英語で起動
aict web --no-browser  # ブラウザを開かない
```

### 主要ページ
- `http://localhost:8080/dashboard` - メインダッシュボード
- `http://localhost:8080/contributors` - 貢献者分析
- `http://localhost:8080/files` - ファイル別統計
- `http://localhost:8080/timeline` - 開発タイムライン

### REST API
```bash
GET  /api/health              # ヘルスチェック
GET  /api/stats               # 統計データ（JSON）
GET  /api/contributors        # 貢献者リスト
GET  /api/timeline            # タイムライン
GET  /api/period/{start}/{end} # 期間別分析
GET  /api/files               # ファイル統計
GET  /api/blame/{file}        # ファイル別blame情報
POST /api/analysis/custom     # カスタム分析クエリ
```

## 🛡️ セキュリティ機能

### 基本設定
```bash
# データ暗号化を有効化
export AICT_ENCRYPT_DATA=true

# 監査ログを有効化  
export AICT_AUDIT_LOG=true

# セキュリティスキャン実行
aict security scan
```

## 📚 詳細ドキュメント

- **[要求定義書（RDD.md）](RDD.md)** - 完全な機能仕様
- **[使用方法ガイド](docs/USAGE.md)** - 詳細な使用方法とオプション
- **[セキュリティガイド](docs/SECURITY.md)** - セキュリティ機能詳細
- **[API リファレンス](docs/API.md)** - REST API完全仕様
- **[開発者ガイド](docs/DEVELOPMENT.md)** - 開発・カスタマイズ情報
- **[トラブルシューティング](docs/TROUBLESHOOTING.md)** - よくある問題と解決方法

## ⚙️ 設定

### 環境変数
```bash
export AICT_LANGUAGE=ja        # 言語設定（ja/en）
export AICT_DEBUG=true         # デバッグモード
export AICT_ENCRYPT_DATA=true  # データ暗号化
```

### 設定ファイル
- `~/.aict/config.json` - グローバル設定
- `.git/ai-tracker/config.json` - プロジェクト設定

## 🎯 対応環境

- Go 1.19以上
- Git 2.20以上  
- Claude Code（最新版）
- OS: Linux/macOS/Windows(WSL2)

## 📈 パフォーマンス

### DuckDB高速化実装
| 操作 | 実測値 | パフォーマンス | 備考 |
|------|-------|-------------|------|
| データ書き込み | ~500ns/op | 🚀 超高速 | SQLベース最適化 |
| 期間別分析 | <200ms | 🚀 高速 | 大規模データ対応 |
| 統計計算 | <100ms | 🚀 高速 | 複雑クエリ最適化 |
| メモリ効率 | 最適化済み | 🚀 効率的 | インデックス活用 |

### 基本機能パフォーマンス
| 操作 | 実測値 | 目標 | 状態 |
|------|-------|------|------|
| 追跡記録 | 45ms | 100ms | ✅ |
| Blame表示 | 280ms | 500ms | ✅ |
| 統計計算 | 750ms | 1000ms | ✅ |

## 🤝 サポート

- **Issues**: [GitHub Issues](https://github.com/ai-code-tracker/aict/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ai-code-tracker/aict/discussions)
- **Email**: contact@ai-code-tracker.dev

## 📄 ライセンス

MIT License - 詳細は[LICENSE](LICENSE)を参照

---

**AI Code Tracker v0.2.0** - DuckDB高性能ストレージ搭載で、AIと人間のコラボレーションを高速分析・可視化