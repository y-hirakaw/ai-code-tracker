package web

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Config はWebサーバーの設定
type Config struct {
	DataDir string
	Debug   bool
	Lang    string
}

// UpdateEvent はリアルタイム更新イベント
type UpdateEvent struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// StandaloneServer は独立したWebサーバー（依存関係なし）
type StandaloneServer struct {
	config      *Config
	dataDir     string
	
	// リアルタイム更新用
	subscribers map[string]chan *UpdateEvent
	subsMutex   sync.RWMutex
	
	// 簡易キャッシュ
	statsCache    *StandaloneStats
	cacheMutex    sync.RWMutex
	cacheExpiry   time.Time
}

// StandaloneStats は独立した統計構造
type StandaloneStats struct {
	TotalLines    int                      `json:"total_lines"`
	AILines       int                      `json:"ai_lines"`
	HumanLines    int                      `json:"human_lines"`
	FileCount     int                      `json:"file_count"`
	LastUpdated   time.Time                `json:"last_updated"`
	FileStats     []StandaloneFileInfo     `json:"file_stats"`
	Contributors  []StandaloneContributor  `json:"contributors"`
}

// StandaloneFileInfo は独立したファイル情報
type StandaloneFileInfo struct {
	Path       string `json:"path"`
	TotalLines int    `json:"total_lines"`
	AILines    int    `json:"ai_lines"`
	HumanLines int    `json:"human_lines"`
}

// StandaloneContributor は独立した貢献者情報
type StandaloneContributor struct {
	Name  string `json:"name"`
	Type  string `json:"type"` // "ai" or "human"
	Lines int    `json:"lines"`
}

// StandaloneBlameResult は独立したBlame結果
type StandaloneBlameResult struct {
	FilePath string                 `json:"file_path"` 
	Lines    []StandaloneBlameLine  `json:"lines"`
	Summary  StandaloneBlameSummary `json:"summary"`
}

// StandaloneBlameLine は独立したBlame行情報  
type StandaloneBlameLine struct {
	LineNumber int       `json:"line_number"`
	Content    string    `json:"content"`
	Author     string    `json:"author"`
	AuthorType string    `json:"author_type"`
	Timestamp  time.Time `json:"timestamp"`
}

// StandaloneBlameSummary は独立したBlame要約
type StandaloneBlameSummary struct {
	TotalLines   int            `json:"total_lines"`
	AILines      int            `json:"ai_lines"`
	HumanLines   int            `json:"human_lines"`
	Contributors map[string]int `json:"contributors"`
}

// StandaloneEvent は独立したイベント情報
type StandaloneEvent struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Files       []string  `json:"files"`
}

// NewSimpleWebServer は新しい独立Webサーバーを作成する
func NewSimpleWebServer(config *Config) *StandaloneServer {
	// データディレクトリの確認
	dataDir := config.DataDir
	if dataDir == "" {
		dataDir = findDefaultDataDir()
	}

	server := &StandaloneServer{
		config:      config,
		dataDir:     dataDir,
		subscribers: make(map[string]chan *UpdateEvent),
		statsCache:  &StandaloneStats{},
		cacheExpiry: time.Now(),
	}

	// バックグラウンドでキャッシュ更新
	go server.updateCacheLoop()

	return server
}

