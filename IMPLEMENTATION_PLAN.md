# SPEC.md準拠への移行実装計画

**作成日**: 2025-11-23
**対象バージョン**: v0.7.0 (SPEC.md準拠版)
**現在バージョン**: v0.6.1

---

## 📋 実装概要

現在の軽量実装（`.ai_code_tracking/` + JSONL形式）からSPEC.md完全準拠（`.git/aict/` + Git notes統合）への段階的移行を行います。3つのフェーズで実装を進め、既存機能との互換性を保ちながら新機能を追加します。

### 主要な変更点

| 項目 | 現在実装 | SPEC.md準拠 |
|-----|---------|------------|
| **保存場所** | `.ai_code_tracking/` | `.git/aict/` |
| **チェックポイント形式** | JSONL（軽量） | JSON配列（詳細） |
| **行範囲追跡** | なし | `[[start, end]]` 形式 |
| **Authorship Log** | `AIEditNote` (簡易) | `AuthorshipLog` (完全) |
| **Git notes ref** | `refs/notes/aict` | `refs/aict/authorship/` |
| **コマンド** | `aict track` | `aict checkpoint` |

---

## Phase 1: 基盤整備（データ構造・保存場所の変更）

### 1.1 新しい型定義の追加

**ファイル**: `internal/tracker/types.go`

#### 追加する型定義

```go
// SPEC.md § 主要な型定義

// AuthorType represents the type of code author
type AuthorType string

const (
    AuthorTypeHuman AuthorType = "human"
    AuthorTypeAI    AuthorType = "ai"
)

// Change represents file-level changes with line ranges
type Change struct {
    Added   int     `json:"added"`
    Deleted int     `json:"deleted"`
    Lines   [][]int `json:"lines"` // [[start, end], [single], ...]
}

// CheckpointV2 represents a development checkpoint (SPEC.md準拠)
type CheckpointV2 struct {
    Timestamp time.Time          `json:"timestamp"`
    Author    string             `json:"author"`
    Type      AuthorType         `json:"type"`
    Metadata  map[string]string  `json:"metadata,omitempty"`
    Changes   map[string]Change  `json:"changes"` // filepath -> Change
}

// AuthorshipLog represents commit-level authorship information
type AuthorshipLog struct {
    Version   string                `json:"version"`
    Commit    string                `json:"commit"`
    Timestamp time.Time             `json:"timestamp"`
    Files     map[string]FileInfo   `json:"files"`
}

// FileInfo contains author information for a single file
type FileInfo struct {
    Authors []AuthorInfo `json:"authors"`
}

// AuthorInfo represents a single author's contribution to a file
type AuthorInfo struct {
    Name     string            `json:"name"`
    Type     AuthorType        `json:"type"`
    Lines    [][]int           `json:"lines"`    // [[start, end], ...]
    Metadata map[string]string `json:"metadata,omitempty"`
}

// Report represents generated code generation report
type Report struct {
    Range        string             `json:"range,omitempty"`
    Branch       string             `json:"branch,omitempty"`
    Commits      int                `json:"commits,omitempty"`
    Period       *Period            `json:"period,omitempty"`
    Summary      SummaryStats       `json:"summary"`
    ByFile       []FileStats        `json:"by_file,omitempty"`
    ByAuthor     []AuthorStats      `json:"by_author,omitempty"`
}

type Period struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type SummaryStats struct {
    TotalLines    int     `json:"total_lines"`
    AILines       int     `json:"ai_lines"`
    HumanLines    int     `json:"human_lines"`
    AIPercentage  float64 `json:"ai_percentage"`
}

type FileStats struct {
    Path         string  `json:"path"`
    TotalLines   int     `json:"total_lines"`
    AILines      int     `json:"ai_lines"`
    HumanLines   int     `json:"human_lines"`
    AIPercentage float64 `json:"ai_percentage"`
}

type AuthorStats struct {
    Name       string     `json:"name"`
    Type       AuthorType `json:"type"`
    Lines      int        `json:"lines"`
    Percentage float64    `json:"percentage"`
    Commits    int        `json:"commits,omitempty"`
}
```

#### 既存の型との共存

- **既存**: `CheckpointRecord` (JSONL形式) - 後方互換性のため保持
- **新規**: `CheckpointV2` (SPEC.md準拠) - 新しいチェックポイントシステム用

---

### 1.2 保存場所の変更

#### 新規作成: `internal/storage/aict_storage.go`

`.git/aict/` ディレクトリ配下の操作を管理する新しいストレージレイヤー。

