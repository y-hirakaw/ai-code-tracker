# AI Code Tracker - 使用方法ガイド

## 📋 目次

- [基本コマンド](#基本コマンド)
- [詳細な統計表示](#詳細な統計表示)
- [Blame機能](#blame機能)
- [設定管理](#設定管理)
- [言語切り替え](#言語切り替え)
- [Webダッシュボード](#webダッシュボード)
- [データ管理](#データ管理)

## 基本コマンド

### プロジェクト初期化
```bash
# プロジェクトを初期化
aict init

# ウィザードによる対話的設定
aict wizard
```

### Hook設定
```bash
# Git hooks と Claude Code hooks を自動設定
aict setup hooks

# 設定状況の確認
aict config --list
```

### 手動トラッキング
```bash
# 人間による編集を記録
aict track --author "開発者名"

# AI による編集を記録（通常は自動）
aict track --ai --model "claude-sonnet-4"

# テストモード
aict track --test
```

## 詳細な統計表示

### 基本統計
```bash
# サマリー表示
aict stats --summary

# 表形式で詳細表示
aict stats --format table

# JSON形式で出力
aict stats --format json
```

### 期間別統計
```bash
# 特定期間の統計
aict stats --since "2024-01-01" --until "2024-01-31"

# 最近7日間
aict stats --since "7 days ago"

# 日別統計
aict stats --format daily
```

### 分類別統計
```bash
# ファイル別統計（上位10件）
aict stats --by-file --top 10

# 作成者別統計
aict stats --by-author

# 貢献者形式
aict stats --format contributors
```

### 統計出力例
```
📊 AI Code Tracker - プロジェクト統計 (2024-01-01 〜 2024-01-31)

🎯 全体サマリー
┌─────────────────┬──────────┬─────────┬──────────────┐
│     項目        │   行数   │   割合  │   ファイル数 │
├─────────────────┼──────────┼─────────┼──────────────┤
│ 🤖 AI生成       │  1,248   │   62%   │      18      │
│ 👤 人間作成     │    752   │   38%   │      15      │
│ 📄 総行数       │  2,000   │  100%   │      25      │
└─────────────────┴──────────┴─────────┴──────────────┘

🏆 主な貢献者
┌──────────────────┬──────────┬─────────┬──────────────┐
│     作成者       │   行数   │   割合  │  最終更新    │
├──────────────────┼──────────┼─────────┼──────────────┤
│ 🤖 Claude Sonnet │  1,248   │   62%   │  2024-01-31  │
│ 👤 開発者A       │    480   │   24%   │  2024-01-30  │
│ 👤 開発者B       │    272   │   14%   │  2024-01-29  │
└──────────────────┴──────────┴─────────┴──────────────┘

📈 最近の活動 (7日間)
• AI編集: 24セッション、268行追加
• 人間編集: 16セッション、142行追加
• 平均セッション時間: 18分
```

## Blame機能

### 基本的なBlame表示
```bash
# ファイルのblame表示
aict blame src/main.go

# 行範囲を指定
aict blame src/main.go --lines 10-50

# 詳細表示
aict blame src/main.go --verbose
```

### Blame出力例
```
📁 src/handlers/api.go

   1  👤 開発者A        2024-01-15 09:30  package handlers
   2  
   3  👤 開発者A        2024-01-15 09:30  import (
   4  🤖 Claude Sonnet  2024-01-20 14:15      "encoding/json"
   5  👤 開発者A        2024-01-15 09:30      "net/http"
   6  🤖 Claude Sonnet  2024-01-20 14:15      "strconv"
   7  👤 開発者A        2024-01-15 09:30  )
   8  
   9  👤 開発者A        2024-01-15 09:30  type APIHandler struct {
  10  🤖 Claude Sonnet  2024-01-20 14:15      server *Server
  11  🤖 Claude Sonnet  2024-01-20 14:15      logger *Logger
  12  👤 开发者A        2024-01-15 09:30  }
  13  
  14  🤖 Claude Sonnet  2024-01-20 14:15  func (h *APIHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
  15  🤖 Claude Sonnet  2024-01-20 14:15      stats, err := h.server.GetStats()
  16  🤖 Claude Sonnet  2024-01-20 14:15      if err != nil {
  17  🤖 Claude Sonnet  2024-01-20 14:15          http.Error(w, err.Error(), http.StatusInternalServerError)
  18  🤖 Claude Sonnet  2024-01-20 14:15          return
  19  🤖 Claude Sonnet  2024-01-20 14:15      }
  20  🤖 Claude Sonnet  2024-01-20 14:15      json.NewEncoder(w).Encode(stats)
  21  🤖 Claude Sonnet  2024-01-20 14:15  }

📈 このファイルの統計:
• 全21行中 AI: 12行 (57%) / 人間: 9行 (43%)
• 🤖 Claude Sonnet: 12行
• 👤 開発者A: 9行
```

## 設定管理

### 設定の確認
```bash
# 全設定を表示
aict config --list

# 特定の設定を確認
aict config --get language
aict config --get security.encryption
```

### 設定の変更
```bash
# 言語設定
aict config --set language ja

# セキュリティ設定
aict config --set security.encryption true
aict config --set security.audit_log true
```

### 設定ファイルの場所
- **グローバル設定**: `~/.aict/config.json`
- **プロジェクト設定**: `.git/ai-tracker/config.json`
- **セキュリティ設定**: `.git/ai-tracker/security-config.json`

## 言語切り替え

### 動的言語切り替え
```bash
# 現在の言語設定を確認
aict lang

# 利用可能な言語一覧
aict lang --list

# 日本語に切り替え
aict lang ja

# 英語に切り替え
aict lang en

# 永続化して切り替え
aict lang ja --persistent
```

### 環境変数による設定
```bash
# セッション単位での言語設定
export AICT_LANGUAGE=ja

# 設定を確認
aict lang
```

## Webダッシュボード

### 起動オプション
```bash
# 基本起動
aict web

# カスタムポートで起動
aict web --port 3000
aict web -p 3000

# 言語指定で起動
aict web --lang en
aict web -l ja

# デバッグモードで起動
aict web --debug

# ブラウザを自動で開かない
aict web --no-browser

# データディレクトリを指定
aict web --data-dir /custom/path
```

### ダッシュボード機能
- **リアルタイム統計**: WebSocketによる即座な更新
- **多言語インターフェース**: 日本語・英語対応
- **レスポンシブデザイン**: モバイル・デスクトップ対応
- **インタラクティブチャート**: Chart.jsによるグラフ表示

### 主要ページ
| URL | 説明 |
|-----|------|
| `/dashboard` | メインダッシュボード |
| `/contributors` | 貢献者分析ページ |
| `/files` | ファイル別統計ページ |
| `/timeline` | 開発タイムライン |
| `/settings` | 設定ページ |

## データ管理

### データクリーンアップ
```bash
# 古いデータを削除（90日以前）
aict clean --older-than 90d

# 特定期間のデータを削除
aict clean --before "2024-01-01"

# 確認後に削除
aict clean --older-than 30d --confirm
```

### バックアップ
```bash
# データをバックアップ
aict backup --output backup.tar.gz

# 特定期間のデータをバックアップ
aict backup --since "2024-01-01" --output jan-backup.tar.gz
```

### データ復元
```bash
# バックアップから復元
aict restore --input backup.tar.gz

# 既存データをマージして復元
aict restore --input backup.tar.gz --merge
```

## 高度な使用法

### フィルタリング
```bash
# 特定のモデルのみの統計
aict stats --model "claude-sonnet-4"

# 特定の作成者のみの統計  
aict stats --author "開発者A"

# 特定のファイル拡張子のみ
aict stats --extension ".go" ".js"
```

### エクスポート
```bash
# CSV形式でエクスポート
aict export --format csv --output stats.csv

# JSON形式でエクスポート
aict export --format json --output stats.json
```

### デバッグ
```bash
# デバッグモードで実行
AICT_DEBUG=true aict stats

# ログレベルを設定
AICT_LOG_LEVEL=debug aict blame src/main.go
```