// findDefaultDataDir はデフォルトのデータディレクトリを探す
func findDefaultDataDir() string {
	// 現在のディレクトリから .git/ai-tracker を探す
	currentDir, _ := os.Getwd()
	for dir := currentDir; dir != "/" && dir != ""; {
		trackerDir := filepath.Join(dir, ".git", "ai-tracker")
		if _, err := os.Stat(trackerDir); err == nil {
			return trackerDir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	
	// 見つからない場合は現在のディレクトリを使用
	return currentDir
}

// GetStats は統計データを取得する
func (s *StandaloneServer) GetStats(ctx context.Context, filter interface{}) (*StandaloneStats, error) {
	s.cacheMutex.RLock()
	
	// キャッシュが有効（1分以内）かチェック
	if time.Now().Before(s.cacheExpiry) {
		defer s.cacheMutex.RUnlock()
		return s.statsCache, nil
	}
	s.cacheMutex.RUnlock()

	// 新しい統計を生成
	stats, err := s.generateStats()
	if err != nil {
		return nil, err
	}

	// キャッシュを更新
	s.cacheMutex.Lock()
	s.statsCache = stats
	s.cacheExpiry = time.Now().Add(1 * time.Minute)
	s.cacheMutex.Unlock()

	return stats, nil
}

// GetContributors は貢献者リストを取得する
func (s *StandaloneServer) GetContributors(ctx context.Context) ([]StandaloneContributor, error) {
	stats, err := s.GetStats(ctx, nil)
	if err != nil {
		return nil, err
	}
	return stats.Contributors, nil
}

// GetBlame はファイルのblame情報を取得する
func (s *StandaloneServer) GetBlame(ctx context.Context, filePath string) (*StandaloneBlameResult, error) {
	// 簡易実装（デモデータ）
	return &StandaloneBlameResult{
		FilePath: filePath,
		Lines: []StandaloneBlameLine{
			{1, "package main", "Developer", "human", time.Now().Add(-24 * time.Hour)},
			{2, "", "Developer", "human", time.Now().Add(-24 * time.Hour)},
			{3, "import \"fmt\"", "Claude Sonnet 4", "ai", time.Now().Add(-12 * time.Hour)},
			{4, "", "Claude Sonnet 4", "ai", time.Now().Add(-12 * time.Hour)},
			{5, "func main() {", "Developer", "human", time.Now().Add(-24 * time.Hour)},
			{6, "    fmt.Println(\"Hello, World!\")", "Claude Sonnet 4", "ai", time.Now().Add(-12 * time.Hour)},
			{7, "}", "Developer", "human", time.Now().Add(-24 * time.Hour)},
		},
		Summary: StandaloneBlameSummary{
			TotalLines:   7,
			AILines:      3,
			HumanLines:   4,
			Contributors: map[string]int{"Developer": 4, "Claude Sonnet 4": 3},
		},
	}, nil
}

// GetTimeline はタイムライン情報を取得する
func (s *StandaloneServer) GetTimeline(ctx context.Context, limit int) ([]StandaloneEvent, error) {
	// 簡易実装（デモデータ）
	events := []StandaloneEvent{
		{
			ID: "event-1", 
			Timestamp: time.Now().Add(-1 * time.Hour),
			Type: "ai_edit",
			Author: "Claude Sonnet 4", 
			Description: "Refactored main function",
			Files: []string{"src/main.go"},
		},
		{
			ID: "event-2",
			Timestamp: time.Now().Add(-3 * time.Hour),
			Type: "human_edit",
			Author: "Developer",
			Description: "Fixed bug in error handling",
			Files: []string{"src/handler.go"},
		},
		{
			ID: "event-3",
			Timestamp: time.Now().Add(-6 * time.Hour),
			Type: "ai_edit",
			Author: "Claude Sonnet 4",
			Description: "Added utility functions",
			Files: []string{"src/utils.go"},
		},
	}
	
	if limit > 0 && limit < len(events) {
		events = events[:limit]
	}
	
	return events, nil
}

// GetFileStats はファイル統計を取得する
func (s *StandaloneServer) GetFileStats(ctx context.Context) ([]StandaloneFileInfo, error) {
	stats, err := s.GetStats(ctx, nil)
	if err != nil {
		return nil, err
	}
	return stats.FileStats, nil
}

// generateStats は統計を生成する
func (s *StandaloneServer) generateStats() (*StandaloneStats, error) {
	stats := &StandaloneStats{
		LastUpdated:  time.Now(),
		FileStats:    []StandaloneFileInfo{},
		Contributors: []StandaloneContributor{},
	}

	// JSONLファイルを読み込み（存在する場合）
	tracksFile := filepath.Join(s.dataDir, "tracks.jsonl")
	if _, err := os.Stat(tracksFile); err == nil {
		if err := s.loadFromJSONL(tracksFile, stats); err != nil {
			log.Printf("Warning: Failed to load JSONL: %v", err)
		}
	}

	// デモデータ（実際のデータがない場合）
	if stats.TotalLines == 0 {
		stats.TotalLines = 1500
		stats.AILines = 900  
		stats.HumanLines = 600
		stats.FileCount = 25
		
		stats.FileStats = []StandaloneFileInfo{
			{"src/main.go", 250, 150, 100},
			{"src/handler.go", 180, 120, 60},
			{"src/utils.go", 95, 45, 50},
			{"internal/web/server.go", 200, 150, 50},
			{"internal/cli/app.go", 150, 80, 70},
		}
		
		stats.Contributors = []StandaloneContributor{
			{"Claude Sonnet 4", "ai", 900},
			{"Developer", "human", 600},
		}
	}

	return stats, nil
}

// loadFromJSONL は JSONLファイルから統計を読み込む
func (s *StandaloneServer) loadFromJSONL(filename string, stats *StandaloneStats) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	
	fileMap := make(map[string]*StandaloneFileInfo)
	authorMap := make(map[string]*StandaloneContributor)
	
	for decoder.More() {
		var event map[string]interface{}
		if err := decoder.Decode(&event); err != nil {
			continue
		}
		
		// ファイル情報を抽出
		if files, ok := event["files"].([]interface{}); ok {
			for _, f := range files {
				if fileData, ok := f.(map[string]interface{}); ok {
					path, _ := fileData["path"].(string)
					linesAdded, _ := fileData["lines_added"].(float64)
					
					if path != "" {
						if _, exists := fileMap[path]; !exists {
							fileMap[path] = &StandaloneFileInfo{Path: path}
						}
						fileMap[path].TotalLines += int(linesAdded)
						
						// 作成者タイプに基づいて分類
						author, _ := event["author"].(string)
						if isAIAuthor(author) {
							fileMap[path].AILines += int(linesAdded)
						} else {
							fileMap[path].HumanLines += int(linesAdded)
						}
					}
				}
			}
		}
		
		// 作成者情報を抽出
		if author, ok := event["author"].(string); ok && author != "" {
			linesAdded := 0
			if files, ok := event["files"].([]interface{}); ok {
				for _, f := range files {
					if fileData, ok := f.(map[string]interface{}); ok {
						if lines, ok := fileData["lines_added"].(float64); ok {
							linesAdded += int(lines)
						}
					}
				}
			}
			
			if _, exists := authorMap[author]; !exists {
				authorType := "human"
				if isAIAuthor(author) {
					authorType = "ai"
				}
				authorMap[author] = &StandaloneContributor{
					Name: author,
					Type: authorType,
					Lines: 0,
				}
			}
			authorMap[author].Lines += linesAdded
		}
	}
	
	// マップからスライスに変換
	for _, fileInfo := range fileMap {
		stats.FileStats = append(stats.FileStats, *fileInfo)
		stats.TotalLines += fileInfo.TotalLines
		stats.AILines += fileInfo.AILines
		stats.HumanLines += fileInfo.HumanLines
	}
	
	for _, contributor := range authorMap {
		stats.Contributors = append(stats.Contributors, *contributor)
	}
	
	stats.FileCount = len(stats.FileStats)
	
	return nil
}

// isAIAuthor はAI作成者かどうかを判定する
func isAIAuthor(author string) bool {
	aiKeywords := []string{"claude", "ai", "gpt", "copilot", "sonnet", "opus"}
	authorLower := strings.ToLower(author)
	for _, keyword := range aiKeywords {
		if strings.Contains(authorLower, keyword) {
			return true
		}
	}
	return false
}

// Subscribe はリアルタイム更新を購読する
func (s *StandaloneServer) Subscribe(clientID string) <-chan *UpdateEvent {
	s.subsMutex.Lock()
	defer s.subsMutex.Unlock()
	
	ch := make(chan *UpdateEvent, 100)
	s.subscribers[clientID] = ch
	return ch
}

// Unsubscribe はリアルタイム更新を停止する
func (s *StandaloneServer) Unsubscribe(clientID string) {
	s.subsMutex.Lock()
	defer s.subsMutex.Unlock()
	
	if ch, exists := s.subscribers[clientID]; exists {
		close(ch)
		delete(s.subscribers, clientID)
	}
}

// Broadcast は全購読者にイベントを送信する
func (s *StandaloneServer) Broadcast(event *UpdateEvent) {
	s.subsMutex.RLock()
	defer s.subsMutex.RUnlock()
	
	for _, ch := range s.subscribers {
		select {
		case ch <- event:
		default:
			// バッファが満杯の場合はスキップ
		}
	}
}

// updateCacheLoop はキャッシュを定期更新する
func (s *StandaloneServer) updateCacheLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if stats, err := s.generateStats(); err == nil {
				s.cacheMutex.Lock()
				oldStats := s.statsCache
				s.statsCache = stats
				s.cacheExpiry = time.Now().Add(1 * time.Minute)
				s.cacheMutex.Unlock()

				// 変更があれば通知
				if oldStats == nil || hasStandaloneStatsChanged(oldStats, stats) {
					s.Broadcast(&UpdateEvent{
						Type:      "stats_updated",
						Timestamp: time.Now(),
						Data:      stats,
					})
				}
			}
		}
	}
}

// hasStandaloneStatsChanged は統計に変更があったかチェックする
func hasStandaloneStatsChanged(old, new *StandaloneStats) bool {
	if old == nil || new == nil {
		return true
	}
	
	return old.TotalLines != new.TotalLines ||
		   old.AILines != new.AILines ||
		   old.HumanLines != new.HumanLines ||
		   len(old.FileStats) != len(new.FileStats)
}

// IsHealthy はサーバーの健全性をチェックする
func (s *StandaloneServer) IsHealthy() bool {
	return true // 簡易版では常に健全
}

// GetConfig は設定を取得する
func (s *StandaloneServer) GetConfig() *Config {
	return s.config
}