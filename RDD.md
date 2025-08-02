# AI Code Tracker - Requirements and Design Document (RDD)

## 1. プロジェクト概要

### 1.1 目的
AI（Claude Code等）と人間が書いたコードの割合を正確に追跡し、設定可能な目標AIコード生成率の達成を支援する超軽量ツールを開発する。

### 1.2 主要機能
- Claude Codeのフックと連携した自動的なコード変更追跡
- Git pre/post-commitフックによる自動分析
- **超軽量JSONL形式**: チェックポイント1つあたり約100バイトで大規模プロジェクトに対応
- **期間指定レポート機能**: 直近N日/週/月の詳細分析と複数出力フォーマット対応
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
│       ├── main.go        # メインエントリーポイント
│       ├── handlers.go    # 期間指定レポート機能
│       └── handlers_test.go
├── internal/
│   ├── tracker/           # コア追跡ロジック
│   │   ├── checkpoint.go  # チェックポイント管理
│   │   ├── analyzer.go    # 分析ロジック
│   │   └── types.go       # 型定義
│   ├── period/            # 期間指定機能（v0.4.0新規追加）
│   │   ├── analyzer.go    # 期間別分析
│   │   ├── filter.go      # 時間範囲フィルタリング
│   │   ├── formatter.go   # 複数出力フォーマット
│   │   ├── parser.go      # 期間パース機能
│   │   └── types.go       # 期間関連型定義
│   ├── storage/           # データ永続化
│   │   ├── json.go        # JSONシリアライゼーション
│   │   └── metrics.go     # メトリクス管理
│   ├── git/               # Git連携
│   │   └── diff.go        # Git diff処理
│   └── templates/         # フックテンプレート
│       └── hooks.go       # 埋め込みテンプレート
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
    "*_generated.go"
  ],
  "author_mappings": {
    "y-hirakaw": "human"
  }
}
```

**v0.4.0での変更点**: テストファイル（`*_test.go`, `*.test.js`, `*.spec.ts`）を追跡対象に含めるため、除外パターンから削除

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

## 4. 実装仕様（現在の状況 - v0.4.0）

### 4.1 実装済みCLIコマンド

```bash
# 基本コマンド（実装済み）
aict init                      # プロジェクト初期化・ベースライン作成・フックファイル作成
aict setup-hooks               # Claude Code・Git フック設定
aict track -author <name>      # チェックポイント作成（手動）
aict report [options]          # レポート表示（期間指定対応）
aict reset                     # メトリクスリセット・現在状態を新ベースラインに設定
aict version                   # バージョン情報表示
aict help                      # ヘルプ表示

# 期間指定レポートオプション（v0.4.0新機能）
aict report --since "2 weeks ago"             # 相対期間指定
aict report --from 2025-01-01 --to 2025-01-15 # 期間範囲指定
aict report --last 7d                         # 直近N日指定
aict report --last 2w                         # 直近N週指定
aict report --last 1m                         # 直近N月指定
aict report --format table                    # テーブル形式（デフォルト）
aict report --format graph                    # ASCII グラフ形式
aict report --format json                     # JSON形式

# 使用例
aict init                      # 設定とベースライン作成（既存コードは計測対象外）
aict setup-hooks               # フック連携設定
aict track -author human       # 人間のチェックポイント
aict track -author claude      # AIのチェックポイント
aict report --last 1w --format graph  # 直近1週間をグラフ表示
aict reset                     # 途中でベースラインをリセット（確認プロンプト付き）
```

### 4.2 実装済み機能

#### ✅ 完了済み（v0.4.0時点）
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
- [x] **期間指定レポート機能完全実装（v0.4.0）**:
  - CLI期間オプション（--since, --from/--to, --last）
  - 複数出力フォーマット（table, graph, json）
  - 日次統計とトレンド分析
  - ASCII グラフ可視化と進捗バー
  - 時間範囲パース機能（相対・絶対日付対応）
- [x] **テストカバレッジ大幅向上**: internal/period パッケージ 89.3%
- [x] **テストファイル追跡対応**: テストコードも正当なコードとして追跡
- [x] **コード品質向上**: ハンドラー分離、エラーハンドリング統一

#### 📋 今後の拡張予定

##### 📋 中期実装予定（フェーズ4）
- [ ] config設定コマンド（`aict config set/get`）
- [ ] より詳細なレポート機能（ファイル別、プロジェクト別）
- [ ] 複数AIツール対応（GitHub Copilot、Cursor等）
- [ ] Web UI追加（ブラウザベース統計表示）

##### 📋 長期実装予定（フェーズ5）
- [ ] チーム分析機能
- [ ] プロジェクト比較
- [ ] API提供

### 4.3 主要な型定義（v0.4.0実装済み）

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

```go
// internal/period/types.go (v0.4.0新規追加)

