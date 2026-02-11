package git

import (
	"fmt"
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

func TestParseRangeNumstat(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedCommits []string
		expectedStats   map[string]map[string][2]int
	}{
		{
			name:            "empty input",
			input:           "",
			expectedCommits: nil,
			expectedStats:   map[string]map[string][2]int{},
		},
		{
			name:            "single commit with files",
			input:           "__AICT_COMMIT__abc123\n\n10\t5\tfile1.go\n3\t1\tfile2.go",
			expectedCommits: []string{"abc123"},
			expectedStats: map[string]map[string][2]int{
				"abc123": {
					"file1.go": {10, 5},
					"file2.go": {3, 1},
				},
			},
		},
		{
			name:            "multiple commits",
			input:           "__AICT_COMMIT__abc123\n\n10\t5\tfile1.go\n__AICT_COMMIT__def456\n\n7\t2\tfile3.go\n3\t0\tfile4.go",
			expectedCommits: []string{"abc123", "def456"},
			expectedStats: map[string]map[string][2]int{
				"abc123": {
					"file1.go": {10, 5},
				},
				"def456": {
					"file3.go": {7, 2},
					"file4.go": {3, 0},
				},
			},
		},
		{
			name:            "commit with no file changes",
			input:           "__AICT_COMMIT__abc123\n\n10\t5\tfile1.go\n__AICT_COMMIT__merge456\n__AICT_COMMIT__def789\n\n2\t1\tfile2.go",
			expectedCommits: []string{"abc123", "merge456", "def789"},
			expectedStats: map[string]map[string][2]int{
				"abc123":   {"file1.go": {10, 5}},
				"merge456": {},
				"def789":   {"file2.go": {2, 1}},
			},
		},
		{
			name:            "binary file skipped",
			input:           "__AICT_COMMIT__abc123\n\n-\t-\tbinary.dat\n10\t5\tfile.go",
			expectedCommits: []string{"abc123"},
			expectedStats: map[string]map[string][2]int{
				"abc123": {"file.go": {10, 5}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, commits := ParseRangeNumstat(tt.input)

			// コミット数の確認
			if len(commits) != len(tt.expectedCommits) {
				t.Errorf("commits count: got %d, want %d", len(commits), len(tt.expectedCommits))
			}
			for i, expected := range tt.expectedCommits {
				if i < len(commits) && commits[i] != expected {
					t.Errorf("commit[%d]: got %q, want %q", i, commits[i], expected)
				}
			}

			// numstat データの確認
			if len(stats) != len(tt.expectedStats) {
				t.Errorf("stats count: got %d, want %d", len(stats), len(tt.expectedStats))
			}
			for commitHash, expectedFiles := range tt.expectedStats {
				gotFiles, exists := stats[commitHash]
				if !exists {
					t.Errorf("missing commit %q in stats", commitHash)
					continue
				}
				if len(gotFiles) != len(expectedFiles) {
					t.Errorf("commit %q: got %d files, want %d", commitHash, len(gotFiles), len(expectedFiles))
				}
				for filePath, expectedNums := range expectedFiles {
					gotNums, exists := gotFiles[filePath]
					if !exists {
						t.Errorf("commit %q: missing file %q", commitHash, filePath)
						continue
					}
					if gotNums != expectedNums {
						t.Errorf("commit %q file %q: got %v, want %v", commitHash, filePath, gotNums, expectedNums)
					}
				}
			}
		})
	}
}

func TestGetRangeNumstat(t *testing.T) {
	mockExecutor := gitexec.NewMockExecutor()
	mockExecutor.RunFunc = func(args ...string) (string, error) {
		return "__AICT_COMMIT__abc123\n\n10\t5\tfile.go\n__AICT_COMMIT__def456\n\n3\t1\tfile2.go", nil
	}

	stats, commits, err := GetRangeNumstat(mockExecutor, "HEAD~2..HEAD")
	if err != nil {
		t.Fatalf("GetRangeNumstat() error = %v", err)
	}

	if len(commits) != 2 {
		t.Errorf("commits count: got %d, want 2", len(commits))
	}
	if len(stats) != 2 {
		t.Errorf("stats count: got %d, want 2", len(stats))
	}

	// git log引数の確認
	calls := mockExecutor.GetCalls("Run")
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}
	expectedArgs := []string{"log", "--numstat", "--format=__AICT_COMMIT__%H", "--end-of-options", "HEAD~2..HEAD"}
	for i, arg := range expectedArgs {
		if i < len(calls[0].Args) && calls[0].Args[i] != arg {
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

func TestParseRangeNumstat_FileRename(t *testing.T) {
	input := "__AICT_COMMIT__abc123\n\n5\t2\told/path.go => new/path.go"
	stats, commits := ParseRangeNumstat(input)

	if len(commits) != 1 {
		t.Fatalf("commits count: got %d, want 1", len(commits))
	}

	files := stats["abc123"]
	if _, exists := files["new/path.go"]; !exists {
		t.Error("expected renamed file to use new path")
	}
	if _, exists := files["old/path.go"]; exists {
		t.Error("old path should not be present")
	}
	if files["new/path.go"] != [2]int{5, 2} {
		t.Errorf("got %v, want [5, 2]", files["new/path.go"])
	}
}

func TestGetRangeNumstat_Error(t *testing.T) {
	mockExecutor := gitexec.NewMockExecutor()
	mockExecutor.RunFunc = func(args ...string) (string, error) {
		return "", fmt.Errorf("fatal: bad revision 'invalid..range'")
	}

	_, _, err := GetRangeNumstat(mockExecutor, "invalid..range")
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}
