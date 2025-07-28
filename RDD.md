# AI Code Tracker (AICT) 要求定義書

## 1. システム概要

### 1.1 目的
Claude Codeを中心としたAIコーディングツールの利用において、AIが生成したコードと人間が書いたコードを自動的に区別・追跡するシステムを構築する。

### 1.2 基本コンセプト
- **自動追跡**: Claude Code hooksとGit hooksを活用した完全自動化
- **透明性**: 開発フローを妨げない非侵襲的な追跡
- **Git統合**: 既存のGitワークフローとシームレスに統合
- **セキュリティ重視**: 企業環境で安全に利用できる包括的なセキュリティ機能

### 1.3 動作原理
1. Claude Codeがファイルを編集する前に、現在の状態を人間の作業として記録
2. Claude Codeの編集後、変更をAIの作業として記録
3. 人間のコミット時に最終状態を記録
4. 全ての記録はDuckDB形式で高速に保存・検索
5. 期間別分析により柔軟な開発振り返りを実現
6. セキュリティ機能により機密データを保護

## 2. 機能要件

### 2.1 自動トラッキング機能

#### 2.1.1 Claude Code連携
- **PreToolUse Hook**: ファイル編集前の状態を自動記録
- **PostToolUse Hook**: AI編集後の変更を自動記録
- **Stop Hook**: セッション終了時の統計表示
- **Notification Hook**: アイドル状態と権限エラーのハンドリング

#### 2.1.2 Git連携
- **post-commit Hook**: コミット時の最終状態を記録
- **重複防止**: 5秒以内の重複記録を自動スキップ
- **Claude Code検出**: コミットメッセージからのAI生成コード判別

### 2.2 データ記録仕様

#### 2.2.1 DuckDBスキーマ設計
```sql
-- メインテーブル: トラッキングイベント
CREATE TABLE tracks (
    id VARCHAR PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    event_type VARCHAR NOT NULL,
    author VARCHAR NOT NULL,
    author_type VARCHAR NOT NULL, -- 'ai' or 'human'
    model VARCHAR,
    commit_hash VARCHAR,
    session_id VARCHAR,
    message TEXT,
    date_partition DATE GENERATED ALWAYS AS (DATE(timestamp)),
    INDEX idx_timestamp (timestamp),
    INDEX idx_author_type (author_type),
    INDEX idx_date_partition (date_partition)
);

-- ファイル変更テーブル
CREATE TABLE file_changes (
    id VARCHAR PRIMARY KEY,
    track_id VARCHAR NOT NULL,
    file_path VARCHAR NOT NULL,
    lines_added INTEGER DEFAULT 0,
    lines_modified INTEGER DEFAULT 0,
    lines_deleted INTEGER DEFAULT 0,
    file_hash VARCHAR,
    FOREIGN KEY (track_id) REFERENCES tracks(id),
    INDEX idx_file_path (file_path),
    INDEX idx_track_id (track_id)
);
```

#### 2.2.2 保存場所
```
.git/
└── ai-tracker/
    ├── aict.duckdb          # メインデータベース
    ├── stats-cache.json     # 統計キャッシュ
    ├── audit.jsonl          # 監査ログ（セキュリティ機能）
    ├── security-config.json # セキュリティ設定
    └── backups/             # 自動バックアップ
        ├── daily/
        └── weekly/
```

### 2.3 コマンドラインインターフェース

```bash
# 手動トラッキング（自動化により通常は不要）
aict track [--ai] [--author <name>] [--model <model>]

# 期間別分析（新機能）
aict period "2024-01-01" "2024-03-31"     # 日付指定
aict period "Q1 2024"                      # 四半期指定
aict period "Jan-Mar 2024"                 # 月名指定
aict period "last 3 months"                # 相対指定

# 拡張blame表示
aict blame <file>
  10  John Doe    2024-01-01  func main() {
  11  Claude Code 2024-01-01  ├─ claude-sonnet-4
  12  Claude Code 2024-01-01  │  if err != nil {
  13  John Doe    2024-01-02      log.Fatal(err) // 修正

# 統計表示
aict stats [--format json|table|summary]
aict stats --since "2024-01-01" --author "John Doe"
aict stats --by-file --top 10

# レポートエクスポート（新機能）
aict period "Q1 2024" --export markdown --output "Q1_report.md"
aict period "2024-01-01" "2024-03-31" --export csv --output "analysis.csv"
aict period "last quarter" --export pdf --output "report.pdf"
aict period "this year" --export-all --output-dir "./reports/"

# セキュリティ機能
aict security scan              # セキュリティスキャン実行
aict security status           # セキュリティ状況確認
aict security config           # セキュリティ設定管理

# 管理コマンド
aict init                      # 初期設定
aict setup hooks              # Hook設定
aict config --list            # 設定確認
aict clean --older-than 90d   # 古いデータのクリーンアップ
aict web                       # Webダッシュボード起動
```

