package git

import (
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

func TestParseNumstat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string][2]int
	}{
		{
			name:  "empty input",
			input: "",
			expected: map[string][2]int{},
		},
		{
			name:  "single file",
			input: "10\t5\tfile.go",
			expected: map[string][2]int{
				"file.go": {10, 5},
			},
		},
		{
			name: "multiple files",
			input: "10\t5\tfile1.go\n20\t15\tfile2.go\n30\t25\tfile3.go",
			expected: map[string][2]int{
				"file1.go": {10, 5},
				"file2.go": {20, 15},
				"file3.go": {30, 25},
			},
		},
		{
			name:  "file with spaces in path",
			input: "10\t5\tpath/to/my file.go",
			expected: map[string][2]int{
				"path/to/my file.go": {10, 5},
			},
		},
		{
			name:  "file rename",
			input: "10\t5\told/path.go => new/path.go",
			expected: map[string][2]int{
				"new/path.go": {10, 5},
			},
		},
		{
			name: "binary file (skipped)",
			input: "-\t-\tbinary.dat\n10\t5\tfile.go",
			expected: map[string][2]int{
				"file.go": {10, 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNumstat(tt.input)
			if err != nil {
				t.Errorf("ParseNumstat() error = %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ParseNumstat() got %d entries, want %d", len(result), len(tt.expected))
				return
			}

			for filepath, expected := range tt.expected {
				got, exists := result[filepath]
				if !exists {
					t.Errorf("ParseNumstat() missing file %q", filepath)
					continue
				}

				if got[0] != expected[0] || got[1] != expected[1] {
					t.Errorf("ParseNumstat() for %q got [%d, %d], want [%d, %d]",
						filepath, got[0], got[1], expected[0], expected[1])
				}
			}
		})
	}
}

func TestGetNumstatBetweenCommits(t *testing.T) {
	mockExecutor := gitexec.NewMockExecutor()
	mockExecutor.RunFunc = func(args ...string) (string, error) {
		return "10\t5\tfile.go\n20\t15\tfile2.go", nil
	}

	result, err := GetNumstatBetweenCommits(mockExecutor, "commit1", "commit2")
	if err != nil {
		t.Fatalf("GetNumstatBetweenCommits() error = %v", err)
	}

	expected := map[string][2]int{
		"file.go":  {10, 5},
		"file2.go": {20, 15},
	}

	if len(result) != len(expected) {
		t.Errorf("GetNumstatBetweenCommits() got %d entries, want %d", len(result), len(expected))
	}

	for filepath, expectedStats := range expected {
		got, exists := result[filepath]
		if !exists {
			t.Errorf("GetNumstatBetweenCommits() missing file %q", filepath)
			continue
		}

		if got != expectedStats {
			t.Errorf("GetNumstatBetweenCommits() for %q got %v, want %v", filepath, got, expectedStats)
		}
	}

	// Verify the executor was called with correct arguments
	calls := mockExecutor.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	expectedArgs := []string{"diff", "--numstat", "commit1", "commit2"}
	if len(calls[0].Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(calls[0].Args))
	}
	for i, arg := range expectedArgs {
		if calls[0].Args[i] != arg {
			t.Errorf("Arg %d: got %q, want %q", i, calls[0].Args[i], arg)
		}
	}
}

func TestGetNumstatFromHead(t *testing.T) {
	mockExecutor := gitexec.NewMockExecutor()
	mockExecutor.RunFunc = func(args ...string) (string, error) {
		return "5\t3\tchanged.go", nil
	}

	result, err := GetNumstatFromHead(mockExecutor)
	if err != nil {
		t.Fatalf("GetNumstatFromHead() error = %v", err)
	}

	expected := map[string][2]int{
		"changed.go": {5, 3},
	}

	if len(result) != len(expected) {
		t.Errorf("GetNumstatFromHead() got %d entries, want %d", len(result), len(expected))
	}

	// Verify the executor was called with correct arguments
	calls := mockExecutor.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	expectedArgs := []string{"diff", "HEAD", "--numstat"}
	if len(calls[0].Args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(calls[0].Args))
	}
	for i, arg := range expectedArgs {
		if calls[0].Args[i] != arg {
			t.Errorf("Arg %d: got %q, want %q", i, calls[0].Args[i], arg)
		}
	}
}
