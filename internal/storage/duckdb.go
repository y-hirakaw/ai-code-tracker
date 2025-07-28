package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/y-hirakaw/ai-code-tracker/pkg/types"
	_ "github.com/marcboeker/go-duckdb"
)

// DuckDBStorage は DuckDB を使用した高速ストレージ実装
type DuckDBStorage struct {
	db      *sql.DB
	dbPath  string
	debug   bool
}

// NewDuckDBStorage は新しい DuckDB ストレージインスタンスを作成する
func NewDuckDBStorage(dataDir string, debug bool) (*DuckDBStorage, error) {
	// データディレクトリの作成
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "aict.duckdb")
	
	// DuckDB 接続文字列の構築
	dsn := dbPath
	if debug {
		dsn += "?access_mode=read_write&threads=4"
	}

	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DuckDB: %w", err)
	}

	storage := &DuckDBStorage{
		db:     db,
		dbPath: dbPath,
		debug:  debug,
	}

	// スキーマの初期化
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	if debug {
		log.Printf("🦆 DuckDB storage initialized: %s", dbPath)
	}

	return storage, nil
}

// initSchema はデータベーススキーマを初期化する
func (s *DuckDBStorage) initSchema() error {
	schema := `
	-- メインテーブル: トラッキングイベント
	CREATE TABLE IF NOT EXISTS tracks (
		id VARCHAR PRIMARY KEY,
		timestamp TIMESTAMP NOT NULL,
		event_type VARCHAR NOT NULL,
		author VARCHAR NOT NULL,
		author_type VARCHAR NOT NULL, -- 'ai' or 'human'
		model VARCHAR,
		commit_hash VARCHAR,
		session_id VARCHAR,
		message TEXT,
		-- パーティショニング用
		date_partition DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- ファイル変更テーブル
	CREATE TABLE IF NOT EXISTS file_changes (
		id VARCHAR PRIMARY KEY,
		track_id VARCHAR NOT NULL,
		file_path VARCHAR NOT NULL,
		lines_added INTEGER DEFAULT 0,
		lines_modified INTEGER DEFAULT 0,
		lines_deleted INTEGER DEFAULT 0,
		file_hash VARCHAR,
		language VARCHAR,
		FOREIGN KEY (track_id) REFERENCES tracks(id)
	);

	-- インデックスの作成（パフォーマンス最適化）
	CREATE INDEX IF NOT EXISTS idx_tracks_timestamp ON tracks(timestamp);
	CREATE INDEX IF NOT EXISTS idx_tracks_author_type ON tracks(author_type);
	CREATE INDEX IF NOT EXISTS idx_tracks_date_partition ON tracks(date_partition);
	CREATE INDEX IF NOT EXISTS idx_tracks_session_id ON tracks(session_id);
	
	CREATE INDEX IF NOT EXISTS idx_file_changes_track_id ON file_changes(track_id);
	CREATE INDEX IF NOT EXISTS idx_file_changes_file_path ON file_changes(file_path);
	CREATE INDEX IF NOT EXISTS idx_file_changes_language ON file_changes(language);

	-- 統計キャッシュテーブル（オプション）
	CREATE TABLE IF NOT EXISTS stats_cache (
		cache_key VARCHAR PRIMARY KEY,
		data JSON NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_stats_cache_expires ON stats_cache(expires_at);
	
	-- 言語判定用のViewを作成
	CREATE OR REPLACE VIEW file_changes_with_language AS
	SELECT *,
		CASE 
			WHEN file_path LIKE '%.go' THEN 'Go'
			WHEN file_path LIKE '%.js' THEN 'JavaScript'
			WHEN file_path LIKE '%.ts' THEN 'TypeScript'
			WHEN file_path LIKE '%.tsx' THEN 'TypeScript'
			WHEN file_path LIKE '%.py' THEN 'Python'
			WHEN file_path LIKE '%.java' THEN 'Java'
			WHEN file_path LIKE '%.cpp' THEN 'C++'
			WHEN file_path LIKE '%.c' THEN 'C'
			WHEN file_path LIKE '%.rs' THEN 'Rust'
			WHEN file_path LIKE '%.php' THEN 'PHP'
			WHEN file_path LIKE '%.rb' THEN 'Ruby'
			WHEN file_path LIKE '%.swift' THEN 'Swift'
			WHEN file_path LIKE '%.kt' THEN 'Kotlin'
			WHEN file_path LIKE '%.cs' THEN 'C#'
			WHEN file_path LIKE '%.html' THEN 'HTML'
			WHEN file_path LIKE '%.css' THEN 'CSS'
			WHEN file_path LIKE '%.scss' THEN 'SCSS'
			WHEN file_path LIKE '%.vue' THEN 'Vue'
			WHEN file_path LIKE '%.jsx' THEN 'React'
			WHEN file_path LIKE '%.md' THEN 'Markdown'
			WHEN file_path LIKE '%.yaml' OR file_path LIKE '%.yml' THEN 'YAML'
			WHEN file_path LIKE '%.json' THEN 'JSON'
			WHEN file_path LIKE '%.xml' THEN 'XML'
			WHEN file_path LIKE '%.sql' THEN 'SQL'
			ELSE 'Other'
		END as computed_language
	FROM file_changes;
	`

	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// 古いキャッシュエントリを削除
	_, err = s.db.Exec("DELETE FROM stats_cache WHERE expires_at < CURRENT_TIMESTAMP")
	if err != nil && s.debug {
		log.Printf("Warning: failed to clean old cache entries: %v", err)
	}

	return nil
}

