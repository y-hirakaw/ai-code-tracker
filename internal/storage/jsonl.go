package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ai-code-tracker/aict/pkg/types"
)

const (
	// DefaultDataDir はトラッキングデータを保存するデフォルトディレクトリ
	DefaultDataDir = ".git/ai-tracker"
	// DataFileName はメインデータファイル名
	DataFileName = "events.jsonl"
	// IndexFileName は高速クエリ用のインデックスファイル名
	IndexFileName = "index.json"
	// BackupSuffix はバックアップファイルに付加される接尾辞
	BackupSuffix = ".backup"
)

// Storage handles JSONL file operations for tracking events
type Storage struct {
	dataDir    string
	dataFile   string
	indexFile  string
	mutex      sync.RWMutex
	index      *Index
}

// Index represents an in-memory index for fast queries
type Index struct {
	// EventsByDate maps date (YYYY-MM-DD) to event IDs
	EventsByDate map[string][]string `json:"events_by_date"`
	// EventsByAuthor maps author to event IDs
	EventsByAuthor map[string][]string `json:"events_by_author"`
	// EventsByType maps event type to event IDs
	EventsByType map[string][]string `json:"events_by_type"`
	// TotalEvents is the total number of events
	TotalEvents int `json:"total_events"`
	// LastUpdated is when the index was last updated
	LastUpdated time.Time `json:"last_updated"`
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		EventsByDate:   make(map[string][]string),
		EventsByAuthor: make(map[string][]string),
		EventsByType:   make(map[string][]string),
		TotalEvents:    0,
		LastUpdated:    time.Now(),
	}
}

// NewStorage creates a new Storage instance
func NewStorage(dataDir string) (*Storage, error) {
	if dataDir == "" {
		dataDir = DefaultDataDir
	}
	
	dataFile := filepath.Join(dataDir, DataFileName)
	indexFile := filepath.Join(dataDir, IndexFileName)
	
	storage := &Storage{
		dataDir:   dataDir,
		dataFile:  dataFile,
		indexFile: indexFile,
		index:     NewIndex(),
	}
	
	// Create data directory if it doesn't exist
	if err := storage.ensureDataDir(); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Load existing index
	if err := storage.loadIndex(); err != nil {
		// If index doesn't exist or is corrupted, rebuild it
		if err := storage.rebuildIndex(); err != nil {
			return nil, fmt.Errorf("failed to rebuild index: %w", err)
		}
	}
	
	return storage, nil
}

// ensureDataDir creates the data directory if it doesn't exist
func (s *Storage) ensureDataDir() error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", s.dataDir, err)
	}
	return nil
}

// WriteEvent writes a tracking event to the JSONL file
func (s *Storage) WriteEvent(event *types.TrackEvent) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Validate event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	
	// Convert to JSON
	jsonStr, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to convert event to JSON: %w", err)
	}
	
	// Write to file atomically
	if err := s.writeLineAtomic(jsonStr); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	
	// Update index
	s.updateIndex(event)
	
	// Save index
	if err := s.saveIndex(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}
	
	return nil
}

// writeLineAtomic writes a line to the JSONL file atomically
func (s *Storage) writeLineAtomic(line string) error {
	// Create temporary file
	tempFile := s.dataFile + ".tmp"
	
	// Open existing file for reading (if it exists)
	var existingContent []byte
	if _, err := os.Stat(s.dataFile); err == nil {
		existingContent, err = os.ReadFile(s.dataFile)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	}
	
	// Write to temporary file
	file, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer file.Close()
	
	// Write existing content
	if len(existingContent) > 0 {
		if _, err := file.Write(existingContent); err != nil {
			return fmt.Errorf("failed to write existing content: %w", err)
		}
	}
	
	// Write new line
	if _, err := file.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write new line: %w", err)
	}
	
	// Sync to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	
	// Close file
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempFile, s.dataFile); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}
	
	return nil
}

// ReadEvents reads all events from the JSONL file
func (s *Storage) ReadEvents() ([]*types.TrackEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	file, err := os.Open(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []*types.TrackEvent{}, nil
		}
		return nil, fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()
	
	var events []*types.TrackEvent
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		event, err := types.FromJSON(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}
		
		events = append(events, event)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	return events, nil
}

// ReadEventsByDateRange reads events within a date range
func (s *Storage) ReadEventsByDateRange(start, end time.Time) ([]*types.TrackEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	var eventIDs []string
	
	// Collect event IDs from index
	current := start
	for !current.After(end) {
		dateKey := current.Format("2006-01-02")
		if ids, exists := s.index.EventsByDate[dateKey]; exists {
			eventIDs = append(eventIDs, ids...)
		}
		current = current.AddDate(0, 0, 1)
	}
	
	if len(eventIDs) == 0 {
		return []*types.TrackEvent{}, nil
	}
	
	// Read events by IDs
	return s.readEventsByIDs(eventIDs)
}

