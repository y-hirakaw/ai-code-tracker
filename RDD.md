# AI Code Tracker - Requirements and Design Document (RDD)

## 1. プロジェクト概要

### 1.1 目的
AI（Claude Code等）と人間が書いたコードの割合を正確に追跡し、設定可能な目標AIコード生成率の達成を支援するツールを開発する。

### 1.2 主要機能
- Claude Codeのフックと連携した自動的なコード変更追跡
- Git post-commitフックによる自動分析
- 高速なバイナリ形式でのデータ保存
- リアルタイムの進捗表示と目標達成率の可視化

### 1.3 技術スタック
- 実装言語: Go
- データ形式: JSON
- 連携: Claude Code hooks, Git hooks
- 対象ファイル: 設定可能（任意のプログラミング言語に対応）

## 2. アーキテクチャ設計

### 2.1 システム構成

```
ai-code-tracker/
├── cmd/
│   └── aict/              # メインCLIツール
│       └── main.go
├── internal/
│   ├── tracker/           # コア追跡ロジック
│   │   ├── checkpoint.go  # チェックポイント管理
│   │   ├── analyzer.go    # 分析ロジック
│   │   └── types.go       # 型定義
│   ├── storage/           # データ永続化
│   │   ├── json.go        # JSONシリアライゼーション
│   │   └── metrics.go     # メトリクス管理
│   └── git/               # Git連携
│       └── diff.go        # Git diff処理
├── hooks/
│   ├── aict-pre-tool-use.sh   # Claude Code Pre hook
│   ├── aict-post-tool-use.sh  # Claude Code Post hook
│   └── post-commit             # Git post-commit hook
├── .claude-code/
│   └── config.json        # Claude Code設定
└── .gitignore
```

### 2.2 データフロー

```
1. Claude Code Pre-Tool Hook
   ↓
   [Human状態記録] → checkpoint_<timestamp>_human.json
                     author: "John Doe"
   ↓
2. Claude Codeによる編集
   ↓
3. Claude Code Post-Tool Hook
   ↓
   [AI状態記録] → checkpoint_<timestamp>_ai.json
                   author: "Claude Code"
                   ai_tool: "claude_code"
   ↓
4. 人間による編集
   ↓
5. Git Commit
   ↓
6. Git Post-Commit Hook
   ↓
   [分析実行] → metrics.json更新
               各authorごとの統計を集計
```

## 3. データ仕様

### 3.1 設定ファイル形式

```json
// .ai_tracking/config.json
{
  "target_ai_percentage": 80.0,
  "tracking_dir": ".ai_tracking",
  "include_extensions": [
    ".go", ".rs", ".py", ".js", ".ts", ".jsx", ".tsx",
    ".java", ".kt", ".swift", ".c", ".cpp", ".cs", ".rb",
    ".php", ".scala", ".r", ".m", ".mm", ".vue", ".dart"
  ],
  "exclude_patterns": [
    "*_test.go",
    "*.generated.*",
    "vendor/*",
    "node_modules/*"
  ],
  "authors": {
    "human": "${git.user.name}",  // 自動的にgitのuser.nameを使用
    "ai_tools": {
      "claude_code": "Claude Code",
      "cursor": "Cursor",
      "copilot": "GitHub Copilot"
    },
    "default_ai_tool": "claude_code"
  }
}
```

**設定項目の説明:**
- `target_ai_percentage`: AI生成コードの目標割合
- `include_extensions`: 計測対象とする拡張子（コードファイルのみを対象とし、ドキュメントを除外）
- `exclude_patterns`: 除外するファイルパターン（テストファイルや自動生成ファイルを除外可能）
- `authors.human`: 人間の作者名（${git.user.name}で自動取得、または固定値）
- `authors.ai_tools`: 使用可能なAIツールの定義
- `authors.default_ai_tool`: デフォルトのAIツール

### 3.2 チェックポイントJSON形式

```json
{
  "timestamp": "2025-01-30T10:30:00Z",
  "author": "John Doe",
  "author_type": "human",
  "files": {
    "src/main.go": 45,
    "src/utils.go": 23,
    "pkg/handler/auth.go": 67
  },
  "summary": {
    "total_lines": 135,
    "files_count": 3
  }
}
```

または

```json
{
  "timestamp": "2025-01-30T10:35:00Z",
  "author": "Claude Code",
  "author_type": "ai",
  "ai_tool": "claude_code",
  "files": {
    "src/main.go": 120,
    "src/auth.go": 85
  },
  "summary": {
    "total_lines": 205,
    "files_count": 2
  }
}
```

