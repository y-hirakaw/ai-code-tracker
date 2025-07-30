package test

import (
	"fmt"
	"strings"
)

// OldVerboseFunction - This will be refactored
func OldVerboseFunction(input string) string {
	var result string
	
	// Verbose way to check if string is empty
	if len(input) == 0 {
		result = "Input is empty"
	} else {
		// Verbose way to process string
		chars := []rune(input)
		processedChars := make([]rune, 0, len(chars))
		
		for i := 0; i < len(chars); i++ {
			char := chars[i]
			if char >= 'a' && char <= 'z' {
				processedChars = append(processedChars, char-32)
			} else if char >= 'A' && char <= 'Z' {
				processedChars = append(processedChars, char)
			} else {
				processedChars = append(processedChars, char)
			}
		}
		
		result = string(processedChars)
		
		// Verbose way to add prefix
		if result != "" {
			result = "Processed: " + result
		}
	}
	
	// Verbose way to log
	fmt.Println("Function executed with input:", input)
	fmt.Println("Function result:", result)
	
	return result
}

// AnotherVerboseFunction - Will be simplified
func AnotherVerboseFunction(numbers []int) int {
	if numbers == nil {
		return 0
	}
	
	if len(numbers) == 0 {
		return 0
	}
	
	sum := 0
	for i := 0; i < len(numbers); i++ {
		sum = sum + numbers[i]
	}
	
	return sum
}