package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/interfaces"
	"github.com/y-hirakaw/ai-code-tracker/internal/security"
)

// Ensure JSONStorageV2 implements the Storage interface
var _ interfaces.Storage = (*JSONStorageV2)(nil)

// JSONStorageV2 is an improved version with security enhancements
type JSONStorageV2 struct {
	baseDir      string
	safeOps      *security.SafeFileOperations
	jsonDecoder  *security.SafeJSONDecoder
}

// NewJSONStorageV2 creates a new improved JSON storage instance
func NewJSONStorageV2(baseDir string) (*JSONStorageV2, error) {
	safeOps, err := security.NewSafeFileOperations(baseDir)
	if err != nil {
		return nil, errors.NewStorageError("NewJSONStorageV2", baseDir, err)
	}
	
	return &JSONStorageV2{
		baseDir:     baseDir,
		safeOps:     safeOps,
		jsonDecoder: security.NewSafeJSONDecoder(security.MaxJSONSize),
	}, nil
}

// Save stores data as JSON with enhanced security
func (js *JSONStorageV2) Save(filename string, data interface{}) error {
	// Validate the file path
	filePath := filepath.Join(js.baseDir, filename)
	validPath, err := js.safeOps.ValidatePath(filePath)
	if err != nil {
		return err
	}
	
	// Ensure directory exists
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.NewStorageError("Save", dir, err)
	}
	
	// Marshal data with size awareness
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.NewStorageError("Save", filename, err)
	}
	
	// Check size before writing
	if len(jsonData) > security.MaxJSONSize {
		return errors.NewValidationError(
			"Save",
			filename,
			fmt.Sprintf("JSON data size %d exceeds maximum allowed size %d", len(jsonData), security.MaxJSONSize),
		)
	}
	
	// Write to file
	if err := os.WriteFile(validPath, jsonData, 0644); err != nil {
		return errors.NewStorageError("Save", filename, err)
	}
	
	return nil
}

// Load retrieves and unmarshals JSON data with security checks
func (js *JSONStorageV2) Load(filename string, data interface{}) error {
	// Validate the file path
	filePath := filepath.Join(js.baseDir, filename)
	validPath, err := js.safeOps.ValidatePath(filePath)
	if err != nil {
		return err
	}
	
	// Open file
	file, err := os.Open(validPath)
	if err != nil {
		return errors.NewStorageError("Load", filename, err)
	}
	defer file.Close()
	
	// Use safe JSON decoder
	if err := js.jsonDecoder.Decode(file, data); err != nil {
		return errors.NewStorageError("Load", filename, err)
	}
	
	return nil
}

// Exists checks if a file exists with path validation
func (js *JSONStorageV2) Exists(filename string) bool {
	filePath := filepath.Join(js.baseDir, filename)
	validPath, err := js.safeOps.ValidatePath(filePath)
	if err != nil {
		return false
	}
	
	_, err = os.Stat(validPath)
	return err == nil
}

// Delete removes a file with security checks
func (js *JSONStorageV2) Delete(filename string) error {
	filePath := filepath.Join(js.baseDir, filename)
	validPath, err := js.safeOps.ValidatePath(filePath)
	if err != nil {
		return err
	}
	
	if err := os.Remove(validPath); err != nil {
		return errors.NewStorageError("Delete", filename, err)
	}
	
	return nil
}

// List returns files matching the pattern with security validation
func (js *JSONStorageV2) List(pattern string) ([]string, error) {
	// Validate pattern doesn't contain dangerous characters
	if strings.Contains(pattern, "..") {
		return nil, errors.NewValidationError("List", "pattern", "path traversal not allowed")
	}
	
	searchPath := filepath.Join(js.baseDir, pattern)
	matches, err := filepath.Glob(searchPath)
	if err != nil {
		return nil, errors.NewStorageError("List", pattern, err)
	}
	
	// Convert to relative paths and validate each
	var result []string
	for _, match := range matches {
		// Validate each matched path
		if _, err := js.safeOps.ValidatePath(match); err != nil {
			continue // Skip invalid paths
		}
		
		rel, err := filepath.Rel(js.baseDir, match)
		if err != nil {
			continue
		}
		result = append(result, rel)
	}
	
	return result, nil
}