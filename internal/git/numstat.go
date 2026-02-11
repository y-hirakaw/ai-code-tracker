package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/y-hirakaw/ai-code-tracker/internal/gitexec"
)

// NumstatEntry represents a single file's statistics from git diff --numstat
type NumstatEntry struct {
	Filepath string
	Added    int
	Deleted  int
}

// ParseNumstat parses git diff --numstat output into structured data
// Input format: "added\tdeleted\tfilepath" (one per line)
// Handles binary files (shows "-") and file renames ("path1 => path2")
func ParseNumstat(output string) (map[string][2]int, error) {
	result := make(map[string][2]int)
	lines := strings.Split(output, "\n")

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

// GetNumstatBetweenCommits runs git diff --numstat between two commits
func GetNumstatBetweenCommits(executor gitexec.Executor, fromCommit, toCommit string) (map[string][2]int, error) {
	if err := gitexec.ValidateRevisionArg(fromCommit); err != nil {
		return nil, err
	}
	if err := gitexec.ValidateRevisionArg(toCommit); err != nil {
		return nil, err
	}
	output, err := executor.Run("diff", "--numstat", fromCommit, toCommit)
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

	return ParseNumstat(output)
}

// GetNumstatFromHead runs git diff --numstat from HEAD (for uncommitted changes)
func GetNumstatFromHead(executor gitexec.Executor) (map[string][2]int, error) {
	output, err := executor.Run("diff", "HEAD", "--numstat")
	if err != nil {
		return nil, fmt.Errorf("failed to run git diff --numstat: %w", err)
	}

	return ParseNumstat(output)
}

// commitNumstatMarker はgit logの出力でコミットを区切るマーカー
const commitNumstatMarker = "__AICT_COMMIT__"

// GetRangeNumstat はコミット範囲内の全コミットのnumstatを1回のgit呼び出しで取得します。
// 戻り値:
//   - numstats: map[commitHash]map[filepath][2]int ([0]=added, [1]=deleted)
//   - commits: コミットハッシュのリスト（git log順＝新しい順）
func GetRangeNumstat(executor gitexec.Executor, rangeSpec string) (map[string]map[string][2]int, []string, error) {
	if err := gitexec.ValidateRevisionArg(rangeSpec); err != nil {
		return nil, nil, err
	}
	output, err := executor.Run("log", "--numstat", "--format="+commitNumstatMarker+"%H", "--end-of-options", rangeSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get range numstat: %w", err)
	}

	numstats, commits := ParseRangeNumstat(output)
	return numstats, commits, nil
}

// ParseRangeNumstat は git log --numstat --format=__AICT_COMMIT__%H の出力をパースします。
func ParseRangeNumstat(output string) (map[string]map[string][2]int, []string) {
	result := make(map[string]map[string][2]int)
	var commits []string
	var currentCommit string

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, commitNumstatMarker) {
			currentCommit = strings.TrimPrefix(line, commitNumstatMarker)
			commits = append(commits, currentCommit)
			result[currentCommit] = make(map[string][2]int)
			continue
		}

		if currentCommit == "" || line == "" {
			continue
		}

		// numstat行をパース: added\tdeleted\tfilepath
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added, err1 := strconv.Atoi(parts[0])
		deleted, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue // バイナリファイル（"-"表示）はスキップ
		}

		filePath := strings.Join(parts[2:], " ")
		if idx := strings.Index(filePath, " => "); idx != -1 {
			filePath = filePath[idx+4:]
		}

		result[currentCommit][filePath] = [2]int{added, deleted}
	}

	return result, commits
}
