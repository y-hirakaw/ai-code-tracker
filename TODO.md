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

- [x] **3-1**: numstat解析ロジックの統一 (High)
  - `parseNumstatOutput`, `parseNumstatFiles` を削除、`git.ParseNumstat` に集約
  - `getCommitDiff` 内の手動パースも `git.ParseNumstat` に統一
- [x] **3-2**: Executor インスタンスの使い回し (Medium)
  - `convertSinceToRange` 内の冗長な `executor2` を削除
- [x] **3-3**: main.go の gitexec 迂回修正 (Medium)
  - `exec.Command("git", ...)` を `gitexec.NewExecutor()` 経由に変更
  - `os/exec` import を削除
- [x] **3-4**: デバッグ出力の制御化 (Medium)
  - `debugf()` ヘルパー関数を追加、`AICT_DEBUG` 環境変数で制御
  - 8箇所のハードコードされた `[DEBUG]` 出力を `debugf` に置き換え
- [x] **3-5**: マジックストリング/ナンバーの定数化 (Low)
  - `storage` パッケージにエクスポート定数4つ追加（AictDirName, CheckpointsDirName, LatestFileName, ConfigFileName）
  - `handlers_debug.go` のハードコードパスを定数参照に置き換え

## Phase 3.5: 動作確認

Phase 3 の変更（特にnumstat統一・gitexec移行）がデータパスに影響するため、実環境での動作確認を行う。

- [x] **3.5-1**: レポート生成の動作確認
  - `aict report --since 7d` → 正常出力（10コミット、作業量1561行）
  - `aict report --since 1d` → 正常出力（4コミット、作業量1414行）
  - numstat統一後もデータが正しく集計されていることを確認
- [x] **3.5-2**: チェックポイント操作の動作確認
  - `aict checkpoint --author human` → 正常に保存
  - `aict debug show` → 保存データの表示正常（タイムスタンプ、作成者、メタデータ）
  - `aict debug clean` → 削除正常
- [x] **3.5-3**: getGitUserName() の動作確認
  - gitexec移行後も `git config user.name` の値 (`y-hirakaw`) が正しく取得されていることをレポートのBy Author欄で確認

## Phase 4: ハンドラリファクタリング

- [x] **4-1**: ハンドラの error 返却パターンへの移行 (High)
  - 全7ハンドラファイルを `error` 返却に変更
  - `main()` で一元的にエラー表示 + `exitFunc(1)` に統一
  - `os.Exit(1)` 39箇所を完全除去（main.go の `exitFunc(1)` のみ残存）
- [x] **4-2**: handleRangeReportWithOptions の分割 (Medium)
  - 170行 → `collectAuthorStats()` + `buildReport()` + `authorStatsResult` 構造体に分割
  - `handleRangeReportWithOptions` は22行に削減（87%削減）
- [x] **4-3**: buildAuthorshipLogFromDiff の重複初期化修正 (High)
  - `handleCommit()` で `LoadConfig()` を呼び、`cfg` をパラメータとして渡すように変更
  - `buildAuthorshipLogFromDiff` 内の `NewAIctStorage()` / `LoadConfig()` を除去

## Phase 4.5: 動作確認

Phase 4 の変更（error返却パターン・関数分割・Config読み込み変更）の実環境動作確認。

- [x] **4.5-1**: レポート生成の動作確認（4-2 分割後）
  - `aict report --since 7d` → 正常出力（12コミット、作業量1859行）
  - `aict report --since 1d` → 正常出力（6コミット、作業量1712行）
  - collectAuthorStats/buildReport 分割後もデータが正しく集計されていることを確認
- [x] **4.5-2**: エラーハンドリングの動作確認（4-1 一元化後）
  - `aict report`（引数なし）→ usage表示 + `Error:` + exit 1
  - エラーが `main()` で一元的に表示されることを確認
- [x] **4.5-3**: チェックポイント操作の動作確認
  - `aict checkpoint --author human` → 正常に保存
  - `aict debug show` → 保存データの表示正常
  - `aict debug clean` → 削除正常
- [x] **4.5-4**: バージョン表示
  - `aict version` → `1.3.1` 正常出力

## Phase 5: パフォーマンス改善

