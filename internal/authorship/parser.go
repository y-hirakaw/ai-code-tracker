package authorship

import (
	"fmt"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// ValidateAuthorshipLog validates an AuthorshipLog structure
func ValidateAuthorshipLog(log *tracker.AuthorshipLog) error {
	if log.Version == "" {
		return fmt.Errorf("version is required")
	}

	if log.Commit == "" {
		return fmt.Errorf("commit hash is required")
	}

	if log.Version != AuthorshipLogVersion {
		return fmt.Errorf("unsupported version: %s (expected: %s)", log.Version, AuthorshipLogVersion)
	}

	// Validate file info
	for filepath, fileInfo := range log.Files {
		if len(fileInfo.Authors) == 0 {
			return fmt.Errorf("file %s has no authors", filepath)
		}

		for _, author := range fileInfo.Authors {
			if author.Name == "" {
				return fmt.Errorf("file %s has author with empty name", filepath)
			}

			if author.Type != tracker.AuthorTypeHuman && author.Type != tracker.AuthorTypeAI {
				return fmt.Errorf("file %s has invalid author type: %s", filepath, author.Type)
			}
		}
	}

	return nil
}
