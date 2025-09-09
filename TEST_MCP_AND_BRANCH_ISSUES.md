# AI Code Tracker (AICT) - ブランチ連動機能の課題

## 🔀 ブランチ連動機能の課題と解決策

### 現在の課題

#### 課題1: ブランチ情報が記録されない
**問題**: `CheckpointRecord`にブランチ情報が含まれていない
```go
// 現在の構造
CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Commit    string    `json:"commit,omitempty"`
    Added     int       `json:"added"`
    Deleted   int       `json:"deleted"`
    // Branch情報がない
}
```

#### 課題2: ブランチ別のメトリクス分離ができない
- 機能ブランチでのAI利用率
- ブランチ間の開発パターン比較
- マージ時のメトリクス統合

### 解決策の提案

#### 1. データ構造の拡張
```go
CheckpointRecord struct {
    Timestamp time.Time `json:"timestamp"`
    Author    string    `json:"author"`
    Branch    string    `json:"branch"`    // 新規追加
    Commit    string    `json:"commit,omitempty"`
    Added     int       `json:"added"`
    Deleted   int       `json:"deleted"`
}
```

#### 2. ブランチ別レポート機能
```bash
# 提案する新しいコマンド
aict report --branch feature/new-ui
aict report --all-branches
aict report --branch-comparison main..feature/new-ui
```

#### 3. Git統合の活用
- 既存の`GetCurrentBranch()`関数を活用
- チェックポイント記録時にブランチ情報を自動取得
- ブランチ固有のメトリクス計算

### 実装優先度
1. **高優先度**: ブランチ連動機能（新機能として価値が高い）

### ブランチ課題検証テスト
- [ ] 現在のブランチ情報取得テスト
- [ ] 異なるブランチでの動作確認
- [ ] ブランチ切り替え時の動作確認

---

## 📋 次のステップ

1. **ブランチ課題の実装** - データ構造の拡張とレポート機能の追加
2. **バージョンアップ** - ブランチ連動機能実装後の更新

---

**作成日**: 2025年01月26日  
**対象バージョン**: AICT v0.5.3+ (ブランチ連動対応版)