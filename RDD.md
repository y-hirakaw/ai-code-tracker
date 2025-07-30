# AI Code Tracker - Requirements and Design Document (RDD)

## 1. プロジェクト概要

### 1.1 目的
AI（Claude Code等）と人間が書いたコードの割合を正確に追跡し、設定可能な目標AIコード生成率の達成を支援するツールを開発する。

### 1.2 主要機能
- Claude Codeのフックと連携した自動的なコード変更追跡
- Git post-commitフックによる自動分析
- JSON形式でのデータ保存
- リアルタイムの進捗表示と目標達成率の可視化

### 1.3 技術スタック
- 実装言語: Go
- データ形式: JSON
- 連携: Claude Code hooks, Git hooks
- 対象ファイル: 設定可能（任意のプログラミング言語に対応）

## 2. アーキテクチャ設計

### 2.1 システム構成

```
ai-code-tracker/
├── cmd/
│   └── aict/              # メインCLIツール
│       └── main.go
├── internal/
│   ├── tracker/           # コア追跡ロジック
│   │   ├── checkpoint.go  # チェックポイント管理
│   │   ├── analyzer.go    # 分析ロジック
│   │   └── types.go       # 型定義
│   ├── storage/           # データ永続化
│   │   ├── json.go        # JSONシリアライゼーション
│   │   └── metrics.go     # メトリクス管理
│   └── git/               # Git連携
│       └── diff.go        # Git diff処理
├── hooks/
│   ├── aict-pre-tool-use.sh   # Claude Code Pre hook
│   ├── aict-post-tool-use.sh  # Claude Code Post hook
│   └── post-commit             # Git post-commit hook
├── .claude-code/
│   └── config.json        # Claude Code設定
└── .ai_code_tracking/     # AI追跡データディレクトリ
    ├── config.json        # 追跡設定
    ├── checkpoints/       # コードスナップショット
    └── metrics/           # 追跡メトリクス
```

### 2.2 データフロー（実装済み）

```
1. 初期状態記録
   ↓
   aict track -author human  → checkpoint_<id>_human.json
   ↓
2. AI編集後記録
   ↓
   aict track -author claude → checkpoint_<id>_claude.json
   ↓
3. 分析実行・メトリクス更新
   ↓
   metrics/current.json更新
```

## 3. データ仕様（実装済み）

### 3.1 設定ファイル形式

```json
// .ai_code_tracking/config.json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [
    ".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs", ".swift"
  ],
  "exclude_patterns": [
    "*_test.go", "*.test.js", "*.spec.ts", "*_generated.go"
  ],
  "author_mappings": {
    "y-hirakaw": "human"
  }
}
```

### 3.2 チェックポイントJSON形式（実装済み）

```json
{
  "id": "abc12345",
  "timestamp": "2025-07-30T15:52:30.252106+09:00",
  "author": "claude",
  "files": {
    "test/example.swift": {
      "path": "test/example.swift", 
      "lines": ["import Foundation", "// Simple vocabulary...", ...]
    }
  }
}
```

### 3.3 メトリクスJSON形式（実装済み）

```json
{
  "total_lines": 817,
  "ai_lines": 14,
  "human_lines": 803,
  "percentage": 1.7135862913096693,
  "last_updated": "2025-07-30T15:52:30.252106+09:00"
}
```

## 4. 実装仕様（現在の状況）

### 4.1 実装済みCLIコマンド

```bash
# 基本コマンド（実装済み）
aict init                      # プロジェクト初期化
aict track -author <name>      # チェックポイント作成
aict report                    # レポート表示

# 使用例
aict track -author human       # 人間のチェックポイント
aict track -author claude      # AIのチェックポイント
```

### 4.2 実装済み機能

#### ✅ 完了済み
- [x] プロジェクト基盤構築（go.mod、ディレクトリ構造）
- [x] コア機能実装（checkpoint.go, analyzer.go, types.go）
- [x] Git統合（diff.go）
- [x] ストレージ層（json.go, metrics.go）
- [x] CLI実装（init, track, reportコマンド）
- [x] 基本的な動作確認とテスト
- [x] メトリクスの累積ロジック修正
- [x] ディレクトリ名を.ai_code_trackingに変更

#### 📋 今後の拡張予定
- [ ] Claude Codeフック実装
- [ ] Git post-commitフック実装
- [ ] config設定コマンド
- [ ] より詳細なレポート機能
- [ ] 複数AIツール対応

### 4.3 主要な型定義（実装済み）

```go
// internal/tracker/types.go
type Checkpoint struct {
    ID        string                 `json:"id"`
    Timestamp time.Time              `json:"timestamp"`
    Author    string                 `json:"author"`
    Files     map[string]FileContent `json:"files"`
}

type FileContent struct {
    Path  string   `json:"path"`
    Lines []string `json:"lines"`
}

type AnalysisResult struct {
    TotalLines  int     `json:"total_lines"`
    AILines     int     `json:"ai_lines"`
    HumanLines  int     `json:"human_lines"`
    Percentage  float64 `json:"percentage"`
    LastUpdated time.Time `json:"last_updated"`
}

type Config struct {
    TargetAIPercentage float64           `json:"target_ai_percentage"`
    TrackedExtensions  []string          `json:"tracked_extensions"`
    ExcludePatterns    []string          `json:"exclude_patterns"`
    AuthorMappings     map[string]string `json:"author_mappings"`
}
```

## 5. UI/UX設計（実装済み）

### 5.1 現在の進捗表示

```
AI Code Tracking Report
======================
Total Lines: 817
AI Lines: 14 (1.7%)
Human Lines: 803 (98.3%)

Target: 80.0% AI code
Progress: 2.1%

Last Updated: 2025-07-30 15:52:30
```

## 6. パフォーマンス（実装済み）

- チェックポイント記録: 高速（JSON形式）
- 分析処理: リアルタイム
- メモリ使用量: 軽量
- ファイルサイズ: 効率的

## 7. セットアップ手順（実装済み）

```bash
# 1. ビルド
go build -o bin/aict ./cmd/aict

# 2. 初期化
./bin/aict init

# 3. 使用開始
./bin/aict track -author human    # 人間のベースライン
./bin/aict track -author claude   # AI編集後
./bin/aict report                 # レポート表示
```

## 8. 検証結果

プロジェクトは正常に動作し、以下が確認済み：

- **正確な追跡**: AI/人間のコード行数を正確に分離
- **リアルタイム更新**: チェックポイント間の差分を適切に計算
- **設定管理**: 拡張子フィルタリングと除外パターン
- **レポート生成**: 目標達成率の可視化

**テスト結果例**:
- 初期: 人間 801行
- AI追加: 人間 801行、AI 14行、合計 815行
- 人間追加: 人間 803行、AI 14行、合計 817行

目標値（80% AIコード）に対する進捗率: 2.1%

## 9. 今後の拡張計画

### 短期（フェーズ2）
- Claude Codeフック統合
- Git post-commitフック
- 設定管理コマンド拡張

### 中期（フェーズ3）
- 複数AIツール対応
- より詳細なレポート
- Web UI追加

### 長期（フェーズ4）
- チーム分析機能
- プロジェクト比較
- API提供