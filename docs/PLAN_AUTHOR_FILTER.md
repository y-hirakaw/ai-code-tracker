# Implementation Plan: --author Filter Flag

## 概要

`aict report`コマンドに`--author`フラグを追加し、特定の作者（複数可）でレポートをフィルタリングできるようにする。

**バージョン**: v1.2.0

**目的**: チーム開発環境で「自分とClaude」など、特定の作者のコード貢献度のみを確認できるようにする。

---

## 背景

現在の`aict report`コマンドは、指定した期間のすべてのコミット・すべての作者を表示する。チーム開発では以下のニーズがある：

1. **個人の振り返り**: 自分とAIアシスタントのコード生成率のみを確認したい
2. **ペアプログラミング**: 特定の2人の貢献度を確認したい
3. **チーム分析**: 特定のサブチームのメトリクスを確認したい

---

## 要件定義

### 機能要件

#### FR-1: 単一作者フィルタリング
```bash
aict report --since 7d --author "Aさん"
```
- 指定した作者のコミットのみをレポートに含める
- AI/人間の分類は維持する

#### FR-2: 複数作者フィルタリング
```bash
aict report --since 7d --author "Aさん,Claude"
```
- カンマ区切りで複数の作者を指定可能
- スペースは自動的にトリムされる
- 大文字小文字は区別する（既存の作者名と完全一致）

#### FR-3: 既存フラグとの互換性
```bash
# --range との組み合わせ
aict report --range origin/main..HEAD --author "Aさん,Claude"

# --since との組み合わせ
aict report --since 2w --author "Aさん"

# --format との組み合わせ
aict report --since 7d --author "Aさん,Claude" --format json
```

#### FR-4: フィルタリング後のメトリクス再計算
- Summary統計はフィルタリング後の作者のみで再計算
- AI生成率はフィルタリング後の行数で計算
- 詳細メトリクスもフィルタリング後のデータで再計算

### 非機能要件

#### NFR-1: パフォーマンス
- フィルタリングによるパフォーマンス劣化は最小限（<5%）
- 既存の処理フローを大きく変更しない

#### NFR-2: 後方互換性
- `--author`フラグなしの動作は完全に維持
- 既存のJSON出力フォーマットを変更しない

#### NFR-3: エラーハンドリング
- 存在しない作者名を指定した場合でもエラーにしない（空のレポート）
- 無効なフラグの組み合わせは検出しない（すべて許可）

---

## 設計

### アーキテクチャ

```
┌─────────────────────────────────────┐
│ CLI Flag Parsing                    │
│ --author "Alice,Bob,Claude"         │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ Author Filter Initialization        │
│ - Split by comma                    │
│ - Trim whitespace                   │
│ - Create map[string]bool            │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ Report Generation Loop              │
│ for each commit:                    │
│   for each file:                    │
│     for each author:                │
│       if authorFilter[author.Name]: │
│         aggregate metrics           │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ Metrics Recalculation               │
│ - Total lines                       │
│ - AI percentage                     │
│ - Detailed metrics                  │
└─────────────────────────────────────┘
```

### データ構造

#### 修正: ReportOptions
```go
type ReportOptions struct {
    Range  string
    Since  string
    Format string
    Author string  // 新規追加: カンマ区切りの作者名リスト
}
```

#### 内部: Author Filter
```go
var authorFilter map[string]bool
if opts.Author != "" {
    authorFilter = make(map[string]bool)
    for _, name := range strings.Split(opts.Author, ",") {
        authorFilter[strings.TrimSpace(name)] = true
    }
}
```

### フィルタリングロジック

#### 疑似コード
```go
// 各ファイルの作者情報をループ
for _, author := range fileInfo.Authors {
    // Author filtering
    if authorFilter != nil && !authorFilter[author.Name] {
        continue  // この作者をスキップ
    }

    // 既存の集計処理
    if author.Type == tracker.AuthorTypeAI {
        aiLines += countLines(author.Lines)
    } else {
        humanLines += countLines(author.Lines)
    }

    // 詳細メトリクスの集計
    // ...
}
```

---

## 実装計画

### Phase 1: 基本実装（必須）

#### タスク1.1: フラグ定義の追加
**ファイル**: `cmd/aict/handlers_range.go`

