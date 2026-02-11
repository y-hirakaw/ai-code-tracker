package tracker

import "strings"

// MatchesPattern performs simple wildcard pattern matching.
// Supports prefix wildcard (*_test.go), suffix wildcard (vendor/*), and exact match.
func MatchesPattern(fpath, pattern string) bool {
	if pattern == "" {
		return false
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(fpath, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(fpath, pattern[:len(pattern)-1])
	}
	return fpath == pattern
}

// IsTrackedFile checks if a file should be tracked based on config.
// A file is tracked if it has a tracked extension and does not match any exclude pattern.
func IsTrackedFile(fpath string, cfg *Config) bool {
	hasValidExt := false
	for _, ext := range cfg.TrackedExtensions {
		if strings.HasSuffix(fpath, ext) {
			hasValidExt = true
			break
		}
	}
	if !hasValidExt {
		return false
	}

	for _, pattern := range cfg.ExcludePatterns {
		if MatchesPattern(fpath, pattern) {
			return false
		}
	}
	return true
}
