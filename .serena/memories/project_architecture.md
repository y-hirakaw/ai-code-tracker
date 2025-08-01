# プロジェクト構造とアーキテクチャ

## ディレクトリ構造
```
ai-code-tracker/
├── cmd/aict/                  # CLI エントリーポイント
│   ├── main.go               # メイン CLI ロジック
│   ├── handlers.go           # 期間レポートハンドラー (v0.4.0)
│   └── *_test.go            # CLI テスト
├── internal/                 # プライベートパッケージ
│   ├── tracker/             # コアトラッキングロジック
│   │   ├── checkpoint.go    # チェックポイント管理
│   │   ├── checkpoint_jsonl.go # JSONL記録システム  
│   │   ├── analyzer.go      # 分析ロジック
│   │   ├── analyzer_jsonl.go # JSONL分析
│   │   └── types.go         # 型定義
│   ├── period/              # 期間分析 (v0.4.0)
│   │   ├── analyzer.go      # 期間分析ロジック
│   │   ├── formatter.go     # 出力フォーマット
│   │   ├── parser.go        # 時間範囲パーサー
│   │   ├── filter.go        # レコードフィルタリング
│   │   └── types.go         # 期間分析型
│   ├── storage/             # データ永続化
│   │   ├── json.go          # JSON シリアライゼーション
│   │   └── metrics.go       # メトリクス管理
│   ├── git/                 # Git統合
│   │   └── diff.go          # Git diff処理
│   └── templates/           # フックテンプレート
│       └── hooks.go         # 埋め込みフックスクリプト
├── bin/aict                 # CLI実行ファイル
└── .ai_code_tracking/       # 実行時データディレクトリ
    ├── config.json          # トラッキング設定
    ├── checkpoints.jsonl    # 軽量レコード
    └── hooks/               # 生成されたフックスクリプト
```

## 主要コンポーネント

### 1. CLI Layer (cmd/aict/)
- **main.go**: CLI コマンド処理、フラグ解析
- **handlers.go**: 期間レポート機能 (v0.4.0新機能)
- **機能**: init, track, report, setup-hooks, reset, config

### 2. Core Tracking (internal/tracker/)
- **checkpoint.go**: レガシーチェックポイントシステム
- **checkpoint_jsonl.go**: 新しい軽量JSONL記録システム
- **analyzer.go**: チェックポイント分析
- **analyzer_jsonl.go**: JSONL レコード分析
- **types.go**: 共通データ構造

### 3. Period Analysis (internal/period/)
- **analyzer.go**: 時間範囲指定分析
- **formatter.go**: テーブル/グラフ/JSON出力
- **parser.go**: 時間表現解析 (--last 7d, --since "2 weeks ago")
- **filter.go**: 日付範囲フィルタリング

### 4. Data Layer (internal/storage/)
- **json.go**: 汎用JSON永続化
- **metrics.go**: メトリクス特化ストレージ

### 5. Git Integration (internal/git/)
- **diff.go**: Git リポジトリ分析、diff処理、コミット情報取得

### 6. Hook System (internal/templates/)
- **hooks.go**: Claude Code/Git フック用埋め込みスクリプトテンプレート

## データフロー

### 自動トラッキング
1. **Claude Code Pre-Tool Hook** → 人間状態記録
2. **Claude Code 編集実行**
3. **Claude Code Post-Tool Hook** → AI状態記録
4. **人間による追加編集**
5. **Git commit** → post-commitフックでメトリクス更新

### 手動トラッキング
1. `aict track -author human` → 現在状態を人間作成として記録
2. `aict track -author claude` → 現在状態をAI作成として記録

### レポート生成
1. JSONL レコード読み込み
2. 期間フィルタリング（オプション）
3. AI/人間比率計算
4. 指定フォーマットで出力（table/graph/json）

## 設計パターン
- **Factory Pattern**: `New*()` コンストラクタ
- **Strategy Pattern**: 複数の出力フォーマット
- **Repository Pattern**: ストレージ抽象化
- **Template Method**: フック生成システム