### 3.3 メトリクスJSON形式

```json
{
  "config": {
    "target_ai_percentage": 95.0
  },
  "total_stats": {
    "by_author": {
      "John Doe": {
        "lines": 1234,
        "percentage": 5.0,
        "type": "human"
      },
      "Claude Code": {
        "lines": 20456,
        "percentage": 82.8,
        "type": "ai",
        "ai_tool": "claude_code"
      },
      "Cursor": {
        "lines": 3000,
        "percentage": 12.2,
        "type": "ai",
        "ai_tool": "cursor"
      }
    },
    "summary": {
      "human_lines": 1234,
      "ai_lines": 23456,
      "total_lines": 24690,
      "ai_percentage": 95.0
    }
  },
  "daily_stats": {
    "2025-01-30": {
      "by_author": {
        "John Doe": {"lines": 100},
        "Claude Code": {"lines": 1900}
      },
      "commits": 5
    }
  },
  "commits": [
    {
      "hash": "abc123",
      "timestamp": "2025-01-30T10:30:00Z",
      "stats": {
        "John Doe": {"lines": 50, "type": "human"},
        "Claude Code": {"lines": 950, "type": "ai"}
      },
      "ai_percentage": 95.0,
      "target_met": true,
      "target_value": 95.0
    }
  ]
}
```

## 4. 実装仕様

### 4.1 CLIコマンド

```bash
# 基本コマンド
aict track human           # 人間の作業開始を記録（gitのuser.name使用）
aict track ai              # AIの作業開始を記録（デフォルトAIツール使用）
aict track ai cursor       # 特定のAIツール（Cursor）を指定
aict analyze               # 現在までの変更を分析
aict report                # レポート表示
aict reset                 # チェックポイントをリセット
aict config                # 設定管理

# 内部コマンド（フックから呼ばれる）
aict hook pre-edit         # Claude Code Pre hook用
aict hook post-edit        # Claude Code Post hook用
aict hook post-edit cursor # 特定のAIツールを指定

# 設定コマンド
aict config set target 95              # 目標AI率を95%に設定
aict config set human "John Doe"       # 人間の名前を設定（デフォルト: git user.name）
aict config add ai-tool cursor "Cursor"     # AIツールを追加
aict config get target                 # 現在の目標値を表示
aict config list                       # 全設定を表示
```

### 4.2 フック実装

#### aict-pre-tool-use.sh
```bash
#!/bin/bash
# Claude Codeの編集前に人間の最終状態を記録
/Users/username/git/ai-code-tracker/bin/aict hook pre-edit
```

#### aict-post-tool-use.sh
```bash
#!/bin/bash
# Claude Codeの編集後にAIの状態を記録
/Users/username/git/ai-code-tracker/bin/aict hook post-edit
```

#### Git post-commit hook
```bash
#!/bin/bash
# コミット時に分析を実行
aict analyze --commit
```

### 4.3 主要な型定義

```go
// internal/tracker/types.go
package tracker

import "time"

type AuthorType string

const (
    Human AuthorType = "human"
    AI    AuthorType = "ai"
)

type Config struct {
    TargetAIPercentage float64           `json:"target_ai_percentage"`
    TrackingDir        string            `json:"tracking_dir"`
    IncludeExtensions  []string          `json:"include_extensions"`
    ExcludePatterns    []string          `json:"exclude_patterns"`
    Authors            AuthorsConfig     `json:"authors"`
}

type AuthorsConfig struct {
    Human         string            `json:"human"`
    AITools       map[string]string `json:"ai_tools"`
    DefaultAITool string           `json:"default_ai_tool"`
}

type Checkpoint struct {
    Timestamp  time.Time         `json:"timestamp"`
    Author     string            `json:"author"`      // "y-hirakawa" or "Claude Code"
    AuthorType AuthorType        `json:"author_type"` // "human" or "ai"
    AITool     string            `json:"ai_tool,omitempty"`
    Files      map[string]int32  `json:"files"`
    Summary    CheckpointSummary `json:"summary"`
}

type CheckpointSummary struct {
    TotalLines int32 `json:"total_lines"`
    FilesCount int   `json:"files_count"`
}

type AnalysisResult struct {
    CommitHash    string              `json:"commit_hash,omitempty"`
    Timestamp     time.Time           `json:"timestamp"`
    Stats         map[string]AuthorStats `json:"stats"` // key: author name
    TotalLines    int32               `json:"total_lines"`
    AIPercentage  float64             `json:"ai_percentage"`
    TargetMet     bool                `json:"target_met"`
    TargetValue   float64             `json:"target_value"`
    Sessions      []Session           `json:"sessions"`
}

type AuthorStats struct {
    Lines      int32   `json:"lines"`
    Percentage float64 `json:"percentage"`
    Type       AuthorType `json:"type"`
}

type Session struct {
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Author      string    `json:"author"`
    AuthorType  AuthorType `json:"author_type"`
    LinesAdded  int32     `json:"lines_added"`
    FilesCount  int       `json:"files_count"`
}
```

