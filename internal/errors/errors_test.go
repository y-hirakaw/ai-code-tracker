package errors

import (
	"errors"
	"testing"
)

func TestAICTError(t *testing.T) {
	tests := []struct {
		name     string
		err      *AICTError
		expected string
	}{
		{
			name: "StorageError with path",
			err:  NewStorageError("Save", "/tmp/test.json", errors.New("permission denied")),
			expected: "Save: storage operation failed [/tmp/test.json]: permission denied",
		},
		{
			name: "GitError without path",
			err:  NewGitError("GetDiff", "repository not found", errors.New("exit status 128")),
			expected: "GetDiff: repository not found: exit status 128",
		},
		{
			name: "ValidationError",
			err:  NewValidationError("Validate", "percentage", "must be between 0 and 100"),
			expected: "Validate: must be between 0 and 100 [percentage]: <nil>",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAICTError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	aictErr := NewStorageError("Test", "file.json", baseErr)
	
	if unwrapped := aictErr.Unwrap(); unwrapped != baseErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, baseErr)
	}
}

func TestAICTError_Is(t *testing.T) {
	err1 := &AICTError{Type: ErrTypeStorage}
	err2 := &AICTError{Type: ErrTypeStorage}
	err3 := &AICTError{Type: ErrTypeGit}
	
	tests := []struct {
		name   string
		err    *AICTError
		target error
		want   bool
	}{
		{
			name:   "Same type",
			err:    err1,
			target: err2,
			want:   true,
		},
		{
			name:   "Different type",
			err:    err1,
			target: err3,
			want:   false,
		},
		{
			name:   "Nil target",
			err:    err1,
			target: nil,
			want:   false,
		},
		{
			name:   "Non-AICTError target",
			err:    err1,
			target: errors.New("other error"),
			want:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Is(tt.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorConstructors(t *testing.T) {
	t.Run("NewStorageError", func(t *testing.T) {
		err := NewStorageError("Read", "data.json", errors.New("file not found"))
		if err.Type != ErrTypeStorage {
			t.Errorf("Type = %v, want %v", err.Type, ErrTypeStorage)
		}
		if err.Op != "Read" {
			t.Errorf("Op = %v, want %v", err.Op, "Read")
		}
		if err.Path != "data.json" {
			t.Errorf("Path = %v, want %v", err.Path, "data.json")
		}
	})
	
	t.Run("NewGitError", func(t *testing.T) {
		err := NewGitError("Commit", "failed to commit", errors.New("no changes"))
		if err.Type != ErrTypeGit {
			t.Errorf("Type = %v, want %v", err.Type, ErrTypeGit)
		}
		if err.Message != "failed to commit" {
			t.Errorf("Message = %v, want %v", err.Message, "failed to commit")
		}
	})
	
	t.Run("NewConfigError", func(t *testing.T) {
		err := NewConfigError("Load", "invalid format", errors.New("json error"))
		if err.Type != ErrTypeConfig {
			t.Errorf("Type = %v, want %v", err.Type, ErrTypeConfig)
		}
	})
	
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("Validate", "email", "invalid email format")
		if err.Type != ErrTypeValidation {
			t.Errorf("Type = %v, want %v", err.Type, ErrTypeValidation)
		}
		if err.Path != "email" {
			t.Errorf("Path = %v, want %v", err.Path, "email")
		}
		if err.Err != nil {
			t.Errorf("Err = %v, want nil", err.Err)
		}
	})
}