**変更内容**:
```go
type ReportOptions struct {
    Range  string
    Since  string
    Format string
    Author string  // 追加
}

func handleRangeReport() {
    var opts ReportOptions
    fs := flag.NewFlagSet("report", flag.ExitOnError)

    // 既存フラグ
    fs.StringVar(&opts.Range, "range", "", "...")
    fs.StringVar(&opts.Since, "since", "", "...")
    fs.StringVar(&opts.Format, "format", "table", "...")

    // 新規フラグ
    fs.StringVar(&opts.Author, "author", "", "Filter by author name(s), comma-separated (e.g., \"Alice,Bob,Claude\")")

    // ...
}
```

**工数**: 0.5時間

---

#### タスク1.2: Author Filterの初期化
**ファイル**: `cmd/aict/handlers_range.go`

**変更内容**:
```go
func handleRangeReportWithOptions(opts ReportOptions) {
    // ... 既存処理 ...

    // Author filter parsing
    var authorFilter map[string]bool
    if opts.Author != "" {
        authorFilter = make(map[string]bool)
        for _, name := range strings.Split(opts.Author, ",") {
            authorFilter[strings.TrimSpace(name)] = true
        }
    }

    // ... レポート生成処理 ...
}
```

**工数**: 0.5時間

---

#### タスク1.3: フィルタリングロジックの追加
**ファイル**: `cmd/aict/handlers_range.go`

**変更箇所**: 行162-250付近のAuthorshipLog処理ループ内

**変更内容**:
```go
// 既存: for _, author := range fileInfo.Authors {
for _, author := range fileInfo.Authors {
    // ★ 新規追加: Author filtering
    if authorFilter != nil && !authorFilter[author.Name] {
        continue
    }

    // 既存の処理はそのまま
    numstat, found := numstatMap[filepath]
    if !found {
        continue
    }

    added := numstat[0]
    deleted := numstat[1]

    if author.Type == tracker.AuthorTypeAI {
        aiLines += added
        humanLines += 0

        // 詳細メトリクス
        detailedMetrics.WorkVolume.AIAdded += added
        detailedMetrics.WorkVolume.AIDeleted += deleted
        detailedMetrics.WorkVolume.AIChanges += added + deleted
        detailedMetrics.Contributions.AIAdditions += added
    } else {
        aiLines += 0
        humanLines += added

        // 詳細メトリクス
        detailedMetrics.WorkVolume.HumanAdded += added
        detailedMetrics.WorkVolume.HumanDeleted += deleted
        detailedMetrics.WorkVolume.HumanChanges += added + deleted
        detailedMetrics.Contributions.HumanAdditions += added
    }

    // ... 残りの処理 ...
}
```

**工数**: 1時間

---

#### タスク1.4: By Authorセクションのフィルタリング
**ファイル**: `cmd/aict/handlers_range.go`

**変更箇所**: 行252-280付近のBy Author集計部分

**変更内容**:
```go
// Build by-author stats
authorStatsMap := make(map[string]*tracker.AuthorStats)
for _, author := range fileInfo.Authors {
    // ★ 新規追加: Author filtering
    if authorFilter != nil && !authorFilter[author.Name] {
        continue
    }

    // 既存の集計処理
    if stats, exists := authorStatsMap[author.Name]; exists {
        stats.Lines += countLines(author.Lines)
    } else {
        authorStatsMap[author.Name] = &tracker.AuthorStats{
            Name:  author.Name,
            Type:  author.Type,
            Lines: countLines(author.Lines),
        }
    }
}
```

**工数**: 0.5時間

---

#### タスク1.5: Top Filesセクションのフィルタリング
**ファイル**: `cmd/aict/handlers_range.go`

**変更箇所**: 行282-310付近のTop Files集計部分

**変更内容**:
```go
// Build by-file stats
fileStatsMap := make(map[string]*tracker.FileStats)
for _, author := range fileInfo.Authors {
    // ★ 新規追加: Author filtering
    if authorFilter != nil && !authorFilter[author.Name] {
        continue
    }

    // 既存の集計処理
    if stats, exists := fileStatsMap[filepath]; exists {
        if author.Type == tracker.AuthorTypeAI {
            stats.AILines += countLines(author.Lines)
        } else {
            stats.HumanLines += countLines(author.Lines)
        }
        stats.TotalLines = stats.AILines + stats.HumanLines
    } else {
        // 新規作成
        // ...
    }
}
```

