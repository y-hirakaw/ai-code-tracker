package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
	"github.com/y-hirakaw/ai-code-tracker/internal/testutil"
	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

func TestGetLatestCommitHash(t *testing.T) {
	// Setup test git repository
	tmpDir := testutil.TempGitRepo(t)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Create a commit
	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n")
	testutil.GitCommit(t, tmpDir, "Test commit")

	// Test getLatestCommitHash
	hash, err := getLatestCommitHash()

	if err != nil {
		t.Fatalf("getLatestCommitHash() error = %v", err)
	}

	if hash == "" {
		t.Error("getLatestCommitHash() returned empty hash")
	}

	// Verify hash format (40 hex characters)
	if len(hash) != 40 {
		t.Errorf("getLatestCommitHash() hash length = %d, want 40", len(hash))
	}

	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("getLatestCommitHash() invalid character in hash: %c", c)
			break
		}
	}
}

func TestHandleCommit_EnvironmentSetup(t *testing.T) {
	// このテストはhandleCommit()の統合テストの前提条件（環境セットアップ）を検証する
	tmpDir := testutil.TempGitRepo(t)
	testutil.InitAICT(t, tmpDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	testutil.CreateTestFile(t, tmpDir, "test.go", "package main\n\nfunc main() {}\n")
	testutil.GitCommit(t, tmpDir, "Initial commit")

	hash, err := getLatestCommitHash()
	if err != nil {
		t.Fatalf("Failed to get commit hash: %v", err)
	}

	if hash == "" {
		t.Error("Expected commit hash after git commit")
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		fpath    string
		pattern  string
		expected bool
	}{
		// サフィックスワイルドカード（*で始まるパターン）
		{"suffix match _test.go", "handlers_commit_test.go", "*_test.go", true},
		{"suffix match nested", "internal/foo/bar_test.go", "*_test.go", true},
		{"suffix no match", "handlers_commit.go", "*_test.go", false},

		// プレフィックスワイルドカード（*で終わるパターン）
		{"prefix match vendor", "vendor/lib/foo.go", "vendor/*", true},
		{"prefix match node_modules", "node_modules/pkg/index.js", "node_modules/*", true},
		{"prefix no match", "src/main.go", "vendor/*", false},

		// 完全一致
		{"exact match", "Makefile", "Makefile", true},
		{"exact no match", "makefile", "Makefile", false},

		// 空文字列
		{"empty pattern", "foo.go", "", false},
		{"empty fpath", "", "*_test.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.MatchesPattern(tt.fpath, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.fpath, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsTrackedFile(t *testing.T) {
	cfg := &tracker.Config{
		TrackedExtensions: []string{".go", ".py", ".js"},
		ExcludePatterns:   []string{"*_test.go", "vendor/*"},
	}

	tests := []struct {
		name     string
		fpath    string
		expected bool
	}{
		// 追跡対象
		{"go file tracked", "main.go", true},
		{"py file tracked", "script.py", true},
		{"js file tracked", "app.js", true},
		{"nested go file", "internal/pkg/handler.go", true},

		// 除外パターンに該当
		{"test file excluded", "main_test.go", false},
		{"nested test excluded", "pkg/handler_test.go", false},
		{"vendor excluded", "vendor/lib/foo.go", false},

		// 追跡対象外の拡張子
		{"md not tracked", "README.md", false},
		{"txt not tracked", "notes.txt", false},
		{"yaml not tracked", "config.yaml", false},
		{"no extension", "Makefile", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tracker.IsTrackedFile(tt.fpath, cfg)
			if result != tt.expected {
				t.Errorf("IsTrackedFile(%q) = %v, want %v", tt.fpath, result, tt.expected)
			}
		})
	}
}

func TestIsTrackedFile_EmptyConfig(t *testing.T) {
	cfg := &tracker.Config{
		TrackedExtensions: []string{},
		ExcludePatterns:   []string{},
	}

	if tracker.IsTrackedFile("main.go", cfg) {
		t.Error("IsTrackedFile should return false when no extensions are configured")
	}
}



func TestGetCommitDiff(t *testing.T) {
	tests := []struct {
		name           string
		mockRunFunc    func(args ...string) (string, error)
		expectedFiles  map[string]tracker.Change
		expectError    bool
	}{
		{
			name: "normal commit with numstat output",
			mockRunFunc: func(args ...string) (string, error) {
				if len(args) >= 2 && args[0] == "rev-parse" && args[1] == "HEAD~1" {
					// HEAD~1 exists (normal commit)
					return "abc123", nil
				}
				if len(args) >= 4 && args[0] == "diff" && args[1] == "--numstat" {
					return "10\t2\tmain.go\n5\t0\thelper.go", nil
				}
				return "", nil
			},
			expectedFiles: map[string]tracker.Change{
				"main.go": {
					Added:   10,
					Deleted: 2,
					Lines:   [][]int{{1, 10}},
				},
				"helper.go": {
					Added:   5,
					Deleted: 0,
					Lines:   [][]int{{1, 5}},
				},
			},
			expectError: false,
		},
		{
			name: "initial commit fallback to show",
			mockRunFunc: func(args ...string) (string, error) {
				if len(args) >= 2 && args[0] == "rev-parse" && args[1] == "HEAD~1" {
					// HEAD~1 does not exist (initial commit)
					return "", fmt.Errorf("fatal: bad revision 'HEAD~1'")
				}
				if len(args) >= 1 && args[0] == "show" {
					return "3\t0\tnew_file.go", nil
				}
				return "", nil
			},
			expectedFiles: map[string]tracker.Change{
				"new_file.go": {
					Added:   3,
					Deleted: 0,
					Lines:   [][]int{{1, 3}},
				},
			},
			expectError: false,
		},
		{
			name: "empty numstat output",
			mockRunFunc: func(args ...string) (string, error) {
				if len(args) >= 2 && args[0] == "rev-parse" && args[1] == "HEAD~1" {
					return "abc123", nil
				}
				if len(args) >= 4 && args[0] == "diff" && args[1] == "--numstat" {
					return "", nil
				}
				return "", nil
			},
			expectedFiles: map[string]tracker.Change{},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DI: mock executor
			origExecutor := newExecutor
			defer func() { newExecutor = origExecutor }()

			mock := gitexec.NewMockExecutor()
			mock.RunFunc = tt.mockRunFunc
			newExecutor = func() gitexec.Executor { return mock }

			result, err := getCommitDiff("abc123def456")

			if tt.expectError && err == nil {
				t.Fatal("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expectedFiles) {
				t.Errorf("got %d files, want %d", len(result), len(tt.expectedFiles))
			}

			for fpath, expected := range tt.expectedFiles {
				got, ok := result[fpath]
				if !ok {
					t.Errorf("missing file %q in result", fpath)
					continue
				}
				if got.Added != expected.Added {
					t.Errorf("file %q: Added = %d, want %d", fpath, got.Added, expected.Added)
				}
				if got.Deleted != expected.Deleted {
					t.Errorf("file %q: Deleted = %d, want %d", fpath, got.Deleted, expected.Deleted)
				}
				if len(got.Lines) != len(expected.Lines) {
					t.Errorf("file %q: Lines count = %d, want %d", fpath, len(got.Lines), len(expected.Lines))
				}
			}
		})
	}
}

func TestGetCommitDiff_VerifiesCommandArgs(t *testing.T) {
	// Verify that initial commit uses "show" and normal commit uses "diff"
	t.Run("normal commit uses diff command", func(t *testing.T) {
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		mock.RunFunc = func(args ...string) (string, error) {
			if args[0] == "rev-parse" {
				return "parent_hash", nil
			}
			return "1\t0\tfile.go", nil
		}
		newExecutor = func() gitexec.Executor { return mock }

		_, _ = getCommitDiff("abc123")

		// Check that "diff" was called (not "show")
		calls := mock.GetCalls("Run")
		foundDiff := false
		for _, call := range calls {
			if len(call.Args) > 0 && call.Args[0] == "diff" {
				foundDiff = true
				// Verify args contain --numstat HEAD~1 HEAD
				argsStr := strings.Join(call.Args, " ")
				if !strings.Contains(argsStr, "--numstat") {
					t.Error("diff command missing --numstat flag")
				}
			}
		}
		if !foundDiff {
			t.Error("expected 'diff' command for normal commit, but it was not called")
		}
	})

	t.Run("initial commit uses show command", func(t *testing.T) {
		origExecutor := newExecutor
		defer func() { newExecutor = origExecutor }()

		mock := gitexec.NewMockExecutor()
		mock.RunFunc = func(args ...string) (string, error) {
			if args[0] == "rev-parse" {
				return "", fmt.Errorf("HEAD~1 not found")
			}
			return "2\t0\tinit.go", nil
		}
		newExecutor = func() gitexec.Executor { return mock }

		_, _ = getCommitDiff("first_commit")

		// Check that "show" was called (not "diff")
		calls := mock.GetCalls("Run")
		foundShow := false
		for _, call := range calls {
			if len(call.Args) > 0 && call.Args[0] == "show" {
				foundShow = true
				argsStr := strings.Join(call.Args, " ")
				if !strings.Contains(argsStr, "--numstat") {
					t.Error("show command missing --numstat flag")
				}
				if !strings.Contains(argsStr, "first_commit") {
					t.Error("show command missing commit hash")
				}
			}
		}
		if !foundShow {
			t.Error("expected 'show' command for initial commit, but it was not called")
		}
	})
}
