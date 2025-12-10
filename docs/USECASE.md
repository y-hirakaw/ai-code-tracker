# AI Code Tracker (AICT) ユースケース集

AI Code Trackerのレポート確認タイミングと実践的な利用パターンをまとめたガイドです。

## レポート確認タイミングの分類

### 1. 日常的な確認（デイリー）

**タイミング**: 毎日の業務終了時、朝のスタンドアップ前

```bash
aict report --since 1d
```

**目的**:
- その日の作業内容の振り返り
- AI/人間の作業配分の把握
- 予想外のAI生成率の検出

**ペルソナ**: 個人開発者、アクティブに開発中のチームメンバー

**活用例**:
```bash
# 毎日の業務終了時に確認
aict report --since 1d

# 出力例:
# 📊 AI Code Generation Report
# Range: since 1d (3 commits)
#
# Summary:
#   Total Lines:        45
#   🤖 AI Generated:    30 (66.7%)
#   👤 Human Written:   15 (33.3%)
```

---

### 2. スプリント管理（ウィークリー/バイウィークリー）

**タイミング**: スプリントレビュー、振り返りミーティング前

```bash
# 1週間スプリント
aict report --since 1w

# 2週間スプリント
aict report --since 2w
```

**目的**:
- スプリント期間中の生産性分析
- AIアシスタント活用効果の測定
- チーム内での共有と改善議論

**ペルソナ**: スクラムマスター、テックリード、開発チーム全体

**活用例**:
```bash
# スプリントレビュー前に2週間分のレポートを生成
aict report --since 2w

# JSONフォーマットでエクスポートしてチームに共有
aict report --since 2w --format json > sprint-report.json
```

---

### 3. プルリクエスト作成時

**タイミング**: PR作成前、レビュー依頼前

```bash
aict report --range origin/main..HEAD
```

**目的**:
- PRに含まれる変更のオーサーシップ情報を添付
- レビュアーへのコンテキスト提供
- AI生成コードの品質確認促進

**ペルソナ**: PR作成者、コードレビュアー

**活用例**:
```bash
# フィーチャーブランチのレポート生成
aict report --range origin/main..HEAD

# PR説明文に含めるレポート生成
aict report --range origin/main..HEAD > pr-authorship.txt
```

**GitHub Actions統合例**:
```yaml
# .github/workflows/pr-report.yml
name: AICT PR Report

on:
  pull_request:
    types: [opened]

jobs:
  aict-report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          fetch-depth: 0  # 全履歴取得

      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'

      - name: Install AICT
        run: go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest

      - name: Generate Report
        run: |
          aict report --range origin/${{ github.base_ref }}..HEAD > pr-report.txt
          gh pr comment ${{ github.event.number }} --body "$(cat pr-report.txt)"
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### 4. マイルストーン/リリース前

**タイミング**: バージョンリリース前、四半期末、プロジェクト完了時

```bash
# 過去1ヶ月
aict report --since 1m

# 特定バージョンタグ間
aict report --range v1.0.0..v2.0.0
```

**目的**:
- 長期的な開発トレンド分析
- リリースノートへのメトリクス追加
- ステークホルダーへの報告資料作成

**ペルソナ**: プロダクトマネージャー、エンジニアリングマネージャー

**活用例**:
```bash
# リリース前の1ヶ月間のレポート
aict report --since 1m

# タグ間のレポートをリリースノートに含める
aict report --range v1.0.0..v2.0.0 > release-metrics.txt

# Git tagに自動追加
git tag -a v1.2.0 -m "$(cat <<EOF
Release v1.2.0

$(aict report --range v1.1.0..HEAD)
EOF
)"
```

---

### 5. プロジェクト評価・監査

**タイミング**: 月次レポート、四半期レビュー、年次評価

```bash
# 四半期レビュー（3ヶ月）
aict report --since 3m --format json > quarterly-report.json

# 年次レビュー（1年）
aict report --since 1y --format json > annual-report.json
```

**目的**:
- AI活用ROIの定量評価
- 開発プロセスの改善点特定
- 予算計画・ツール投資判断

**ペルソナ**: CTO、エンジニアリングディレクター、経営層

**活用例**:
```bash
# 四半期レポートをJSON形式でエクスポート
aict report --since 3m --format json > Q1-2025-report.json