**工数**: 0.5時間

---

### Phase 2: テスト実装（必須）

#### タスク2.1: ユニットテストの追加
**ファイル**: `cmd/aict/handlers_range_test.go`

**テストケース**:
```go
func TestAuthorFilter_SingleAuthor(t *testing.T) {
    // 単一作者フィルタリング
}

func TestAuthorFilter_MultipleAuthors(t *testing.T) {
    // 複数作者フィルタリング
}

func TestAuthorFilter_NoMatch(t *testing.T) {
    // 該当作者なしの場合（空のレポート）
}

func TestAuthorFilter_WithWhitespace(t *testing.T) {
    // スペースのトリム確認
}

func TestAuthorFilter_CaseSensitive(t *testing.T) {
    // 大文字小文字の区別確認
}

func TestAuthorFilter_WithoutFlag(t *testing.T) {
    // フラグなしの動作確認（後方互換性）
}
```

**工数**: 2時間

---

#### タスク2.2: 統合テストの追加
**ファイル**: 新規 `test_author_filter.sh`

**テストシナリオ**:
```bash
#!/bin/bash
# test_author_filter.sh

# 1. Setup: 複数作者のコミットを作成
# 2. Test: --author "Aさん" でフィルタリング
# 3. Verify: Aさんのみが表示される
# 4. Test: --author "Aさん,Claude" でフィルタリング
# 5. Verify: AさんとClaudeのみが表示される
# 6. Test: --author "存在しない人"
# 7. Verify: 空のレポート
```

**工数**: 1.5時間

---

### Phase 3: ドキュメント更新（必須）

#### タスク3.1: USAGE.md更新
**ファイル**: `docs/USAGE.md`

**追加箇所**: レポートコマンドのオプションセクション

**追加内容**:
```markdown
### オプション

| オプション | 説明 | デフォルト |
|----------|------|-----------|
| `--range <range>` | コミット範囲を指定 | - |
| `--since <date>` | 指定日時以降のコミット | - |
| `--format <format>` | 出力フォーマット（`table` または `json`） | `table` |
| `--author <names>` | 作者名でフィルタリング（カンマ区切り） | すべて |

### 作者フィルタリング

特定の作者のコード貢献度のみを確認できます：

#### 単一作者
```bash
# 自分のコード生成率のみ
aict report --since 7d --author "Your Name"
```

#### 複数作者
```bash
# 自分とClaude
aict report --since 7d --author "Your Name,Claude"

# チームのサブグループ
aict report --since 2w --author "Alice,Bob,Carol"
```

#### 使用例

**個人の振り返り**:
```bash
# 過去1週間の自分とAIの貢献度
aict report --since 1w --author "$(git config user.name),Claude"
```

**ペアプログラミング**:
```bash
# 特定の2人のペアの成果
aict report --since 1d --author "Alice,Bob"
```

**PR作成時**:
```bash
# このブランチでの自分とAIの貢献
aict report --range origin/main..HEAD --author "$(git config user.name),Claude"
```
```

**工数**: 1時間

---

#### タスク3.2: USECASE.md更新
**ファイル**: `docs/USECASE.md`

**追加セクション**: "個人のAI活用率確認"

**工数**: 0.5時間

---

#### タスク3.3: PLAN更新
**ファイル**: このファイル（`docs/PLAN_AUTHOR_FILTER.md`）

**最終更新**: 実装完了時にステータス更新

**工数**: 0.5時間

---

## 工数見積もり

| フェーズ | タスク | 工数 | 累計 |
|---------|-------|------|------|
| Phase 1 | 1.1 フラグ定義 | 0.5h | 0.5h |
| Phase 1 | 1.2 Filter初期化 | 0.5h | 1.0h |
| Phase 1 | 1.3 フィルタリングロジック | 1.0h | 2.0h |
| Phase 1 | 1.4 By Author集計 | 0.5h | 2.5h |
| Phase 1 | 1.5 Top Files集計 | 0.5h | 3.0h |
| Phase 2 | 2.1 ユニットテスト | 2.0h | 5.0h |
| Phase 2 | 2.2 統合テスト | 1.5h | 6.5h |
| Phase 3 | 3.1 USAGE.md | 1.0h | 7.5h |
| Phase 3 | 3.2 USECASE.md | 0.5h | 8.0h |
| Phase 3 | 3.3 PLAN更新 | 0.5h | 8.5h |
| - | **合計** | **8.5h** | - |

