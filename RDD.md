# AI Code Tracker (AICT) 要求定義書

## 1. システム概要

### 1.1 目的
Claude Codeを中心としたAIコーディングツールの利用において、AIが生成したコードと人間が書いたコードを自動的に区別・追跡するシステムを構築する。

### 1.2 基本コンセプト
- **自動追跡**: Claude Code hooksとGit hooksを活用した完全自動化
- **透明性**: 開発フローを妨げない非侵襲的な追跡
- **Git統合**: 既存のGitワークフローとシームレスに統合

### 1.3 動作原理
1. Claude Codeがファイルを編集する前に、現在の状態を人間の作業として記録
2. Claude Codeの編集後、変更をAIの作業として記録
3. 人間のコミット時に最終状態を記録
4. 全ての記録はJSONL形式で時系列に保存

## 2. 機能要件

### 2.1 自動トラッキング機能

#### 2.1.1 Claude Code連携
- **PreToolUse Hook**: ファイル編集前の状態を自動記録
- **PostToolUse Hook**: AI編集後の変更を自動記録
- **Stop Hook**: セッション終了時の統計表示

#### 2.1.2 Git連携
- **post-commit Hook**: コミット時の最終状態を記録
- **重複防止**: 5秒以内の重複記録を自動スキップ

### 2.2 データ記録仕様

#### 2.2.1 JSONL形式
```jsonl
{"id":"track-001","timestamp":"2024-01-01T10:00:00Z","event_type":"pre_edit","author":"John Doe","files":{"main.go":{"lines_before":100}}}
{"id":"track-002","timestamp":"2024-01-01T10:05:00Z","event_type":"ai_edit","author":"Claude Code","model":"claude-sonnet-4","files":{"main.go":{"lines_added":50,"lines_modified":10,"lines_deleted":5}}}
```

#### 2.2.2 保存場所
```
.git/
└── ai-tracker/
    ├── tracks.jsonl         # メインの追跡ファイル
    ├── index.json          # 高速検索用インデックス
    └── stats-cache.json    # 統計キャッシュ
```

### 2.3 コマンドラインインターフェース

```bash
# 手動トラッキング（自動化により通常は不要）
act track [--ai] [--author <name>] [--model <model>]

# 拡張blame表示
act blame <file>
  10  John Doe    2024-01-01  func main() {
  11  Claude Code 2024-01-01  ├─ claude-sonnet-4
  12  Claude Code 2024-01-01  │  if err != nil {
  13  John Doe    2024-01-02      log.Fatal(err) // 修正

# 統計表示
act stats [--format json|table|summary]
act stats --since "2024-01-01" --author "John Doe"
act stats --by-file --top 10

# 管理コマンド
act init                    # 初期設定
act config --list          # 設定確認
act clean --older-than 90d # 古いデータのクリーンアップ
```

### 2.4 レポート機能

#### 2.4.1 基本統計
- 全体のAI/人間コード比率
- ファイル別、ディレクトリ別の統計
- 時系列での推移グラフ

#### 2.4.2 詳細分析
- 開発者別のAI活用率
- AIモデル別の利用統計
- コード品質メトリクス（将来拡張）

## 3. 非機能要件

### 3.1 パフォーマンス
- トラッキング作成: 100ms以内
- blame表示: 1000行のファイルで500ms以内
- 統計計算: 10万行のコードベースで1秒以内

### 3.2 信頼性
- JSONLの追記失敗時の自動リトライ
- 部分的な破損に対する耐性
- 自動バックアップ機能（オプション）

### 3.3 互換性
- Git 2.20以上
- Go 1.19以上
- Claude Code 最新版
- Linux/macOS/Windows(WSL2)対応

## 4. 実装仕様

### 4.1 ディレクトリ構造
```
ai-code-tracker/
├── cmd/
│   └── aict/
│       └── main.go         # CLIエントリポイント
├── internal/
│   ├── tracker/           # コアトラッキングロジック
│   ├── hooks/             # Hook処理
│   ├── blame/             # Blame機能
│   ├── stats/             # 統計処理
│   └── storage/           # JSONL/Index管理
├── pkg/
│   └── types/             # 共通型定義
├── scripts/
│   ├── install.sh         # インストーラー
│   └── setup-hooks.sh     # Hook設定スクリプト
├── Makefile
└── README.md
```

### 4.2 主要データ構造
```go
type TrackEvent struct {
    ID          string              `json:"id"`
    Timestamp   time.Time          `json:"timestamp"`
    EventType   string             `json:"event_type"`
    Author      string             `json:"author"`
    Model       string             `json:"model,omitempty"`
    CommitRef   string             `json:"commit_ref,omitempty"`
    Files       map[string]FileInfo `json:"files"`
}

type FileInfo struct {
    Path          string   `json:"path"`
    LinesAdded    int      `json:"lines_added,omitempty"`
    LinesModified int      `json:"lines_modified,omitempty"`
    LinesDeleted  int      `json:"lines_deleted,omitempty"`
    LinesBefore   int      `json:"lines_before,omitempty"`
    ChangedLines  []int    `json:"changed_lines,omitempty"`
}
```

## 5. 導入手順

### 5.1 インストール
```bash
# 方法1: インストーラー使用
curl -sSL https://your-domain/install-act.sh | bash

# 方法2: 手動インストール
git clone https://github.com/your-org/ai-tracker
cd ai-tracker
make install
```

### 5.2 初期設定
```bash
# プロジェクトで初期化
cd your-project
act init

# Claude Code hooks設定（自動）
act setup claude-hooks

# Git hooks設定（自動）
act setup git-hooks
```

### 5.3 動作確認
```bash
# テストトラッキング
act track --test

# 設定確認
act config --verify
```

## 6. セキュリティ考慮事項

### 6.1 データ保護
- トラッキングデータは`.gitignore`に含めない（共有のため）
- 機密情報を含まないメタデータのみ記録
- 個人情報のマスキング機能（将来）

### 6.2 アクセス制御
- ローカルファイルシステムのみ使用
- ネットワーク通信なし（将来のダッシュボード機能を除く）

## 7. 今後の拡張計画

### Phase 1（MVP - 1ヶ月）
- 基本的な自動トラッキング
- blame機能
- 簡易統計表示

### Phase 2（3ヶ月）
- Webダッシュボード
- VSCode拡張機能
- より詳細な差分解析

### Phase 3（6ヶ月）
- 他のAIツール対応（GitHub Copilot等）
- チーム分析機能
- AIコード品質評価