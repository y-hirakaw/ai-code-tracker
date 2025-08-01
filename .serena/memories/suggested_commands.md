# 推奨コマンドとシステムユーティリティ

## 開発で使用する主要コマンド

### プロジェクトビルドとテスト
```bash
# プロジェクトビルド
go build -o bin/aict ./cmd/aict

# 全テスト実行
go test ./...

# テストカバレッジ付き
go test -cover ./...

# コードフォーマット
go fmt ./...

# リンター実行
golangci-lint run

# 依存関係整理
go mod tidy
```

### AICT CLI コマンド
```bash
# プロジェクト初期化
aict init

# フック設定
aict setup-hooks

# 手動トラッキング
aict track -author human
aict track -author claude

# レポート表示
aict report                              # 基本レポート
aict report --last 7d                    # 過去7日間
aict report --since "2 weeks ago"        # 2週間前から
aict report --from 2025-01-01 --to 2025-01-15  # 期間指定
aict report --format graph               # グラフ表示
aict report --format json                # JSON出力

# 設定管理
aict config                              # 設定表示
aict reset                               # リセット（確認付き）
aict version                             # バージョン表示
```

### Git コマンド（Darwin システム用）
```bash
# 基本的なGitコマンド
git status                               # ステータス確認
git add .                                # 全変更をステージング
git commit -m "commit message"           # コミット
git push                                 # プッシュ
git pull                                 # プル
git log --oneline                        # ログ表示
git diff                                 # 差分表示
git branch                               # ブランチ一覧
```

### ファイル操作（macOS/Darwin用）
```bash
# ディレクトリ操作
ls -la                                   # ファイル一覧（詳細）
ls -la .ai_code_tracking/                # トラッキングディレクトリ確認
find . -name "*.go" -type f              # Goファイル検索
find . -name "*test*.go" -type f         # テストファイル検索

# ファイル内容確認
cat README.md                            # ファイル内容表示
head -n 20 file.go                       # 先頭20行表示
tail -n 20 file.go                       # 末尾20行表示
grep -r "function_name" ./internal/      # 文字列検索（再帰）

# ディスク使用量
du -sh .ai_code_tracking/                # トラッキングデータサイズ
du -sh bin/                              # バイナリサイズ
```

### パフォーマンス確認
```bash
# ベンチマーク実行
go test -bench=. ./internal/period/

# プロファイリング
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# メモリ使用量確認
go test -memprofile=mem.prof ./...
```

### システム情報（Darwin固有）
```bash
# システム情報
uname -a                                 # システム情報
sw_vers                                  # macOSバージョン
which go                                 # Go実行ファイルパス
go version                               # Goバージョン
echo $GOPATH                             # Go作業ディレクトリ
echo $PATH                               # 環境変数PATH
```

### デバッグとトラブルシューティング
```bash
# 詳細ログ付き実行
AICT_DEBUG=1 aict report

# 設定ファイル確認
cat .ai_code_tracking/config.json
cat .claude/settings.json

# フックファイル確認
ls -la .ai_code_tracking/hooks/
cat .ai_code_tracking/hooks/pre-tool-use.sh

# JSONLファイル確認
head -n 5 .ai_code_tracking/checkpoints.jsonl
tail -n 5 .ai_code_tracking/checkpoints.jsonl
```

## 重要な注意事項
- **Darwin特有**: macOSでは `find` コマンドの構文が Linux と若干異なる場合があります
- **パス区切り**: macOS では `/` を使用（Windows の `\` と異なる）
- **実行権限**: フックファイルは `chmod +x` で実行権限を付与
- **環境変数**: `$GOPATH/bin` が `$PATH` に含まれていることを確認