# AI Code Tracker (AICT) - 開発タスク

## 現状の問題

### 🐛 バグ: AI生成コード割合が39%と誤表示される

**症状**:
- ほぼ全てのコードをAIが生成しているのに、レポートでは39%と表示される
- 実際は80%以上のはずが、61%が人間作成と表示される

**原因調査結果**:

1. **チェックポイント記録の問題** (`internal/tracker/checkpoint_jsonl.go:96`)
   ```go
   cmd := exec.Command("git", "diff", "HEAD", "--numstat")
   ```
   - `git diff HEAD` は「HEADと作業ディレクトリの差分」を取得
   - コミット前の変更がずっと記録され続ける
   - 同じコミット内で `added=32, deleted=31` が3回記録される

2. **チェックポイントの重複記録**
   ```json
   {"author":"human", "added":32, "deleted":31}   // pre-tool-use
   {"author":"claude", "added":32, "deleted":31}  // post-tool-use
   {"author":"human", "added":32, "deleted":31}   // post-commit
   ```
   - 同じ差分が3回記録されている
   - `AnalyzeRecords` は連続レコード間の差分を計算するため、差分≈0になる

3. **分析ロジックの問題** (`internal/tracker/analyzer_jsonl.go:19-32`)
   ```go
   for i := 1; i < len(records); i++ {
       diff := a.calculateRecordDiff(records[i-1], records[i])
       // diff = 0 になってしまう
   }
   ```

## 設計の根本的見直し

### 議論の結果

1. **測定したいもの**: コードベース全体のAI生成コード vs 人間が書いたコードの割合
2. **Git運用の考慮**: ブランチ切り替え、マージ、rebase、cherry-pick
3. **ユーザビリティ**: ユーザーに「AIのコミットか人間のコミットか」を意識させない

### 参考実装: git-ai

**リポジトリ**: https://github.com/acunniffe/git-ai

**核心的なアイデア**:
- ✅ **git notes を使用** - rebase/merge/cherry-pickに強い
- ✅ **エージェントが自動マーク** - Cursor, Claude Code, Copilotが自動的にCLIを呼ぶ
- ✅ **行レベル追跡** - どの行がAI生成かを記録
- ✅ **パーリポジトリ設定不要** - グローバルインストールで動作

**技術詳細**:
- Rust実装（90.2%）
- git plumbing commands使用
- パフォーマンス: <100ms (Chromiumリポジトリでテスト済み)
- `git-ai blame` コマンドでAI生成行を可視化

## 次のステップ: 2つのオプション

### Option A: 既存実装のバグ修正（短期）

**修正方針**:
1. `RecordCheckpoint` を修正して、チェックポイント間の差分を記録
2. スキップロジックを改善（同じ値の重複記録を防ぐ）
3. `AnalyzeRecords` の分析ロジックを修正

**メリット**: 既存ユーザーへの影響が少ない
**デメリット**: Git運用の複雑さに対応しきれない

### Option B: git notes方式への移行（推奨・長期）

**新アーキテクチャ**:

```
1. Claude Code Hooks (.claude/settings.json):
   {
     "hooks": {
       "PostToolUse": [{
         "command": "aict mark-ai-edit"
       }]
     }
   }

2. aict mark-ai-edit:
   - git diff で変更行を検出
   - git notes --ref=aict に記録
   - JSON形式: {"lines": [15,16,17], "tool": "claude", "files": ["file.go"]}

3. post-commit hook:
   - git notes を読んで永続記録に変換
   - コードベース統計を更新

4. aict report:
   - git blame + git notes でAI%を正確に計算
```

**データ構造**:
```go
// git notes に保存
type AIEditNote struct {
    Timestamp   time.Time
    Tool        string              // "claude", "copilot", etc
    Files       map[string][]int    // filepath -> line numbers
}

// 永続記録
type CommitSnapshot struct {
    Commit      string
    Branch      string
    TotalLines  int
    AILines     int
    HumanLines  int
    Percentage  float64
}
```

**git notes 名前空間**:
- `refs/notes/aict` を使用
- 他のツールと衝突しない（検証済み）
- rebase/merge/cherry-pickで保持される

**実装フェーズ**:
1. ✅ Phase 1: git notes マーキング機能
2. ✅ Phase 2: git notes 記録機能の実装
3. ✅ Phase 3: git blame ベースの分析基礎実装
4. 🔄 Phase 4: git notes と git blame を組み合わせて正確なAI%を計算

## 技術的検証結果

### git notes の安全性（検証済み ✅）

