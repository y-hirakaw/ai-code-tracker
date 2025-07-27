package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// EventType はトラッキングイベントの種類を表す
type EventType string

const (
	// EventTypeAI はAIによって生成された変更を表す
	EventTypeAI EventType = "ai"
	// EventTypeHuman は人間によって書かれた変更を表す
	EventTypeHuman EventType = "human"
	// EventTypeCommit はgitコミットイベントを表す
	EventTypeCommit EventType = "commit"
	// EventTypeUnknown は識別できない変更を表す
	EventTypeUnknown EventType = "unknown"
)

// IsValid はEventTypeが有効かどうかをチェックする
func (e EventType) IsValid() bool {
	switch e {
	case EventTypeAI, EventTypeHuman, EventTypeCommit, EventTypeUnknown:
		return true
	default:
		return false
	}
}

// String はEventTypeの文字列表現を返す
func (e EventType) String() string {
	return string(e)
}

// FileInfo は特定のファイルの変更に関する情報を表す
type FileInfo struct {
	// Path はリポジトリルートからの相対パス
	Path string `json:"path"`
	// LinesAdded は追加された行数
	LinesAdded int `json:"lines_added"`
	// LinesModified は変更された行数
	LinesModified int `json:"lines_modified"`
	// LinesDeleted は削除された行数
	LinesDeleted int `json:"lines_deleted"`
	// Hash は変更後のファイル内容のハッシュ
	Hash string `json:"hash,omitempty"`
}

// TotalChanges は行変更の総数を返す
func (f *FileInfo) TotalChanges() int {
	return f.LinesAdded + f.LinesModified + f.LinesDeleted
}

// Validate はFileInfoが有効かどうかをチェックする
func (f *FileInfo) Validate() error {
	if f.Path == "" {
		return fmt.Errorf("ファイルパスは空にできません")
	}
	if f.LinesAdded < 0 || f.LinesModified < 0 || f.LinesDeleted < 0 {
		return fmt.Errorf("行数は負の値にできません")
	}
	return nil
}

// TrackEvent は単一のトラッキングイベントを表す
type TrackEvent struct {
	// ID はこのイベントの一意識別子
	ID string `json:"id"`
	// Timestamp はイベントが発生した時刻
	Timestamp time.Time `json:"timestamp"`
	// EventType はイベントの種類（AI、人間、コミットなど）を示す
	EventType EventType `json:"event_type"`
	// Author は変更を行った人またはAI
	Author string `json:"author"`
	// Model はAIモデル名（AIイベントのみ）
	Model string `json:"model,omitempty"`
	// CommitHash はgitコミットハッシュ（コミットイベントのみ）
	CommitHash string `json:"commit_hash,omitempty"`
	// Files は変更されたファイルの情報を含む
	Files []FileInfo `json:"files"`
	// Message は任意の説明またはコミットメッセージ
	Message string `json:"message,omitempty"`
	// SessionID は関連するイベントをグループ化する
	SessionID string `json:"session_id,omitempty"`
}

// TotalLinesAdded はすべてのファイルにわたって追加された行の総数を返す
func (t *TrackEvent) TotalLinesAdded() int {
	total := 0
	for _, file := range t.Files {
		total += file.LinesAdded
	}
	return total
}

// TotalLinesModified はすべてのファイルにわたって変更された行の総数を返す
func (t *TrackEvent) TotalLinesModified() int {
	total := 0
	for _, file := range t.Files {
		total += file.LinesModified
	}
	return total
}

// TotalLinesDeleted はすべてのファイルにわたって削除された行の総数を返す
func (t *TrackEvent) TotalLinesDeleted() int {
	total := 0
	for _, file := range t.Files {
		total += file.LinesDeleted
	}
	return total
}

// TotalChanges はすべてのファイルにわたって変更された行の総数を返す
func (t *TrackEvent) TotalChanges() int {
	return t.TotalLinesAdded() + t.TotalLinesModified() + t.TotalLinesDeleted()
}