**バッファ含む**: 10時間（約1.5日）

---

## テスト計画

### ユニットテスト

#### Test 1: 単一作者フィルタリング
```go
func TestAuthorFilter_SingleAuthor(t *testing.T) {
    // Setup
    opts := ReportOptions{
        Since:  "7d",
        Author: "Alice",
    }

    // Execute
    report := handleRangeReportWithOptions(opts)

    // Verify
    assert.Equal(t, 1, len(report.ByAuthor))
    assert.Equal(t, "Alice", report.ByAuthor[0].Name)
}
```

#### Test 2: 複数作者フィルタリング
```go
func TestAuthorFilter_MultipleAuthors(t *testing.T) {
    opts := ReportOptions{
        Since:  "7d",
        Author: "Alice,Bob,Claude",
    }

    report := handleRangeReportWithOptions(opts)

    assert.LessOrEqual(t, len(report.ByAuthor), 3)
    for _, author := range report.ByAuthor {
        assert.Contains(t, []string{"Alice", "Bob", "Claude"}, author.Name)
    }
}
```

#### Test 3: スペーストリム
```go
func TestAuthorFilter_WithWhitespace(t *testing.T) {
    opts := ReportOptions{
        Since:  "7d",
        Author: " Alice , Bob , Claude ",  // スペースあり
    }

    report := handleRangeReportWithOptions(opts)

    // スペースがトリムされて正しくフィルタリングされる
    assert.LessOrEqual(t, len(report.ByAuthor), 3)
}
```

#### Test 4: 該当作者なし
```go
func TestAuthorFilter_NoMatch(t *testing.T) {
    opts := ReportOptions{
        Since:  "7d",
        Author: "NonExistentAuthor",
    }

    report := handleRangeReportWithOptions(opts)

    // 空のレポート
    assert.Equal(t, 0, len(report.ByAuthor))
    assert.Equal(t, 0, report.Summary.TotalLines)
}
```

#### Test 5: 後方互換性
```go
func TestAuthorFilter_WithoutFlag(t *testing.T) {
    opts := ReportOptions{
        Since:  "7d",
        Author: "",  // フラグなし
    }

    report := handleRangeReportWithOptions(opts)

    // すべての作者が表示される（既存動作）
    assert.Greater(t, len(report.ByAuthor), 0)
}
```

---

### 統合テスト

#### シナリオ1: 実際のGitリポジトリでのテスト
```bash
#!/bin/bash
# test_author_filter.sh

set -e

# Setup
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

git init
git config user.name "TestUser"
git config user.email "test@example.com"

# AICT初期化
aict init
aict setup-hooks

# コミット1: TestUserの人間コード
echo "line 1" > test.go
git add test.go
git commit -m "Human commit 1"
aict checkpoint --author "TestUser"
aict commit

# コミット2: Claudeのコード（シミュレート）
echo "line 2" >> test.go
git add test.go
git commit -m "AI commit 1"
aict checkpoint --author "Claude"
aict commit

# コミット3: 別の人（Bob）
echo "line 3" >> test.go
git add test.go
git commit -m "Bob commit 1"
aict checkpoint --author "Bob"
aict commit

# Test 1: TestUserのみ
echo "Test 1: Filter by TestUser"
REPORT=$(aict report --since 1d --author "TestUser")
echo "$REPORT" | grep -q "TestUser"
echo "$REPORT" | grep -qv "Bob"
echo "$REPORT" | grep -qv "Claude"
echo "✓ Test 1 passed"

# Test 2: TestUserとClaude
echo "Test 2: Filter by TestUser,Claude"
REPORT=$(aict report --since 1d --author "TestUser,Claude")
echo "$REPORT" | grep -q "TestUser"
echo "$REPORT" | grep -q "Claude"
echo "$REPORT" | grep -qv "Bob"
echo "✓ Test 2 passed"

# Test 3: 存在しない人
echo "Test 3: Filter by nonexistent author"
REPORT=$(aict report --since 1d --author "NonExistent")
[[ $(echo "$REPORT" | grep -c "Total Lines:.*0") -eq 1 ]]
echo "✓ Test 3 passed"

# Cleanup
cd -
rm -rf "$TEMP_DIR"

echo "All tests passed!"
```

