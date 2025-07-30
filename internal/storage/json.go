package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type JSONStorage struct {
	baseDir string
}

func NewJSONStorage(baseDir string) *JSONStorage {
	return &JSONStorage{
		baseDir: baseDir,
	}
}

func (js *JSONStorage) Save(filename string, data interface{}) error {
	filePath := filepath.Join(js.baseDir, filename)
	
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (js *JSONStorage) Load(filename string, data interface{}) error {
	filePath := filepath.Join(js.baseDir, filename)
	
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(fileData, data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

func (js *JSONStorage) Exists(filename string) bool {
	filePath := filepath.Join(js.baseDir, filename)
	_, err := os.Stat(filePath)
	return err == nil
}

func (js *JSONStorage) Delete(filename string) error {
	filePath := filepath.Join(js.baseDir, filename)
	return os.Remove(filePath)
}

func (js *JSONStorage) List(pattern string) ([]string, error) {
	searchPattern := filepath.Join(js.baseDir, pattern)
	matches, err := filepath.Glob(searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	result := make([]string, len(matches))
	for i, match := range matches {
		result[i], _ = filepath.Rel(js.baseDir, match)
	}

	return result, nil
}