```go
package storage

import (
    "encoding/json"
    "os"
    "path/filepath"
)

// AIctStorage manages .git/aict/ directory
type AIctStorage struct {
    gitDir string // .git/aict/
}

// NewAIctStorage creates a new AIctStorage instance
func NewAIctStorage() (*AIctStorage, error) {
    // 1. .git ディレクトリを検出
    gitDir, err := findGitDir()
    if err != nil {
        return nil, err
    }

    // 2. .git/aict/ を作成
    aictDir := filepath.Join(gitDir, "aict")
    if err := os.MkdirAll(aictDir, 0755); err != nil {
        return nil, err
    }

    return &AIctStorage{gitDir: aictDir}, nil
}

// SaveCheckpoint appends a checkpoint to latest.json
func (s *AIctStorage) SaveCheckpoint(cp *CheckpointV2) error {
    // .git/aict/checkpoints/latest.json に追記（配列形式）
    checkpointsDir := filepath.Join(s.gitDir, "checkpoints")
    os.MkdirAll(checkpointsDir, 0755)

    checkpointsFile := filepath.Join(checkpointsDir, "latest.json")

    // 既存のチェックポイントを読み込み
    checkpoints, _ := s.LoadCheckpoints()

    // 新しいチェックポイントを追加
    checkpoints = append(checkpoints, cp)

    // JSON配列として保存
    data, err := json.MarshalIndent(checkpoints, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(checkpointsFile, data, 0644)
}

// LoadCheckpoints loads all checkpoints from latest.json
func (s *AIctStorage) LoadCheckpoints() ([]*CheckpointV2, error) {
    checkpointsFile := filepath.Join(s.gitDir, "checkpoints", "latest.json")

    data, err := os.ReadFile(checkpointsFile)
    if err != nil {
        if os.IsNotExist(err) {
            return []*CheckpointV2{}, nil
        }
        return nil, err
    }

    var checkpoints []*CheckpointV2
    if err := json.Unmarshal(data, &checkpoints); err != nil {
        return nil, err
    }

    return checkpoints, nil
}

// ClearCheckpoints removes all checkpoints
func (s *AIctStorage) ClearCheckpoints() error {
    checkpointsFile := filepath.Join(s.gitDir, "checkpoints", "latest.json")
    return os.Remove(checkpointsFile)
}

// SaveConfig saves config.json
func (s *AIctStorage) SaveConfig(cfg *Config) error {
    configFile := filepath.Join(s.gitDir, "config.json")
    data, err := json.MarshalIndent(cfg, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(configFile, data, 0644)
}

// LoadConfig loads config.json
func (s *AIctStorage) LoadConfig() (*Config, error) {
    configFile := filepath.Join(s.gitDir, "config.json")
    data, err := os.ReadFile(configFile)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// findGitDir finds .git directory from current directory
func findGitDir() (string, error) {
    dir, err := os.Getwd()
    if err != nil {
        return "", err
    }

    for {
        gitDir := filepath.Join(dir, ".git")
        if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
            return gitDir, nil
        }

        parent := filepath.Dir(dir)
        if parent == dir {
            return "", fmt.Errorf(".git directory not found")
        }
        dir = parent
    }
}
```

#### 変更: `cmd/aict/main.go`

```go
// 変更前
const defaultBaseDir = ".ai_code_tracking"

// 変更後
const defaultBaseDir = ".git/aict"

// 初期化時に既存データから移行
func handleInit() {
    // 新しいストレージを作成
    storage, err := storage.NewAIctStorage()
    if err != nil {
        log.Fatal(err)
    }

    // 既存の .ai_code_tracking/ からデータを移行（存在する場合）
    if err := migrateFromLegacyStorage(); err != nil {
        log.Printf("Warning: migration failed: %v", err)
    }
}
```

---

### 1.3 Git notes形式の統一

#### 新規作成: `internal/authorship/` パッケージ

**ファイル構成**:
- `log.go` - AuthorshipLog構造体と基本操作
- `builder.go` - チェックポイント群→AuthorshipLog変換
- `parser.go` - JSON解析とバリデーション

#### `internal/authorship/log.go`

```go
package authorship

import (
    "encoding/json"
    "time"
)

const AuthorshipLogVersion = "1.0"

// AuthorshipLog represents commit-level authorship information
// SPEC.md § Authorship Log
type AuthorshipLog struct {
    Version   string                `json:"version"`
    Commit    string                `json:"commit"`
    Timestamp time.Time             `json:"timestamp"`
    Files     map[string]FileInfo   `json:"files"`
}

// ToJSON converts AuthorshipLog to JSON bytes
func (l *AuthorshipLog) ToJSON() ([]byte, error) {
    return json.MarshalIndent(l, "", "  ")
}

// FromJSON parses JSON bytes to AuthorshipLog
func FromJSON(data []byte) (*AuthorshipLog, error) {
    var log AuthorshipLog
    if err := json.Unmarshal(data, &log); err != nil {
        return nil, err
    }
    return &log, nil
}
```

#### `internal/authorship/builder.go`

```go
package authorship

import (
    "time"
    "github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// BuildAuthorshipLog converts checkpoints to AuthorshipLog
// SPEC.md § チェックポイント → Authorship Log変換
func BuildAuthorshipLog(checkpoints []*tracker.CheckpointV2, commitHash string) (*AuthorshipLog, error) {
    log := &AuthorshipLog{
        Version:   AuthorshipLogVersion,
        Commit:    commitHash,
        Timestamp: time.Now(),
        Files:     make(map[string]FileInfo),
    }

    // ファイルごとに作成者情報を集約
    for _, cp := range checkpoints {
        for filepath, change := range cp.Changes {
            fileInfo, exists := log.Files[filepath]
            if !exists {
                fileInfo = FileInfo{Authors: []AuthorInfo{}}
            }

            // 同じ作成者が既に存在するか確認
            authorIdx := -1
            for i, author := range fileInfo.Authors {
                if author.Name == cp.Author && author.Type == cp.Type {
                    authorIdx = i
                    break
                }
            }

            if authorIdx >= 0 {
                // 既存の作成者に行範囲を追加
                fileInfo.Authors[authorIdx].Lines = append(
                    fileInfo.Authors[authorIdx].Lines,
                    change.Lines...,
                )
            } else {
                // 新しい作成者を追加
                fileInfo.Authors = append(fileInfo.Authors, AuthorInfo{
                    Name:     cp.Author,
                    Type:     cp.Type,
                    Lines:    change.Lines,
                    Metadata: cp.Metadata,
                })
            }

            log.Files[filepath] = fileInfo
        }
    }

    return log, nil
}
```

#### 変更: `internal/gitnotes/notes.go`