- [x] **5-1**: レポート生成の N+1 問題解消 (High)
  - `git log --numstat --format=__AICT_COMMIT__%H` でバッチnumstat取得（`GetRangeNumstat`）
  - `git log --notes=refs/aict/authorship --format=__AICT_HASH__%H%n%N` でバッチnotes取得（`GetAuthorshipLogsForRange`）
  - `collectAuthorStats` をバッチ化、gitプロセス起動を2N+1回→2回に削減
- [x] **5-2**: チェックポイント保存の JSONL 化 (Medium)
  - `SaveCheckpoint`: JSON配列の全読み→追加→全書き（O(n)）→ JSONL追記（O(1)）に変更
  - `LoadCheckpoints`: JSON配列（旧形式）とJSONL（新形式）の自動判別
  - `migrateToJSONLIfNeeded`: 旧JSON配列ファイルの自動マイグレーション
  - `handlers_debug.go`: 直接ファイル操作をstorage APIに統一
  - テスト追加: 後方互換性、マイグレーション、破損ファイル、JSONL形式確認
- [x] **5-3**: captureSnapshot のメモリ効率改善 (Medium)
  - `captureSnapshot` の行数カウント: `len(strings.Split(string(content), "\n"))` → `bytes.Count(content, []byte{'\n'}) + 1`
  - `getDetailedDiff` の新規ファイル行数カウント: 同様に `bytes.Count` + `bytes.TrimSpace` に変更
  - 大きなファイルで不要なスライス生成を回避し、メモリ使用量を削減
- [x] **5-4**: analyzeFromNumstat の二重ループ修正 (High)
  - 2つの `for filepath, afterStats := range after.NumstatData` ループを1つに統合
  - 新規ファイルの行数が二重カウントされていたバグを修正
  - NewFilesメトリクスを統合ループ内で処理

## Phase 6: セキュリティ強化

- [x] **6-1**: Git引数のオプション注入防止 (Medium)
  - `git notes add/show`: commit引数の前に `--` を追加
  - `git log`: `--end-of-options` を追加（`notes.go`, `numstat.go`, `handlers_range.go`）
  - `git diff --numstat`: `ValidateRevisionArg` バリデーション追加
  - `gitexec.ValidateRevisionArg()`: `-` で始まるリビジョン引数を拒否する関数を追加
- [x] **6-2**: setup-hooks のリポジトリルート検出 (Low)
  - `git rev-parse --show-toplevel` で絶対パス化
- [x] **6-3**: セキュリティモジュールの統合判断 (Medium)
  - 判断: 再実装不要。Phase 6-1のインライン対策（ValidateRevisionArg, `--`, `--end-of-options`）で十分
  - Phase 2-3で削除した`internal/security/`, `internal/validation/`は現在の規模では不要

## Phase 7: テスト品質向上

- [x] **7-1**: gitnotes パッケージのテスト追加 (Medium)
  - `notes_test.go` に7テスト追加（GitError, SecurityArgs, InvalidJSON, NoNotes, isNoteNotFound等）
- [x] **7-2**: templates パッケージのテスト追加 (Low)
  - `hooks_test.go` に4テスト追加（ExitCleanly, BinaryDetection, GitRevParse, AICTInitialized）
- [x] **7-3**: 既存テストの改善 (Low)
  - `handlers_range_test.go` に2テスト追加（ExpandShorthandDate 13ケース, IsNumeric 7ケース, AuthorCommitCountAccuracy）

## Phase 8: その他

- [x] **8-1**: コメント言語の統一 (Low)
- [x] **8-2**: AuthorMappings の初期化 (Low)
- [x] **8-3**: 変数名 filepath のシャドウイング解消 (Low)

---

## 今後の改善候補

テストレビュー・実動作テスト（v1.4.0）で特定された改善候補。優先度順。

### テストカバレッジ向上

- [x] **T-1**: `cmd/aict/` パッケージのテストカバレッジ向上 (High)
  - `buildReport`: 3テスト（AI/人間混在、since表示、単一作成者）
  - `buildAuthorshipLogFromDiff`: 5テスト（チェックポイント一致、デフォルト作成者、除外ファイル、未追跡拡張子、空diff）
  - `formatRangeReport`: 3テスト（JSON/table/不正フォーマット）

