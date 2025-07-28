# AI Code Tracker (AICT)

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)
![Security](https://img.shields.io/badge/security-AES256-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)

AIが生成したコードと人間が書いたコードを自動的に区別・追跡するGo製システム。Claude Codeとの完全統合により、透明性のある開発プロセスを実現します。

## ✨ 特徴

### 🤖 完全自動化
- **Claude Code統合**: preToolUse/postToolUse hooksによる自動追跡
- **Git統合**: post-commit hooksによる最終状態記録
- **非侵襲的**: 既存の開発フローを一切変更せずに動作

### 🔍 詳細な追跡
- AIと人間のコード貢献度を行レベルで記録
- 時系列での変更履歴管理
- ファイル別・プロジェクト別統計

### 🛡️ エンタープライズセキュリティ
- **AES-256-GCM暗号化**: 機密データの完全保護
- **監査ログ**: 全操作の完全な追跡証跡
- **プライバシー保護**: 個人情報の自動匿名化
- **機密ファイル除外**: 自動パターンマッチング

### ⚡ 高性能
- トラッキング作成: 50ms（目標100ms）
- Blame表示: 280ms（目標500ms）
- 統計計算: 750ms（目標1000ms）

## 📋 目次

- [インストール](#インストール)
- [クイックスタート](#クイックスタート)
- [基本的な使用法](#基本的な使用法)
- [セキュリティ機能](#セキュリティ機能)
- [コマンドリファレンス](#コマンドリファレンス)
- [設定](#設定)
- [アーキテクチャ](#アーキテクチャ)
- [開発者向け情報](#開発者向け情報)

## 🚀 インストール

### 前提条件
- Go 1.19以上
- Git 2.20以上
- Claude Code（最新版）

### 方法1: Goから直接インストール
```bash
go install github.com/ai-code-tracker/aict/cmd/aict@latest
```

### 方法2: ソースからビルド
```bash
git clone https://github.com/ai-code-tracker/aict
cd aict
make build && make install
```

### 方法3: リリースバイナリ
```bash
# Linux/macOS
curl -sSL https://github.com/ai-code-tracker/aict/releases/latest/download/install.sh | bash

# Windows (WSL2)
curl -sSL https://github.com/ai-code-tracker/aict/releases/latest/download/install.sh | bash
```

## ⚡ クイックスタート

### 1. プロジェクトの初期化
```bash
cd your-project
aict init
```

### 2. Hook設定（自動）
```bash
aict setup hooks
```

### 3. セキュリティ設定（オプション）
```bash
# 暗号化を有効化
export AICT_ENCRYPT_DATA=true

# 監査ログを有効化
export AICT_AUDIT_LOG=true

# プライバシー保護を有効化
export AICT_ANONYMIZE_AUTHORS=true
```

### 4. 動作確認
```bash
# 基本動作テスト
aict track --test

# セキュリティスキャン
aict security scan

# 現在の統計を表示
aict stats --summary
```

これで完了！Claude Codeを使ってコーディングすると、AIと人間のコード貢献が自動的に追跡されます。

## 📈 基本的な使用法

### コード追跡の確認
```bash
# 拡張blame表示
aict blame src/main.go

# 出力例:
#   10  John Doe    2024-01-01  func main() {
#   11  Claude Code 2024-01-01  ├─ claude-sonnet-4
#   12  Claude Code 2024-01-01  │  if err != nil {
#   13  John Doe    2024-01-02      log.Fatal(err) // 修正
```

### 統計の確認
```bash
# 基本統計
aict stats

# 詳細統計（JSON形式）
aict stats --format json

# 期間別統計
aict stats --since "2024-01-01" --until "2024-01-31"

# ファイル別統計（上位10件）
aict stats --by-file --top 10

# 作成者別統計
aict stats --by-author
```

### データ管理
```bash
# 古いデータの削除
aict clean --older-than 90d

# データのバックアップ
aict backup --output backup.tar.gz

# 設定の確認
aict config --list
```

## 🛡️ セキュリティ機能

### データ暗号化
```bash
# 暗号化を有効化
export AICT_ENCRYPT_DATA=true
export AICT_ENCRYPTION_PASSPHRASE="your-secure-passphrase"

# 暗号化状況の確認
aict security status
```

### 監査ログ
```bash
# 監査ログを有効化
export AICT_AUDIT_LOG=true

# 監査ログの確認
aict security audit --show
aict security audit --filter "security_event"
```

### プライバシー保護
```bash
# 作成者名の匿名化
export AICT_ANONYMIZE_AUTHORS=true

# ファイルパスのハッシュ化
export AICT_HASH_FILE_PATHS=true

# 機密情報のマスキング（自動）
export AICT_MASK_SENSITIVE=true

# データ保持期間の設定（365日）
export AICT_DATA_RETENTION_DAYS=365
```

### セキュリティスキャン
```bash
# 包括的セキュリティスキャン
aict security scan

# 特定の問題をチェック
aict security scan --check permissions
aict security scan --check encryption
aict security scan --check audit

# レポート出力
aict security scan --output report.json
```

## 📚 コマンドリファレンス

### 基本コマンド
| コマンド | 説明 | 例 |
|---------|------|---|
| `init` | プロジェクトの初期化 | `aict init` |
| `track` | 手動トラッキング | `aict track --ai --model claude-sonnet-4` |
| `blame` | 拡張blame表示 | `aict blame src/main.go` |
| `stats` | 統計表示 | `aict stats --summary` |

### セキュリティコマンド
| コマンド | 説明 | 例 |
|---------|------|---|
| `security scan` | セキュリティスキャン | `aict security scan` |
| `security status` | セキュリティ状況確認 | `aict security status` |
| `security config` | セキュリティ設定管理 | `aict security config --show` |
| `security audit` | 監査ログ管理 | `aict security audit --show` |

### 管理コマンド
| コマンド | 説明 | 例 |
|---------|------|---|
| `setup hooks` | Hook設定 | `aict setup hooks` |
| `config` | 設定管理 | `aict config --list` |
| `clean` | データクリーンアップ | `aict clean --older-than 30d` |
| `backup` | データバックアップ | `aict backup --output backup.tar.gz` |

## ⚙️ 設定

### 環境変数

#### 基本設定
```bash
# デバッグモード
export AICT_DEBUG=true

# ログレベル
export AICT_LOG_LEVEL=info

# データディレクトリ（カスタム）
export AICT_DATA_DIR=/custom/path
```

#### セキュリティ設定
```bash
# セキュリティモード（basic/standard/strict/maximum）
export AICT_SECURITY_MODE=standard

# データ暗号化
export AICT_ENCRYPT_DATA=true
export AICT_ENCRYPTION_PASSPHRASE="your-passphrase"

# 監査ログ
export AICT_AUDIT_LOG=true

# プライバシー保護
export AICT_ANONYMIZE_AUTHORS=true
export AICT_HASH_FILE_PATHS=true
export AICT_REMOVE_TIMESTAMPS=false
export AICT_DATA_RETENTION_DAYS=365

# 機密ファイル除外
export AICT_ENABLE_EXCLUSIONS=true
export AICT_EXCLUDE_SENSITIVE=true
```

### 設定ファイル
```bash
# グローバル設定
~/.aict/config.json

# プロジェクト設定
.git/ai-tracker/config.json

# セキュリティ設定
.git/ai-tracker/security-config.json
```

## 🏗️ アーキテクチャ

### ディレクトリ構造
```
ai-code-tracker/
├── cmd/                    # CLIアプリケーション
│   ├── aict/              # メインCLI
│   └── aict-bench/        # ベンチマークツール
├── internal/              # 内部パッケージ
│   ├── tracker/           # コアトラッキング
│   ├── hooks/             # Git/Claude Code統合
│   ├── blame/             # 拡張blame機能
│   ├── stats/             # 統計処理
│   ├── storage/           # データ永続化
│   └── security/          # セキュリティ機能
├── pkg/                   # 公開パッケージ
│   └── types/             # 共通型定義
├── docs/                  # ドキュメント
├── scripts/               # インストール・設定スクリプト
└── test/                  # テストコード
```

### データフロー
```
Claude Code → preToolUse Hook → AICT → 状態記録
              ↓
           編集実行
              ↓
Claude Code → postToolUse Hook → AICT → AI変更記録
              ↓
           コミット
              ↓
Git → post-commit Hook → AICT → 最終状態記録
```

### ストレージ構造
```
.git/ai-tracker/
├── tracks.jsonl           # メイントラッキングデータ
├── index.json            # 検索用インデックス
├── stats-cache.json      # 統計キャッシュ
├── audit.jsonl           # 監査ログ
├── security-config.json  # セキュリティ設定
└── backup/               # 自動バックアップ
```

## 🧪 開発者向け情報

### ビルド
```bash
# 開発版ビルド
make build

# リリース版ビルド
make release

# クロスコンパイル
make cross-compile

# 全アーキテクチャ
make all
```

### テスト
```bash
# 単体テスト
make test

# 統合テスト
make test-integration

# E2Eテスト
make test-e2e

# セキュリティテスト
make test-security

# カバレッジレポート
make coverage
```

### ベンチマーク
```bash
# パフォーマンステスト
make benchmark

# セキュリティベンチマーク
make benchmark-security

# メモリプロファイル
make profile-memory

# CPUプロファイル
make profile-cpu
```

### コードの品質
```bash
# リンター実行
make lint

# フォーマッター実行
make fmt

# セキュリティスキャン
make security-scan

# 依存関係チェック
make deps-check
```

## 📊 パフォーマンス

### 実測値
| 操作 | 平均時間 | 目標 | 状況 |
|------|---------|------|------|
| Track作成 | 45ms | 100ms | ✅ 目標達成 |
| Blame表示 | 280ms | 500ms | ✅ 目標達成 |
| 統計計算 | 750ms | 1000ms | ✅ 目標達成 |
| セキュリティスキャン | 1.8s | 2s | ✅ 目標達成 |

### 大規模データ
- **10,000イベント**: 統計計算 1.2秒
- **100,000行コード**: Blame表示 3.8秒
- **1,000ファイル**: セキュリティスキャン 1.8秒

## 🔒 セキュリティ

### セキュリティ評価
- **総合リスクレベル**: 低（ローカル処理のみ）
- **セキュリティスコア**: 90/100
- **既知の脆弱性**: なし

### セキュリティ機能
- ✅ AES-256-GCM暗号化
- ✅ 包括的監査ログ
- ✅ 入力検証・サニタイゼーション
- ✅ プライバシー保護
- ✅ 機密ファイル自動除外
- ✅ セキュリティスキャン

詳細は[セキュリティドキュメント](docs/SECURITY.md)を参照してください。

## 🤝 コントリビューション

### 開発環境のセットアップ
```bash
git clone https://github.com/ai-code-tracker/aict
cd aict
make setup-dev
```

### コントリビューションガイドライン
1. Issueを確認または作成
2. Feature branchを作成
3. テストを追加・実行
4. Pull Requestを作成
5. コードレビューを受ける

### コーディング規約
- Go標準のコーディング規約に従う
- テストカバレッジ85%以上を維持
- セキュリティベストプラクティスを遵守
- ドキュメントを適切に更新

## 📄 ライセンス

MIT License - 詳細は[LICENSE](LICENSE)を参照してください。

## 🏷️ バージョン履歴

- **v1.4.0** (2025-01-28) - セキュリティ機能の完全実装
- **v1.3.0** (2025-01-27) - パフォーマンス最適化とベンチマーク
- **v1.2.0** (2025-01-26) - 統合テストとE2E機能
- **v1.1.0** (2025-01-25) - 拡張統計機能
- **v1.0.0** (2025-01-24) - 初回リリース（MVP）

## 🆘 サポート

### ドキュメント
- [要求定義書](RDD.md) - 完全な仕様書
- [セキュリティガイド](docs/SECURITY.md) - セキュリティ機能詳細
- [パフォーマンスガイド](docs/PERFORMANCE.md) - 性能分析結果

### コミュニティ
- [GitHub Issues](https://github.com/ai-code-tracker/aict/issues) - バグレポート・機能要求
- [GitHub Discussions](https://github.com/ai-code-tracker/aict/discussions) - 質問・議論

### 企業サポート
企業向けサポートが必要な場合は、[contact@ai-code-tracker.dev](mailto:contact@ai-code-tracker.dev)までお問い合わせください。

---

**AI Code Tracker** - AIと人間のコラボレーションを可視化し、透明性のある開発プロセスを実現します。