```go
package gitnotes

const (
    // 変更前
    // DefaultNotesRef = "refs/notes/aict"

    // 変更後: SPEC.md § Git Notes統合
    AuthorshipNotesRef = "refs/aict/authorship"
)

// AddAuthorshipLog adds an AuthorshipLog to Git notes
func (nm *NotesManager) AddAuthorshipLog(log *authorship.AuthorshipLog) error {
    data, err := log.ToJSON()
    if err != nil {
        return err
    }

    // refs/aict/authorship/{commit-sha} に保存
    cmd := exec.Command("git", "notes", "--ref="+AuthorshipNotesRef, "add",
        "-m", string(data), log.Commit)
    return cmd.Run()
}

// GetAuthorshipLog retrieves an AuthorshipLog from Git notes
func (nm *NotesManager) GetAuthorshipLog(commitHash string) (*authorship.AuthorshipLog, error) {
    cmd := exec.Command("git", "notes", "--ref="+AuthorshipNotesRef, "show", commitHash)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    return authorship.FromJSON(output)
}
```

---

### Phase 1 完了条件

- [ ] `internal/tracker/types.go` に新しい型定義が追加され、既存コードと共存
- [ ] `internal/storage/aict_storage.go` が実装され、`.git/aict/` ディレクトリ操作が可能
- [ ] `internal/authorship/` パッケージが作成され、AuthorshipLog操作が可能
- [ ] `internal/gitnotes/notes.go` が `refs/aict/authorship/` 形式に対応
- [ ] 既存の `.ai_code_tracking/` からの移行ロジックが実装
- [ ] 単体テストが作成され、すべてパス

---

## Phase 2: コマンド実装

### 2.1 `aict checkpoint` 完全実装

**新規作成**: `cmd/aict/handlers_checkpoint.go`

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"

    "github.com/y-hirakaw/ai-code-tracker/internal/storage"
    "github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func handleCheckpoint() {
    fs := flag.NewFlagSet("checkpoint", flag.ExitOnError)
    author := fs.String("author", "", "作成者名（デフォルト: config.default_author）")
    model := fs.String("model", "", "AIモデル名（AIエージェントの場合）")
    message := fs.String("message", "", "メモ（オプション）")
    fs.Parse(os.Args[2:])

    // ストレージを初期化
    store, err := storage.NewAIctStorage()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // 設定を読み込み
    config, err := store.LoadConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    // 作成者名を決定
    authorName := *author
    if authorName == "" {
        authorName = config.DefaultAuthor
    }

    // 作成者タイプを判定
    authorType := tracker.AuthorTypeHuman
    if isAIAgent(authorName, config.AIAgents) {
        authorType = tracker.AuthorTypeAI
    }

    // 前回のチェックポイント以降の変更を検出
    changes, err := detectChanges()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error detecting changes: %v\n", err)
        os.Exit(1)
    }

    // チェックポイントを作成
    checkpoint := &tracker.CheckpointV2{
        Timestamp: time.Now(),
        Author:    authorName,
        Type:      authorType,
        Metadata:  make(map[string]string),
        Changes:   changes,
    }

    // メタデータを追加
    if *model != "" {
        checkpoint.Metadata["model"] = *model
    }
    if *message != "" {
        checkpoint.Metadata["message"] = *message
    }

    // チェックポイントを保存
    if err := store.SaveCheckpoint(checkpoint); err != nil {
        fmt.Fprintf(os.Stderr, "Error saving checkpoint: %v\n", err)
        os.Exit(1)
    }

    // 変更行数をカウント
    totalAdded := 0
    for _, change := range changes {
        totalAdded += change.Added
    }

    fmt.Printf("✓ Checkpoint created (%s, %d lines added)\n", authorName, totalAdded)
}

// detectChanges detects file changes since last checkpoint
func detectChanges() (map[string]tracker.Change, error) {
    // git diff --unified=0 --numstat で変更を取得
    cmd := exec.Command("git", "diff", "--unified=0", "--numstat", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    changes := make(map[string]tracker.Change)

    // 各ファイルの変更を解析
    for _, line := range strings.Split(string(output), "\n") {
        if line == "" {
            continue
        }

        parts := strings.Fields(line)
        if len(parts) < 3 {
            continue
        }

        added, _ := strconv.Atoi(parts[0])
        deleted, _ := strconv.Atoi(parts[1])
        filepath := parts[2]

        // 行範囲を取得
        lineRanges, err := getLineRanges(filepath)
        if err != nil {
            continue
        }

        changes[filepath] = tracker.Change{
            Added:   added,
            Deleted: deleted,
            Lines:   lineRanges,
        }
    }

    return changes, nil
}

// getLineRanges extracts line ranges from git diff output
func getLineRanges(filepath string) ([][]int, error) {
    cmd := exec.Command("git", "diff", "--unified=0", "HEAD", "--", filepath)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var ranges [][]int

    // @@ -1,2 +3,4 @@ 形式の行範囲を解析
    for _, line := range strings.Split(string(output), "\n") {
        if !strings.HasPrefix(line, "@@") {
            continue
        }

        // +3,4 の部分を抽出
        parts := strings.Split(line, "@@")
        if len(parts) < 2 {
            continue
        }

        rangePart := strings.TrimSpace(parts[1])
        plusIdx := strings.Index(rangePart, "+")
        if plusIdx == -1 {
            continue
        }

        rangeStr := strings.Fields(rangePart[plusIdx+1:])[0]
        rangeNums := strings.Split(rangeStr, ",")

        if len(rangeNums) == 1 {
            // 単一行: +10
            lineNum, _ := strconv.Atoi(rangeNums[0])
            ranges = append(ranges, []int{lineNum})
        } else if len(rangeNums) == 2 {
            // 範囲: +10,5 (10行目から5行)
            start, _ := strconv.Atoi(rangeNums[0])
            count, _ := strconv.Atoi(rangeNums[1])
            ranges = append(ranges, []int{start, start + count - 1})
        }
    }

    return ranges, nil
}

// isAIAgent checks if author is an AI agent
func isAIAgent(author string, aiAgents []string) bool {
    for _, agent := range aiAgents {
        if author == agent {
            return true
        }
    }
    return false
}
```

---

### 2.2 `aict commit` 新規実装

**新規作成**: `cmd/aict/handlers_commit.go`

```go
package main

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/y-hirakaw/ai-code-tracker/internal/authorship"
    "github.com/y-hirakaw/ai-code-tracker/internal/gitnotes"
    "github.com/y-hirakaw/ai-code-tracker/internal/storage"
)

