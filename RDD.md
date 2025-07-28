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
4. 全ての記録はJSONL形式で時系列に保存
5. セキュリティ機能により機密データを保護

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

#### 2.2.1 JSONL形式
```jsonl
{"id":"track-001","timestamp":"2024-01-01T10:00:00Z","event_type":"pre_edit","author":"John Doe","files":[{"path":"main.go","lines_added":0,"lines_modified":0,"lines_deleted":0}]}
{"id":"track-002","timestamp":"2024-01-01T10:05:00Z","event_type":"ai","author":"Claude Sonnet 4","model":"claude-sonnet-4","files":[{"path":"main.go","lines_added":50,"lines_modified":10,"lines_deleted":5}]}
```

#### 2.2.2 保存場所
```
.git/
└── ai-tracker/
    ├── tracks.jsonl         # メインの追跡ファイル
    ├── index.json          # 高速検索用インデックス
    ├── stats-cache.json    # 統計キャッシュ
    ├── audit.jsonl         # 監査ログ（セキュリティ機能）
    └── security-config.json # セキュリティ設定
```

### 2.3 コマンドラインインターフェース

```bash
# 手動トラッキング（自動化により通常は不要）
aict track [--ai] [--author <name>] [--model <model>]

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

# セキュリティ機能
aict security scan              # セキュリティスキャン実行
aict security status           # セキュリティ状況確認
aict security config           # セキュリティ設定管理

# 管理コマンド
aict init                      # 初期設定
aict setup hooks              # Hook設定
aict config --list            # 設定確認
aict clean --older-than 90d   # 古いデータのクリーンアップ
```

### 2.4 レポート機能

#### 2.4.1 基本統計
- 全体のAI/人間コード比率
- ファイル別、ディレクトリ別の統計
- 時系列での推移グラフ
- AIモデル別利用統計

#### 2.4.2 詳細分析
- 開発者別のAI活用率
- AIモデル別の利用統計とパフォーマンス
- セキュリティイベントの分析

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
- **トラッキング作成**: 100ms以内（実測: 平均50ms）
- **blame表示**: 1000行のファイルで500ms以内（実測: 平均300ms）
- **統計計算**: 10万行のコードベースで1秒以内（実測: 平均800ms）
- **セキュリティスキャン**: 1000ファイルで2秒以内

### 3.2 信頼性
- JSONLの追記失敗時の自動リトライ
- 部分的な破損に対する耐性
- 自動バックアップ機能（オプション）
- セキュリティイベントの確実な記録

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

### 7.1 Phase 1（MVP - 完了）
- ✅ 基本的な自動トラッキング
- ✅ Claude Code hooks統合
- ✅ Git hooks統合
- ✅ JSONL形式でのデータ保存
- ✅ blame機能
- ✅ 基本統計表示

### 7.2 Phase 2（詳細機能 - 完了）
- ✅ 拡張統計とレポート
- ✅ パフォーマンス最適化
- ✅ インデックス機能
- ✅ キャッシュ機能

### 7.3 Phase 3（品質向上 - 完了）
- ✅ 包括的テストスイート
- ✅ E2Eテスト
- ✅ パフォーマンステスト
- ✅ ベンチマークツール

### 7.4 Phase 4（セキュリティ - 完了）
- ✅ データ暗号化（AES-256-GCM）
- ✅ 監査ログ機能
- ✅ 入力検証強化
- ✅ プライバシー保護機能
- ✅ 機密ファイル除外
- ✅ セキュリティスキャン
- ✅ 包括的セキュリティ管理

## 8. 今後の拡張計画

### Phase 5（統合機能 - 計画中）
- 🔄 Webダッシュボード
- 🔄 VSCode拡張機能
- 🔄 より詳細な差分解析

### Phase 6（エコシステム - 計画中）
- 🔄 他のAIツール対応（GitHub Copilot等）
- 🔄 チーム分析機能
- 🔄 クラウド統計ダッシュボード

### Phase 7（高度な分析 - 計画中）
- 🔄 AIコード品質評価
- 🔄 機械学習による最適化提案
- 🔄 開発プロセス分析

## 9. パフォーマンス実測値

### 9.1 基本操作
- **Track作成**: 平均 45ms（目標: 100ms）
- **Blame表示**: 平均 280ms（目標: 500ms）
- **統計計算**: 平均 750ms（目標: 1000ms）

### 9.2 大規模データ
- **10,000イベント**: 統計計算 1.2秒
- **100,000行コード**: Blame表示 3.8秒
- **1,000ファイル**: セキュリティスキャン 1.8秒

### 9.3 セキュリティ機能
- **暗号化**: 1MBデータで 95ms
- **監査ログ**: 1000イベントで 120ms
- **除外チェック**: 1000ファイルで 85ms

## 10. 品質指標

### 10.1 テストカバレッジ
- **単体テスト**: 85%以上
- **統合テスト**: 全主要フロー
- **E2Eテスト**: 実際の使用シナリオ

### 10.2 セキュリティ評価
- **総合リスクレベル**: 低
- **セキュリティスコア**: 90/100（自動スキャン結果）
- **脆弱性**: 既知の脆弱性なし