```bash
# カスタム名前空間でテスト
git notes --ref=test-aict add -m "test note" HEAD
git notes --ref=test-aict show HEAD  # → "test note"
ls .git/refs/notes/  # → test-aict ファイルが作成される

# 他のツールと衝突しない
refs/notes/commits      # デフォルト
refs/notes/ai-edits     # git-ai
refs/notes/aict         # AICT（提案）
```

## 最近の変更履歴

### v0.6.1 での改善
- ✅ デフォルト設定を `storage.GetDefaultConfig()` に一元化
- ✅ `.kt` (Kotlin) と `.swift` (Swift) を追跡対象に追加
- ✅ 重複したデフォルト設定コードを削除（DRY原則）
- ✅ テストの期待値を更新（17拡張子、9除外パターン、3 author mappings）

### 設定ファイル構造
```json
{
  "target_ai_percentage": 80.0,
  "tracked_extensions": [
    ".go", ".py", ".js", ".ts", ".java", ".cs", ".cpp", ".c", ".h",
    ".rb", ".php", ".swift", ".kt", ".rs", ".scala", ".r", ".m"
  ],
  "exclude_patterns": [
    "*_test.go", "*.test.js", "*.spec.ts", "*_test.py",
    "vendor/*", "node_modules/*", ".git/*", "dist/*", "build/*"
  ],
  "author_mappings": {
    "AI Assistant": "ai",
    "Claude": "ai",
    "GitHub Copilot": "ai",
    "y-hirakaw\n": "human"
  }
}
```

## 推奨される次のアクション

1. **Option Bを採用** - git notes方式への移行
   - より正確な測定が可能
   - Git運用に強い
   - git-aiとの思想的一貫性

2. **実装手順**:
   ```
   Step 1: aict mark-ai-edit コマンド実装
   Step 2: git notes 記録機能
   Step 3: post-commit での集計
   Step 4: git blame + notes による分析
   Step 5: レポート機能の改善
   Step 6: 既存データのマイグレーション
   ```

3. **マイルストーン**:
   - [x] MVP: git notes マーキング動作確認
   - [x] git blame ベースの分析基礎実装
   - [x] Claude Code hooks 統合
   - [ ] git notes + git blame で正確なAI%計算
   - [ ] テストとドキュメント整備

**注**: 既存データのマイグレーションは不要（新方式で再スタート）

## 参考リンク

- git-ai: https://github.com/acunniffe/git-ai
- Claude Code Hooks: https://docs.claude.com/en/docs/claude-code/hooks
- Git Notes Documentation: https://git-scm.com/docs/git-notes

## 開発環境

- Go version: 1.21+
- Current version: v0.6.1
- Main branch: main
- 最終コミット: 0d73a48 (feat: Centralize default config and add .kt/.swift support)

## 実装進捗 (2025-11-23)

### ✅ 完了
1. **aict mark-ai-edit コマンド** - AI編集をgit notesに記録
2. **git notes 統合** - `internal/gitnotes` パッケージ作成
3. **aict snapshot コマンド** - git blameベースのコードベース分析
4. **Claude Code hooks 更新** - PostToolUseフックで自動マーク
5. **git blame 分析基盤** - `internal/blame` パッケージ作成
6. **Phase 4: git notes + git blame 統合** - 正確なAI%計算の実装完了

#### Phase 4 実装詳細:
- ✅ git blameでコミットハッシュを正しく解析（40文字の16進数判定）
- ✅ 各コミットのgit notes (`refs/notes/aict`) をクエリ
- ✅ notesにファイルが記録されていればAI、なければauthor判定
- ✅ パフォーマンス最適化（git notesクエリ結果のキャッシング）
- ✅ `--post-commit`フラグ実装（コミット後のマーク機能）
- ✅ post-tool-use → post-commit ワークフロー実装
  - PostToolUseフックで`.pending_ai_edit`マーカー作成
  - PostCommitフックでマーカー読み込み、`mark-ai-edit --post-commit`実行
  - git notesをHEADコミットに記録、マーカー削除
- ✅ 動作確認完了（git notes自動記録、snapshotでの正確な計算）

### 🔄 次のタスク
**Phase 5**: テストとドキュメント更新

**残課題**:
- 過去のコミット（git notesなし）は人間としてカウントされる
  - これは設計通り：新方式で再スタート、過去データのマイグレーション不要
  - 今後のClaude Code編集から正確に追跡される

---

**作成日**: 2025-11-23
**最終更新**: 2025-11-23
**ステータス**: MVP実装完了、正確性向上フェーズ