- [x] **T-2**: `handlers_checkpoint.go` の純粋関数テスト追加 (Medium)
  - `getFileList`: 3テスト（空マップ、単一ファイル、複数ファイル）
  - `detectChangesFromSnapshot`: 混合変更エッジケース追加（新規+変更+削除+未変更の複合テスト）

### 入力バリデーション

- [x] **V-1**: `--since` フラグの入力バリデーション (Low)
  - `validateSinceInput()` 関数追加: 既知の日付形式かチェックし、未知の場合に警告を出力
  - 対応形式: 7d/2w/1m/1y、YYYY-MM-DD、yesterday/today、N days ago 等
  - テスト: 18ケース（正常形式13、異常形式5）

- [x] **V-2**: `--format` フラグの不正値エラー改善 (Low)
  - エラーメッセージに `(available: table, json)` を追加
  - テスト: 利用可能フォーマット一覧が含まれることを検証

---

## 多角的評価で検出された課題（v1.4.1）

4つの専門エージェント（コード品質・テスト品質・セキュリティ・ドキュメント）による並行分析で検出。

### Phase 9: データ整合性・バグリスク (Critical)

- [x] **9-1**: `isTrackedFile` / `shouldTrackFile` の統一 (Critical)
  - `internal/tracker/file_filter.go` に `IsTrackedFile()` / `MatchesPattern()` を作成
  - `analyzer.go` の `shouldTrackFile()` を `IsTrackedFile()` に委譲（`strings.Contains` バグを修正）
  - `handlers_commit.go` のローカル実装を削除し `tracker.IsTrackedFile()` を使用
  - テスト: `file_filter_test.go` 追加、既存テストをワイルドカードパターンに修正

- [x] **9-2**: JSONパースエラーのログ出力追加 (High)
  - `internal/storage/aict_storage.go`: JSONL行パースエラー時に `log.Printf` でwarning出力
  - `internal/gitnotes/notes.go`: Authorship Log JSONパースエラー・取得エラー時に `log.Printf` でwarning出力
  - 変数名 `log` → `alog` に変更（`log` パッケージとの衝突回避）

- [x] **9-3**: エラー二重出力の解消 (Medium)
  - `handlers_range.go`: 排他チェックの `fmt.Println("Error: ...")` を削除、エラーメッセージを統合
  - `handlers_debug.go`: サブコマンド未指定・不明時の `fmt.Println("エラー: ...")` を削除、`fmt.Errorf` に一本化

- [x] **9-4**: `handleDebugClearNotes()` の ref フィルタリング改善 (Medium)
  - `strings.Contains(line, "aict")` → プレフィックスベースのマッチングに変更
  - 対象: `refs/aict/*`, `refs/notes/aict*`, `refs/notes/refs/aict/*`
  - ブランチ名に"aict"を含むrefの誤削除を防止（動作確認済み）

- [x] **9-5**: `buildReport()` のゼロ除算ガード追加 (Medium)
  - byAuthorループ内に `report.Summary.TotalLines > 0` ガードを追加
  - `TotalLines == 0` 時の `+Inf` によるJSON出力エラーを防止

### Phase 10: コード品質・アーキテクチャ改善

- [x] **10-1**: `collectAuthorStats()` の分割 (High, CC=12)
  - `processCommitFiles()`, `processFileAuthors()`, `calculateAuthorContribution()`, `accumulateMetrics()` に分割
  - CC=12 → CC≤5 に削減

- [x] **10-2**: `cmd/aict/` のビジネスロジック分離 (High)
  - `buildAuthorshipMap()`, `buildAuthorshipLogFromDiff()` を `internal/authorship/builder.go` に移動
  - テストも `internal/authorship/builder_test.go` に移動
  - CLIハンドラーはパラメータ解析とロジック呼び出しに専念

- [x] **10-3**: `gitexec.NewExecutor()` のDIパターン化 (Medium)
  - `cmd/aict/executor.go` にパッケージレベル `newExecutor` ファクトリ変数を導入
  - 全16箇所の `gitexec.NewExecutor()` を `newExecutor()` に置換
  - テスト時にモック注入可能、各ファイルから `gitexec` インポートを削除

