# AI Code Tracker - Requirements and Design Document (RDD)

## 1. プロジェクト概要

### 1.1 目的
AI（Claude Code等）と人間が書いたコードの割合を正確に追跡し、設定可能な目標AIコード生成率の達成を支援する超軽量ツールを開発する。

### 1.2 主要機能
- Claude Codeのフックと連携した自動的なコード変更追跡
- Git pre/post-commitフックによる自動分析
- **超軽量JSONL形式**: チェックポイント1つあたり約100バイトで大規模プロジェクトに対応
- **シンプルなアーキテクチャ**: ベースライン概念を廃止し、差分追跡のみに集中
- インタラクティブな既存設定マージ機能
- 高速な差分計算とリアルタイム進捗表示

### 1.3 技術スタック
- 実装言語: Go
- データ形式: JSONL（JSON Lines）
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
├── internal/
│   └── templates/          # フックテンプレート
│       └── hooks.go        # 埋め込みテンプレート
├── .claude/
│   └── settings.json      # Claude Code設定
└── .ai_code_tracking/     # AI追跡データディレクトリ
    ├── config.json        # 追跡設定
    ├── checkpoints.jsonl  # 超軽量チェックポイント記録（JSONL形式）
    ├── hooks/             # フックスクリプト（自動生成）
    │   ├── pre-tool-use.sh
    │   ├── post-tool-use.sh
    │   ├── pre-commit
    │   └── post-commit
    └── metrics/           # レガシーメトリクス（下位互換用）
```

### 2.2 超軽量追跡データフロー（v0.3.1実装済み）

```
1. PreToolUse Hook → JSONL記録: {"author":"human","added":X,"deleted":Y}
   ↓
   Claude Code編集実行
   ↓
2. PostToolUse Hook → JSONL記録: {"author":"claude","added":X+n,"deleted":Y+m}
   ↓
   人間の追加編集（optional）
   ↓
3. Pre-commit Hook → 現在状態記録
   ↓
4. Post-commit Hook → 分析・アーカイブ（レガシー互換用）
```

**JSONL形式の利点**：
- **超軽量**: 1レコード約100バイト（従来の70%削減）
- **高速処理**: シンプルな数値計算のみ
- **大規模対応**: 数万ファイルでも軽量
- **フィルタリング対応**: tracked_extensionsによる拡張子フィルタリング機能付き
- **スマートスキップ**: 対象外ファイルのみの変更時は記録をスキップして効率化

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
  "baseline_lines": 803,
  "ai_lines": 14,
  "human_lines": 0,
  "percentage": 100.0,
  "last_updated": "2025-07-30T15:52:30.252106+09:00"
}
```

**重要**: `percentage`はベースラインを除いた追加行（`ai_lines + human_lines`）に対するAI行の割合を示します。

## 4. 実装仕様（現在の状況）

### 4.1 実装済みCLIコマンド

```bash
# 基本コマンド（実装済み）
aict init                      # プロジェクト初期化・ベースライン作成・フックファイル作成
aict setup-hooks               # Claude Code・Git フック設定
aict track -author <name>      # チェックポイント作成（手動）
aict report                    # レポート表示（ベースラインからの変更を表示）
aict reset                     # メトリクスリセット・現在状態を新ベースラインに設定

# 使用例
aict init                      # 設定とベースライン作成（既存コードは計測対象外）
aict setup-hooks               # フック連携設定
aict track -author human       # 人間のチェックポイント
aict track -author claude      # AIのチェックポイント
aict reset                     # 途中でベースラインをリセット（確認プロンプト付き）
```

### 4.2 実装済み機能