func handleCommit() {
    // ストレージを初期化
    store, err := storage.NewAIctStorage()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // チェックポイントを読み込み
    checkpoints, err := store.LoadCheckpoints()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading checkpoints: %v\n", err)
        os.Exit(1)
    }

    if len(checkpoints) == 0 {
        // チェックポイントがない場合は何もしない
        return
    }

    // 最新のコミットハッシュを取得
    commitHash, err := getLatestCommitHash()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting commit hash: %v\n", err)
        os.Exit(1)
    }

    // チェックポイント群をAuthorship Logに変換
    log, err := authorship.BuildAuthorshipLog(checkpoints, commitHash)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error building authorship log: %v\n", err)
        os.Exit(1)
    }

    // Git notesに保存
    nm := gitnotes.NewNotesManager()
    if err := nm.AddAuthorshipLog(log); err != nil {
        fmt.Fprintf(os.Stderr, "Error saving authorship log: %v\n", err)
        os.Exit(1)
    }

    // チェックポイントをクリア
    if err := store.ClearCheckpoints(); err != nil {
        fmt.Fprintf(os.Stderr, "Warning: failed to clear checkpoints: %v\n", err)
    }

    fmt.Println("✓ Authorship log created")
}

// getLatestCommitHash retrieves the latest commit hash
func getLatestCommitHash() (string, error) {
    cmd := exec.Command("git", "rev-parse", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}
```

**Git hook統合**: `.git/hooks/post-commit`

```bash
#!/bin/sh
# Post-commit hook to generate Authorship Log

# aict commit を実行
aict commit

exit 0
```

---

### 2.3 `aict report --range` 追加

**変更**: `cmd/aict/handlers.go`

```go
type ReportOptions struct {
    // 既存オプション
    Since       string
    From        string
    To          string
    Last        string
    Branch      string
    BranchRegex string
    BranchPattern string
    AllBranches bool
    Format      string

    // 新規追加
    Range       string // "origin/main..feature-branch"
}

func handleReport() {
    // ... 既存のフラグ定義 ...

    // --range フラグを追加
    rangeFlag := flag.String("range", "", "Commit range (e.g., origin/main..HEAD)")

    flag.Parse()

    opts := &ReportOptions{
        // ... 既存のオプション ...
        Range: *rangeFlag,
    }

    // --range が指定された場合
    if opts.Range != "" {
        handleRangeReport(opts)
        return
    }

    // 既存のレポート処理
    // ...
}

func handleRangeReport(opts *ReportOptions) {
    // 1. git log <range> でコミット一覧を取得
    commits, err := getCommitsInRange(opts.Range)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // 2. 各コミットのAuthorship Logを読み込み
    nm := gitnotes.NewNotesManager()

    totalAI := 0
    totalHuman := 0
    byAuthor := make(map[string]*AuthorStats)
    byFile := make(map[string]*FileStats)

    for _, commitHash := range commits {
        log, err := nm.GetAuthorshipLog(commitHash)
        if err != nil {
            // Authorship Logがないコミットはスキップ
            continue
        }

        // 3. 集計
        for filepath, fileInfo := range log.Files {
            for _, author := range fileInfo.Authors {
                lineCount := countLines(author.Lines)

                // 作成者別集計
                stats, exists := byAuthor[author.Name]
                if !exists {
                    stats = &AuthorStats{
                        Name: author.Name,
                        Type: author.Type,
                    }
                    byAuthor[author.Name] = stats
                }
                stats.Lines += lineCount
                stats.Commits++

                // ファイル別集計
                fileStats, exists := byFile[filepath]
                if !exists {
                    fileStats = &FileStats{Path: filepath}
                    byFile[filepath] = fileStats
                }
                fileStats.TotalLines += lineCount

                if author.Type == tracker.AuthorTypeAI {
                    totalAI += lineCount
                    fileStats.AILines += lineCount
                } else {
                    totalHuman += lineCount
                    fileStats.HumanLines += lineCount
                }
            }
        }
    }

    // 4. レポート生成
    report := &Report{
        Range:   opts.Range,
        Commits: len(commits),
        Summary: SummaryStats{
            TotalLines:   totalAI + totalHuman,
            AILines:      totalAI,
            HumanLines:   totalHuman,
            AIPercentage: float64(totalAI) / float64(totalAI+totalHuman) * 100,
        },
    }

    // ByAuthor, ByFile を追加
    for _, stats := range byAuthor {
        stats.Percentage = float64(stats.Lines) / float64(totalAI+totalHuman) * 100
        report.ByAuthor = append(report.ByAuthor, *stats)
    }

    for _, stats := range byFile {
        stats.AIPercentage = float64(stats.AILines) / float64(stats.TotalLines) * 100
        report.ByFile = append(report.ByFile, *stats)
    }

    // 5. フォーマットに応じて出力
    formatReport(report, opts.Format)
}

// getCommitsInRange retrieves commit hashes in the given range
func getCommitsInRange(rangeSpec string) ([]string, error) {
    cmd := exec.Command("git", "log", "--format=%H", rangeSpec)
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var commits []string
    for _, line := range strings.Split(string(output), "\n") {
        line = strings.TrimSpace(line)
        if line != "" {
            commits = append(commits, line)
        }
    }

    return commits, nil
}

// countLines counts total lines from line ranges
func countLines(ranges [][]int) int {
    total := 0
    for _, r := range ranges {
        if len(r) == 1 {
            total++
        } else if len(r) == 2 {
            total += r[1] - r[0] + 1
        }
    }
    return total
}
```

---

### 2.4 `aict sync` 新規実装

**新規作成**: `cmd/aict/handlers_sync.go`

```go
package main

import (
    "fmt"
    "os"
    "os/exec"
)

func handleSync() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: aict sync [push|fetch]")
        os.Exit(1)
    }

    subcommand := os.Args[2]

    switch subcommand {
    case "push":
        handleSyncPush()
    case "fetch":
        handleSyncFetch()
    default:
        fmt.Printf("Unknown subcommand: %s\n", subcommand)
        fmt.Println("Usage: aict sync [push|fetch]")
        os.Exit(1)
    }
}

