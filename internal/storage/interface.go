package storage

import (
	"context"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
)

// StorageInterface は統一されたストレージインターフェース
type StorageInterface interface {
	// 基本のトラッキング機能
	StoreTrackEvent(event *types.TrackEvent) error
	ReadEvents() ([]*types.TrackEvent, error)
	ReadEventsByDateRange(start, end time.Time) ([]*types.TrackEvent, error)
	ReadEventsByAuthor(author string) ([]*types.TrackEvent, error)
	ReadEventsByType(eventType types.EventType) ([]*types.TrackEvent, error)
	GetStatistics() (*types.Statistics, error)
	Close() error
}

// AdvancedStorageInterface は高度な分析機能を提供するインターフェース
type AdvancedStorageInterface interface {
	StorageInterface
	
	// 期間別分析機能（DuckDBのみ）
	GetPeriodAnalysis(ctx context.Context, startDate, endDate time.Time) (*PeriodAnalysis, error)
	GetBasicStats(ctx context.Context) (*BasicStats, error)
	
	// データベース情報取得
	GetDatabaseInfo() (*DatabaseInfo, error)
	TestConnection() error
}

// StorageType はストレージタイプを表す
type StorageType int

const (
	StorageTypeJSONL StorageType = iota
	StorageTypeDuckDB
)

// StorageConfig はストレージ設定
type StorageConfig struct {
	Type    StorageType `json:"type"`
	DataDir string      `json:"data_dir"`
	Debug   bool        `json:"debug"`
}

// NewStorageByType はタイプに応じたストレージを作成
func NewStorageByType(config StorageConfig) (StorageInterface, error) {
	switch config.Type {
	case StorageTypeDuckDB:
		return NewDuckDBStorage(config.DataDir, config.Debug)
	case StorageTypeJSONL:
		fallthrough
	default:
		return NewStorage(config.DataDir)
	}
}

// NewAdvancedStorageByType は高度な機能を持つストレージを作成
func NewAdvancedStorageByType(config StorageConfig) (AdvancedStorageInterface, error) {
	switch config.Type {
	case StorageTypeDuckDB:
		return NewDuckDBStorage(config.DataDir, config.Debug)
	default:
		// JSONLストレージには高度な機能がないため、ラッパーを返す
		basicStorage, err := NewStorage(config.DataDir)
		if err != nil {
			return nil, err
		}
		return &JSONLStorageWrapper{basicStorage}, nil
	}
}

// JSONLStorageWrapper はJSONLストレージを高度な機能インターフェースに適合させる
type JSONLStorageWrapper struct {
	*Storage
}

// GetPeriodAnalysis はJSONLストレージでは簡単な分析のみを提供
func (w *JSONLStorageWrapper) GetPeriodAnalysis(ctx context.Context, startDate, endDate time.Time) (*PeriodAnalysis, error) {
	events, err := w.ReadEventsByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	return buildBasicPeriodAnalysis(events, startDate, endDate), nil
}

// GetBasicStats はJSONLストレージでは基本統計のみを提供
func (w *JSONLStorageWrapper) GetBasicStats(ctx context.Context) (*BasicStats, error) {
	stats, err := w.GetStatistics()
	if err != nil {
		return nil, err
	}
	
	basicStats := &BasicStats{
		TotalEvents: stats.TotalEvents,
		AILines:     stats.TotalLinesAdded, // 簡略化された計算
		HumanLines:  0,                     // JSONLでは区別が困難
		TotalLines:  stats.TotalLinesAdded,
	}
	
	if stats.FirstEvent != nil {
		basicStats.FirstEvent = *stats.FirstEvent
	}
	if stats.LastEvent != nil {
		basicStats.LastEvent = *stats.LastEvent
	}
	
	return basicStats, nil
}

// GetDatabaseInfo はJSONLストレージの情報を返す
func (w *JSONLStorageWrapper) GetDatabaseInfo() (*DatabaseInfo, error) {
	return &DatabaseInfo{
		Path:       w.dataFile,
		TrackCount: w.index.TotalEvents,
	}, nil
}

// TestConnection はJSONLストレージでは常に成功
func (w *JSONLStorageWrapper) TestConnection() error {
	return nil
}

// buildBasicPeriodAnalysis はJSONLデータから基本的な期間分析を構築
func buildBasicPeriodAnalysis(events []*types.TrackEvent, startDate, endDate time.Time) *PeriodAnalysis {
	analysis := &PeriodAnalysis{
		StartDate: startDate,
		EndDate:   endDate,
	}
	
	fileMap := make(map[string]*FileAnalysis)
	
	for _, event := range events {
		analysis.SessionCount++
		
		for _, file := range event.Files {
			// ファイル分析
			if fa, exists := fileMap[file.Path]; exists {
				fa.TotalLines += file.LinesAdded
				if event.EventType == types.EventTypeAI {
					fa.AILines += file.LinesAdded
				} else {
					fa.HumanLines += file.LinesAdded
				}
			} else {
				fa := &FileAnalysis{
					FilePath:   file.Path,
					Language:   detectLanguage(file.Path),
					FirstEdit:  event.Timestamp,
					LastEdit:   event.Timestamp,
				}
				if event.EventType == types.EventTypeAI {
					fa.AILines = file.LinesAdded
				} else {
					fa.HumanLines = file.LinesAdded
				}
				fa.TotalLines = fa.AILines + fa.HumanLines
				fileMap[file.Path] = fa
			}
			
			// 全体統計に追加
			analysis.TotalLines += file.LinesAdded
			if event.EventType == types.EventTypeAI {
				analysis.AILines += file.LinesAdded
			} else {
				analysis.HumanLines += file.LinesAdded
			}
		}
	}
	
	// パーセンテージ計算
	if analysis.TotalLines > 0 {
		analysis.AIPercentage = float64(analysis.AILines) / float64(analysis.TotalLines) * 100
	}
	
	// ファイルのパーセンテージ計算
	for _, fa := range fileMap {
		if fa.TotalLines > 0 {
			fa.AIPercentage = float64(fa.AILines) / float64(fa.TotalLines) * 100
		}
		analysis.FileBreakdown = append(analysis.FileBreakdown, *fa)
	}
	
	analysis.FileCount = len(fileMap)
	
	return analysis
}

// detectLanguage はファイルパスから言語を推測（簡略版）
func detectLanguage(filePath string) string {
	switch {
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".go":
		return "Go"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".js":
		return "JavaScript"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".ts":
		return "TypeScript"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".tsx":
		return "TypeScript"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".py":
		return "Python"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".java":
		return "Java"
	default:
		return "Other"
	}
}