#### ✅ 完了済み
- [x] プロジェクト基盤構築（go.mod、ディレクトリ構造）
- [x] コア機能実装（checkpoint.go, analyzer.go, types.go）
- [x] Git統合（diff.go）
- [x] ストレージ層（json.go, metrics.go）
- [x] CLI実装（init, track, report, resetコマンド）
- [x] **ベースライン機能実装**: 既存コードを基準点として設定、追加分のみ追跡
- [x] **リセット機能実装**: メトリクスクリア・新ベースライン設定（確認プロンプト付き）
- [x] 基本的な動作確認とテスト
- [x] メトリクスの累積ロジック修正
- [x] ディレクトリ名を.ai_code_trackingに変更
- [x] Claude Codeフック実装（PreToolUse/PostToolUse）
- [x] Git post-commitフック実装
- [x] フックテンプレート埋め込み機能
- [x] setup-hooksコマンド実装
- [x] インタラクティブ既存設定マージ機能
- [x] 既存GitフックとClaude設定の安全な統合
- [x] **超軽量JSONL形式実装**: チェックポイント記録を約100バイトに軽量化
- [x] **tracked_extensionsフィルタリング**: 設定された拡張子のみを追跡対象に
- [x] **スマートスキップ機能**: 対象外ファイルのみの変更時は記録をスキップして効率化

#### 📋 今後の拡張予定

##### 🔥 優先実装項目
- [ ] **期間指定レポート機能** - 直近N日/週/月のAI/人間割合分析
  - `aict report --since "2 weeks ago"`
  - `aict report --from 2025-01-01 --to 2025-01-15`
  - `aict report --last 14d`
  - 時系列での進捗変化グラフ表示
  - 期間別統計（日次、週次、月次集計）
  
##### 📋 中期実装予定
- [ ] config設定コマンド（`aict config set/get`）
- [ ] より詳細なレポート機能（ファイル別、プロジェクト別）
- [ ] 複数AIツール対応（GitHub Copilot、Cursor等）
- [ ] Web UI追加（ブラウザベース統計表示）

### 4.3 主要な型定義（v0.3.1実装済み）

```go
// internal/tracker/types.go

// 新しい超軽量JSONL形式（v0.3.1）
type CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Commit    string    `json:"commit,omitempty"`
    Added     int       `json:"added"`   // 全ファイルの追加行数合計
    Deleted   int       `json:"deleted"` // 全ファイルの削除行数合計
}

// レガシー形式（下位互換用）
type Checkpoint struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    Author      string                 `json:"author"`
    CommitHash  string                 `json:"commit_hash,omitempty"`
    Files       map[string]FileContent `json:"files"`
    NumstatData map[string][2]int      `json:"numstat_data,omitempty"`
}

// シンプル化されたメトリクス（ベースライン概念を削除）
type AnalysisResult struct {
    TotalLines  int       `json:"total_lines"`
    AILines     int       `json:"ai_lines"`
    HumanLines  int       `json:"human_lines"`
    Percentage  float64   `json:"percentage"`
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
Total Lines: 817 (including 803 baseline)
Added Lines: 14
  AI Lines: 14 (100.0%)
  Human Lines: 0 (0.0%)

Target: 80.0% AI code
Progress: 125.0%

Last Updated: 2025-07-30 15:52:30
```

**変更点**: ベースライン分を明示し、追加された行のみでAI/人間の割合を計算表示

### 5.2 インタラクティブ設定マージ

```
$ aict setup-hooks
Warning: Git post-commit hook already exists at .git/hooks/post-commit
Do you want to merge AI Code Tracker functionality? (y/N): y
✓ Git post-commit hook merged with existing hook

Warning: Claude settings already exist at .claude/settings.json  
Do you want to merge AI Code Tracker hooks? (y/N): y
✓ Claude Code hooks merged with existing settings
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

# 2. 初期化（設定、ベースライン作成、フックファイル作成）
./bin/aict init                   # 既存コードを基準点（ベースライン）として設定

# 3. フック設定（Claude CodeとGit連携）
./bin/aict setup-hooks

# 4. 自動追跡開始
# Claude Codeで編集すると自動的に追跡される（ベースラインからの差分のみ）

# 5. 手動追跡（必要に応じて）
./bin/aict track -author human    # 人間の追加編集
./bin/aict track -author claude   # AI編集後
./bin/aict report                 # レポート表示（ベースラインを除く）

# 6. 途中でベースラインをリセット（必要に応じて）
./bin/aict reset                  # 確認プロンプト付きで現在状態を新ベースラインに
```

## 8. 検証結果

プロジェクトは正常に動作し、以下が確認済み：

- **正確な追跡**: AI/人間のコード行数を正確に分離
- **リアルタイム更新**: チェックポイント間の差分を適切に計算
- **設定管理**: 拡張子フィルタリングと除外パターン
- **レポート生成**: 目標達成率の可視化