func handleSyncPush() {
    // refs/aict/authorship/* をリモートにpush
    cmd := exec.Command("git", "push", "origin", "refs/aict/authorship/*:refs/aict/authorship/*")
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error pushing authorship logs: %v\n%s\n", err, output)
        os.Exit(1)
    }

    fmt.Println("✓ Authorship logs pushed to remote")
}

func handleSyncFetch() {
    // リモートから refs/aict/authorship/* をfetch
    cmd := exec.Command("git", "fetch", "origin", "refs/aict/authorship/*:refs/aict/authorship/*")
    output, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching authorship logs: %v\n%s\n", err, output)
        os.Exit(1)
    }

    fmt.Println("✓ Authorship logs fetched from remote")
}
```

**Git hook統合（オプション）**:

`.git/hooks/pre-push`:
```bash
#!/bin/sh
# Pre-push hook to sync authorship logs

git push origin "refs/aict/authorship/*:refs/aict/authorship/*"

exit 0
```

`.git/hooks/post-merge`:
```bash
#!/bin/sh
# Post-merge hook to sync authorship logs

aict sync fetch

exit 0
```

---

### 2.5 `cmd/aict/main.go` の変更

```go
func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    command := os.Args[1]

    switch command {
    case "init":
        handleInit()
    case "track":
        handleTrack()
    case "checkpoint":  // 新規追加
        handleCheckpoint()
    case "commit":      // 新規追加
        handleCommit()
    case "report":
        handleReport()
    case "sync":        // 新規追加
        handleSync()
    case "snapshot":
        handleSnapshot()
    case "reset":
        handleReset()
    case "mark-ai-edit":
        handleMarkAIEdit()
    case "setup-hooks":
        handleSetupHooks()
    case "config":
        handleConfig()
    case "version":
        handleVersion()
    default:
        fmt.Printf("Unknown command: %s\n", command)
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("AI Code Tracker (aict) - Track AI vs Human code contributions")
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  aict init                           Initialize tracking in repository")
    fmt.Println("  aict checkpoint [options]           Record development checkpoint (SPEC.md)")
    fmt.Println("  aict commit                         Generate Authorship Log (auto-run by hook)")
    fmt.Println("  aict track [options]                Record checkpoint (legacy)")
    fmt.Println("  aict report [options]               Display code generation report")
    fmt.Println("  aict sync [push|fetch]              Sync authorship logs with remote")
    fmt.Println("  aict snapshot                       Analyze entire codebase with git blame")
    fmt.Println("  aict reset                          Reset all metrics")
    fmt.Println("  aict mark-ai-edit [options]         Mark AI edit in git notes")
    fmt.Println("  aict setup-hooks                    Setup Git and Claude Code hooks")
    fmt.Println("  aict config                         Edit configuration")
    fmt.Println("  aict version                        Show version")
    fmt.Println()
    fmt.Println("Checkpoint Options:")
    fmt.Println("  --author <name>                     Author name (default: config.default_author)")
    fmt.Println("  --model <model>                     AI model name (for AI agents)")
    fmt.Println("  --message <msg>                     Optional message")
    fmt.Println()
    fmt.Println("Report Options:")
    fmt.Println("  --range <base>..<head>              Commit range (e.g., origin/main..HEAD)")
    fmt.Println("  --branch <name>                     Filter by branch name")
    fmt.Println("  --last <period>                     Relative period (7d, week, month)")
    fmt.Println("  --format <format>                   Output format (table, graph, json)")
    fmt.Println()
}
```

---

### Phase 2 完了条件

- [ ] `aict checkpoint` がSPEC.md仕様通りに動作
- [ ] `aict commit` がAuthorship Logを生成し、Git notesに保存
- [ ] `aict report --range` がコミット範囲レポートを表示
- [ ] `aict sync push/fetch` がGit notesを同期
- [ ] Git hooksが統合され、自動化が動作
- [ ] すべてのコマンドに単体テストが存在

---

## Phase 3: 統合とテスト

### 3.1 既存機能との互換性

#### 共存戦略

- **`aict track`** (既存): JSONL形式で記録継続、`.ai_code_tracking/` を使用
- **`aict checkpoint`** (新): CheckpointV2形式で記録、`.git/aict/` を使用
- **`aict report`**: 両形式のデータを読み込み可能に

#### 変更: `internal/tracker/checkpoint_jsonl.go`

```go
// レガシーデータの読み込みサポート
func LoadLegacyCheckpoints(baseDir string) ([]CheckpointRecord, error) {
    // .ai_code_tracking/checkpoints.jsonl を読み込み
}