### 4.4 コア機能の実装方針

#### 設定管理
1. `.ai_tracking/config.json`に設定を保存
2. デフォルト値:
   - target_ai_percentage: 80.0
   - tracking_dir: ".ai_tracking"
   - include_extensions: [主要なプログラミング言語の拡張子]
   - authors.human: "${git.user.name}" (自動取得)
   - authors.default_ai_tool: "claude_code"
3. 設定変更は即座に反映
4. git user.nameは初回実行時に自動取得してキャッシュ

#### チェックポイント記録
1. Git diff --numstat を実行して変更行数を取得
2. 設定された拡張子のファイルのみを対象とする（ドキュメントや設定ファイルを除外し、純粋なコード生成率を測定）
3. JSON形式でシリアライズして保存
4. ファイル名: `checkpoint_<unix_timestamp>_<author>.json`

#### 差分計算
1. 連続する2つのチェックポイント間の差分を計算
2. ファイルごとの行数増加分を著者に帰属
3. 削除行は考慮しない（純粋な追加行数のみ）

#### 分析処理
1. 全チェックポイントを時系列順に読み込み
2. セッションごとの貢献度を計算
3. 累積統計を更新
4. 結果を表示・保存

## 5. UI/UX設計

### 5.1 進捗表示

```
📊 AI CODE GENERATION ANALYSIS
============================================================

AI Generated: [████████████████████████████████░░░░░░░░] 85.3%
Target:       [██████████████████████████████████████░░] 95.0%

📈 Statistics by Author:
   Claude Code:  7,230 lines (72.3%)
   Cursor:       1,300 lines (13.0%)
   John Doe:     1,470 lines (14.7%)
   Total:       10,000 lines

📊 AI Tools Breakdown:
   Claude Code: 84.8% of AI code
   Cursor:      15.2% of AI code

❌ Target not met. Need 9.7% more AI-generated code
   Suggestion: Next 970 lines should be AI-generated

💡 Tip: Change target with 'aict config set target <value>'
```

### 5.2 エラーハンドリング

- Gitリポジトリ外での実行時: 明確なエラーメッセージ
- チェックポイント不足時: 分析に必要なデータ不足を通知
- 破損したデータ: 自動的にスキップして続行

## 6. パフォーマンス要件

- チェックポイント記録: 50ms以下
- 1000ファイルの分析: 200ms以下
- メモリ使用量: 50MB以下
- JSONファイルサイズ: 通常10KB以下/チェックポイント

## 7. テスト計画

### 7.1 単体テスト
- JSONシリアライゼーション
- 差分計算ロジック
- Git diff解析
- ファイルフィルタリング（拡張子、除外パターン）

### 7.2 統合テスト
- Claude Codeフック連携
- Git commit フロー
- 大規模リポジトリでの動作

## 8. セットアップ手順

```bash
# 1. リポジトリのクローン
git clone https://github.com/yourusername/ai-code-tracker.git
cd ai-code-tracker

# 2. ビルド
go build -o bin/aict ./cmd/aict

# 3. パスを通す
export PATH=$PATH:$(pwd)/bin

# 4. フックの設定
# .claude-code/config.json を配置（既に記載の内容）

# 5. Git post-commit hookの設定
cp hooks/post-commit .git/hooks/
chmod +x .git/hooks/post-commit

# 6. 初期化
aict init

# 7. 目標値を設定（デフォルト: 80%）
aict config set target 95
```

## 9. 開発ガイドライン

### コーディング規約
- Go標準のフォーマッティング（gofmt）
- エラーは必ず処理する
- パッケージは機能ごとに分離
- テストカバレッジ80%以上

### コミットメッセージ
```
feat: 新機能追加
fix: バグ修正
docs: ドキュメント更新
refactor: リファクタリング
test: テスト追加・修正
chore: その他の変更
```