// StoreTrackEvent はトラッキングイベントをDuckDBに保存する
func (s *DuckDBStorage) StoreTrackEvent(event *types.TrackEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// EventTypeからAuthorTypeを推定
	var authorType string
	switch event.EventType {
	case types.EventTypeAI:
		authorType = "ai"
	case types.EventTypeHuman:
		authorType = "human"
	case types.EventTypeCommit:
		// コミットは通常は人間由来だが、より詳細な判別が可能ならそれを使用
		authorType = "human"
	default:
		authorType = "unknown"
	}

	// tracksテーブルにイベントを挿入
	trackQuery := `
		INSERT INTO tracks (id, timestamp, event_type, author, author_type, model, commit_hash, session_id, message, date_partition)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err = tx.ExecContext(ctx, trackQuery,
		event.ID,
		event.Timestamp,
		event.EventType,
		event.Author,
		authorType,
		event.Model,
		event.CommitHash,
		event.SessionID,
		event.Message,
		event.Timestamp.Truncate(24*time.Hour), // date_partitionを設定
	)
	if err != nil {
		return fmt.Errorf("failed to insert track event: %w", err)
	}

	// file_changesテーブルにファイル変更を挿入
	if len(event.Files) > 0 {
		fileQuery := `
			INSERT INTO file_changes (id, track_id, file_path, lines_added, lines_modified, lines_deleted, file_hash, language)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`
		
		for i, file := range event.Files {
			fileID := fmt.Sprintf("%s_file_%d", event.ID, i)
			language := detectLanguageFromPath(file.Path)
			_, err = tx.ExecContext(ctx, fileQuery,
				fileID,
				event.ID,
				file.Path,
				file.LinesAdded,
				file.LinesModified,
				file.LinesDeleted,
				file.Hash,
				language,
			)
			if err != nil {
				return fmt.Errorf("failed to insert file change: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if s.debug {
		log.Printf("🦆 Stored track event: %s (%s by %s)", event.ID, event.EventType, event.Author)
	}

	return nil
}

// GetBasicStats は基本統計を高速取得する
func (s *DuckDBStorage) GetBasicStats(ctx context.Context) (*BasicStats, error) {
	query := `
	WITH stats_summary AS (
		SELECT 
			COUNT(*) as total_events,
			COUNT(DISTINCT t.session_id) as total_sessions,
			COUNT(DISTINCT fc.file_path) as total_files,
			COALESCE(SUM(CASE WHEN t.author_type = 'ai' THEN fc.lines_added ELSE 0 END), 0) as ai_lines,
			COALESCE(SUM(CASE WHEN t.author_type = 'human' THEN fc.lines_added ELSE 0 END), 0) as human_lines,
			MIN(t.timestamp) as first_event,
			MAX(t.timestamp) as last_event
		FROM tracks t
		LEFT JOIN file_changes fc ON t.id = fc.track_id
		WHERE t.timestamp >= CURRENT_DATE - INTERVAL '30 days'
	)
	SELECT 
		COALESCE(total_events, 0) as total_events,
		COALESCE(total_sessions, 0) as total_sessions,
		COALESCE(total_files, 0) as total_files,
		COALESCE(ai_lines, 0) as ai_lines,
		COALESCE(human_lines, 0) as human_lines,
		COALESCE(ai_lines, 0) + COALESCE(human_lines, 0) as total_lines,
		CASE 
			WHEN (ai_lines + human_lines) > 0 
			THEN ai_lines::FLOAT / (ai_lines + human_lines) * 100
			ELSE 0 
		END as ai_percentage,
		first_event,
		last_event
	FROM stats_summary
	`

	row := s.db.QueryRowContext(ctx, query)
	
	var stats BasicStats
	err := row.Scan(
		&stats.TotalEvents,
		&stats.TotalSessions,
		&stats.TotalFiles,
		&stats.AILines,
		&stats.HumanLines,
		&stats.TotalLines,
		&stats.AIPercentage,
		&stats.FirstEvent,
		&stats.LastEvent,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get basic stats: %w", err)
	}

	return &stats, nil
}

// Close はデータベース接続を閉じる
func (s *DuckDBStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetDatabasePath はデータベースファイルのパスを返す
func (s *DuckDBStorage) GetDatabasePath() string {
	return s.dbPath
}

// GetStatistics は統合された統計情報を返す（StorageInterface実装）
func (s *DuckDBStorage) GetStatistics() (*types.Statistics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	basicStats, err := s.GetBasicStats(ctx)
	if err != nil {
		return nil, err
	}
	
	// DuckDBの詳細統計をtypes.Statisticsに変換
	stats := &types.Statistics{
		TotalEvents:        basicStats.TotalEvents,
		TotalLinesAdded:    basicStats.TotalLines,
		TotalLinesModified: 0, // DuckDBでは正確な計算が必要
		TotalLinesDeleted:  0, // DuckDBでは正確な計算が必要
	}
	
	// AI/Human イベント数の計算（イベントタイプベース）
	query := `
	SELECT 
		SUM(CASE WHEN event_type = 'ai' THEN 1 ELSE 0 END) as ai_events,
		SUM(CASE WHEN event_type = 'human' THEN 1 ELSE 0 END) as human_events,
		SUM(CASE WHEN event_type = 'commit' THEN 1 ELSE 0 END) as commit_events
	FROM tracks
	`
	
	row := s.db.QueryRowContext(ctx, query)
	err = row.Scan(&stats.AIEvents, &stats.HumanEvents, &stats.CommitEvents)
	if err != nil {
		if s.debug {
			log.Printf("Warning: failed to get event counts: %v", err)
		}
	}
	
	// 最初と最後のイベントの取得
	if !basicStats.FirstEvent.IsZero() {
		stats.FirstEvent = &basicStats.FirstEvent
	}
	if !basicStats.LastEvent.IsZero() {
		stats.LastEvent = &basicStats.LastEvent
	}
	
	return stats, nil
}

// ReadEvents はすべてのイベントを読み取る（StorageInterface実装）
func (s *DuckDBStorage) ReadEvents() ([]*types.TrackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := `
	SELECT 
		t.id, t.timestamp, t.event_type, t.author, t.author_type, 
		t.model, t.commit_hash, t.session_id, t.message
	FROM tracks t
	ORDER BY t.timestamp
	`
	
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracks: %w", err)
	}
	defer rows.Close()
	
	var events []*types.TrackEvent
	
	for rows.Next() {
		event := &types.TrackEvent{}
		var authorType, eventTypeStr string
		
		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&eventTypeStr,
			&event.Author,
			&authorType,
			&event.Model,
			&event.CommitHash,
			&event.SessionID,
			&event.Message,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}
		
		// EventTypeの変換
		event.EventType = types.EventType(eventTypeStr)
		
		// ファイル変更情報を取得
		fileQuery := `
		SELECT file_path, lines_added, lines_modified, lines_deleted, file_hash
		FROM file_changes
		WHERE track_id = ?
		`
		
		fileRows, err := s.db.QueryContext(ctx, fileQuery, event.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query file changes: %w", err)
		}
		
		for fileRows.Next() {
			var file types.FileInfo
			err := fileRows.Scan(
				&file.Path,
				&file.LinesAdded,
				&file.LinesModified,
				&file.LinesDeleted,
				&file.Hash,
			)
			if err != nil {
				fileRows.Close()
				return nil, fmt.Errorf("failed to scan file change: %w", err)
			}
			
			event.Files = append(event.Files, file)
		}
		fileRows.Close()
		
		events = append(events, event)
	}
	
	return events, rows.Err()
}

// ReadEventsByDateRange は期間内のイベントを読み取る（StorageInterface実装）
func (s *DuckDBStorage) ReadEventsByDateRange(start, end time.Time) ([]*types.TrackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := `
	SELECT 
		t.id, t.timestamp, t.event_type, t.author, t.author_type, 
		t.model, t.commit_hash, t.session_id, t.message
	FROM tracks t
	WHERE t.timestamp >= ? AND t.timestamp <= ?
	ORDER BY t.timestamp
	`
	
	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracks by date range: %w", err)
	}
	defer rows.Close()
	
	return s.parseTrackRows(ctx, rows)
}

// ReadEventsByAuthor は作成者別のイベントを読み取る（StorageInterface実装）
func (s *DuckDBStorage) ReadEventsByAuthor(author string) ([]*types.TrackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := `
	SELECT 
		t.id, t.timestamp, t.event_type, t.author, t.author_type, 
		t.model, t.commit_hash, t.session_id, t.message
	FROM tracks t
	WHERE t.author = ?
	ORDER BY t.timestamp
	`
	
	rows, err := s.db.QueryContext(ctx, query, author)
	if err != nil {
		return nil, fmt.Errorf("failed to query tracks by author: %w", err)
	}
	defer rows.Close()
	
	return s.parseTrackRows(ctx, rows)
}

// ReadEventsByType はイベントタイプ別のイベントを読み取る（StorageInterface実装）
func (s *DuckDBStorage) ReadEventsByType(eventType types.EventType) ([]*types.TrackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := `
	SELECT 
		t.id, t.timestamp, t.event_type, t.author, t.author_type, 
		t.model, t.commit_hash, t.session_id, t.message
	FROM tracks t
	WHERE t.event_type = ?
	ORDER BY t.timestamp
	`
	
	rows, err := s.db.QueryContext(ctx, query, string(eventType))
	if err != nil {
		return nil, fmt.Errorf("failed to query tracks by type: %w", err)
	}
	defer rows.Close()
	
	return s.parseTrackRows(ctx, rows)
}

// parseTrackRows はクエリ結果をTrackEventに変換するヘルパー
func (s *DuckDBStorage) parseTrackRows(ctx context.Context, rows *sql.Rows) ([]*types.TrackEvent, error) {
	var events []*types.TrackEvent
	
	for rows.Next() {
		event := &types.TrackEvent{}
		var authorType, eventTypeStr string
		
		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&eventTypeStr,
			&event.Author,
			&authorType,
			&event.Model,
			&event.CommitHash,
			&event.SessionID,
			&event.Message,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan track: %w", err)
		}
		
		// EventTypeの変換
		event.EventType = types.EventType(eventTypeStr)
		
		// ファイル変更情報を取得
		fileQuery := `
		SELECT file_path, lines_added, lines_modified, lines_deleted, file_hash
		FROM file_changes
		WHERE track_id = ?
		`
		
		fileRows, err := s.db.QueryContext(ctx, fileQuery, event.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query file changes: %w", err)
		}
		
		for fileRows.Next() {
			var file types.FileInfo
			err := fileRows.Scan(
				&file.Path,
				&file.LinesAdded,
				&file.LinesModified,
				&file.LinesDeleted,
				&file.Hash,
			)
			if err != nil {
				fileRows.Close()
				return nil, fmt.Errorf("failed to scan file change: %w", err)
			}
			
			event.Files = append(event.Files, file)
		}
		fileRows.Close()
		
		events = append(events, event)
	}
	
	return events, rows.Err()
}

// BasicStats は基本統計の構造体
type BasicStats struct {
	TotalEvents   int       `json:"total_events"`
	TotalSessions int       `json:"total_sessions"`
	TotalFiles    int       `json:"total_files"`
	AILines       int       `json:"ai_lines"`
	HumanLines    int       `json:"human_lines"`
	TotalLines    int       `json:"total_lines"`
	AIPercentage  float64   `json:"ai_percentage"`
	FirstEvent    time.Time `json:"first_event"`
	LastEvent     time.Time `json:"last_event"`
}

// TestConnection はデータベース接続をテストする
func (s *DuckDBStorage) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version string
	row := s.db.QueryRowContext(ctx, "SELECT version()")
	if err := row.Scan(&version); err != nil {
		return fmt.Errorf("failed to test connection: %w", err)
	}

	if s.debug {
		log.Printf("🦆 DuckDB version: %s", version)
	}

	return nil
}

// GetDatabaseInfo はデータベースの情報を取得する
func (s *DuckDBStorage) GetDatabaseInfo() (*DatabaseInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info := &DatabaseInfo{
		Path: s.dbPath,
	}

	// ファイルサイズ
	if stat, err := os.Stat(s.dbPath); err == nil {
		info.Size = stat.Size()
		info.ModTime = stat.ModTime()
	}

	// テーブル数とレコード数
	countQuery := `
	SELECT 
		(SELECT COUNT(*) FROM tracks) as track_count,
		(SELECT COUNT(*) FROM file_changes) as file_change_count,
		(SELECT COUNT(*) FROM stats_cache) as cache_count
	`
	
	row := s.db.QueryRowContext(ctx, countQuery)
	err := row.Scan(&info.TrackCount, &info.FileChangeCount, &info.CacheCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get database info: %w", err)
	}

	return info, nil
}

// DatabaseInfo はデータベースの情報
type DatabaseInfo struct {
	Path            string    `json:"path"`
	Size            int64     `json:"size"`
	ModTime         time.Time `json:"mod_time"`
	TrackCount      int       `json:"track_count"`
	FileChangeCount int       `json:"file_change_count"`
	CacheCount      int       `json:"cache_count"`
}

// detectLanguageFromPath はファイルパスから言語を判定する
func detectLanguageFromPath(filePath string) string {
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
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".cpp":
		return "C++"
	case len(filePath) >= 2 && filePath[len(filePath)-2:] == ".c":
		return "C"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".rs":
		return "Rust"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".php":
		return "PHP"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".rb":
		return "Ruby"
	case len(filePath) >= 6 && filePath[len(filePath)-6:] == ".swift":
		return "Swift"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".kt":
		return "Kotlin"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".cs":
		return "C#"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".html":
		return "HTML"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".css":
		return "CSS"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".scss":
		return "SCSS"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".vue":
		return "Vue"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".jsx":
		return "React"
	case len(filePath) >= 3 && filePath[len(filePath)-3:] == ".md":
		return "Markdown"
	case len(filePath) >= 5 && (filePath[len(filePath)-5:] == ".yaml" || filePath[len(filePath)-4:] == ".yml"):
		return "YAML"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".json":
		return "JSON"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".xml":
		return "XML"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".sql":
		return "SQL"
	default:
		return "Other"
	}
}