### 2.4 レポート機能

#### 2.4.1 期間別分析機能
- **柔軟な期間指定**: 日付範囲、四半期、月名、相対期間
- **ファイル別AI実装率**: 期間内の各ファイルでのAI/人間コード比率
- **言語別統計**: プログラミング言語ごとのAI活用パターン
- **時系列トレンド**: 期間内での日別・週別の実装推移
- **編集セッション分析**: AI・人間の編集パターンと効率性

#### 2.4.2 エクスポート機能
- **Markdown形式**: 美しいフォーマットでの読みやすいレポート
- **CSV形式**: データ分析・Excel連携用の構造化データ
- **JSON形式**: プログラマティックアクセス用の完全データ
- **PDF形式**: 企業向けプレゼンテーション用の高品質レポート
- **一括エクスポート**: 全形式での同時出力機能

#### 2.4.3 高度分析
- **AI協働パターン**: 効率的なAI活用パターンの特定
- **生産性指標**: 期間内での開発速度とAI寄与度
- **ファイル継続編集**: 同一ファイルの期間前後での変化追跡
- **スキル成長分析**: 人間の自立度とAI依存度の変化
- **効率性ベンチマーク**: 過去期間との比較分析

### 2.5 セキュリティ機能

#### 2.5.1 データ暗号化
- **アルゴリズム**: AES-256-GCM
- **キー導出**: PBKDF2 (10,000 iterations)
- **設定**: 環境変数 `AICT_ENCRYPT_DATA=true` で有効化

#### 2.5.2 監査ログ
- **形式**: JSONL形式の構造化ログ
- **記録内容**: ファイルアクセス、データ操作、セキュリティイベント
- **設定**: 環境変数 `AICT_AUDIT_LOG=true` で有効化

#### 2.5.3 入力検証
- **パストラバーサル対策**: `../` パターンの検出と除去
- **ファイルパス検証**: 有効なUTF-8、長さ制限、危険文字チェック
- **バッチ操作検証**: 重複ファイル検出、数量制限

#### 2.5.4 プライバシー保護
- **作成者匿名化**: SHA-256ハッシュによる匿名化
- **ファイルパスハッシュ化**: 機密パスの保護
- **機密情報マスキング**: パスワード、APIキー等の自動マスク
- **データ保持期間**: 設定可能な自動削除機能

#### 2.5.5 機密ファイル除外
- **デフォルトパターン**: 秘密鍵、環境変数、ログファイル等
- **カスタムルール**: プロジェクト固有の除外パターン
- **機密度レベル**: Critical/High/Medium/Low の4段階

## 3. 非機能要件

### 3.1 パフォーマンス

#### 3.1.1 DuckDB移行後の性能目標
- **トラッキング作成**: 50ms以内（JSONLから50%高速化）
- **blame表示**: 1000行のファイルで30ms以内（90%高速化）
- **統計計算**: 10万行のコードベースで100ms以内（90%高速化）
- **期間別分析**: 任意期間の分析を200ms以内で完了
- **複雑クエリ**: 多次元分析・集約処理を500ms以内
- **レポート生成**: 全形式エクスポートを2秒以内

#### 3.1.2 スケーラビリティ目標
- **データ規模**: 100万イベント以上でも高速処理
- **メモリ効率**: 現在比70%削減
- **並行処理**: 複数クエリの同時実行対応
- **キャッシュ最適化**: 頻繁なクエリの自動キャッシュ

### 3.2 信頼性
- **データベース整合性**: DuckDBによるACID特性保証
- **自動バックアップ**: 日次・週次の自動バックアップ
- **障害回復**: データ破損時の自動修復機能
- **トランザクション管理**: 複雑な操作の原子性保証
- **セキュリティイベントの確実な記録**: 監査ログの冗長化

