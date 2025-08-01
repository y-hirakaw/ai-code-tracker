package tracker

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CheckpointManager struct {
	baseDir string
}

func NewCheckpointManager(baseDir string) *CheckpointManager {
	return &CheckpointManager{
		baseDir: baseDir,
	}
}

func (cm *CheckpointManager) CreateCheckpoint(author string, extensions []string) (*Checkpoint, error) {
	checkpoint := &Checkpoint{
		ID:          cm.generateID(),
		Timestamp:   time.Now(),
		Author:      author,
		Files:       make(map[string]FileContent),
		NumstatData: make(map[string][2]int),
	}

	// Try to get current commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err == nil {
		checkpoint.CommitHash = strings.TrimSpace(string(output))
	}

	// Collect numstat data (diff from HEAD)
	if err := cm.collectNumstatData(checkpoint); err != nil {
		// Continue even if numstat fails (might not be a git repo)
		fmt.Printf("Warning: Could not collect numstat data: %v\n", err)
	}

	err = cm.scanCodeFiles(".", extensions, checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to scan code files: %w", err)
	}

	return checkpoint, nil
}

func (cm *CheckpointManager) SaveCheckpoint(checkpoint *Checkpoint) error {
	checkpointDir := filepath.Join(cm.baseDir, "checkpoints")
	if err := os.MkdirAll(checkpointDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	filename := filepath.Join(checkpointDir, fmt.Sprintf("%s_%s.json", checkpoint.Author, checkpoint.ID))
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	return nil
}

func (cm *CheckpointManager) LoadCheckpoint(filename string) (*Checkpoint, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

func (cm *CheckpointManager) GetLatestCheckpoints(author string, count int) ([]*Checkpoint, error) {
	checkpointDir := filepath.Join(cm.baseDir, "checkpoints")
	var pattern string
	if author == "*" {
		pattern = filepath.Join(checkpointDir, "*.json")
	} else {
		pattern = filepath.Join(checkpointDir, fmt.Sprintf("%s_*.json", author))
	}

	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}

	checkpoints := make([]*Checkpoint, 0, len(files))
	for _, file := range files {
		cp, err := cm.LoadCheckpoint(file)
		if err != nil {
			continue
		}
		checkpoints = append(checkpoints, cp)
	}

	if len(checkpoints) > count && count > 0 {
		checkpoints = checkpoints[len(checkpoints)-count:]
	}

	return checkpoints, nil
}

// collectNumstatData collects current git diff --numstat data from HEAD
func (cm *CheckpointManager) collectNumstatData(checkpoint *Checkpoint) error {
	cmd := exec.Command("git", "diff", "HEAD", "--numstat")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

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

		checkpoint.NumstatData[filepath] = [2]int{added, deleted}
	}

	return nil
}

func (cm *CheckpointManager) scanCodeFiles(root string, extensions []string, checkpoint *Checkpoint) error {
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || strings.HasPrefix(path, ".ai_code_tracking") || strings.HasPrefix(path, ".git") {
			return nil
		}

		ext := filepath.Ext(path)
		if !extMap[ext] {
			return nil
		}

		content, err := cm.readFileLines(path)
		if err != nil {
			return err
		}

		checkpoint.Files[path] = FileContent{
			Path:  path,
			Lines: content,
		}

		return nil
	})
}

func (cm *CheckpointManager) readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	return lines, nil
}

func (cm *CheckpointManager) generateID() string {
	h := md5.New()
	io.WriteString(h, time.Now().String())
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}
