package blame

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/storage"
	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// BlameInfo は1行のblame情報を表す
type BlameInfo struct {
	// LineNumber は行番号
	LineNumber int
	// Author は作成者（人間またはAI）
	Author string
	// Date は最終変更日時
	Date time.Time
	// IsAI はAIによる変更かどうか
	IsAI bool
	// Model はAIモデル名（AIの場合のみ）
	Model string
	// CommitHash はコミットハッシュ
	CommitHash string
	// Content は行の内容
	Content string
}

// FileBlameResult はファイル全体のblame結果を表す
type FileBlameResult struct {
	// FilePath はファイルパス
	FilePath string
	// Lines は各行のblame情報
	Lines []BlameInfo
	// Statistics は統計情報
	Statistics BlameStatistics
}

// BlameStatistics はblame統計情報を表す
type BlameStatistics struct {
	// TotalLines は総行数
	TotalLines int
	// AILines はAIによる行数
	AILines int
	// HumanLines は人間による行数
	HumanLines int
	// AIPercentage はAIの貢献率
	AIPercentage float64
	// HumanPercentage は人間の貢献率
	HumanPercentage float64
	// TopAIModel は最も使用されたAIモデル
	TopAIModel string
	// TopHumanAuthor は最も貢献した人間の作成者
	TopHumanAuthor string
}

// Blamer は拡張blame機能を提供する
type Blamer struct {
	storage *storage.Storage
	gitRepo string
}

// NewBlamer は新しいBlamerインスタンスを作成する
func NewBlamer(storage *storage.Storage, gitRepo string) *Blamer {
	return &Blamer{
		storage: storage,
		gitRepo: gitRepo,
	}
}

// BlameFile はファイルの拡張blame情報を取得する
func (b *Blamer) BlameFile(filePath string) (*FileBlameResult, error) {
	// Git blameを実行
	gitBlameLines, err := b.getGitBlame(filePath)
	if err != nil {
		return nil, fmt.Errorf("Git blameの取得に失敗しました: %w", err)
	}

	// トラッキング情報を取得
	events, err := b.storage.ReadEvents()
	if err != nil {
		return nil, fmt.Errorf("トラッキング情報の取得に失敗しました: %w", err)
	}

	// Blame情報とトラッキング情報を結合
	blameLines, err := b.combineBlameWithTracking(gitBlameLines, events, filePath)
	if err != nil {
		return nil, fmt.Errorf("Blame情報の結合に失敗しました: %w", err)
	}

	// 統計情報を計算
	stats := b.calculateStatistics(blameLines)

	return &FileBlameResult{
		FilePath:   filePath,
		Lines:      blameLines,
		Statistics: stats,
	}, nil
}

// getGitBlame はGit blameを実行して基本情報を取得する
func (b *Blamer) getGitBlame(filePath string) ([]GitBlameLine, error) {
	// git blame --porcelain でより詳細な情報を取得
	cmd := exec.Command("git", "blame", "--porcelain", filePath)
	cmd.Dir = b.gitRepo

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blameコマンドの実行に失敗しました: %w", err)
	}

	return b.parseGitBlameOutput(string(output))
}

// GitBlameLine はGit blameの1行分の情報を表す
type GitBlameLine struct {
	CommitHash string
	Author     string
	Date       time.Time
	LineNumber int
	Content    string
}

// parseGitBlameOutput はGit blameの出力をパースする
func (b *Blamer) parseGitBlameOutput(output string) ([]GitBlameLine, error) {
	if strings.TrimSpace(output) == "" {
		return []GitBlameLine{}, nil
	}
	
	lines := strings.Split(output, "\n")
	var result []GitBlameLine
	var currentCommit GitBlameLine
	
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		
		// 空行をスキップ
		if line == "" {
			i++
			continue
		}

		// コミットハッシュ行を検出（40文字のハッシュで始まり、スペースが含まれる）
		if len(line) >= 40 && strings.Contains(line, " ") && !strings.HasPrefix(line, "\t") {
			parts := strings.Fields(line)
			if len(parts) >= 3 && len(parts[0]) >= 40 {
				// 新しいブロックの開始
				currentCommit = GitBlameLine{}
				currentCommit.CommitHash = parts[0]
				
				// 行番号を取得（3番目の要素、または2番目の要素）
				var lineNumStr string
				if len(parts) >= 3 {
					lineNumStr = parts[2]
				} else if len(parts) >= 2 {
					lineNumStr = parts[1]
				}
				
				if lineNum, err := strconv.Atoi(lineNumStr); err == nil {
					currentCommit.LineNumber = lineNum
				}
				
				i++
				
				// メタデータ行を処理
				for i < len(lines) {
					metaLine := lines[i]
					if strings.HasPrefix(metaLine, "author ") {
						currentCommit.Author = strings.TrimPrefix(metaLine, "author ")
					} else if strings.HasPrefix(metaLine, "author-time ") {
						timestampStr := strings.TrimPrefix(metaLine, "author-time ")
						if timestamp, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
							currentCommit.Date = time.Unix(timestamp, 0)
						}
					} else if strings.HasPrefix(metaLine, "\t") {
						// コード行に到達
						currentCommit.Content = strings.TrimPrefix(metaLine, "\t")
						result = append(result, currentCommit)
						break
					}
					i++
				}
			}
		}
		i++
	}

	return result, nil
}