### 3.3 互換性
- **Git**: 2.20以上
- **Go**: 1.19以上
- **Claude Code**: 最新版（Claude Sonnet 4 / Claude Opus 4 対応）
- **OS**: Linux/macOS/Windows(WSL2)対応

### 3.4 セキュリティ要件
- **リスクレベル**: 低リスク（ローカル処理のみ）
- **データ保護**: AES-256暗号化対応
- **アクセス制御**: ファイルシステム権限ベース
- **監査証跡**: 包括的な操作ログ
- **プライバシー**: GDPR準拠のデータ処理

## 4. 実装仕様

### 4.1 ディレクトリ構造
```
ai-code-tracker/
├── cmd/
│   ├── aict/                  # CLIエントリポイント
│   │   └── main.go
│   └── aict-bench/           # パフォーマンスベンチマーク
│       └── main.go
├── internal/
│   ├── tracker/              # コアトラッキングロジック
│   ├── hooks/                # Hook処理
│   ├── blame/                # Blame機能
│   ├── stats/                # 統計処理
│   ├── storage/              # JSONL/Index管理
│   └── security/             # セキュリティ機能
│       ├── manager.go        # セキュリティ統合管理
│       ├── encryption.go     # データ暗号化
│       ├── audit.go          # 監査ログ
│       ├── validation.go     # 入力検証
│       ├── privacy.go        # プライバシー管理
│       └── exclusion.go      # 機密ファイル除外
├── pkg/
│   └── types/                # 共通型定義
├── scripts/
│   ├── install.sh            # インストーラー
│   ├── setup-hooks.sh        # Hook設定スクリプト
│   └── security-scan.sh      # セキュリティスキャン
├── docs/
│   ├── SECURITY.md           # セキュリティドキュメント
│   └── PERFORMANCE.md        # パフォーマンステスト結果
├── test/
│   └── integration/          # E2Eテスト
├── setting_doc/
│   ├── ClaudeCodeHooks.md    # Claude Code設定
│   └── GitPostHook.md        # Git Hook設定
├── Makefile
└── README.md
```

### 4.2 主要データ構造
```go
type TrackEvent struct {
    ID          string              `json:"id"`
    Timestamp   time.Time          `json:"timestamp"`
    EventType   EventType          `json:"event_type"`
    Author      string             `json:"author"`
    Model       string             `json:"model,omitempty"`
    CommitHash  string             `json:"commit_hash,omitempty"`
    Files       []FileInfo         `json:"files"`
    Message     string             `json:"message,omitempty"`
    SessionID   string             `json:"session_id,omitempty"`
}

type FileInfo struct {
    Path          string   `json:"path"`
    LinesAdded    int      `json:"lines_added"`
    LinesModified int      `json:"lines_modified"`
    LinesDeleted  int      `json:"lines_deleted"`
    Hash          string   `json:"hash,omitempty"`
}

type SecurityConfig struct {
    Mode                  SecurityMode          `json:"mode"`
    EnableAuditLog        bool                  `json:"enable_audit_log"`
    EncryptSensitiveData  bool                  `json:"encrypt_sensitive_data"`
    ValidateFilePaths     bool                  `json:"validate_file_paths"`
    Privacy               PrivacyConfig         `json:"privacy"`
    Exclusions            ExclusionConfig       `json:"exclusions"`
}
```

## 5. 導入手順

### 5.1 インストール
```bash
# 方法1: Goから直接インストール
go install github.com/ai-code-tracker/aict/cmd/aict@latest

# 方法2: ソースからビルド
git clone https://github.com/ai-code-tracker/aict
cd aict
make build && make install
```

### 5.2 初期設定
```bash
# プロジェクトで初期化
cd your-project
aict init

# Hook設定
aict setup hooks

# セキュリティ設定（オプション）
export AICT_ENCRYPT_DATA=true
export AICT_AUDIT_LOG=true
export AICT_ANONYMIZE_AUTHORS=true
```

### 5.3 動作確認
```bash
# 基本動作テスト
aict track --test

# セキュリティ機能テスト
aict security scan

# 統計表示テスト
aict stats --summary
```

## 6. セキュリティ考慮事項

