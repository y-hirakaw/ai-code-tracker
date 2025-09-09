# AI Code Tracker (AICT) - ブランチ連動機能 開発計画

**プロジェクト**: AI Code Tracker (AICT) v0.5.4+
**作成日**: 2025年01月26日
**対象機能**: ブランチ連動統計機能 + 正規表現マッチング機能

---

## 📋 **プロジェクト概要**

### **現在の状況**
- **総統計**: 640行（AI: 493行77.0%, Human: 147行23.0%）
- **既存レコード**: 92件（ブランチ情報なし）
- **MCPツール対応**: ✅ 完了（v0.5.3）
- **次の課題**: ブランチ別統計とパターンマッチング機能

### **目標機能**
1. **基本ブランチ統計**: 各ブランチでのAI/人間コード比率分析
2. **正規表現マッチング**: パターンベースのブランチグループ分析
3. **レポート機能拡張**: 柔軟な統計表示とフィルタリング
4. **後方互換性**: 既存92件のレコードとの互換性維持

---

## 🎯 **新機能仕様**

### **CLI拡張仕様**

```bash
# 基本ブランチ統計
aict report --branch main                    # 完全一致
aict report --branch feature/ui-improve      # 特定ブランチ

# 正規表現マッチング
aict report --branch-regex "^feature/"       # feature系全て
aict report --branch-regex "(hotfix|bugfix)" # 緊急修正系
aict report --branch-pattern "feature/*"     # glob風パターン

# 複合条件
aict report --branch-regex "^feature/" --last 7d  # パターン+期間
aict report --all-branches --format json          # 全ブランチJSON出力

# グループ統計
aict report --branch-group                   # 定義済みグループ統計
```

### **レポート表示例**

#### **単一ブランチ**
```
Branch Report: feature/ui-improve
================================
Records: 15 (2025-01-20 to 2025-01-25)
Added Lines: 342 (AI: 267, Human: 75)
AI Ratio: 78.1%
Progress: 97.6% (target: 80.0%)
```

#### **正規表現グループ**
```
Branch Pattern Report: "^feature/"
==================================
Matching Branches: feature/ui-improve, feature/api-v2, feature/auth
Total Records: 45
Added Lines: 1,234 (AI: 987, Human: 247)
Group AI Ratio: 79.9%

Per-Branch Breakdown:
  feature/ui-improve:   AI 78.1% (267/342 lines) [15 records]
  feature/api-v2:       AI 82.5% (412/499 lines) [18 records]
  feature/auth:         AI 78.4% (308/393 lines) [12 records]
```

---

## 🏗️ **技術設計**

### **データ構造拡張**

#### **CheckpointRecord構造体**
```go
// Before (current)
type CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Commit    string    `json:"commit,omitempty"`
    Added     int       `json:"added"`
    Deleted   int       `json:"deleted"`
}

// After (v0.5.4+)
type CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Branch    string    `json:"branch,omitempty"`    // 新規追加
    Commit    string    `json:"commit,omitempty"`
    Added     int       `json:"added"`
    Deleted   int       `json:"deleted"`
}
```

#### **互換性メソッド**
```go
// 後方互換性のためのヘルパーメソッド
func (r *CheckpointRecord) GetBranch() string {
    if r.Branch == "" {
        return "main" // デフォルトブランチとして扱う
    }
    return r.Branch
}

func (r *CheckpointRecord) HasBranchInfo() bool {
    return r.Branch != ""
}
```

### **新規パッケージ設計**

#### **internal/branch/パッケージ**
```go
// filter.go - ブランチフィルタリング
type BranchFilter struct {
    Pattern     string
    IsRegex     bool
    IsGlob      bool
}

func (f *BranchFilter) Matches(branch string) bool

// analyzer.go - ブランチ別分析
type BranchAnalyzer struct {
    records []tracker.CheckpointRecord
}

func (a *BranchAnalyzer) AnalyzeByBranch(branch string) (*BranchReport, error)
func (a *BranchAnalyzer) AnalyzeByPattern(pattern string) (*GroupReport, error)

// report.go - レポート生成
type BranchReporter struct{}

func (r *BranchReporter) GenerateBranchReport(analysis *BranchReport) string
func (r *BranchReporter) GenerateGroupReport(analysis *GroupReport) string
```

