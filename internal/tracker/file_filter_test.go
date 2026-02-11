package tracker

import "testing"

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		fpath    string
		pattern  string
		expected bool
	}{
		{"suffix wildcard match", "main_test.go", "*_test.go", true},
		{"suffix wildcard nested", "pkg/foo_test.go", "*_test.go", true},
		{"suffix wildcard no match", "main.go", "*_test.go", false},
		{"prefix wildcard match", "vendor/lib/foo.go", "vendor/*", true},
		{"prefix wildcard no match", "src/main.go", "vendor/*", false},
		{"exact match", "Makefile", "Makefile", true},
		{"exact no match", "makefile", "Makefile", false},
		{"empty pattern", "foo.go", "", false},
		{"empty fpath suffix", "", "*_test.go", false},
		{"empty fpath prefix", "", "vendor/*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesPattern(tt.fpath, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.fpath, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsTrackedFile(t *testing.T) {
	cfg := &Config{
		TrackedExtensions: []string{".go", ".py", ".js"},
		ExcludePatterns:   []string{"*_test.go", "vendor/*"},
	}

	tests := []struct {
		name     string
		fpath    string
		expected bool
	}{
		{"go file tracked", "main.go", true},
		{"py file tracked", "script.py", true},
		{"js file tracked", "app.js", true},
		{"nested go tracked", "internal/pkg/handler.go", true},
		{"test excluded", "main_test.go", false},
		{"nested test excluded", "pkg/handler_test.go", false},
		{"vendor excluded", "vendor/lib/foo.go", false},
		{"md not tracked", "README.md", false},
		{"no extension", "Makefile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTrackedFile(tt.fpath, cfg)
			if result != tt.expected {
				t.Errorf("IsTrackedFile(%q) = %v, want %v", tt.fpath, result, tt.expected)
			}
		})
	}
}

func TestIsTrackedFile_EmptyConfig(t *testing.T) {
	cfg := &Config{
		TrackedExtensions: []string{},
		ExcludePatterns:   []string{},
	}

	if IsTrackedFile("main.go", cfg) {
		t.Error("IsTrackedFile should return false with empty extensions")
	}
}