### 6.1 データ保護
- **暗号化**: AES-256-GCMによる機密データ保護
- **アクセス制御**: ファイルシステム権限（700/600）
- **機密情報除外**: 自動パターンマッチングによる除外
- **プライバシー**: 個人情報の匿名化とマスキング

### 6.2 監査とコンプライアンス
- **完全な操作ログ**: 全てのアクセスと変更を記録
- **GDPR準拠**: データ保持期間の管理と削除権の実装
- **セキュリティスキャン**: 定期的な脆弱性チェック

### 6.3 脅威モデル
- **対象脅威**: 機密コードの漏洩、不正アクセス、データ改ざん
- **対策**: 入力検証、暗号化、監査ログ、アクセス制御
- **残存リスク**: 低（ローカル環境での処理のみ）

## 7. 実装完了状況

### 7.1 Phase 1-11（MVP・基盤機能 - 完了）
- ✅ 基本的な自動トラッキング
- ✅ Claude Code hooks統合
- ✅ Git hooks統合
- ✅ JSONL形式でのデータ保存
- ✅ blame機能
- ✅ 基本統計表示
- ✅ 拡張統計とレポート
- ✅ パフォーマンス最適化
- ✅ インデックス機能
- ✅ キャッシュ機能
- ✅ 包括的テストスイート
- ✅ E2Eテスト
- ✅ パフォーマンステスト
- ✅ ベンチマークツール
- ✅ データ暗号化（AES-256-GCM）
- ✅ 監査ログ機能
- ✅ 入力検証強化
- ✅ プライバシー保護機能
- ✅ 機密ファイル除外
- ✅ セキュリティスキャン
- ✅ 包括的セキュリティ管理

### 7.2 Phase 12-19（品質向上・UX向上 - 完了）
- ✅ **Phase 12**: 実際の使用環境でのテストと検証
- ✅ **Phase 13**: ユーザビリティとDX（開発体験）の向上
- ✅ **Phase 14**: エラーメッセージの改善と多言語対応
- ✅ **Phase 15**: コードベースのリファクタリングと責務分離
- ✅ **Phase 16**: モジュール間の依存関係整理と最適化
- ✅ **Phase 17**: 共通ユーティリティの抽出と統合
- ✅ **Phase 18**: コンテキストアウェアなヘルプメッセージ
- ✅ **Phase 19**: 動的言語切り替え機能
- ✅ **Phase 20**: Webダッシュボード統合機能

### 7.3 完成した主要機能
- ✅ **多言語対応**: 日本語・英語の完全対応（i18n）
- ✅ **動的言語切り替え**: `aict lang` コマンドによるリアルタイム切り替え
- ✅ **ユーザーフレンドリーエラー**: コンテキスト対応エラーメッセージ
- ✅ **コンテキストヘルプ**: 状況に応じた適切なヘルプ提供
- ✅ **モジュラー設計**: 責務分離されたクリーンなアーキテクチャ
- ✅ **設定ウィザード**: インタラクティブセットアップ機能
- ✅ **CLI UX**: 絵文字・カラー対応の見やすいインターフェース
- ✅ **Webダッシュボード**: ブラウザベースのリアルタイム統計表示

## 8. Webダッシュボード機能詳細

### 8.1 Phase 20完成機能
- ✅ **独立Webサーバー**: `aict-web` バイナリによる完全独立動作
- ✅ **リアルタイム統計**: WebSocket経由でのライブ更新
- ✅ **多言語Webインターフェース**: 日本語・英語のWebUI
- ✅ **レスポンシブデザイン**: Bootstrap 5 + Chart.js統合
- ✅ **RESTful API**: 全統計データへのAPI アクセス
- ✅ **セキュリティ機能**: CORS、セキュリティヘッダー、入力検証

### 8.2 Webダッシュボード仕様

#### アーキテクチャ
```
Browser ←→ aict-web (HTTP/WebSocket) ←→ .git/ai-tracker/
                ↓
        Bootstrap UI + Chart.js
```

#### API エンドポイント
```
GET  /api/health                    # ヘルスチェック
GET  /api/stats                     # 統計データ（JSON）
GET  /api/contributors              # 貢献者リスト
GET  /api/timeline                  # タイムライン情報
GET  /api/files                     # ファイル統計
GET  /api/blame/{file}              # ファイル別blame情報
GET  /api/period/{start}/{end}      # 期間別分析（新機能）
GET  /api/export/{format}           # レポートエクスポート（新機能）
POST /api/analysis/custom           # カスタム分析クエリ
WS   /ws                            # WebSocketリアルタイム更新
```

