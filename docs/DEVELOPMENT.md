# AI Code Tracker - 開発者ガイド

## 📋 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [ビルドとテスト](#ビルドとテスト)
- [アーキテクチャ](#アーキテクチャ)
- [カスタマイズ](#カスタマイズ)
- [コントリビューション](#コントリビューション)

## 開発環境のセットアップ

### 前提条件
- Go 1.19以上
- Git 2.20以上
- Make

### プロジェクトのクローン
```bash
git clone https://github.com/ai-code-tracker/aict
cd aict
```

### 依存関係のインストール
```bash
go mod download
```

### 開発用ビルド
```bash
make dev
```

## ビルドとテスト

### ビルドコマンド
```bash
# 開発版ビルド
make build

# リリース版ビルド
make release

# クロスコンパイル
make cross-compile
```

### テストの実行
```bash
# 単体テスト
make test

# 統合テスト
make test-integration

# E2Eテスト
make test-e2e

# カバレッジレポート
make coverage
```

### 品質チェック
```bash
# リンター実行
make lint

# フォーマッター実行
make fmt

# セキュリティスキャン
make security-scan
```

## アーキテクチャ

### ディレクトリ構造
```
ai-code-tracker/
├── cmd/                    # CLIアプリケーション
│   ├── aict/              # メインCLI
│   ├── aict-web/          # Webダッシュボード
│   └── aict-bench/        # ベンチマークツール
├── internal/              # 内部パッケージ
│   ├── tracker/           # コアトラッキング
│   ├── hooks/             # Git/Claude Code統合
│   ├── blame/             # 拡張blame機能
│   ├── stats/             # 統計処理
│   ├── storage/           # データ永続化
│   ├── security/          # セキュリティ機能
│   ├── web/               # Webダッシュボード機能
│   ├── i18n/              # 国際化システム
│   ├── ui/                # ヘルプ・UI機能
│   ├── cli/               # CLIコマンドハンドラー
│   ├── errors/            # エラーハンドリング
│   └── utils/             # 共通ユーティリティ
├── pkg/                   # 公開パッケージ
│   └── types/             # 共通型定義
└── docs/                  # ドキュメント
```

### 主要コンポーネント

#### トラッキングシステム
- **tracker**: コアトラッキングロジック
- **hooks**: Git/Claude Code統合
- **storage**: JSONL形式でのデータ保存

#### 分析システム
- **blame**: 拡張blame機能
- **stats**: 統計計算とレポート

#### セキュリティシステム
- **security**: 暗号化・監査・プライバシー保護

#### Webシステム
- **web**: 独立Webサーバー
- **handlers**: API・ダッシュボードハンドラー
- **middleware**: セキュリティ・ログ・CORS対応

## カスタマイズ

### 新機能の追加

#### 1. 新しいコマンドの追加
```go
// internal/cli/new_command.go
package cli

func HandleNewCommand(args []string) error {
    // コマンドロジックを実装
    return nil
}
```

#### 2. 新しい統計形式の追加
```go
// internal/stats/formatters.go
func FormatCustom(stats *Stats) (string, error) {
    // カスタムフォーマット実装
    return "", nil
}
```

#### 3. 新しいセキュリティ機能の追加
```go
// internal/security/custom_feature.go
func NewCustomSecurityFeature() SecurityFeature {
    // セキュリティ機能実装
    return nil
}
```

### 設定のカスタマイズ

#### 環境変数の追加
```go
// internal/utils/config.go
type Config struct {
    CustomSetting string `env:"AICT_CUSTOM_SETTING"`
}
```

#### 多言語メッセージの追加
```go
// internal/i18n/messages.go
var Messages = map[string]map[string]string{
    "ja": {
        "custom_message": "カスタムメッセージ",
    },
    "en": {
        "custom_message": "Custom message",
    },
}
```

## コントリビューション

### 開発フロー
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

### テストの書き方
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "expected1"},
        {"case2", "input2", "expected2"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := NewFeature(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### パフォーマンステスト
```go
func BenchmarkNewFeature(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewFeature("test input")
    }
}
```

詳細な開発情報については [RDD.md](../RDD.md) を参照してください。