// ReadEventsByAuthor reads events by author
func (s *Storage) ReadEventsByAuthor(author string) ([]*types.TrackEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	eventIDs, exists := s.index.EventsByAuthor[author]
	if !exists || len(eventIDs) == 0 {
		return []*types.TrackEvent{}, nil
	}
	
	return s.readEventsByIDs(eventIDs)
}

// ReadEventsByType reads events by type
func (s *Storage) ReadEventsByType(eventType types.EventType) ([]*types.TrackEvent, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	eventIDs, exists := s.index.EventsByType[string(eventType)]
	if !exists || len(eventIDs) == 0 {
		return []*types.TrackEvent{}, nil
	}
	
	return s.readEventsByIDs(eventIDs)
}

// readEventsByIDs reads events by their IDs
func (s *Storage) readEventsByIDs(eventIDs []string) ([]*types.TrackEvent, error) {
	idSet := make(map[string]bool)
	for _, id := range eventIDs {
		idSet[id] = true
	}
	
	allEvents, err := s.ReadEvents()
	if err != nil {
		return nil, err
	}
	
	var filteredEvents []*types.TrackEvent
	for _, event := range allEvents {
		if idSet[event.ID] {
			filteredEvents = append(filteredEvents, event)
		}
	}
	
	return filteredEvents, nil
}

// updateIndex updates the in-memory index with a new event
func (s *Storage) updateIndex(event *types.TrackEvent) {
	// Update by date
	dateKey := event.Timestamp.Format("2006-01-02")
	s.index.EventsByDate[dateKey] = append(s.index.EventsByDate[dateKey], event.ID)
	
	// Update by author
	s.index.EventsByAuthor[event.Author] = append(s.index.EventsByAuthor[event.Author], event.ID)
	
	// Update by type
	typeKey := string(event.EventType)
	s.index.EventsByType[typeKey] = append(s.index.EventsByType[typeKey], event.ID)
	
	// Update totals
	s.index.TotalEvents++
	s.index.LastUpdated = time.Now()
}

// loadIndex loads the index from disk
func (s *Storage) loadIndex() error {
	file, err := os.Open(s.indexFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	return decoder.Decode(s.index)
}

// saveIndex saves the index to disk
func (s *Storage) saveIndex() error {
	// Create temporary file
	tempFile := s.indexFile + ".tmp"
	
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary index file: %w", err)
	}
	defer file.Close()
	
	// Encode index
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(s.index); err != nil {
		return fmt.Errorf("failed to encode index: %w", err)
	}
	
	// Sync to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync index file: %w", err)
	}
	
	// Close file
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close index file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tempFile, s.indexFile); err != nil {
		return fmt.Errorf("failed to rename temporary index file: %w", err)
	}
	
	return nil
}

// rebuildIndex rebuilds the index from the data file
func (s *Storage) rebuildIndex() error {
	s.index = NewIndex()
	
	events, err := s.ReadEvents()
	if err != nil {
		return fmt.Errorf("failed to read events for index rebuild: %w", err)
	}
	
	for _, event := range events {
		s.updateIndex(event)
	}
	
	return s.saveIndex()
}

// GetStatistics returns aggregated statistics
func (s *Storage) GetStatistics() (*types.Statistics, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	events, err := s.ReadEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to read events for statistics: %w", err)
	}
	
	stats := &types.Statistics{}
	
	for _, event := range events {
		stats.TotalEvents++
		stats.TotalLinesAdded += event.TotalLinesAdded()
		stats.TotalLinesModified += event.TotalLinesModified()
		stats.TotalLinesDeleted += event.TotalLinesDeleted()
		
		switch event.EventType {
		case types.EventTypeAI:
			stats.AIEvents++
		case types.EventTypeHuman:
			stats.HumanEvents++
		case types.EventTypeCommit:
			stats.CommitEvents++
		}
		
		if stats.FirstEvent == nil || event.Timestamp.Before(*stats.FirstEvent) {
			stats.FirstEvent = &event.Timestamp
		}
		if stats.LastEvent == nil || event.Timestamp.After(*stats.LastEvent) {
			stats.LastEvent = &event.Timestamp
		}
	}
	
	return stats, nil
}

// Close closes the storage (currently no-op but good for future cleanup)
func (s *Storage) Close() error {
	return nil
}