// 新形式への変換
func ConvertToCheckpointV2(record *CheckpointRecord) (*CheckpointV2, error) {
    // CheckpointRecord → CheckpointV2 変換
}
```

#### 変更: `cmd/aict/handlers.go` (report)

```go
func handleReport() {
    // 1. 新形式のデータを読み込み (.git/aict/)
    store, _ := storage.NewAIctStorage()
    newCheckpoints, _ := store.LoadCheckpoints()

    // 2. レガシー形式のデータを読み込み (.ai_code_tracking/)
    legacyRecords, _ := tracker.LoadLegacyCheckpoints(".ai_code_tracking")

    // 3. 統合してレポート生成
    // ...
}
```

---

### 3.2 データ移行

#### 新規作成: `cmd/aict/handlers_migrate.go`

```go
package main

import (
    "fmt"
    "os"

    "github.com/y-hirakaw/ai-code-tracker/internal/storage"
    "github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func handleMigrate() {
    fmt.Println("Starting migration from .ai_code_tracking/ to .git/aict/...")

    // 1. 新しいストレージを作成
    store, err := storage.NewAIctStorage()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    // 2. レガシーデータを読み込み
    legacyRecords, err := tracker.LoadLegacyCheckpoints(".ai_code_tracking")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading legacy data: %v\n", err)
        os.Exit(1)
    }

    // 3. 新形式に変換
    fmt.Printf("Converting %d legacy checkpoints...\n", len(legacyRecords))
    for _, record := range legacyRecords {
        checkpoint, err := tracker.ConvertToCheckpointV2(&record)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to convert checkpoint: %v\n", err)
            continue
        }

        if err := store.SaveCheckpoint(checkpoint); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to save checkpoint: %v\n", err)
        }
    }

    // 4. 設定ファイルをコピー
    legacyConfig, err := tracker.LoadConfig(".ai_code_tracking")
    if err == nil {
        if err := store.SaveConfig(legacyConfig); err != nil {
            fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
        }
    }

    fmt.Println("✓ Migration completed")
    fmt.Println()
    fmt.Println("Next steps:")
    fmt.Println("  1. Review .git/aict/ directory")
    fmt.Println("  2. Test with 'aict report'")
    fmt.Println("  3. Remove .ai_code_tracking/ if everything works")
}
```

#### 新規作成: `scripts/migrate.sh`

```bash
#!/bin/bash
# Migration script from .ai_code_tracking/ to .git/aict/

set -e

echo "AI Code Tracker Migration Script"
echo "================================="
echo

# Check if .ai_code_tracking exists
if [ ! -d ".ai_code_tracking" ]; then
    echo "Error: .ai_code_tracking/ directory not found"
    exit 1
fi

# Check if .git exists
if [ ! -d ".git" ]; then
    echo "Error: Not a git repository"
    exit 1
fi

# Run migration
echo "Running aict migrate..."
aict migrate

echo
echo "Migration completed!"
echo
read -p "Do you want to remove .ai_code_tracking/? (y/N) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf .ai_code_tracking
    echo "✓ .ai_code_tracking/ removed"
else
    echo "Keeping .ai_code_tracking/ for now"
fi
```

---

### 3.3 テスト実装

#### 新規作成: `internal/authorship/builder_test.go`

```go
package authorship

import (
    "testing"
    "time"

    "github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestBuildAuthorshipLog(t *testing.T) {
    checkpoints := []*tracker.CheckpointV2{
        {
            Timestamp: time.Now(),
            Author:    "Alice",
            Type:      tracker.AuthorTypeHuman,
            Changes: map[string]tracker.Change{
                "main.go": {
                    Added:   10,
                    Deleted: 2,
                    Lines:   [][]int{{1, 10}},
                },
            },
        },
        {
            Timestamp: time.Now(),
            Author:    "Claude Code",
            Type:      tracker.AuthorTypeAI,
            Metadata:  map[string]string{"model": "claude-sonnet-4"},
            Changes: map[string]tracker.Change{
                "main.go": {
                    Added:   50,
                    Deleted: 5,
                    Lines:   [][]int{{21, 70}},
                },
            },
        },
    }

    log, err := BuildAuthorshipLog(checkpoints, "abc123")
    if err != nil {
        t.Fatalf("BuildAuthorshipLog failed: %v", err)
    }

    if log.Version != AuthorshipLogVersion {
        t.Errorf("Expected version %s, got %s", AuthorshipLogVersion, log.Version)
    }

    if log.Commit != "abc123" {
        t.Errorf("Expected commit abc123, got %s", log.Commit)
    }

    fileInfo, exists := log.Files["main.go"]
    if !exists {
        t.Fatal("main.go not found in files")
    }

    if len(fileInfo.Authors) != 2 {
        t.Errorf("Expected 2 authors, got %d", len(fileInfo.Authors))
    }

    // Check Alice's contribution
    found := false
    for _, author := range fileInfo.Authors {
        if author.Name == "Alice" && author.Type == tracker.AuthorTypeHuman {
            found = true
            if len(author.Lines) != 1 || author.Lines[0][0] != 1 || author.Lines[0][1] != 10 {
                t.Errorf("Alice's line ranges incorrect: %v", author.Lines)
            }
        }
    }
    if !found {
        t.Error("Alice not found in authors")
    }
}
```

#### 新規作成: `internal/storage/aict_storage_test.go`

```go
package storage

