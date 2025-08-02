package security

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
)

const (
	// MaxJSONSize is the maximum size for JSON parsing (10MB)
	MaxJSONSize = 10 * 1024 * 1024
	
	// MaxPathDepth is the maximum allowed path depth
	MaxPathDepth = 20
)

// SafeCommandExecutor provides secure command execution
type SafeCommandExecutor struct {
	allowedCommands map[string]bool
}

// NewSafeCommandExecutor creates a new safe command executor
func NewSafeCommandExecutor() *SafeCommandExecutor {
	return &SafeCommandExecutor{
		allowedCommands: map[string]bool{
			"git": true,
		},
	}
}

// IsCommandAllowed checks if a command is allowed to execute
func (e *SafeCommandExecutor) IsCommandAllowed(cmd string) bool {
	return e.allowedCommands[cmd]
}

// ValidateCommandArgs validates command arguments for safety
func (e *SafeCommandExecutor) ValidateCommandArgs(args []string) error {
	for _, arg := range args {
		// Check for shell injection attempts
		if containsShellMetacharacters(arg) {
			return errors.NewValidationError(
				"ValidateCommandArgs",
				"argument",
				fmt.Sprintf("potentially unsafe argument: %s", arg),
			)
		}
		
		// Check for path traversal attempts
		if strings.Contains(arg, "..") {
			return errors.NewValidationError(
				"ValidateCommandArgs",
				"argument",
				"path traversal not allowed in arguments",
			)
		}
	}
	
	return nil
}

// containsShellMetacharacters checks for shell metacharacters
func containsShellMetacharacters(s string) bool {
	// Check for common shell metacharacters
	dangerous := []string{
		";", "&", "|", "`", "$", "(", ")", "{", "}", "<", ">", "!", "\\", "'", "\"",
		"\n", "\r", "\x00",
	}
	
	for _, char := range dangerous {
		if strings.Contains(s, char) {
			return true
		}
	}
	
	return false
}

// SafeJSONDecoder provides size-limited JSON decoding
type SafeJSONDecoder struct {
	maxSize int64
}

// NewSafeJSONDecoder creates a new safe JSON decoder
func NewSafeJSONDecoder(maxSize int64) *SafeJSONDecoder {
	if maxSize <= 0 {
		maxSize = MaxJSONSize
	}
	return &SafeJSONDecoder{
		maxSize: maxSize,
	}
}

// Decode safely decodes JSON with size limits
func (d *SafeJSONDecoder) Decode(r io.Reader, v interface{}) error {
	// Limit the reader to prevent excessive memory usage
	limitedReader := io.LimitReader(r, d.maxSize)
	
	// Decode the JSON
	decoder := json.NewDecoder(limitedReader)
	if err := decoder.Decode(v); err != nil {
		return errors.NewStorageError("Decode", "", err)
	}
	
	// Check if we hit the size limit
	var dummy json.RawMessage
	if err := decoder.Decode(&dummy); err != io.EOF {
		return errors.NewValidationError(
			"Decode",
			"json",
			fmt.Sprintf("JSON size exceeds maximum allowed size of %d bytes", d.maxSize),
		)
	}
	
	return nil
}

// SafeFileOperations provides secure file operations
type SafeFileOperations struct {
	baseDir string
}

// NewSafeFileOperations creates a new safe file operations handler
func NewSafeFileOperations(baseDir string) (*SafeFileOperations, error) {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, errors.NewStorageError("NewSafeFileOperations", baseDir, err)
	}
	
	return &SafeFileOperations{
		baseDir: absBase,
	}, nil
}

// ValidatePath ensures the path is safe and within bounds
func (f *SafeFileOperations) ValidatePath(path string) (string, error) {
	// Clean the path
	cleaned := filepath.Clean(path)
	
	// Make it absolute
	absPath, err := filepath.Abs(cleaned)
	if err != nil {
		return "", errors.NewStorageError("ValidatePath", path, err)
	}
	
	// Ensure it's within the base directory
	if !strings.HasPrefix(absPath, f.baseDir) {
		return "", errors.NewValidationError(
			"ValidatePath",
			"path",
			"path is outside the allowed directory",
		)
	}
	
	// Check path depth
	relPath, _ := filepath.Rel(f.baseDir, absPath)
	depth := strings.Count(relPath, string(os.PathSeparator))
	if depth > MaxPathDepth {
		return "", errors.NewValidationError(
			"ValidatePath",
			"path",
			fmt.Sprintf("path depth %d exceeds maximum allowed depth %d", depth, MaxPathDepth),
		)
	}
	
	return absPath, nil
}

// SafeRemoveAll safely removes a directory after validation
func (f *SafeFileOperations) SafeRemoveAll(path string) error {
	// Validate the path
	validPath, err := f.ValidatePath(path)
	if err != nil {
		return err
	}
	
	// Check if the path exists
	info, err := os.Stat(validPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return errors.NewStorageError("SafeRemoveAll", path, err)
	}
	
	// Only remove if it's a directory
	if !info.IsDir() {
		return errors.NewValidationError(
			"SafeRemoveAll",
			"path",
			"path is not a directory",
		)
	}
	
	// Additional check: don't remove critical directories
	if isCriticalDirectory(validPath) {
		return errors.NewValidationError(
			"SafeRemoveAll",
			"path",
			"cannot remove critical directory",
		)
	}
	
	// Perform the removal
	if err := os.RemoveAll(validPath); err != nil {
		return errors.NewStorageError("SafeRemoveAll", path, err)
	}
	
	return nil
}

// isCriticalDirectory checks if a directory is critical and should not be removed
func isCriticalDirectory(path string) bool {
	base := filepath.Base(path)
	critical := []string{
		".git",
		".ai_code_tracking",
		"node_modules",
		"vendor",
	}
	
	for _, c := range critical {
		if base == c {
			return true
		}
	}
	
	return false
}