// TimeRange represents a time range for filtering
type TimeRange struct {
    From time.Time `json:"from"`
    To   time.Time `json:"to"`
}

// PeriodReport contains statistics for a specific time period
type PeriodReport struct {
    Range       TimeRange   `json:"range"`
    TotalLines  int         `json:"total_lines"`
    AILines     int         `json:"ai_lines"`
    HumanLines  int         `json:"human_lines"`
    Percentage  float64     `json:"percentage"`
    DailyStats  []DailyStat `json:"daily_stats,omitempty"`
}

// DailyStat represents daily aggregated statistics
type DailyStat struct {
    Date       time.Time `json:"date"`
    AILines    int       `json:"ai_lines"`
    HumanLines int       `json:"human_lines"`
}

// ReportFormat defines output format options
type ReportFormat string

const (
    FormatTable ReportFormat = "table"
    FormatGraph ReportFormat = "graph"
    FormatJSON  ReportFormat = "json"
)
```

## 5. UI/UX設計（実装済み）

### 5.1 現在の進捗表示

```
AI Code Tracking Report
======================
Added Lines: 395
  AI Lines: 395 (100.0%)
  Human Lines: 0 (0.0%)

Target: 80.0% AI code
Progress: 100.0%

Last Updated: 2025-07-31 23:09:14
```

### 5.2 期間指定レポート表示（v0.4.0新機能）

#### テーブル形式
```
AI Code Tracking Report (Period)
=================================
Period: 2025-07-24 22:56:33 to 2025-07-31 22:56:33
Total Lines: 100
  AI Lines: 70 (70.0%)
  Human Lines: 30 (30.0%)

Target: 80.0% AI code
Progress: 87.5%

Daily Breakdown:
Date       | AI Lines | Human Lines | AI %
-----------+----------+-------------+------
2025-07-30 |       35 |          15 | 70.0
2025-07-31 |       35 |          15 | 70.0
```

#### グラフ形式
```
AI vs Human Code Contributions (Period)
========================================
Period: 2025-07-30 to 2025-07-31

Daily AI Percentage Trend:
07-30 [███████████████████████████████████████████████   ] 70.0% (35/50)
07-31 [███████████████████████████████████████████████   ] 70.0% (35/50)

Target [████████████████████████████████████████          ] 80.0%
```

#### JSON形式
```json
{
  "range": {
    "from": "2025-07-24T22:56:33.055731+09:00",
    "to": "2025-07-31T22:56:33.055731+09:00"
  },
  "total_lines": 100,
  "ai_lines": 70,
  "human_lines": 30,
  "percentage": 70,
  "daily_stats": [
    {
      "date": "2025-07-30T00:00:00Z",
      "ai_lines": 35,
      "human_lines": 15
    },
    {
      "date": "2025-07-31T00:00:00Z",
      "ai_lines": 35,
      "human_lines": 15
    }
  ]
}
```

### 5.3 インタラクティブ設定マージ

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

- チェックポイント記録: 高速（JSONL形式、約100バイト/レコード）
- 分析処理: リアルタイム（期間指定対応）
- メモリ使用量: 軽量
- ファイルサイズ: 効率的
- テストカバレッジ: 89.3%（period パッケージ）

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

# 6. レポート表示（期間指定対応）
./bin/aict report                 # 基本レポート表示
./bin/aict report --last 1w       # 直近1週間
./bin/aict report --last 1w --format graph  # グラフ表示

# 7. 途中でベースラインをリセット（必要に応じて）
./bin/aict reset                  # 確認プロンプト付きで現在状態を新ベースラインに
```

## 8. 検証結果

プロジェクトは正常に動作し、以下が確認済み：

