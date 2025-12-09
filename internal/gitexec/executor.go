package gitexec

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Executor defines the interface for executing git commands
type Executor interface {
	// Run executes a git command with the given arguments in the current directory
	Run(args ...string) (string, error)

	// RunInDir executes a git command with the given arguments in a specific directory
	RunInDir(dir string, args ...string) (string, error)
}

// RealExecutor implements Executor for actual git command execution
type RealExecutor struct{}

// NewExecutor creates a new RealExecutor instance
func NewExecutor() Executor {
	return &RealExecutor{}
}

// Run executes a git command in the current directory
func (e *RealExecutor) Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w\nstderr: %s",
			strings.Join(args, " "), err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunInDir executes a git command in a specific directory
func (e *RealExecutor) RunInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git %s failed in %s: %w\nstderr: %s",
			strings.Join(args, " "), dir, err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// MockExecutor implements Executor for testing
type MockExecutor struct {
	// RunFunc is called when Run is invoked
	RunFunc func(args ...string) (string, error)

	// RunInDirFunc is called when RunInDir is invoked
	RunInDirFunc func(dir string, args ...string) (string, error)

	// CallLog stores all calls for verification
	CallLog []MockCall
}

// MockCall represents a recorded call to the mock executor
type MockCall struct {
	Method string
	Dir    string
	Args   []string
	Output string
	Error  error
}

// NewMockExecutor creates a new MockExecutor with default behavior
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		CallLog: []MockCall{},
	}
}

// Run executes the mock Run function
func (m *MockExecutor) Run(args ...string) (string, error) {
	var output string
	var err error

	if m.RunFunc != nil {
		output, err = m.RunFunc(args...)
	}

	m.CallLog = append(m.CallLog, MockCall{
		Method: "Run",
		Args:   args,
		Output: output,
		Error:  err,
	})

	return output, err
}

// RunInDir executes the mock RunInDir function
func (m *MockExecutor) RunInDir(dir string, args ...string) (string, error) {
	var output string
	var err error

	if m.RunInDirFunc != nil {
		output, err = m.RunInDirFunc(dir, args...)
	}

	m.CallLog = append(m.CallLog, MockCall{
		Method: "RunInDir",
		Dir:    dir,
		Args:   args,
		Output: output,
		Error:  err,
	})

	return output, err
}

// GetCalls returns all recorded calls with the given method name
func (m *MockExecutor) GetCalls(method string) []MockCall {
	var calls []MockCall
	for _, call := range m.CallLog {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// Reset clears the call log
func (m *MockExecutor) Reset() {
	m.CallLog = []MockCall{}
}
