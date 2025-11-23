package authorship

import (
	"encoding/json"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

const AuthorshipLogVersion = "1.0"

// ToJSON converts AuthorshipLog to JSON bytes
func ToJSON(log *tracker.AuthorshipLog) ([]byte, error) {
	return json.MarshalIndent(log, "", "  ")
}

// FromJSON parses JSON bytes to AuthorshipLog
func FromJSON(data []byte) (*tracker.AuthorshipLog, error) {
	var log tracker.AuthorshipLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}
	return &log, nil
}