### **CLI拡張**

#### **新フラグ追加**
```go
// cmd/aict/main.go
var (
    branchFlag      = flag.String("branch", "", "Filter by specific branch")
    branchRegexFlag = flag.String("branch-regex", "", "Filter by branch regex pattern") 
    branchPatternFlag = flag.String("branch-pattern", "", "Filter by branch glob pattern")
    allBranchesFlag = flag.Bool("all-branches", false, "Show all branches summary")
    branchGroupFlag = flag.Bool("branch-group", false, "Show predefined branch groups")
)
```

---

## 📅 **開発スケジュール**

### **Phase 1: 基盤拡張 (Priority: 高)**
**期間**: 1-2日  
**目標**: データ構造とGit統合の強化

#### **1.1 CheckpointRecord拡張**
- [ ] `internal/tracker/types.go` - Branch フィールド追加
- [ ] 後方互換性メソッド実装
- [ ] 既存テストの更新

#### **1.2 Git統合強化**
- [ ] `internal/git/diff.go` - GetCurrentBranch()改良
- [ ] エラーハンドリング強化
- [ ] ブランチ名正規化処理

#### **1.3 レコード記録時のブランチ取得**
- [ ] `internal/tracker/checkpoint_jsonl.go` 修正
- [ ] RecordCheckpoint()にブランチ情報自動取得
- [ ] フックスクリプトとの統合確認

---

### **Phase 2: フィルタリング機能 (Priority: 高)**
**期間**: 2-3日  
**目標**: ブランチベース統計の基盤

#### **2.1 internal/branch パッケージ作成**
- [ ] `filter.go` - BranchFilter実装
- [ ] 正規表現マッチング機能
- [ ] Globパターンサポート（将来拡張）

#### **2.2 レコードフィルタリング**
- [ ] ブランチ別レコード抽出
- [ ] パターンマッチング統合
- [ ] パフォーマンス最適化

#### **2.3 基本テスト**
- [ ] 単体テスト作成
- [ ] パターンマッチングテスト
- [ ] エラーケーステスト

---

### **Phase 3: 分析機能 (Priority: 高)**
**期間**: 2-3日  
**目標**: ブランチ別統計とメトリクス

#### **3.1 ブランチ分析エンジン**
- [ ] `analyzer.go` - BranchAnalyzer実装
- [ ] 単一ブランチ統計
- [ ] グループ統計計算

#### **3.2 メトリクス計算**
- [ ] AI/Human比率計算
- [ ] 進捗率計算
- [ ] 期間別分析統合

#### **3.3 統合テスト**
- [ ] 分析ロジックテスト
- [ ] 実データでの検証
- [ ] パフォーマンステスト

---

### **Phase 4: レポート機能 (Priority: 中)**
**期間**: 1-2日  
**目標**: 見やすい統計表示

#### **4.1 レポートジェネレーター**
- [ ] `report.go` - BranchReporter実装
- [ ] 単一ブランチレポート
- [ ] グループレポートフォーマット

#### **4.2 表示フォーマット**
- [ ] テキスト形式レポート
- [ ] JSON出力サポート（将来拡張）
- [ ] カラー表示（将来拡張）

---

### **Phase 5: CLI統合 (Priority: 高)**
**期間**: 1-2日  
**目標**: ユーザーインターフェース完成

#### **5.1 新フラグ実装**
- [ ] `cmd/aict/main.go` - フラグ追加
- [ ] `cmd/aict/handlers.go` - ハンドラー拡張
- [ ] フラグ競合チェック

#### **5.2 CLI統合**
- [ ] report サブコマンド拡張
- [ ] 既存機能との統合
- [ ] ヘルプメッセージ更新

---

### **Phase 6: テスト・ドキュメント (Priority: 中)**
**期間**: 1-2日  
**目標**: 品質保証と文書化

