# AICT コードレビュー対応 TODO

コードレビューで検出された課題を安全な単位で対応するためのタスクリスト。
各タスクは1コミット単位で、`go test ./...` + `go build` 確認後に完了とする。

## Phase 1: データ整合性・正確性

- [ ] **1-2**: `fmt.Sscanf` を `strconv.Atoi` に統一 (Critical)
  - `cmd/aict/handlers_range.go:389,392`
  - `cmd/aict/handlers_commit.go:168,197,200`
  - パース失敗時のサイレントゼロ扱いを修正
- [ ] **1-1**: SaveCheckpoint のアトミック書き込み導入 (Critical)
  - `internal/storage/aict_storage.go:45`
  - 一時ファイル + `os.Rename` によるアトミック書き込み
  - `LoadCheckpoints()` エラー無視の修正
- [ ] **1-4**: GetAuthorshipLog のエラー区別 (Medium)
  - `internal/gitnotes/notes.go:156-169`
  - 「ノート未存在」と「実行エラー」の区別
- [ ] **1-3**: AI判定ロジックの統一 (High)
  - `cmd/aict/handlers_checkpoint.go:383-402` (`isAIAgent`)
  - `internal/tracker/analyzer.go:295-310` (`IsAIAuthor`)
  - 共通パッケージに統一関数を作成

## Phase 2: デッドコード削除

- [ ] **2-5**: `test_edit.go` を削除 (Low)
- [ ] **2-4**: 未使用変数・構造体の削除 (Medium)
  - `handlers_range.go:98` の `byFile`
  - `handlers_range.go:261-268` の `FileStatsRange`
- [ ] **2-1**: 未使用パッケージの削除 (High)
  - `internal/interfaces/storage.go`
  - `internal/git/diff_context.go`
  - `internal/git/diff.go` 内の未使用メソッド
- [ ] **2-2**: レガシーコードの削除 (High)
  - `internal/tracker/checkpoint.go`
  - `internal/storage/json.go`
  - `internal/templates/hooks.go:84-87`
  - `internal/errors/errors.go:99-107`
  - `internal/gitnotes/notes.go:21-27`
- [ ] **2-3**: セキュリティ/バリデーションモジュールの判断 (High)
  - `internal/security/safe_operations.go`
  - `internal/validation/config.go`
  - Phase 5/6 で統合するか削除するか判断

## Phase 3: コード重複の解消・品質改善

- [ ] **3-1**: numstat解析ロジックの統一 (High)
  - `internal/git/numstat.go` の `ParseNumstat` に集約
  - ハンドラ内の独自実装を削除
- [ ] **3-2**: Executor インスタンスの使い回し (Medium)
  - 関数・ループ内での冗長な `NewExecutor()` 生成を整理
- [ ] **3-3**: main.go の gitexec 迂回修正 (Medium)
  - `exec.Command("git", ...)` を `gitexec` 経由に変更
- [ ] **3-4**: デバッグ出力の制御化 (Medium)
  - `AICT_DEBUG` 環境変数 or `--verbose` フラグで制御
- [ ] **3-5**: マジックストリング/ナンバーの定数化 (Low)

## Phase 4: ハンドラリファクタリング

- [ ] **4-1**: ハンドラの error 返却パターンへの移行 (High)
  - `os.Exit(1)` 33箇所を段階的に移行
  - 1ファイルずつ移行してテスト確認
- [ ] **4-2**: handleRangeReportWithOptions の分割 (Medium)
  - 185行 → `collectAuthorStats()`, `buildReport()` に分割
- [ ] **4-3**: buildAuthorshipLogFromDiff の重複初期化修正 (High)
  - `NewAIctStorage()` / `LoadConfig()` の重複呼び出し除去

## Phase 5: パフォーマンス改善

- [ ] **5-1**: レポート生成の N+1 問題解消 (High)
  - コミットN件に対する2N回のgitプロセス起動を削減
- [ ] **5-2**: チェックポイント保存の JSONL 化 (Medium)
  - 後方互換性を維持しつつ追記型に移行
- [ ] **5-3**: captureSnapshot のメモリ効率改善 (Medium)
  - `strings.Split` → `bytes.Count` に変更
- [ ] **5-4**: analyzeFromNumstat の二重ループ修正 (High)

## Phase 6: セキュリティ強化

- [ ] **6-1**: Git引数のオプション注入防止 (Medium)
  - `--` (end of options marker) を追加
- [ ] **6-2**: setup-hooks のリポジトリルート検出 (Low)
  - `git rev-parse --show-toplevel` で絶対パス化
- [ ] **6-3**: セキュリティモジュールの統合判断 (Medium)
  - Task 2-3 と連動して判断

## Phase 7: テスト品質向上

- [ ] **7-1**: gitnotes パッケージのテスト追加 (Medium)
  - `notes.go` 205行に対するユニットテスト作成
- [ ] **7-2**: templates パッケージのテスト追加 (Low)
- [ ] **7-3**: 既存テストの改善 (Low)
  - `handlers_range_test.go` のスケルトン実装

## Phase 8: その他

- [ ] **8-1**: コメント言語の統一 (Low)
- [ ] **8-2**: AuthorMappings の初期化 (Low)
- [ ] **8-3**: 変数名 filepath のシャドウイング解消 (Low)