// combineBlameWithTracking はGit blameとトラッキング情報を結合する
func (b *Blamer) combineBlameWithTracking(gitBlameLines []GitBlameLine, events []*types.TrackEvent, filePath string) ([]BlameInfo, error) {
	// コミットハッシュごとのトラッキング情報マップを作成
	commitEventMap := make(map[string]*types.TrackEvent)
	for _, event := range events {
		if event.CommitHash != "" {
			commitEventMap[event.CommitHash] = event
		}
	}

	// ファイルパス関連のAIイベントマップを作成
	aiEventsByTime := make(map[string]*types.TrackEvent)
	for _, event := range events {
		if event.EventType == types.EventTypeAI {
			for _, file := range event.Files {
				if strings.Contains(file.Path, filePath) || strings.Contains(filePath, file.Path) {
					// 時刻をキーとしてマッピング
					timeKey := event.Timestamp.Format("2006-01-02 15:04:05")
					aiEventsByTime[timeKey] = event
				}
			}
		}
	}

	var result []BlameInfo
	for _, gitLine := range gitBlameLines {
		blameInfo := BlameInfo{
			LineNumber: gitLine.LineNumber,
			Author:     gitLine.Author,
			Date:       gitLine.Date,
			CommitHash: gitLine.CommitHash,
			Content:    gitLine.Content,
			IsAI:       false,
		}

		// コミットハッシュから直接的にAI判定
		if event, exists := commitEventMap[gitLine.CommitHash]; exists {
			if event.EventType == types.EventTypeAI {
				blameInfo.IsAI = true
				blameInfo.Author = "Claude Code"
				blameInfo.Model = event.Model
			}
		} else {
			// Claude Codeのコミットパターンを検出
			if b.isClaudeCodeAuthor(gitLine.Author) {
				blameInfo.IsAI = true
				blameInfo.Author = "Claude Code"
				blameInfo.Model = b.guessModelFromDate(gitLine.Date)
			}
		}

		result = append(result, blameInfo)
	}

	return result, nil
}

