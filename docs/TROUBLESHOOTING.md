# AI Code Tracker - トラブルシューティング

## 📋 目次

- [よくある問題](#よくある問題)
- [インストール問題](#インストール問題)
- [設定問題](#設定問題)
- [パフォーマンス問題](#パフォーマンス問題)  
- [Webダッシュボード問題](#webダッシュボード問題)
- [デバッグ方法](#デバッグ方法)

## よくある問題

### Claude Code Hooksが動作しない

**症状**: Claude Codeでファイル編集してもトラッキングされない

**解決方法**:
```bash
# Hook設定を確認
aict config --list

# Hook設定を再実行
aict setup hooks

# Claude Code設定を確認
cat ~/.claude/hooks.json
```

### 統計が表示されない

**症状**: `aict stats` で「データなし」と表示される

**解決方法**:
```bash
# データファイルの存在確認
ls -la .git/ai-tracker/

# 手動でテストデータを作成
aict track --test

# 権限を確認
ls -la .git/ai-tracker/tracks.jsonl
```

### Blame表示で文字化けする

**症状**: 日本語のファイル名やコメントが文字化け

**解決方法**:
```bash
# 言語設定を確認
aict lang

# UTF-8で表示
export LANG=ja_JP.UTF-8
aict blame ファイル名.go
```

## インストール問題

### Go版が古い

**症状**: `go install` でエラーが発生

**解決方法**:
```bash
# Go版を確認
go version

# Go 1.19以上にアップデート
# macOS (Homebrew)
brew install go

# Linux
sudo apt install golang-go
```

### パス設定の問題

**症状**: `aict: command not found`

**解決方法**:
```bash
# GOPATHを確認
echo $GOPATH

# PATHに追加
export PATH=$PATH:$(go env GOPATH)/bin

# 永続化（.bashrc / .zshrc）
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

## 設定問題

### 環境変数が反映されない

**症状**: 設定を変更しても動作が変わらない

**解決方法**:
```bash
# 現在の環境変数を確認
env | grep AICT

# 設定ファイルを確認
aict config --list

# 設定をリセット
aict config --reset
```

### 多言語設定が効かない

**症状**: 日本語に設定しても英語で表示される

**解決方法**:
```bash
# 言語設定を確認
aict lang

# 明示的に設定
aict lang ja --persistent

# 環境変数で設定
export AICT_LANGUAGE=ja
```

## パフォーマンス問題

### 統計計算が遅い

**症状**: `aict stats` の実行に時間がかかる

**解決方法**:
```bash
# キャッシュをクリア
rm .git/ai-tracker/stats-cache.json

# インデックスを再構築
aict index --rebuild

# 古いデータを削除
aict clean --older-than 90d
```

### Blame表示が遅い

**症状**: 大きなファイルでblameが遅い

**解決方法**:
```bash
# 行範囲を指定
aict blame large_file.go --lines 1-100

# キャッシュを有効化
export AICT_ENABLE_CACHE=true
```

## Webダッシュボード問題

### ポートが使用中

**症状**: `bind: address already in use`

**解決方法**:
```bash
# 別のポートを使用
aict web -p 3001

# 使用中のプロセスを確認
lsof -i :8080

# プロセスを終了
kill -9 <PID>
```

### ブラウザで表示されない

**症状**: ダッシュボードが開かない

**解決方法**:
```bash
# 手動でブラウザを開く
aict web --no-browser
# 別途 http://localhost:8080/dashboard にアクセス

# ログを確認
aict web --debug
```

### WebSocketが切断される

**症状**: リアルタイム更新が停止する

**解決方法**:
```bash
# ファイアウォール設定を確認
# デバッグモードで起動
aict web --debug

# 別のポートで試行
aict web -p 8081
```

## デバッグ方法

### デバッグモードの有効化

```bash
# 環境変数で有効化
export AICT_DEBUG=true

# ログレベルを詳細に
export AICT_LOG_LEVEL=debug

# コマンド実行
aict stats --debug
```

### ログファイルの確認

```bash
# ログファイルの場所
ls -la ~/.aict/logs/

# 最新のログを確認
tail -f ~/.aict/logs/aict.log

# エラーログのみ確認
grep ERROR ~/.aict/logs/aict.log
```

### データファイルの診断

```bash
# データファイルの整合性チェック
aict diagnose --check-data

# JSONLファイルの形式チェック
jq . .git/ai-tracker/tracks.jsonl

# ファイルサイズを確認
du -h .git/ai-tracker/
```

### セキュリティ診断

```bash
# セキュリティスキャン実行
aict security scan

# 権限チェック
ls -la .git/ai-tracker/

# 暗号化状況確認
aict security status
```

## エラーメッセージ別対処法

### "Permission denied"

**原因**: ファイル権限の問題

**解決方法**:
```bash
# 権限を修正
chmod -R 700 .git/ai-tracker/
chmod 600 .git/ai-tracker/*.jsonl
```

### "Invalid JSON format"

**原因**: JSONLファイルの破損

**解決方法**:
```bash
# バックアップから復元
aict restore --input backup.tar.gz

# 破損行を特定
jq . .git/ai-tracker/tracks.jsonl
```

### "Hook not found"

**原因**: Claude Code hooks設定の問題

**解決方法**:
```bash
# Hook設定を再実行
aict setup hooks --force

# Claude Code設定ファイルを確認
cat ~/.claude/hooks.json
```

## サポートリソース

### 情報収集コマンド
```bash
# システム情報を収集
aict version --verbose
aict config --list
aict security status

# 環境情報
go version
git --version
uname -a
```

### 問題報告時の情報

問題を報告する際は以下の情報を含めてください：

1. **環境情報**
   - OS・Go・Git のバージョン
   - `aict version` の出力

2. **設定情報**
   - `aict config --list` の出力
   - 関連する環境変数

3. **エラー情報**
   - 具体的なエラーメッセージ
   - 再現手順

4. **ログ情報**
   - デバッグモードでの実行結果
   - ログファイルの該当部分

### コミュニティリソース

- **GitHub Issues**: https://github.com/ai-code-tracker/aict/issues
- **GitHub Discussions**: https://github.com/ai-code-tracker/aict/discussions
- **ドキュメント**: https://ai-code-tracker.dev/docs