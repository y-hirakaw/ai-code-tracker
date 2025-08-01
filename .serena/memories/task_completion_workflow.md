# タスク完了時の実行手順

## 標準的なタスク完了フロー

### 1. コード品質チェック
```bash
# コードフォーマット
go fmt ./...

# リンター実行（エラーがないことを確認）
golangci-lint run

# 依存関係整理
go mod tidy
```

### 2. テスト実行
```bash
# 全テスト実行
go test ./...

# テストカバレッジ確認
go test -cover ./...

# 特定パッケージの詳細テスト（必要に応じて）
go test -v ./internal/period/
```

### 3. ビルド検証
```bash
# プロジェクトビルド
go build -o bin/aict ./cmd/aict

# バージョン確認
./bin/aict version

# 基本機能動作確認
./bin/aict help
```

### 4. 統合テスト（オプション）
```bash
# 統合テスト実行（存在する場合）
go test ./... -tags=integration

# 手動での動作確認
./bin/aict init
./bin/aict report
```

## 新機能追加時の追加手順

### 1. 機能テスト
```bash
# 新機能の単体テスト
go test ./internal/[新機能パッケージ]/

# 新機能のベンチマーク（パフォーマンスが重要な場合）
go test -bench=. ./internal/[新機能パッケージ]/
```

### 2. 互換性確認
```bash
# 既存設定ファイルとの互換性確認
aict config

# 既存のJSONLファイルとの互換性確認
aict report
```

## バージョンリリース時の特別手順

### 1. バージョン更新
```bash
# cmd/aict/main.go の version 定数を更新
# README.md のバージョン番号を更新
```

### 2. 最終検証
```bash
# クリーンビルド
rm -rf bin/
go build -o bin/aict ./cmd/aict

# インストールテスト
go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest
```

### 3. Git操作
```bash
# 変更をコミット
git add .
git commit -m "feat: [機能概要] and bump to v[バージョン]"

# タグ作成
git tag -a v[バージョン] -m "Release v[バージョン]"

# プッシュ
git push origin main
git push origin v[バージョン]
```

## トラブルシューティング

### よくある問題と解決方法
```bash
# テスト失敗時
go test -v ./... # 詳細ログで問題特定

# ビルド失敗時
go mod tidy      # 依存関係を再整理
go clean -cache  # キャッシュクリア

# リンターエラー時
golangci-lint run --fix  # 自動修正可能な問題を修正
```

### 品質基準
- **テストカバレッジ**: 新機能は80%以上を目標
- **リンターエラー**: ゼロエラーを維持
- **ビルドエラー**: ゼロエラーを維持
- **破壊的変更**: 既存機能の互換性を保持

## CLAUDE.md で指定された特別な注意事項

### バージョン管理
- セマンティックバージョニング（major.minor.patch）に従う
- 破壊的変更はメジャーバージョンアップ
- 新機能追加はマイナーバージョンアップ
- バグ修正はパッチバージョンアップ

### 言語規約
- **チャット返答**: 必ず日本語で返答
- **README.md**: 英語で記載維持
- **コメント**: 必要最小限（コードの自己文書化を優先）

### AICT特有の注意
- **トラッキング対象**: 設定された拡張子のファイルのみ（.go, .py, .js等）
- **マークダウンファイル**: MCPサーバーの`mcp-file-editor`を使用して編集
- **設定ファイル**: `.ai_code_tracking/config.json`の互換性維持