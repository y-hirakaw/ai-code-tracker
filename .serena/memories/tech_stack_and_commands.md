# 技術スタックと開発コマンド

## 技術スタック
- **言語**: Go 1.21+
- **依存関係**: 標準ライブラリのみ
- **データ形式**: 超軽量JSONL（レコードあたり約100バイト）
- **フック**: Claude Code フック、Git post-commit
- **対応プラットフォーム**: macOS、Linux、Windows

## 開発コマンド

### ビルド
```bash
# プロジェクトをビルド
go build -o bin/aict ./cmd/aict

# バージョン確認
./bin/aict version
```

### テスト
```bash
# 全テストを実行
go test ./...

# テストカバレッジ付きで実行
go test -cover ./...

# 特定パッケージのテスト
go test ./internal/period/...
```

### コード品質
```bash
# コードフォーマット
go fmt ./...

# リンター実行
golangci-lint run

# 依存関係管理
go mod tidy
```

### インストール
```bash
# 直接インストール（推奨）
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest

# PATHに追加
export PATH=$PATH:$(go env GOPATH)/bin
```

### AICT コマンド
```bash
# 初期化
aict init

# フック設定
aict setup-hooks

# 手動トラッキング
aict track -author human
aict track -author claude

# レポート表示
aict report
aict report --last 7d
aict report --format graph

# リセット
aict reset

# 設定
aict config
```

## システム要件
- Go 1.21以降
- Git（リポジトリ分析用）
- macOS、Linux、またはWindows