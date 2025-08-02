package errors

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType int

const (
	// ErrTypeStorage indicates storage-related errors
	ErrTypeStorage ErrorType = iota
	// ErrTypeGit indicates git-related errors
	ErrTypeGit
	// ErrTypeConfig indicates configuration errors
	ErrTypeConfig
	// ErrTypeAnalysis indicates analysis errors
	ErrTypeAnalysis
	// ErrTypeValidation indicates validation errors
	ErrTypeValidation
)

// AICTError represents a structured error in AI Code Tracker
type AICTError struct {
	Type    ErrorType
	Op      string // Operation that failed
	Path    string // File path or resource involved
	Message string // Human-readable error message
	Err     error  // Underlying error
}

// Error implements the error interface
func (e *AICTError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s [%s]: %v", e.Op, e.Message, e.Path, e.Err)
	}
	return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
}

// Unwrap returns the underlying error
func (e *AICTError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *AICTError) Is(target error) bool {
	if target == nil {
		return false
	}
	
	if te, ok := target.(*AICTError); ok {
		return e.Type == te.Type
	}
	
	return false
}

// NewStorageError creates a new storage error
func NewStorageError(op, path string, err error) *AICTError {
	return &AICTError{
		Type:    ErrTypeStorage,
		Op:      op,
		Path:    path,
		Message: "storage operation failed",
		Err:     err,
	}
}

// NewGitError creates a new git error
func NewGitError(op, message string, err error) *AICTError {
	return &AICTError{
		Type:    ErrTypeGit,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// NewConfigError creates a new configuration error
func NewConfigError(op, message string, err error) *AICTError {
	return &AICTError{
		Type:    ErrTypeConfig,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(op, field, message string) *AICTError {
	return &AICTError{
		Type:    ErrTypeValidation,
		Op:      op,
		Path:    field,
		Message: message,
	}
}

// NewAnalysisError creates a new analysis error
func NewAnalysisError(op, message string, err error) *AICTError {
	return &AICTError{
		Type:    ErrTypeAnalysis,
		Op:      op,
		Message: message,
		Err:     err,
	}
}