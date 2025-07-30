package tracker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CheckpointRecorder handles JSONL-based checkpoint recording
type CheckpointRecorder struct {
	baseDir string
}

func NewCheckpointRecorder(baseDir string) *CheckpointRecorder {
	return &CheckpointRecorder{
		baseDir: baseDir,
	}
}

// RecordCheckpoint records a checkpoint in JSONL format
func (cr *CheckpointRecorder) RecordCheckpoint(author string) error {
	// Get current commit hash
	commit := ""
	if cmd := exec.Command("git", "rev-parse", "HEAD"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			commit = strings.TrimSpace(string(output))
		}
	}

	// Collect numstat data (diff from HEAD)
	numstat, err := cr.collectNumstatData()
	if err != nil {
		// Continue with empty numstat if git fails
		numstat = make(map[string][2]int)
	}

	// Sum up all added and deleted lines
	totalAdded := 0
	totalDeleted := 0
	for _, stats := range numstat {
		totalAdded += stats[0]   // added lines
		totalDeleted += stats[1] // deleted lines
	}

	// Create checkpoint record
	record := CheckpointRecord{
		Timestamp: time.Now(),
		Author:    author,
		Commit:    commit,
		Added:     totalAdded,
		Deleted:   totalDeleted,
	}

	// Append to JSONL file
	return cr.appendRecord(record)
}

// collectNumstatData collects current git diff --numstat data from HEAD
func (cr *CheckpointRecorder) collectNumstatData() (map[string][2]int, error) {
	cmd := exec.Command("git", "diff", "HEAD", "--numstat")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

	result := make(map[string][2]int)
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Format: "added\tdeleted\tfilepath"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added, err := strconv.Atoi(parts[0])
		if err != nil {
			continue // Skip binary files which show "-"
		}

		deleted, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		// Handle renames: "path1 => path2" becomes just "path2"
		filepath := strings.Join(parts[2:], " ")
		if idx := strings.Index(filepath, " => "); idx != -1 {
			filepath = filepath[idx+4:]
		}

		result[filepath] = [2]int{added, deleted}
	}

	return result, nil
}

// appendRecord appends a checkpoint record to the JSONL file
func (cr *CheckpointRecorder) appendRecord(record CheckpointRecord) error {
	checkpointsFile := filepath.Join(cr.baseDir, "checkpoints.jsonl")
	
	// Ensure directory exists
	if err := os.MkdirAll(cr.baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file for appending
	file, err := os.OpenFile(checkpointsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open checkpoints file: %w", err)
	}
	defer file.Close()

	// Marshal record to JSON
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Write JSON line
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write record: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// ReadAllRecords reads all checkpoint records from JSONL file
func (cr *CheckpointRecorder) ReadAllRecords() ([]CheckpointRecord, error) {
	checkpointsFile := filepath.Join(cr.baseDir, "checkpoints.jsonl")
	
	file, err := os.Open(checkpointsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []CheckpointRecord{}, nil // Empty if file doesn't exist
		}
		return nil, fmt.Errorf("failed to open checkpoints file: %w", err)
	}
	defer file.Close()

	var records []CheckpointRecord
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var record CheckpointRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			// Skip malformed lines
			continue
		}

		records = append(records, record)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return records, nil
}

// GetLatestRecords returns the latest N records
func (cr *CheckpointRecorder) GetLatestRecords(count int) ([]CheckpointRecord, error) {
	allRecords, err := cr.ReadAllRecords()
	if err != nil {
		return nil, err
	}

	if count <= 0 || count >= len(allRecords) {
		return allRecords, nil
	}

	// Return last N records
	return allRecords[len(allRecords)-count:], nil
}