// isClaudeCodeAuthor はClaude Codeによる作成者かどうかを判定する
func (b *Blamer) isClaudeCodeAuthor(author string) bool {
	claudePatterns := []string{
		"Claude Code",
		"Claude",
		"claude",
		"AI Assistant",
		"noreply@anthropic.com",
	}

	authorLower := strings.ToLower(author)
	for _, pattern := range claudePatterns {
		if strings.Contains(authorLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// guessModelFromDate は日付からモデルを推測する
func (b *Blamer) guessModelFromDate(date time.Time) string {
	// 2024年後半以降はClaude 4系、それ以前は3系と推測
	if date.After(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)) {
		return "claude-sonnet-4" // デフォルトは最新モデル
	}
	return "claude-3-sonnet"
}

// calculateStatistics は統計情報を計算する
func (b *Blamer) calculateStatistics(lines []BlameInfo) BlameStatistics {
	stats := BlameStatistics{
		TotalLines: len(lines),
	}

	// モデル別・作成者別のカウント
	modelCount := make(map[string]int)
	authorCount := make(map[string]int)

	for _, line := range lines {
		if line.IsAI {
			stats.AILines++
			if line.Model != "" {
				modelCount[line.Model]++
			}
		} else {
			stats.HumanLines++
			authorCount[line.Author]++
		}
	}

	// パーセンテージ計算
	if stats.TotalLines > 0 {
		stats.AIPercentage = float64(stats.AILines) / float64(stats.TotalLines) * 100.0
		stats.HumanPercentage = float64(stats.HumanLines) / float64(stats.TotalLines) * 100.0
	}

	// 最も使用されたモデルを特定
	maxModelCount := 0
	for model, count := range modelCount {
		if count > maxModelCount {
			maxModelCount = count
			stats.TopAIModel = model
		}
	}

	// 最も貢献した人間の作成者を特定
	maxAuthorCount := 0
	for author, count := range authorCount {
		if count > maxAuthorCount {
			maxAuthorCount = count
			stats.TopHumanAuthor = author
		}
	}

	return stats
}

// FormatBlameOutput はblame結果を整形して出力する
func (b *Blamer) FormatBlameOutput(result *FileBlameResult, useColor bool) string {
	var output strings.Builder

	// ヘッダー情報
	output.WriteString(fmt.Sprintf("=== AI Code Tracker Blame: %s ===\n\n", result.FilePath))

	// 統計情報
	output.WriteString("📊 統計情報:\n")
	output.WriteString(fmt.Sprintf("  総行数: %d\n", result.Statistics.TotalLines))
	output.WriteString(fmt.Sprintf("  AI による行: %d (%.1f%%)\n", 
		result.Statistics.AILines, result.Statistics.AIPercentage))
	output.WriteString(fmt.Sprintf("  人間による行: %d (%.1f%%)\n", 
		result.Statistics.HumanLines, result.Statistics.HumanPercentage))
	
	if result.Statistics.TopAIModel != "" {
		output.WriteString(fmt.Sprintf("  主要AIモデル: %s\n", result.Statistics.TopAIModel))
	}
	if result.Statistics.TopHumanAuthor != "" {
		output.WriteString(fmt.Sprintf("  主要貢献者: %s\n", result.Statistics.TopHumanAuthor))
	}
	output.WriteString("\n")

	// 行別blame情報
	output.WriteString("📝 行別情報:\n")
	for _, line := range result.Lines {
		lineStr := b.formatBlameLine(line, useColor)
		output.WriteString(lineStr + "\n")
	}

	return output.String()
}

// formatBlameLine は1行のblame情報を整形する
func (b *Blamer) formatBlameLine(line BlameInfo, useColor bool) string {
	// 日付を短縮形式で表示
	dateStr := line.Date.Format("2006-01-02")
	
	// 作成者を短縮（最大15文字）
	author := line.Author
	if len(author) > 15 {
		author = author[:12] + "..."
	}

	// 基本フォーマット
	prefix := fmt.Sprintf("%4d  %-15s %s", line.LineNumber, author, dateStr)

	// AI/人間の区別
	var indicator string
	if line.IsAI {
		indicator = "🤖"
		if line.Model != "" {
			// モデル名を短縮表示
			model := line.Model
			if strings.Contains(model, "sonnet") {
				model = "S4"
			} else if strings.Contains(model, "opus") {
				model = "O4"
			} else if strings.Contains(model, "claude-3") {
				model = "C3"
			}
			indicator += fmt.Sprintf("(%s)", model)
		}
	} else {
		indicator = "👤"
	}

	// カラー表示
	if useColor {
		if line.IsAI {
			// AI行は青色
			prefix = fmt.Sprintf("\033[34m%s\033[0m", prefix)
		} else {
			// 人間行は緑色
			prefix = fmt.Sprintf("\033[32m%s\033[0m", prefix)
		}
	}

	return fmt.Sprintf("%s %s  %s", prefix, indicator, line.Content)
}

// GetFileContribution はファイルの貢献者別統計を取得する
func (b *Blamer) GetFileContribution(filePath string) (map[string]int, error) {
	result, err := b.BlameFile(filePath)
	if err != nil {
		return nil, err
	}

	contribution := make(map[string]int)
	for _, line := range result.Lines {
		if line.IsAI {
			key := fmt.Sprintf("AI (%s)", line.Model)
			contribution[key]++
		} else {
			contribution[line.Author]++
		}
	}

	return contribution, nil
}

// GetTopContributors は上位貢献者を取得する
func (b *Blamer) GetTopContributors(filePath string, limit int) ([]ContributorInfo, error) {
	contribution, err := b.GetFileContribution(filePath)
	if err != nil {
		return nil, err
	}

	type contributorPair struct {
		name  string
		lines int
	}

	var contributors []contributorPair
	for name, lines := range contribution {
		contributors = append(contributors, contributorPair{name, lines})
	}

	// 行数でソート
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].lines > contributors[j].lines
	})

	// limit適用
	if limit > 0 && len(contributors) > limit {
		contributors = contributors[:limit]
	}

	// 結果変換
	var result []ContributorInfo
	totalLines := 0
	for _, c := range contributors {
		totalLines += c.lines
	}

	for _, c := range contributors {
		percentage := float64(c.lines) / float64(totalLines) * 100.0
		result = append(result, ContributorInfo{
			Name:       c.name,
			Lines:      c.lines,
			Percentage: percentage,
			IsAI:       strings.Contains(c.name, "AI"),
		})
	}

	return result, nil
}

// ContributorInfo は貢献者情報を表す
type ContributorInfo struct {
	// Name は貢献者名
	Name string
	// Lines は行数
	Lines int
	// Percentage は貢献率
	Percentage float64
	// IsAI はAIかどうか
	IsAI bool
}

// ValidateFilePath はファイルパスが有効かどうかを検証する
func (b *Blamer) ValidateFilePath(filePath string) error {
	// ファイルが存在するかチェック
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが存在しません: %s", filePath)
	}

	// Gitで追跡されているかチェック
	cmd := exec.Command("git", "ls-files", "--error-unmatch", filePath)
	cmd.Dir = b.gitRepo
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ファイルがGitで追跡されていません: %s", filePath)
	}

	return nil
}