# 年次レポートの生成
aict report --since 1y > annual-ai-usage.txt
```

**分析スクリプト例**:
```python
# analyze-quarterly.py
import json

with open('Q1-2025-report.json', 'r') as f:
    data = json.load(f)

ai_percentage = data['summary']['ai_percentage']
total_lines = data['summary']['total_lines']

print(f"AI生成率: {ai_percentage:.1f}%")
print(f"総行数: {total_lines:,} 行")

# AI生成率が目標値を達成しているか確認
target = 80.0
if ai_percentage >= target:
    print(f"✅ 目標達成 (目標: {target}%)")
else:
    print(f"⚠️ 目標未達 (目標: {target}%, 差分: {target - ai_percentage:.1f}%)")
```

---

### 6. トラブルシューティング・品質調査

**タイミング**: バグ発見時、パフォーマンス問題発生時、品質懸念時

```bash
# 問題発生期間のコードオーサーシップ確認
aict report --since 2w

# 特定ブランチの調査
aict report --range feature/problematic-feature
```

**目的**:
- 問題の原因箇所（AI生成 vs 人間作成）の特定
- AI生成コードの品質パターン分析
- レビュープロセスの改善提案

**ペルソナ**: QAエンジニア、シニアエンジニア、アーキテクト

**活用例**:
```bash
# バグ発生期間のコードを調査
aict report --since 2w

# 特定コミット範囲のオーサーシップ確認
aict report --range abc123..def456

# 詳細メトリクスで作業量を確認
aict report --since 2w
# → 詳細メトリクスが自動表示され、削除行数も含めた作業量が確認できる
```

---

### 7. チーム共有・透明性確保

**タイミング**: 定期的なチーム全体ミーティング、1on1

```bash
aict report --since 1w
```

**目的**:
- AIアシスタント活用状況の共有
- ベストプラクティスの共有
- チーム内での学び合い促進

**ペルソナ**: 全チームメンバー

**活用例**:
```bash
# 週次ミーティングでチーム全体の状況を共有
aict report --since 1w

# 個別の1on1でAI活用状況をディスカッション
aict report --since 2w
```

---

## 利用パターンまとめ

| タイミング | 頻度 | コマンド例 | 主な利用者 |
|----------|------|-----------|----------|
| デイリー振り返り | 毎日 | `--since 1d` | 個人開発者 |
| スプリントレビュー | 1-2週間 | `--since 1w/2w` | チーム全体 |
| PR作成 | PR毎 | `--range origin/main..HEAD` | PR作成者 |
| リリース前 | 月次 | `--since 1m` | PM/EM |
| 四半期評価 | 3ヶ月 | `--since 3m --format json` | 経営層 |
| 問題調査 | 必要時 | `--since 2w` | QA/シニアEng |

---

## 推奨される自動化タイミング

現在の自動計測（チェックポイント記録）に加えて、以下のレポート自動生成が有用です。

### 1. 自動週次レポート

**cron jobでの自動実行**:
```bash
# 毎週金曜 17:00に週次レポート生成
0 17 * * 5 cd /path/to/project && aict report --since 1w > weekly-report.txt
```

**シェルスクリプト例**:
```bash
#!/bin/bash
# weekly-report.sh

REPORT_DIR="$HOME/aict-reports"
mkdir -p "$REPORT_DIR"

REPORT_FILE="$REPORT_DIR/weekly-$(date +%Y-%m-%d).txt"

cd /path/to/your-project
aict report --since 1w > "$REPORT_FILE"

echo "Weekly report generated: $REPORT_FILE"

# オプション: Slackに通知
# curl -X POST -H 'Content-type: application/json' \
#   --data "{\"text\":\"週次レポートが生成されました: $(cat $REPORT_FILE)\"}" \
#   $SLACK_WEBHOOK_URL
```

### 2. PR作成時の自動レポート添付

**GitHub Actions - Pull Request作成時**:
```yaml
name: AICT PR Report

