package utils

import (
	"fmt"
	"strings"
)

// TruncateString は文字列を指定された長さで切り詰め、必要に応じて省略記号を追加する
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	
	if maxLength <= 3 {
		return strings.Repeat(".", maxLength)
	}
	
	return s[:maxLength-3] + "..."
}

// TruncateStringPrefix は文字列の前方を切り詰め、省略記号を前に付ける
func TruncateStringPrefix(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	
	if maxLength <= 3 {
		return strings.Repeat(".", maxLength)
	}
	
	return "..." + s[len(s)-(maxLength-3):]
}

// SplitAndTrim は文字列を分割し、各要素をトリムする
func SplitAndTrim(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}

// FormatPercentage はパーセンテージを適切にフォーマットする
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// PadString は文字列を指定された幅にパディングする
func PadString(s string, width int) string {
	return fmt.Sprintf("%-*s", width, s)
}

// CreateSeparatorLine は指定された文字で区切り線を作成する
func CreateSeparatorLine(char string, length int) string {
	return strings.Repeat(char, length)
}

// IsEmptyOrWhitespace は文字列が空または空白文字のみかチェックする
func IsEmptyOrWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ContainsIgnoreCase は大文字小文字を無視して文字列が含まれているかチェックする
func ContainsIgnoreCase(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// SanitizeFileName はファイル名として使用できない文字を除去/置換する
func SanitizeFileName(filename string) string {
	// 危険な文字を除去/置換
	replacements := map[string]string{
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}
	
	result := filename
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	
	// 先頭・末尾の空白とピリオドを除去
	result = strings.Trim(result, " .")
	
	return result
}