// Validate はTrackEventが有効かどうかをチェックする
func (t *TrackEvent) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("イベントIDは空にできません")
	}
	if t.Timestamp.IsZero() {
		return fmt.Errorf("タイムスタンプはゼロにできません")
	}
	if !t.EventType.IsValid() {
		return fmt.Errorf("無効なイベントタイプ: %s", t.EventType)
	}
	if t.Author == "" {
		return fmt.Errorf("作者は空にできません")
	}
	if t.EventType == EventTypeAI && t.Model == "" {
		return fmt.Errorf("AIイベントではモデルが必須です")
	}
	if t.EventType == EventTypeCommit && t.CommitHash == "" {
		return fmt.Errorf("コミットイベントではコミットハッシュが必須です")
	}
	for i, file := range t.Files {
		if err := file.Validate(); err != nil {
			return fmt.Errorf("ファイル %d の検証に失敗しました: %w", i, err)
		}
	}
	return nil
}

// ToJSON はTrackEventをJSON文字列に変換する
func (t *TrackEvent) ToJSON() (string, error) {
	if err := t.Validate(); err != nil {
		return "", fmt.Errorf("検証に失敗しました: %w", err)
	}
	data, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("JSONマーシャルに失敗しました: %w", err)
	}
	return string(data), nil
}

// FromJSON はJSON文字列からTrackEventを作成する
func FromJSON(jsonStr string) (*TrackEvent, error) {
	var event TrackEvent
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		return nil, fmt.Errorf("JSONアンマーシャルに失敗しました: %w", err)
	}
	if err := event.Validate(); err != nil {
		return nil, fmt.Errorf("検証に失敗しました: %w", err)
	}
	return &event, nil
}

// GenerateEventID はタイムスタンプと内容に基づいて一意のイベントIDを生成する
func GenerateEventID(timestamp time.Time, eventType EventType, author string) string {
	return fmt.Sprintf("%s_%s_%s_%d", 
		timestamp.Format("20060102_150405"), 
		eventType, 
		author, 
		timestamp.UnixNano()%1000000)
}

// Statistics はトラッキングデータの集計統計を表す
type Statistics struct {
	// TotalEvents はトラッキングイベントの総数
	TotalEvents int `json:"total_events"`
	// AIEvents はAIによって生成されたイベント数
	AIEvents int `json:"ai_events"`
	// HumanEvents は人間によって生成されたイベント数
	HumanEvents int `json:"human_events"`
	// CommitEvents はコミットイベント数
	CommitEvents int `json:"commit_events"`
	// TotalLinesAdded はすべてのイベントにわたって追加された行数
	TotalLinesAdded int `json:"total_lines_added"`
	// TotalLinesModified はすべてのイベントにわたって変更された行数
	TotalLinesModified int `json:"total_lines_modified"`
	// TotalLinesDeleted はすべてのイベントにわたって削除された行数
	TotalLinesDeleted int `json:"total_lines_deleted"`
	// FirstEvent は最初のイベントのタイムスタンプ
	FirstEvent *time.Time `json:"first_event,omitempty"`
	// LastEvent は最後のイベントのタイムスタンプ
	LastEvent *time.Time `json:"last_event,omitempty"`
}

// AIPercentage はAIによって生成された変更のパーセンテージを返す
func (s *Statistics) AIPercentage() float64 {
	if s.TotalEvents == 0 {
		return 0.0
	}
	return float64(s.AIEvents) / float64(s.TotalEvents) * 100.0
}

// HumanPercentage は人間によって生成された変更のパーセンテージを返す
func (s *Statistics) HumanPercentage() float64 {
	if s.TotalEvents == 0 {
		return 0.0
	}
	return float64(s.HumanEvents) / float64(s.TotalEvents) * 100.0
}

// TotalChanges は行変更の総数を返す
func (s *Statistics) TotalChanges() int {
	return s.TotalLinesAdded + s.TotalLinesModified + s.TotalLinesDeleted
}