#### **6.1 総合テスト**
- [ ] 統合テスト拡張
- [ ] エラーケーステスト
- [ ] パフォーマンステスト

#### **6.2 ドキュメント更新**
- [ ] README.md 更新
- [ ] CLAUDE.md 更新  
- [ ] 使用例追加

#### **6.3 バージョンアップ**
- [ ] バージョン番号更新 (v0.5.4)
- [ ] リリースノート作成
- [ ] タグ作成・プッシュ

---

## 🧪 **テスト戦略**

### **テストデータ準備**
```bash
# テスト用ブランチ作成
git checkout -b feature/test-ui
git checkout -b feature/test-api  
git checkout -b hotfix/test-bug
git checkout -b release/v1.0.0
```

### **テストケース**

#### **正規表現マッチング**
- `^feature/` → feature/test-ui, feature/test-api
- `(hotfix|bugfix)` → hotfix/test-bug
- `^release/` → release/v1.0.0
- 無効な正規表現のエラーハンドリング

#### **後方互換性**
- 既存92件レコードの正常処理
- Branchフィールド未設定時のデフォルト動作
- JSON読み込み・書き込みの互換性

#### **統合テスト**
- 異なるブランチでの記録・統計
- 期間フィルターとの組み合わせ
- CLI全オプションの動作確認

---

## 🚨 **リスク管理**

### **技術リスク**
| リスク | 影響度 | 対策 |
|--------|-------|------|
| 既存データ破損 | 高 | バックアップ・段階的移行 |
| パフォーマンス劣化 | 中 | インデックス化・キャッシング |
| 正規表現DoS攻撃 | 低 | タイムアウト・複雑度制限 |
| ブランチ名文字化け | 低 | UTF-8正規化・エスケープ |

### **ミティゲーション**
- **データバックアップ**: 全Phase開始前にcheckpoints.jsonlバックアップ
- **段階リリース**: 各Phaseでの動作確認と修正
- **ロールバック計画**: 問題発生時の復旧手順準備

---

## 📊 **成功指標**

### **機能面**
- ✅ 既存統計機能の完全互換性維持
- ✅ 新規ブランチ統計の正確な表示
- ✅ 正規表現マッチング100%動作
- ✅ 全テストケースPASS

### **パフォーマンス**
- レポート生成時間: <1秒（100レコード以下）
- メモリ使用量: <50MB（通常使用時）
- ディスク使用量増加: <10%（新フィールド追加）

### **ユーザビリティ**
- CLI直感的操作（学習コスト最小）
- エラーメッセージの分かりやすさ
- ヘルプドキュメントの充実

---

## 🔄 **将来拡張計画**

### **Phase 7以降（将来）**
- **ブランチグループ定義**: 設定ファイルでのグループ管理
- **Web UI**: ブラウザベースの統計表示
- **エクスポート機能**: CSV/Excel出力
- **アラート機能**: AI比率閾値アラート
- **Git統合強化**: マージ先推測・プルリクエスト連動

---

## 📋 **開発チェックリスト**

### **Phase 1完了条件**
- [ ] CheckpointRecord.Branch フィールド追加完了
- [ ] 後方互換性テスト全PASS
- [ ] Git統合機能テスト全PASS
- [ ] 既存機能regression無し

### **Phase 2-3完了条件**
- [ ] ブランチフィルタリング機能完成
- [ ] 正規表現マッチング機能完成  
- [ ] 基本統計機能完成
- [ ] 単体テスト全PASS

### **Phase 4-5完了条件**
- [ ] CLI新オプション完成
- [ ] レポート表示機能完成
- [ ] 統合テスト全PASS
- [ ] ドキュメント更新完了

### **最終リリース条件**
- [ ] 全Phase完了
- [ ] パフォーマンス要件満足  
- [ ] セキュリティチェック完了
- [ ] バージョン更新・タグ作成完了

---

**開発計画策定**: 2025年01月26日  
**想定完成**: 2025年02月02日（7日間）  
**対象バージョン**: AICT v0.5.4 (ブランチ連動機能完全版)