- [x] **10-4**: ストレージ初期化+設定読み込みの共通化 (Low)
  - `cmd/aict/helpers.go` に `loadStorageAndConfig()` ヘルパー関数を作成
  - `handlers_checkpoint.go`, `handlers_commit.go` の重複パターンを共通化

- [x] **10-5**: `config.json` 読み込み時のバリデーション追加 (Medium)
  - `internal/storage/aict_storage.go` に `validateConfig()` 関数を追加
  - `TargetAIPercentage` の範囲チェック（0-100）
  - `TrackedExtensions` が空でないことの確認
  - `DefaultAuthor` が空文字列でないことの確認
  - テスト: 8ケース（正常3、異常5）

### Phase 11: テスト品質向上

- [x] **11-1**: `cmd/aict` のCLIハンドラーテスト追加 (High)
  - `handlers_init_test.go`: 3テスト（CreatesConfig, ConfigValues, Idempotent）
  - `handlers_sync_test.go`: 6テスト（MissingSubcommand, UnknownSubcommand, Push, Fetch, PushError, FetchError）
  - `handlers_debug_test.go`: 10テスト（MissingSubcommand, UnknownSubcommand, DispatchShow, ShowNoCheckpoints, ShowWithCheckpoints, CleanNoCheckpoints, CleanWithCheckpoints, DisplayCheckpoint, ClearNotesNoAictRefs, ClearNotesWithAictRefs, ClearNotesShowRefError）
  - `main_test.go`: 10テスト（Debugf Enabled/Disabled, GetGitUserName success/error, main() Version/Help/NoArgs/Unknown/SyncError/DebugError/Checkpoint/VersionFlags/HelpFlag）
  - カバレッジ: 27.3% → 56.7%（目標50%+達成）

- [x] **11-2**: 偽テストの整理 (Medium)
  - `internal/tracker/types_test.go`: 5つの構造体テスト削除（TestCheckpointStructure等）
  - `TestGetDisplayBranch` を追加（4ケース、GetDisplayBranchメソッドの実テスト）

- [x] **11-3**: スキップされたテストの削除 (Medium)
  - `TestDetectChanges`, `TestGetLineRanges` を削除（handlers_checkpoint_v2_test.goに同等テスト存在）

- [x] **11-4**: 空テストの実装 (Low)
  - `TestGetCurrentCommit`: 実Gitリポジトリでコミットハッシュ40文字を検証

- [x] **11-5**: `t.Run` 未使用テストの改善 (Low)
  - `TestIsAIAuthor`: `name`フィールド + `t.Run` 追加
  - `TestIsNoteNotFound`: `name`フィールド + `t.Run` 追加
  - `TestShouldTrackFile` は既に `t.Run` 使用済み（対応不要）

### Phase 12: ドキュメント整合性

- [ ] **12-1**: CLAUDE.md のディレクトリ構造更新 (High)
  - `internal/config/`, `internal/hooks/`, `internal/checkpoint/` は存在しない（削除済み）
  - `internal/git/`（numstat.go）が未記載
  - チェックポイント保存形式を「JSON array」→「JSONL」に修正
  - `--detailed` フラグの記載を常時表示仕様に修正

- [ ] **12-2**: README.md の英語化と充実 (High)
  - CLAUDE.mdの「README.mdだけは英語で記載すること」指示に違反（現在日本語）
  - `go install` コマンド、基本コマンド一覧、バッジ等の追加

- [ ] **12-3**: SPEC.md の整理 (Medium)
  - ロードマップが全て未チェック `[ ]` のまま（実装済み機能多数）
  - 未実装オプション（`--branch`, `--last`, `--by-file` 等）が仕様として残存
  - 初期仕様書として凍結するか、実装状況を反映するか判断

- [ ] **12-4**: IMPLEMENTATION_STATUS.md のアーカイブ化 (Low)
  - v0.7.0で凍結されており現在v1.4.1との乖離が大きい
  - ファイル先頭に注記を追加するか、`docs/archive/` へ移動

- [ ] **12-5**: go.mod のモジュール名確認 (Low)
  - モジュール名 `github.com/y-hirakaw/ai-code-tracker`（`a`なし）
  - ローカルディレクトリ名 `y-hirakawa`（`a`あり）
  - GitHubリポジトリ名との整合性確認が必要