#### ページ構成
```
/                    # インデックス（→ダッシュボード）
/dashboard           # メインダッシュボード
/contributors        # 貢献者ページ
/files              # ファイル統計ページ
/timeline           # タイムラインページ
/period             # 期間別分析ページ（新機能）
/reports            # レポート生成ページ（新機能）
/settings           # 設定ページ
```

### 8.3 使用方法
```bash
# CLI統合起動
aict web                          # デフォルト（ポート8080）
aict web -p 3000                  # カスタムポート
aict web -l en --debug            # 英語+デバッグモード
aict web --no-browser             # ブラウザを開かない

# 直接起動
go build ./cmd/aict-web
./aict-web -port 8080 -lang ja

# アクセス
http://localhost:8080/dashboard   # メインダッシュボード
```

### 8.4 技術スタック
- **Backend**: Go HTTP Server + Gorilla WebSocket + DuckDB
- **Frontend**: Bootstrap 5 + Chart.js + Vanilla JavaScript  
- **データ形式**: JSON API + DuckDB ストレージ
- **リアルタイム**: WebSocket による双方向通信
- **多言語**: i18nシステム統合（日本語・英語）

## 9. Phase 21以降の拡張計画

### Phase 21: DuckDB移行とパフォーマンス革命 ⚡
- ✅ **DuckDB統合**: 高速分析クエリ基盤の構築
- ✅ **期間別分析**: 柔軟な期間指定による開発振り返り
- ✅ **レポートエクスポート**: Markdown/CSV/JSON/PDF対応
- ✅ **クエリ最適化**: 90%以上の性能向上を実現

### Phase 22: 高度個人分析エンジン 📊
- 🔄 **AI協働パターン分析**: 効率的なAI活用パターンの特定
- 🔄 **スキル成長追跡**: 個人の技術習得度とAI依存度の分析
- 🔄 **予測分析**: 開発速度・品質の予測とアラート
- 🔄 **パーソナルコーチング**: AI分析による個人化された改善提案

### Phase 23: IDE統合とエコシステム 🔌
- 🔄 **VSCode Extension**: リアルタイム統計表示・インライン分析
- 🔄 **多AIツール統合**: GitHub Copilot・Tabnineなど他社AI対応
- 🔄 **プラグインシステム**: サードパーティ拡張機能アーキテクチャ
- 🔄 **SDK提供**: Python・JavaScript・Rust SDK

## 10. パフォーマンス実測値・目標

### 10.1 現在の実測値（JSONL実装）
- **Track作成**: 平均 45ms（目標: 100ms）✅
- **Blame表示**: 平均 280ms（目標: 500ms）✅
- **統計計算**: 平均 750ms（目標: 1000ms）✅
- **期間別分析**: 3-5秒（複雑クエリ）

### 10.2 DuckDB移行後の性能目標
- **Track作成**: 25ms以内（50%高速化）
- **Blame表示**: 30ms以内（90%高速化）
- **統計計算**: 100ms以内（85%高速化）
- **期間別分析**: 200ms以内（95%高速化）
- **レポート生成**: 2秒以内（全形式エクスポート）

### 10.3 大規模データ処理能力
- **100万イベント**: 統計計算 500ms以内
- **10万行ファイル**: Blame表示 100ms以内
- **1年間データ**: 期間別分析 300ms以内
- **複雑集約**: 多次元分析 1秒以内

### 10.4 メモリ・リソース効率
- **メモリ使用量**: 現在比70%削減目標
- **ディスク使用量**: DuckDB圧縮により60%削減
- **CPU効率**: ベクトル化実行による最適化
- **並行処理**: 複数クエリの同時実行対応

## 11. 品質指標

### 11.1 テストカバレッジ
- **単体テスト**: 85%以上
- **統合テスト**: 全主要フロー
- **E2Eテスト**: 実際の使用シナリオ

### 11.2 セキュリティ評価
- **総合リスクレベル**: 低
- **セキュリティスコア**: 90/100（自動スキャン結果）
- **脆弱性**: 既知の脆弱性なし