---

## リスク管理

### リスク1: パフォーマンス劣化

**リスク**: 大量のコミット・作者がある場合、フィルタリングが遅くなる可能性

**対策**:
- フィルタリングは`map[string]bool`を使用してO(1)検索
- 既存のループ構造を変更せず、条件追加のみ
- ベンチマークテストで5%以内の劣化を確認

**測定方法**:
```bash
# フィルタなし
time aict report --since 1y > /dev/null

# フィルタあり
time aict report --since 1y --author "Alice,Bob" > /dev/null
```

---

### リスク2: 複雑なフィルタリング要求

**リスク**: ユーザーが正規表現やワイルドカードを期待する可能性

**対策**:
- v1.2.0では完全一致のみサポート
- ドキュメントに明記
- 将来的な拡張は別issue化

---

### リスク3: 大文字小文字の混乱

**リスク**: "Claude" vs "claude" で一致しない

**対策**:
- 既存のAuthorshipLogに記録されている名前と完全一致
- エラーメッセージではなく空のレポート（ユーザーが気づきやすい）
- ドキュメントに大文字小文字を区別することを明記

---

## 成功基準

### 必須条件
- [ ] `--author`フラグが正常に動作する
- [ ] 単一作者・複数作者のフィルタリングが可能
- [ ] すべてのユニットテストが通過
- [ ] 統合テストが通過
- [ ] 既存テスト（`test_since_option.sh`など）が通過
- [ ] ドキュメントが更新されている

### 品質条件
- [ ] パフォーマンス劣化 <5%
- [ ] コードカバレッジ ≥80%
- [ ] `--author`なしの動作が完全に維持されている

### ユーザビリティ条件
- [ ] エラーメッセージが分かりやすい
- [ ] USAGE.mdに実用的な使用例がある
- [ ] JSON出力フォーマットが一貫している

---

## リリース計画

### v1.2.0 リリースノート（案）

```markdown
# AI Code Tracker v1.2.0

## New Features

### 🎯 Author Filtering (`--author` flag)

Filter reports by specific authors to focus on individual or team contributions.

**Examples**:
```bash
# Personal AI usage
aict report --since 7d --author "Your Name,Claude"

# Pair programming
aict report --since 1d --author "Alice,Bob"

# Team subset
aict report --since 2w --author "Alice,Bob,Carol,Claude"
```

**Benefits**:
- Focus on personal AI contribution rates
- Analyze pair programming effectiveness
- Compare team member productivity

See [USAGE.md](docs/USAGE.md) for detailed examples.

## Improvements

- Performance: Author filtering with O(1) lookup
- Documentation: Enhanced usage examples for team scenarios

## Bug Fixes

None

## Breaking Changes

None (fully backward compatible)
```

---

## 次のステップ

### v1.3.0候補機能
1. `--exclude-author` フラグ
2. `--personal` モード（環境変数ベース）
3. 正規表現フィルタリング（`--author-regex`）

### v1.4.0候補機能
1. ファイルパスフィルタリング（`--file`）
2. ブランチフィルタリング（`--branch`）
3. レポートテンプレートカスタマイズ

---

## 参考資料

- [USAGE.md](USAGE.md) - 現在の使い方ドキュメント
- [USECASE.md](USECASE.md) - ユースケース集
- [handlers_range.go](../cmd/aict/handlers_range.go) - 実装対象ファイル
- [types.go](../internal/tracker/types.go) - データ構造定義

---

## 変更履歴

| 日付 | バージョン | 変更内容 | 担当 |
|------|----------|---------|------|
| 2025-01-17 | 1.0 | 初版作成 | Claude |

---

## 承認

- [ ] 要件レビュー
- [ ] 設計レビュー
- [ ] 実装開始承認

**レビュアー**: _____________________

**承認日**: _____________________