import (
    "os"
    "path/filepath"
    "testing"
    "time"

    "github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestAIctStorage(t *testing.T) {
    // Create temporary .git directory
    tmpDir := t.TempDir()
    gitDir := filepath.Join(tmpDir, ".git")
    os.MkdirAll(gitDir, 0755)

    // Change to temp directory
    oldDir, _ := os.Getwd()
    os.Chdir(tmpDir)
    defer os.Chdir(oldDir)

    // Create storage
    store, err := NewAIctStorage()
    if err != nil {
        t.Fatalf("NewAIctStorage failed: %v", err)
    }

    // Test SaveCheckpoint
    checkpoint := &tracker.CheckpointV2{
        Timestamp: time.Now(),
        Author:    "Test",
        Type:      tracker.AuthorTypeHuman,
        Changes: map[string]tracker.Change{
            "test.go": {Added: 10, Deleted: 2, Lines: [][]int{{1, 10}}},
        },
    }

    if err := store.SaveCheckpoint(checkpoint); err != nil {
        t.Fatalf("SaveCheckpoint failed: %v", err)
    }

    // Test LoadCheckpoints
    checkpoints, err := store.LoadCheckpoints()
    if err != nil {
        t.Fatalf("LoadCheckpoints failed: %v", err)
    }

    if len(checkpoints) != 1 {
        t.Errorf("Expected 1 checkpoint, got %d", len(checkpoints))
    }

    if checkpoints[0].Author != "Test" {
        t.Errorf("Expected author Test, got %s", checkpoints[0].Author)
    }

    // Test ClearCheckpoints
    if err := store.ClearCheckpoints(); err != nil {
        t.Fatalf("ClearCheckpoints failed: %v", err)
    }

    checkpoints, _ = store.LoadCheckpoints()
    if len(checkpoints) != 0 {
        t.Errorf("Expected 0 checkpoints after clear, got %d", len(checkpoints))
    }
}
```

#### 新規作成: `cmd/aict/handlers_checkpoint_test.go`

```go
package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
)

func TestHandleCheckpoint(t *testing.T) {
    // Setup test git repository
    tmpDir := t.TempDir()
    os.Chdir(tmpDir)

    exec.Command("git", "init").Run()
    exec.Command("git", "config", "user.name", "Test").Run()
    exec.Command("git", "config", "user.email", "test@example.com").Run()

    // Initialize aict
    os.Args = []string{"aict", "init"}
    handleInit()

    // Create a test file
    os.WriteFile("test.go", []byte("package main\n"), 0644)
    exec.Command("git", "add", "test.go").Run()
    exec.Command("git", "commit", "-m", "initial").Run()

    // Modify the file
    os.WriteFile("test.go", []byte("package main\n\nfunc main() {}\n"), 0644)

    // Run checkpoint
    os.Args = []string{"aict", "checkpoint", "--author", "Alice"}
    handleCheckpoint()

    // Verify checkpoint was created
    checkpointsFile := filepath.Join(".git", "aict", "checkpoints", "latest.json")
    if _, err := os.Stat(checkpointsFile); os.IsNotExist(err) {
        t.Error("Checkpoint file was not created")
    }
}
```

#### 統合テスト

```go
// integration_test.go
package main

import (
    "os"
    "os/exec"
    "testing"
)

func TestFullWorkflow(t *testing.T) {
    // 1. Setup
    tmpDir := t.TempDir()
    os.Chdir(tmpDir)
    exec.Command("git", "init").Run()
    exec.Command("git", "config", "user.name", "Test").Run()
    exec.Command("git", "config", "user.email", "test@example.com").Run()

    // 2. aict init
    os.Args = []string{"aict", "init"}
    handleInit()

    // 3. Create file and checkpoint (human)
    os.WriteFile("main.go", []byte("package main\n"), 0644)
    os.Args = []string{"aict", "checkpoint"}
    handleCheckpoint()

    // 4. Modify and checkpoint (AI)
    os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)
    os.Args = []string{"aict", "checkpoint", "--author", "Claude Code", "--model", "claude-sonnet-4"}
    handleCheckpoint()

    // 5. Git commit (triggers aict commit)
    exec.Command("git", "add", "main.go").Run()
    exec.Command("git", "commit", "-m", "test").Run()
    os.Args = []string{"aict", "commit"}
    handleCommit()

    // 6. Verify Git notes
    cmd := exec.Command("git", "notes", "--ref=refs/aict/authorship", "show", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        t.Fatalf("Git notes not found: %v", err)
    }

    if len(output) == 0 {
        t.Error("Authorship log is empty")
    }

    // 7. Generate report
    os.Args = []string{"aict", "report"}
    handleReport()

    // 8. Test sync (would need remote setup)
    // os.Args = []string{"aict", "sync", "push"}
    // handleSync()
}
```

---

### 3.4 ドキュメント更新

#### 変更: `README.md`

```markdown
# AI Code Tracker (AICT) v0.7.0

SPEC.md準拠版。詳細な行範囲追跡とGit notes統合。

## 新機能 (v0.7.0)

- ✅ `aict checkpoint` - SPEC.md準拠のチェックポイント記録
- ✅ `aict commit` - Authorship Log自動生成
- ✅ `aict report --range` - コミット範囲レポート
- ✅ `aict sync push/fetch` - Git notes同期
- ✅ 行範囲追跡 (`[[start, end]]` 形式)
- ✅ `.git/aict/` への移行

## 基本的な使い方

### 1. 初期化

```bash
$ cd your-project
$ aict init
✓ Initialized aict in /path/to/your-project
```

### 2. チェックポイント記録

```bash
# 人間の作業開始
$ aict checkpoint
✓ Checkpoint created

# AIでコード生成
# (Claude Codeなどで開発)

# AI作業完了を記録
$ aict checkpoint --author "Claude Code" --model "claude-sonnet-4"
✓ Checkpoint created (Claude Code, 50 lines added)
```