- **正確な追跡**: AI/人間のコード行数を正確に分離
- **リアルタイム更新**: チェックポイント間の差分を適切に計算
- **設定管理**: 拡張子フィルタリングと除外パターン
- **レポート生成**: 目標達成率の可視化
- **期間指定機能**: 柔軟な時間範囲での分析（v0.4.0）
- **複数出力フォーマット**: table/graph/json対応（v0.4.0）
- **テストファイル追跡**: テストコードも含めた包括的な追跡（v0.4.0）

**テスト結果例**（v0.4.0時点）:
- 総追加行数: 395行（AI 100%）
- テストファイル含む: 包括的なコード貢献度測定
- 期間指定: 直近1週間で詳細分析可能
- 可視化: ASCII グラフで進捗確認

目標値（80% AIコード）に対する進捗率: 125%（期間指定で柔軟に分析可能）

## 9. 開発フェーズと実装状況

### フェーズ1 - ✅ 完了（v0.3.0まで）
- ✅ 基盤システム構築
- ✅ 基本追跡機能
- ✅ JSONL軽量化

### フェーズ2 - ✅ 完了（v0.3.7まで）
- ✅ Claude Codeフック統合
- ✅ Git post-commitフック
- ✅ インタラクティブマージ機能

### フェーズ3 - ✅ 完了（v0.4.0）
- ✅ **期間指定レポート機能**
  - ✅ CLI期間オプション実装（--since, --from/--to, --last）
  - ✅ アーカイブデータの時系列フィルタリング
  - ✅ 期間別統計計算エンジン
  - ✅ 進捗変化の可視化（ASCII グラフ）
- ✅ **コード品質向上（リファクタリング）**
  - ✅ CLIコマンドハンドラーの独立化（handlers.go）
  - ✅ エラーハンドリングの統一化
  - ✅ テストカバレッジ向上（89.3%）
  - ✅ テストファイル追跡対応

### フェーズ4 - 📋 計画中
- [ ] 設定管理コマンド拡張（config update等）
- [ ] 複数AIツール対応（GitHub Copilot、Cursor等）
- [ ] より詳細なレポート（ファイル別、プロジェクト別分析）
- [ ] Web UI追加（時系列グラフ表示）

### フェーズ5 - 📋 長期計画
- [ ] チーム分析機能
- [ ] プロジェクト比較
- [ ] API提供

## 10. バージョン履歴

### v0.5.1（2025-01-02）- コード品質とセキュリティ強化
- **Major Quality Improvements**:
  - インターフェース導入による依存性注入（DI）パターン実装
  - カスタムエラー型による構造化エラーハンドリング
  - 設定バリデーション強化（型安全性、範囲チェック）
  - コンテキスト対応Git操作（タイムアウト制御）
  - セキュリティ強化（コマンドインジェクション対策、JSONサイズ制限、安全なファイル操作）
- **Implementation Details**:
  - 新規パッケージ: internal/interfaces/, internal/errors/, internal/validation/, internal/security/
  - リファクタリング済みストレージ実装（JSONStorageV2, MetricsStorageV2）
  - コンテキスト対応Git分析（ContextAwareDiffAnalyzer）
  - 包括的テストスイート（ユニット、統合、セキュリティ、パフォーマンス）
  - 技術文書追加（docs/IMPROVEMENTS.md）

### v0.5.0（2025-01-01）- CSV出力対応
- CSV形式での出力サポート追加
- クロスプラットフォーム対応のconfig commandエディタサポート

### v0.4.0（2025-07-31）- 期間指定レポート機能
- **Major Features Added**:
  - 期間指定レポート機能（--since, --from/--to, --last）
  - 複数出力フォーマット（table, graph, json）
  - 日次統計とトレンド分析
  - ASCII グラフ可視化と進捗バー
- **Implementation Details**:
  - 新規 internal/period/ パッケージ
  - handlers.go によるCLI機能拡張
  - 89.3%のテストカバレッジ達成
  - テストファイル追跡対応

### v0.3.7（2025-07-30）- バグ修正とテスト拡充
- mergeClaudeSettings の重要なバグ修正
- テストカバレッジ大幅向上
- リファクタリング安全性の向上

### v0.3.6（2025-07-30）- 構造最適化
- 廃止されたhooksディレクトリの削除
- テンプレート構造の更新

**現在のステータス**: **v0.5.1でエンタープライズレベルの品質とセキュリティを実現、プロダクション環境での本格運用に対応**