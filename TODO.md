# AICT コードレビュー対応 TODO

コードレビューで検出された課題を安全な単位で対応するためのタスクリスト。
各タスクは1コミット単位で、`go test ./...` + `go build` 確認後に完了とする。

## Phase 1: データ整合性・正確性

- [x] **1-2**: `fmt.Sscanf` を `strconv.Atoi` に統一 (Critical)
  - `cmd/aict/handlers_range.go:389,392`
  - `cmd/aict/handlers_commit.go:168,197,200`
  - パース失敗時のサイレントゼロ扱いを修正
- [x] **1-1**: SaveCheckpoint のアトミック書き込み導入 (Critical)
  - `internal/storage/aict_storage.go:45`
  - 一時ファイル + `os.Rename` によるアトミック書き込み
  - `LoadCheckpoints()` エラー無視の修正
- [x] **1-4**: GetAuthorshipLog のエラー区別 (Medium)
  - `internal/gitnotes/notes.go:156-169`
  - 「ノート未存在」と「実行エラー」の区別
- [x] **1-3**: AI判定ロジックの統一 (High)
  - `cmd/aict/handlers_checkpoint.go:383-402` (`isAIAgent`)
  - `internal/tracker/analyzer.go:295-310` (`IsAIAuthor`)
  - `internal/tracker/ai_agent.go` に統一関数 `IsAIAgent` を作成（循環依存回避のためtrackerパッケージに配置）

## Phase 1.5: 実動作確認

Phase 1 の変更がデータ整合性に影響するため、実環境での動作確認を行う。

- [x] **1.5-1**: チェックポイント記録の動作確認
  - human → `human`、Claude Code → `ai` で正常に保存・表示
- [x] **1.5-2**: レポート生成の動作確認
  - `aict report --since 7d` でコードベース貢献・作業量貢献が正常表示
  - 詳細メトリクスは常時表示仕様（`--detailed` フラグ不要）
- [x] **1.5-3**: AI判定の統一確認
  - copilot → `ai`、chatgpt-4 → `ai`、John Doe → `human` を確認
  - 旧実装で判定不可だった名前も正しく分類
- [x] **1.5-4**: Git notes操作の確認
  - `aict commit` → Authorship Log正常生成
  - `git notes show HEAD` → JSON正常出力
  - ノート未存在時 "no note found" メッセージ確認
- [x] **1.5-5**: Phase 1 変更箇所のテスト補強
  - `TestIsAIAgent`: copilot/chatgpt含む14ケース追加
  - `TestIsAIAgentNilMappings`: nil安全性テスト追加
  - `TestSaveCheckpointCorruptedFile`: 破損JSON時のエラー伝播テスト追加
  - `TestSaveCheckpointAtomicWrite`: 一時ファイル残留なし確認テスト追加

## Phase 2: デッドコード削除

- [x] **2-5**: `test_edit.go` を削除 (Low)
  - 既に削除済み（以前のリファクタリングで対応済み）
- [x] **2-4**: 未使用変数・構造体の削除 (Medium)
  - `handlers_range.go` の `byFile` 変数、`FileStatsRange` 構造体、関連ループを削除
- [x] **2-1**: 未使用パッケージの削除 (High)
  - `internal/interfaces/` パッケージ全体を削除
  - `internal/git/diff.go`, `diff_context.go`, `diff_test.go` を削除（DiffAnalyzer系）
- [x] **2-2**: レガシーコードの削除 (High)
  - `internal/tracker/checkpoint.go` + テスト（旧CheckpointManager）
  - `internal/storage/json.go` + テスト（旧JSONStorage）
  - `internal/templates/hooks.go` の `PreCommitHook` 定数
  - `internal/errors/errors.go` の `NewAnalysisError` + `ErrTypeAnalysis`
  - `internal/gitnotes/notes.go` の `AIEditNote` + 旧メソッド群（AddNote, GetNote, ListNotes, RemoveNote）
- [x] **2-3**: セキュリティ/バリデーションモジュールの判断 (High)
  - `internal/security/` パッケージ全体を削除（外部から未使用）
  - `internal/validation/` パッケージ全体を削除（外部から未使用）
  - `internal/errors/` パッケージ全体を削除（上記2パッケージ削除後に未使用化）
  - 必要時にPhase 5/6で適切な形で再実装

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
