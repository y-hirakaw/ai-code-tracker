package validation

import (
	"testing"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestConfigValidator_Validate(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name    string
		config  *tracker.Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go", ".js", ".py"},
				ExcludePatterns:    []string{"*_test.go", "vendor/*"},
				AuthorMappings: map[string]string{
					"AI Assistant": "ai",
					"Claude":       "ai",
				},
			},
			wantErr: false,
		},
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Invalid percentage - negative",
			config: &tracker.Config{
				TargetAIPercentage: -10.0,
				TrackedExtensions:  []string{".go"},
			},
			wantErr: true,
		},
		{
			name: "Invalid percentage - over 100",
			config: &tracker.Config{
				TargetAIPercentage: 150.0,
				TrackedExtensions:  []string{".go"},
			},
			wantErr: true,
		},
		{
			name: "No tracked extensions",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid extension format",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{"go", ".js"}, // "go" is invalid
			},
			wantErr: true,
		},
		{
			name: "Duplicate extensions",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go", ".js", ".go"},
			},
			wantErr: true,
		},
		{
			name: "Reserved pattern in excludes",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go"},
				ExcludePatterns:    []string{".git/*"},
			},
			wantErr: true,
		},
		{
			name: "Invalid glob pattern",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go"},
				ExcludePatterns:    []string{"["},
			},
			wantErr: true,
		},
		{
			name: "Empty author mapping key",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go"},
				AuthorMappings: map[string]string{
					"": "ai",
				},
			},
			wantErr: true,
		},
		{
			name: "Empty author mapping value",
			config: &tracker.Config{
				TargetAIPercentage: 80.0,
				TrackedExtensions:  []string{".go"},
				AuthorMappings: map[string]string{
					"Claude": "",
				},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTargetPercentage(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name       string
		percentage float64
		wantErr    bool
	}{
		{"Valid 0", 0, false},
		{"Valid 50", 50.0, false},
		{"Valid 100", 100.0, false},
		{"Invalid negative", -1.0, true},
		{"Invalid over 100", 100.1, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateTargetPercentage(tt.percentage)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTargetPercentage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExtensions(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name       string
		extensions []string
		wantErr    bool
	}{
		{"Valid extensions", []string{".go", ".js", ".py"}, false},
		{"Empty extensions", []string{}, true},
		{"Invalid format - no dot", []string{"go"}, true},
		{"Invalid format - multiple dots", []string{"..go"}, true},
		{"Invalid format - special chars", []string{".go@"}, true},
		{"Duplicate extensions", []string{".go", ".js", ".go"}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateExtensions(tt.extensions)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExtensions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExcludePatterns(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
	}{
		{"Valid patterns", []string{"*_test.go", "vendor/*", "*.tmp"}, false},
		{"Empty patterns", []string{}, false},
		{"Reserved .git", []string{".git/*"}, true},
		{"Reserved .ai_code_tracking", []string{"*/.ai_code_tracking/*"}, true},
		{"Invalid glob", []string{"["}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateExcludePatterns(tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateExcludePatterns() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}