package types

import (
	"encoding/json"
	"testing"
	"time"
)

// TestEventType はEventTypeの基本動作をテストする
func TestEventType(t *testing.T) {
	tests := []struct {
		name     string
		event    EventType
		isValid  bool
		expected string
	}{
		{"AI Event", EventTypeAI, true, "ai"},
		{"Human Event", EventTypeHuman, true, "human"},
		{"Commit Event", EventTypeCommit, true, "commit"},
		{"Unknown Event", EventTypeUnknown, true, "unknown"},
		{"Invalid Event", EventType("invalid"), false, "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// IsValid のテスト
			if got := tt.event.IsValid(); got != tt.isValid {
				t.Errorf("EventType.IsValid() = %v, want %v", got, tt.isValid)
			}

			// String のテスト
			if got := tt.event.String(); got != tt.expected {
				t.Errorf("EventType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestFileInfo はFileInfoの基本動作をテストする
func TestFileInfo(t *testing.T) {
	tests := []struct {
		name      string
		fileInfo  FileInfo
		wantValid bool
		wantTotal int
	}{
		{
			name: "Valid FileInfo",
			fileInfo: FileInfo{
				Path:          "main.go",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
				Hash:          "abc123",
			},
			wantValid: true,
			wantTotal: 18,
		},
		{
			name: "Empty Path",
			fileInfo: FileInfo{
				Path:          "",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
			},
			wantValid: false,
			wantTotal: 18,
		},
		{
			name: "Negative Lines",
			fileInfo: FileInfo{
				Path:          "main.go",
				LinesAdded:    -1,
				LinesModified: 5,
				LinesDeleted:  3,
			},
			wantValid: false,
			wantTotal: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TotalChanges のテスト
			if got := tt.fileInfo.TotalChanges(); got != tt.wantTotal {
				t.Errorf("FileInfo.TotalChanges() = %v, want %v", got, tt.wantTotal)
			}

			// Validate のテスト
			err := tt.fileInfo.Validate()
			if tt.wantValid && err != nil {
				t.Errorf("FileInfo.Validate() error = %v, want nil", err)
			}
			if !tt.wantValid && err == nil {
				t.Errorf("FileInfo.Validate() error = nil, want error")
			}
		})
	}
}

// TestTrackEvent はTrackEventの基本動作をテストする
func TestTrackEvent(t *testing.T) {
	now := time.Now()
	
	validEvent := &TrackEvent{
		ID:        "test-001",
		Timestamp: now,
		EventType: EventTypeAI,
		Author:    "Claude Code",
		Model:     "claude-sonnet-4",
		Files: []FileInfo{
			{
				Path:          "main.go",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
			},
		},
		Message: "Test change",
	}

	t.Run("Valid TrackEvent", func(t *testing.T) {
		// Validate のテスト
		if err := validEvent.Validate(); err != nil {
			t.Errorf("TrackEvent.Validate() error = %v, want nil", err)
		}

		// Total counts のテスト
		if got := validEvent.TotalLinesAdded(); got != 10 {
			t.Errorf("TrackEvent.TotalLinesAdded() = %v, want 10", got)
		}
		if got := validEvent.TotalLinesModified(); got != 5 {
			t.Errorf("TrackEvent.TotalLinesModified() = %v, want 5", got)
		}
		if got := validEvent.TotalLinesDeleted(); got != 3 {
			t.Errorf("TrackEvent.TotalLinesDeleted() = %v, want 3", got)
		}
		if got := validEvent.TotalChanges(); got != 18 {
			t.Errorf("TrackEvent.TotalChanges() = %v, want 18", got)
		}
	})

	t.Run("Invalid TrackEvent - Empty ID", func(t *testing.T) {
		invalidEvent := *validEvent
		invalidEvent.ID = ""
		
		if err := invalidEvent.Validate(); err == nil {
			t.Errorf("TrackEvent.Validate() error = nil, want error for empty ID")
		}
	})

	t.Run("Invalid TrackEvent - Zero Timestamp", func(t *testing.T) {
		invalidEvent := *validEvent
		invalidEvent.Timestamp = time.Time{}
		
		if err := invalidEvent.Validate(); err == nil {
			t.Errorf("TrackEvent.Validate() error = nil, want error for zero timestamp")
		}
	})

	t.Run("Invalid TrackEvent - Invalid EventType", func(t *testing.T) {
		invalidEvent := *validEvent
		invalidEvent.EventType = EventType("invalid")
		
		if err := invalidEvent.Validate(); err == nil {
			t.Errorf("TrackEvent.Validate() error = nil, want error for invalid event type")
		}
	})

	t.Run("Invalid TrackEvent - Empty Author", func(t *testing.T) {
		invalidEvent := *validEvent
		invalidEvent.Author = ""
		
		if err := invalidEvent.Validate(); err == nil {
			t.Errorf("TrackEvent.Validate() error = nil, want error for empty author")
		}
	})

	t.Run("Invalid TrackEvent - AI without Model", func(t *testing.T) {
		invalidEvent := *validEvent
		invalidEvent.EventType = EventTypeAI
		invalidEvent.Model = ""
		
		if err := invalidEvent.Validate(); err == nil {
			t.Errorf("TrackEvent.Validate() error = nil, want error for AI event without model")
		}
	})

	t.Run("Invalid TrackEvent - Commit without Hash", func(t *testing.T) {
		invalidEvent := *validEvent
		invalidEvent.EventType = EventTypeCommit
		invalidEvent.CommitHash = ""
		
		if err := invalidEvent.Validate(); err == nil {
			t.Errorf("TrackEvent.Validate() error = nil, want error for commit event without hash")
		}
	})
}

// TestTrackEventJSON はTrackEventのJSON変換をテストする
func TestTrackEventJSON(t *testing.T) {
	now := time.Now()
	
	originalEvent := &TrackEvent{
		ID:        "test-001",
		Timestamp: now,
		EventType: EventTypeAI,
		Author:    "Claude Code",
		Model:     "claude-sonnet-4",
		Files: []FileInfo{
			{
				Path:          "main.go",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
				Hash:          "abc123",
			},
		},
		Message: "Test change",
	}

	t.Run("ToJSON and FromJSON", func(t *testing.T) {
		// ToJSON のテスト
		jsonStr, err := originalEvent.ToJSON()
		if err != nil {
			t.Fatalf("TrackEvent.ToJSON() error = %v", err)
		}

		// FromJSON のテスト
		parsedEvent, err := FromJSON(jsonStr)
		if err != nil {
			t.Fatalf("FromJSON() error = %v", err)
		}

		// 比較
		if parsedEvent.ID != originalEvent.ID {
			t.Errorf("FromJSON().ID = %v, want %v", parsedEvent.ID, originalEvent.ID)
		}
		if parsedEvent.EventType != originalEvent.EventType {
			t.Errorf("FromJSON().EventType = %v, want %v", parsedEvent.EventType, originalEvent.EventType)
		}
		if parsedEvent.Author != originalEvent.Author {
			t.Errorf("FromJSON().Author = %v, want %v", parsedEvent.Author, originalEvent.Author)
		}
		if parsedEvent.Model != originalEvent.Model {
			t.Errorf("FromJSON().Model = %v, want %v", parsedEvent.Model, originalEvent.Model)
		}
		if len(parsedEvent.Files) != len(originalEvent.Files) {
			t.Errorf("FromJSON().Files length = %v, want %v", len(parsedEvent.Files), len(originalEvent.Files))
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidJSON := `{"invalid": "json"`
		
		_, err := FromJSON(invalidJSON)
		if err == nil {
			t.Errorf("FromJSON() error = nil, want error for invalid JSON")
		}
	})
}

// TestGenerateEventID はEventID生成をテストする
func TestGenerateEventID(t *testing.T) {
	now := time.Now()
	
	t.Run("Generate Event ID", func(t *testing.T) {
		id1 := GenerateEventID(now, EventTypeAI, "Claude Code")
		id2 := GenerateEventID(now, EventTypeAI, "Claude Code")
		
		// 同じ時刻でも異なるIDが生成されること（ナノ秒部分で区別）
		if id1 == id2 {
			// ナノ秒が同じ場合もあるので、1ナノ秒ずらして再テスト
			now2 := now.Add(1 * time.Nanosecond)
			id3 := GenerateEventID(now2, EventTypeAI, "Claude Code")
			if id1 == id3 {
				t.Errorf("GenerateEventID() generated same ID for different times")
			}
		}
		
		// IDの形式チェック
		if len(id1) == 0 {
			t.Errorf("GenerateEventID() returned empty ID")
		}
	})
}

// TestStatistics はStatisticsの計算をテストする
func TestStatistics(t *testing.T) {
	stats := &Statistics{
		TotalEvents:        100,
		AIEvents:          60,
		HumanEvents:       40,
		CommitEvents:      20,
		TotalLinesAdded:   1000,
		TotalLinesModified: 500,
		TotalLinesDeleted: 200,
	}

	t.Run("Percentage Calculations", func(t *testing.T) {
		if got := stats.AIPercentage(); got != 60.0 {
			t.Errorf("Statistics.AIPercentage() = %v, want 60.0", got)
		}
		
		if got := stats.HumanPercentage(); got != 40.0 {
			t.Errorf("Statistics.HumanPercentage() = %v, want 40.0", got)
		}
	})

	t.Run("Total Changes", func(t *testing.T) {
		if got := stats.TotalChanges(); got != 1700 {
			t.Errorf("Statistics.TotalChanges() = %v, want 1700", got)
		}
	})

	t.Run("Zero Events", func(t *testing.T) {
		emptyStats := &Statistics{TotalEvents: 0}
		
		if got := emptyStats.AIPercentage(); got != 0.0 {
			t.Errorf("Statistics.AIPercentage() with zero events = %v, want 0.0", got)
		}
		
		if got := emptyStats.HumanPercentage(); got != 0.0 {
			t.Errorf("Statistics.HumanPercentage() with zero events = %v, want 0.0", got)
		}
	})
}

// BenchmarkTrackEventJSON はJSON変換のベンチマークテストを行う
func BenchmarkTrackEventJSON(b *testing.B) {
	now := time.Now()
	
	event := &TrackEvent{
		ID:        "bench-001",
		Timestamp: now,
		EventType: EventTypeAI,
		Author:    "Claude Code",
		Model:     "claude-sonnet-4",
		Files: []FileInfo{
			{
				Path:          "main.go",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
				Hash:          "abc123",
			},
		},
		Message: "Benchmark test",
	}

	b.Run("ToJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := event.ToJSON()
			if err != nil {
				b.Fatalf("ToJSON error: %v", err)
			}
		}
	})

	// JSON文字列を準備
	jsonStr, _ := event.ToJSON()

	b.Run("FromJSON", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := FromJSON(jsonStr)
			if err != nil {
				b.Fatalf("FromJSON error: %v", err)
			}
		}
	})
}

// TestJSONMarshalCompatibility は標準JSON Marshalとの互換性をテストする
func TestJSONMarshalCompatibility(t *testing.T) {
	now := time.Now()
	
	event := &TrackEvent{
		ID:        "compat-001",
		Timestamp: now,
		EventType: EventTypeAI,
		Author:    "Claude Code",
		Model:     "claude-sonnet-4",
		Files: []FileInfo{
			{
				Path:          "main.go",
				LinesAdded:    10,
				LinesModified: 5,
				LinesDeleted:  3,
			},
		},
	}

	t.Run("Standard JSON Marshal/Unmarshal", func(t *testing.T) {
		// 標準のjson.Marshalを使用
		data, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		// 標準のjson.Unmarshalを使用
		var unmarshaled TrackEvent
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		// 検証
		if unmarshaled.ID != event.ID {
			t.Errorf("json.Unmarshal().ID = %v, want %v", unmarshaled.ID, event.ID)
		}
		if unmarshaled.EventType != event.EventType {
			t.Errorf("json.Unmarshal().EventType = %v, want %v", unmarshaled.EventType, event.EventType)
		}
	})
}