on:
  pull_request:
    types: [opened]

jobs:
  aict-report:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
        with:
          fetch-depth: 0  # 全履歴取得

      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'

      - name: Install AICT
        run: go install github.com/y-hirakaw/ai-code-tracker/cmd/aict@latest

      - name: Initialize AICT
        run: aict init

      - name: Generate Report
        id: report
        run: |
          REPORT=$(aict report --range origin/${{ github.base_ref }}..HEAD)
          echo "report<<EOF" >> $GITHUB_OUTPUT
          echo "$REPORT" >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Comment PR
        uses: actions/github-script@v6
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '## 📊 AI Code Generation Report\n\n```\n${{ steps.report.outputs.report }}\n```'
            })
```

### 3. リリースノートへの自動追加

**タグ作成時に前回リリースからのレポート生成**:
```bash
# タグ作成スクリプト
#!/bin/bash
# create-release.sh

if [ -z "$1" ]; then
  echo "Usage: $0 <new-version>"
  exit 1
fi

NEW_VERSION=$1
PREV_VERSION=$(git describe --tags --abbrev=0)

# レポート生成
REPORT=$(aict report --range $PREV_VERSION..HEAD)

# タグ作成
git tag -a "$NEW_VERSION" -m "$(cat <<EOF
Release $NEW_VERSION

📊 AI Code Generation Report ($PREV_VERSION..$NEW_VERSION)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
$REPORT
EOF
)"

echo "Created tag: $NEW_VERSION"
echo "Push with: git push origin $NEW_VERSION"
```

### 4. コミット時の自動チェック

**Git pre-commit hookでAI生成率をチェック**:
```bash
#!/bin/bash
# .git/hooks/pre-commit

# AI生成率の閾値（例: 50%以上）
THRESHOLD=50.0

# 現在のブランチのレポート生成
REPORT=$(aict report --range origin/main..HEAD --format json)

# AI生成率を取得
AI_PERCENTAGE=$(echo "$REPORT" | jq -r '.summary.ai_percentage')

# 閾値チェック
if (( $(echo "$AI_PERCENTAGE < $THRESHOLD" | bc -l) )); then
  echo "⚠️ Warning: AI生成率が閾値未満です"
  echo "   現在: ${AI_PERCENTAGE}%"
  echo "   閾値: ${THRESHOLD}%"
  echo ""
  echo "続行しますか? (y/n)"
  read -r response
  if [[ ! "$response" =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

exit 0
```

---

## ベストプラクティス

### 1. 定期的な確認習慣をつける

```bash
# ~/.bashrc や ~/.zshrc に追加
alias aict-daily='aict report --since 1d'
alias aict-weekly='aict report --since 1w'
alias aict-pr='aict report --range origin/main..HEAD'
```

### 2. JSON出力を活用する

```bash
# データ分析用のエクスポート
aict report --since 1m --format json > monthly-$(date +%Y-%m).json

# jq を使った集計
aict report --since 1w --format json | jq '.summary'
```

### 3. チームでの共有

```bash
# Markdown形式のレポート生成
cat <<EOF > WEEKLY_REPORT.md
# Weekly AI Code Generation Report

$(date +%Y-%m-%d)

$(aict report --since 1w)
EOF

git add WEEKLY_REPORT.md
git commit -m "docs: Add weekly AICT report"
```

### 4. 目標値の設定と追跡

```bash
# 目標AI生成率の確認スクリプト
#!/bin/bash
TARGET=80.0

ACTUAL=$(aict report --since 1w --format json | jq -r '.summary.ai_percentage')

echo "目標: ${TARGET}%"
echo "実績: ${ACTUAL}%"

if (( $(echo "$ACTUAL >= $TARGET" | bc -l) )); then
  echo "✅ 目標達成！"
else
  DIFF=$(echo "$TARGET - $ACTUAL" | bc -l)
  echo "⚠️ 目標まであと ${DIFF}%"
fi
```

---

これらのユースケースを参考に、プロジェクトの開発フローに合わせてAI Code Trackerを活用してください。
