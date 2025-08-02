package validation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	
	"github.com/y-hirakaw/ai-code-tracker/internal/errors"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

var (
	// Valid file extension pattern
	fileExtPattern = regexp.MustCompile(`^\.[a-zA-Z0-9]+$`)
	
	// Reserved patterns that should not be excluded
	reservedPatterns = []string{".git", ".ai_code_tracking"}
)

// ConfigValidator validates tracker configuration
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// Validate performs comprehensive validation on the configuration
func (v *ConfigValidator) Validate(config *tracker.Config) error {
	if config == nil {
		return errors.NewValidationError("Validate", "config", "configuration cannot be nil")
	}
	
	// Validate target AI percentage
	if err := v.validateTargetPercentage(config.TargetAIPercentage); err != nil {
		return err
	}
	
	// Validate tracked extensions
	if err := v.validateExtensions(config.TrackedExtensions); err != nil {
		return err
	}
	
	// Validate exclude patterns
	if err := v.validateExcludePatterns(config.ExcludePatterns); err != nil {
		return err
	}
	
	// Validate author mappings
	if err := v.validateAuthorMappings(config.AuthorMappings); err != nil {
		return err
	}
	
	return nil
}

// validateTargetPercentage checks if the target percentage is within valid range
func (v *ConfigValidator) validateTargetPercentage(percentage float64) error {
	if percentage < 0 || percentage > 100 {
		return errors.NewValidationError(
			"validateTargetPercentage",
			"TargetAIPercentage",
			fmt.Sprintf("target percentage must be between 0 and 100, got: %.2f", percentage),
		)
	}
	return nil
}

// validateExtensions checks if file extensions are valid
func (v *ConfigValidator) validateExtensions(extensions []string) error {
	if len(extensions) == 0 {
		return errors.NewValidationError(
			"validateExtensions",
			"TrackedExtensions",
			"at least one file extension must be specified",
		)
	}
	
	seen := make(map[string]bool)
	for _, ext := range extensions {
		// Check format
		if !fileExtPattern.MatchString(ext) {
			return errors.NewValidationError(
				"validateExtensions",
				"TrackedExtensions",
				fmt.Sprintf("invalid extension format: %s (must be like .go, .js)", ext),
			)
		}
		
		// Check for duplicates
		if seen[ext] {
			return errors.NewValidationError(
				"validateExtensions",
				"TrackedExtensions",
				fmt.Sprintf("duplicate extension: %s", ext),
			)
		}
		seen[ext] = true
	}
	
	return nil
}

// validateExcludePatterns checks if exclude patterns are valid
func (v *ConfigValidator) validateExcludePatterns(patterns []string) error {
	for _, pattern := range patterns {
		// Check if pattern is trying to exclude reserved paths
		for _, reserved := range reservedPatterns {
			if strings.Contains(pattern, reserved) {
				return errors.NewValidationError(
					"validateExcludePatterns",
					"ExcludePatterns",
					fmt.Sprintf("cannot exclude reserved pattern: %s", reserved),
				)
			}
		}
		
		// Validate glob pattern syntax
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return errors.NewValidationError(
				"validateExcludePatterns",
				"ExcludePatterns",
				fmt.Sprintf("invalid glob pattern: %s", pattern),
			)
		}
	}
	
	return nil
}

// validateAuthorMappings checks if author mappings are valid
func (v *ConfigValidator) validateAuthorMappings(mappings map[string]string) error {
	if mappings == nil {
		return nil // Nil map is allowed
	}
	
	// Check for empty keys or values
	for key, value := range mappings {
		if strings.TrimSpace(key) == "" {
			return errors.NewValidationError(
				"validateAuthorMappings",
				"AuthorMappings",
				"author mapping key cannot be empty",
			)
		}
		
		if strings.TrimSpace(value) == "" {
			return errors.NewValidationError(
				"validateAuthorMappings",
				"AuthorMappings",
				fmt.Sprintf("author mapping value cannot be empty for key: %s", key),
			)
		}
	}
	
	return nil
}