### 3. コミット（自動でAuthorship Log生成）

```bash
$ git add .
$ git commit -m "Add authentication feature"
✓ Authorship log created
```

### 4. レポート表示

```bash
# 最新のコミット
$ aict report

# コミット範囲
$ aict report --range origin/main..HEAD

# 期間指定
$ aict report --last 7d
```

### 5. リモートと同期

```bash
$ aict sync push
```

## 移行ガイド

v0.6.x以前からの移行:

```bash
$ aict migrate
$ rm -rf .ai_code_tracking  # 確認後
```
```

#### 変更: `CLAUDE.md`

```markdown
## 実装状況 (v0.7.0)

### SPEC.md準拠機能

- ✅ チェックポイントシステム (`.git/aict/checkpoints/`)
- ✅ Authorship Log (Git notes: `refs/aict/authorship/`)
- ✅ 行範囲追跡 (`[[start, end]]` 形式)
- ✅ `aict checkpoint` 完全実装
- ✅ `aict commit` 自動生成
- ✅ `aict report --range` コミット範囲レポート
- ✅ `aict sync` Git notes同期

### データ構造

- `CheckpointV2`: SPEC.md準拠の詳細チェックポイント
- `AuthorshipLog`: Git notes形式の作成者情報
- `Change`: ファイル別の変更と行範囲
```

---

### Phase 3 完了条件

- [ ] 既存の `aict track` と新しい `aict checkpoint` が共存
- [ ] `aict migrate` コマンドが動作し、データ移行が可能
- [ ] すべての新機能に単体テストが存在し、パス
- [ ] 統合テストが成功
- [ ] `README.md`, `CLAUDE.md` が更新
- [ ] バージョンが v0.7.0 に更新

---

## ファイル変更サマリー

### 新規作成（13ファイル）

1. `internal/storage/aict_storage.go` - `.git/aict/` 管理
2. `internal/authorship/log.go` - AuthorshipLog構造体
3. `internal/authorship/builder.go` - チェックポイント→ログ変換
4. `internal/authorship/parser.go` - JSON解析
5. `cmd/aict/handlers_checkpoint.go` - checkpoint コマンド
6. `cmd/aict/handlers_commit.go` - commit コマンド
7. `cmd/aict/handlers_sync.go` - sync コマンド
8. `cmd/aict/handlers_migrate.go` - migrate コマンド
9. `internal/storage/aict_storage_test.go` - テスト
10. `internal/authorship/builder_test.go` - テスト
11. `cmd/aict/handlers_checkpoint_test.go` - テスト
12. `scripts/migrate.sh` - データ移行スクリプト
13. `.git/hooks/post-commit.aict` - Git hookテンプレート

### 変更（8ファイル）

1. `internal/tracker/types.go` - 新型定義追加
2. `internal/gitnotes/notes.go` - refs形式変更、AuthorshipLog対応
3. `cmd/aict/main.go` - 新コマンド追加、baseDir変更
4. `cmd/aict/handlers.go` - `--range` オプション追加
5. `internal/tracker/checkpoint_jsonl.go` - 共存ロジック
6. `.ai_code_tracking/hooks/post-commit` - `aict commit` 呼び出し
7. `README.md` - 新機能ドキュメント
8. `CLAUDE.md` - 実装状況更新

---

## 推奨実装スケジュール

### Week 1: Phase 1（基盤整備）
- **Day 1-2**: 型定義追加 (`types.go`, `authorship/`)
- **Day 3-4**: `.git/aict/` ストレージ実装 (`aict_storage.go`)
- **Day 5**: Git notes形式変更、基礎テスト

### Week 2: Phase 2（コマンド実装）
- **Day 1-2**: `aict checkpoint` 実装とテスト
- **Day 3**: `aict commit` 実装とテスト
- **Day 4**: `aict report --range` 実装
- **Day 5**: `aict sync` 実装、統合テスト

### Week 3: Phase 3（統合とテスト）
- **Day 1-2**: 既存機能との互換性確保、共存ロジック
- **Day 3**: データ移行スクリプト (`migrate`)
- **Day 4**: 統合テスト、エンドツーエンドテスト
- **Day 5**: ドキュメント更新、v0.7.0リリース準備

---

## リスクと対策

### リスク1: 既存ユーザーのデータ損失
**対策**:
- 移行スクリプトを慎重に実装
- `.ai_code_tracking/` を自動削除せず、ユーザー判断に委ねる
- 両形式のデータ読み込みをサポート

### リスク2: Git notes操作の失敗
**対策**:
- Git notes操作の前にバックアップを推奨
- エラーハンドリングを強化
- dry-runモードの提供

### リスク3: 行範囲追跡の複雑さ
**対策**:
- `git diff --unified=0` の出力パースを慎重に実装
- エッジケース（削除のみ、リネーム等）への対応
- 十分なテストカバレッジ

### リスク4: パフォーマンス問題
**対策**:
- 大規模リポジトリでのテスト
- 必要に応じてキャッシュ機構導入
- チェックポイント数の上限設定

---

## 成功基準

1. **機能完全性**: SPEC.mdで定義されたすべてのコマンドが動作
2. **データ整合性**: チェックポイント→Authorship Log変換が正確
3. **後方互換性**: 既存の `aict track` ユーザーが影響を受けない
4. **テストカバレッジ**: 80%以上
5. **ドキュメント**: すべての新機能が文書化
6. **パフォーマンス**: 既存実装と同等以上

---

## 次のステップ

1. このImplementation Planをレビュー
2. Phase 1から実装開始
3. 各フェーズ完了後にレビューとテスト
4. v0.7.0としてリリース
