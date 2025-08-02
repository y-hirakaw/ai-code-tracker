# AI Code Tracker - Requirements and Design Document (RDD)

## 1. プロジェクト概要

**AI Code Tracker (AICT)** は、AI（Claude Code等）と人間が書いたコードの割合を正確に追跡する超軽量ツールです。

### 主要機能
- **自動追跡**: Claude Code・Git hooks 統合による自動コード変更記録
- **超軽量データ**: JSONL形式で1レコード約100バイト（従来比70%削減）
- **期間指定レポート**: 複数出力フォーマット（table/graph/json/csv）対応
- **インテリジェント設定**: 既存フックとの安全なマージ機能
- **高速分析**: リアルタイム差分計算とスマートスキップ

### 技術スタック
- **言語**: Go 1.21+
- **データ**: JSONL（JSON Lines）
- **連携**: Claude Code hooks, Git hooks
- **対象**: 設定可能なファイル拡張子

## 2. システム構成

```
ai-code-tracker/
├── cmd/aict/                  # CLI実装
├── internal/
│   ├── tracker/               # コア追跡ロジック
│   ├── period/                # 期間指定機能（v0.4.0）
│   ├── interfaces/            # DIインターフェース（v0.5.1）
│   ├── errors/                # 構造化エラー（v0.5.1）
│   ├── validation/            # 設定バリデーション（v0.5.1）
│   ├── security/              # セキュリティ強化（v0.5.1）
│   ├── storage/               # データ永続化
│   ├── git/                   # Git連携
│   └── templates/             # フックテンプレート
└── .ai_code_tracking/         # 追跡データ
    ├── config.json            # 設定
    ├── checkpoints.jsonl      # 超軽量記録
    └── hooks/                 # 自動生成フック
```

## 3. データフロー

```
1. PreToolUse Hook → 人間状態記録
   ↓
2. Claude Code編集実行
   ↓  
3. PostToolUse Hook → AI状態記録
   ↓
4. Git commit → メトリクス更新
```

**JSONL記録例**:
```json
{"timestamp":"2025-07-31T23:09:14+09:00","author":"claude","added":395,"deleted":271}
```

## 4. 設定・データ仕様

### 設定ファイル (.ai_code_tracking/config.json)
```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [".go", ".py", ".js", ".ts", ".swift"],
  "exclude_patterns": ["*_generated.go"],
  "author_mappings": {"username": "human"}
}
```

### 主要型定義
```go
// 超軽量JSONL形式
type CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Added     int       `json:"added"`
    Deleted   int       `json:"deleted"`
}

// 期間指定レポート
type PeriodReport struct {
    Range       TimeRange   `json:"range"`
    TotalLines  int         `json:"total_lines"`
    AILines     int         `json:"ai_lines"`
    HumanLines  int         `json:"human_lines"`
    Percentage  float64     `json:"percentage"`
    DailyStats  []DailyStat `json:"daily_stats"`
}
```

## 5. CLI仕様

### 基本コマンド
```bash
aict init                      # プロジェクト初期化
aict setup-hooks               # フック設定
aict track -author <name>      # 手動チェックポイント
aict report [options]          # レポート表示
aict reset                     # メトリクスリセット
aict config                    # 設定表示
aict version                   # バージョン表示
```

### 期間指定オプション（v0.4.0）
```bash
aict report --last 7d                         # 直近7日
aict report --since "2 weeks ago"             # 相対期間
aict report --from 2025-01-01 --to 2025-01-15 # 期間範囲
aict report --format table|graph|json|csv     # 出力形式
```

## 6. レポート出力例

### テーブル形式
```
AI Code Tracking Report
======================
Added Lines: 395
  AI Lines: 395 (100.0%)
  Human Lines: 0 (0.0%)

Target: 80.0% AI code
Progress: 100.0%
```

### グラフ形式
```
Daily AI Percentage Trend:
07-31 [██████████████████████████████████████████████████] 100.0% (395/395)

Target [████████████████████████████████████████          ] 80.0%
```

### CSV形式（v0.5.0）
```csv
Date,AI_Lines,Human_Lines,Total_Lines,AI_Percentage,Target_Percentage,Progress
2025-07-31,395,0,395,100.0,80.0,125.0
```

## 7. 開発履歴

| バージョン | 実装内容 | 完成度 |
|-----------|----------|--------|
| v0.3.0 | 基盤・JSONL軽量化 | ✅ |
| v0.3.7 | Claude/Git hooks統合 | ✅ |
| v0.4.0 | 期間指定・複数出力形式 | ✅ |
| v0.5.0 | CSV出力・config command | ✅ |
| v0.5.1 | DI・セキュリティ強化 | ✅ |
| v0.5.2 | 安定版（現在） | ✅ |

## 8. 品質指標

- **テストカバレッジ**: 89.3%（period パッケージ）
- **セキュリティ**: コマンドインジェクション対策、JSONサイズ制限
- **パフォーマンス**: 100バイト/レコード、リアルタイム分析
- **エラーハンドリング**: 構造化エラー、コンテキスト対応
- **アーキテクチャ**: 依存性注入、インターフェース分離

## 9. セットアップ・使用方法

```bash
# インストール
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest

# プロジェクトで初期化
cd your-project
aict init
aict setup-hooks

# 自動追跡開始（Claude Codeで編集するだけ）
# 手動追跡
aict track -author human
aict track -author claude

# レポート表示
aict report --last 1w --format csv
```

## 10. プロジェクトステータス

**現在**: **v0.5.2 - プロダクション運用可能な安定版**

### 達成済み
- ✅ 全主要機能実装完了
- ✅ エンタープライズレベルの品質・セキュリティ
- ✅ 包括的テストカバレッジ
- ✅ 実用的な性能・軽量性

### 将来拡張（オプション）
- 📋 複数AIツール対応
- 📋 チーム分析機能
- 📋 Web UI
- 📋 API提供

**結論**: 実用レベルに到達、メンテナンスモードに移行可能