**テスト結果例**（ベースライン機能適用後）:
- ベースライン: 803行（計測対象外）
- AI追加: 14行追加（AI 100%）
- 人間追加: 0行追加（Human 0%）
- 合計: 817行（ベースライン803行 + 追加14行）

目標値（80% AIコード）に対する進捗率: 125%（追加分のみで計算）

## 9. 今後の拡張計画

### 短期（フェーズ2）- ✅ 完了
- ✅ Claude Codeフック統合
- ✅ Git post-commitフック
- ✅ インタラクティブマージ機能

### 短期（フェーズ3）- 🔥 優先実装
- [ ] **期間指定レポート機能**
  - CLI期間オプション実装（--since, --from/--to, --last）
  - アーカイブデータの時系列フィルタリング
  - 期間別統計計算エンジン
  - 進捗変化の可視化（ASCII グラフ）
- [ ] **コード品質向上（リファクタリング）**
  - main.go の関数分割（500行超のモノリス構造改善）
  - CLIコマンドハンドラーの独立化
  - エラーハンドリングの統一化
  - テストカバレッジ向上（期間機能と並行実装）
  
### 中期（フェーズ4）
- [ ] 設定管理コマンド拡張（config update等）
- [ ] 複数AIツール対応（GitHub Copilot、Cursor等）
- [ ] より詳細なレポート（ファイル別、プロジェクト別分析）
- [ ] Web UI追加（時系列グラフ表示）

### 長期（フェーズ5）
- [ ] チーム分析機能
- [ ] プロジェクト比較
- [ ] API提供

## 10. 期間指定機能の実装設計（フェーズ3優先項目）

### 10.1 要件定義

**背景**: 現在の実装では累積データのみ表示可能。実用的な分析のため、「直近2週間のAI/人間割合」等の期間指定レポートが必要。

**主要ユースケース**:
- 直近N日間での開発パターン分析
- 月次/週次での進捗レポート生成
- 特定期間でのチーム貢献度評価

### 10.2 新CLIコマンド仕様

```bash
# 期間指定オプション
aict report --since "2 weeks ago"     # 相対期間指定
aict report --since "2025-01-01"      # 絶対日付指定
aict report --from 2025-01-01 --to 2025-01-15  # 期間範囲指定
aict report --last 7d                 # 直近N日指定
aict report --last 2w                 # 直近N週指定
aict report --last 1m                 # 直近N月指定

# 出力形式オプション
aict report --since "1 week ago" --format table   # テーブル表示
aict report --since "1 week ago" --format graph   # ASCII グラフ
aict report --since "1 week ago" --format json    # JSON出力
```

### 10.3 実装計画

#### 10.3.1 データ構造拡張
```go
// 期間フィルタリング用の新しい型
type TimeRange struct {
    From time.Time `json:"from"`
    To   time.Time `json:"to"`
}

type PeriodReport struct {
    Range       TimeRange       `json:"range"`
    TotalLines  int            `json:"total_lines"`
    AILines     int            `json:"ai_lines"`
    HumanLines  int            `json:"human_lines"`
    Percentage  float64        `json:"percentage"`
    DailyStats  []DailyStat    `json:"daily_stats"`
}

type DailyStat struct {
    Date       time.Time `json:"date"`
    AILines    int       `json:"ai_lines"`
    HumanLines int       `json:"human_lines"`
}
```

#### 10.3.2 新機能モジュール
- `internal/period/parser.go` - 期間文字列パース機能
- `internal/period/filter.go` - チェックポイント期間フィルタリング
- `internal/period/analyzer.go` - 期間別統計計算
- `internal/reports/formatter.go` - 複数フォーマット出力対応

### 10.4 実装優先順位
1. **Phase 3.1**: 基本的な期間フィルタリング（--since, --from/--to）
2. **Phase 3.2**: 相対期間指定（--last Nd/Nw/Nm）
3. **Phase 3.3**: 複数出力フォーマット（table, graph, json）
4. **Phase 3.4**: 日次/週次統計とトレンド分析

### 10.5 スマートスキップ機能（実装済み）
- 対象外ファイルのみの変更時は記録をスキップ
- 前回レコードと同じ値の場合は重複記録を防止