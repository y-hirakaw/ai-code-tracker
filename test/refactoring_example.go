package test

import (
	"fmt"
	"strings"
)

// ProcessString - Refactored version
func ProcessString(input string) string {
	if input == "" {
		return "Input is empty"
	}
	
	result := "Processed: " + strings.ToUpper(input)
	fmt.Printf("Processed %q -> %q\n", input, result)
	return result
}

// Sum - Simplified version
func Sum(numbers []int) int {
	sum := 0
	for _, n := range numbers {
		sum += n
	}
	return sum
}