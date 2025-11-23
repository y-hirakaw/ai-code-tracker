package authorship

import (
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// BuildAuthorshipLog converts checkpoints to AuthorshipLog
// SPEC.md § チェックポイント → Authorship Log変換
func BuildAuthorshipLog(checkpoints []*tracker.CheckpointV2, commitHash string) (*tracker.AuthorshipLog, error) {
	log := &tracker.AuthorshipLog{
		Version:   AuthorshipLogVersion,
		Commit:    commitHash,
		Timestamp: time.Now(),
		Files:     make(map[string]tracker.FileInfo),
	}

	// ファイルごとに作成者情報を集約
	for _, cp := range checkpoints {
		for filepath, change := range cp.Changes {
			fileInfo, exists := log.Files[filepath]
			if !exists {
				fileInfo = tracker.FileInfo{Authors: []tracker.AuthorInfo{}}
			}

			// 同じ作成者が既に存在するか確認
			authorIdx := -1
			for i, author := range fileInfo.Authors {
				if author.Name == cp.Author && author.Type == cp.Type {
					authorIdx = i
					break
				}
			}

			if authorIdx >= 0 {
				// 既存の作成者に行範囲を追加
				fileInfo.Authors[authorIdx].Lines = append(
					fileInfo.Authors[authorIdx].Lines,
					change.Lines...,
				)
			} else {
				// 新しい作成者を追加
				fileInfo.Authors = append(fileInfo.Authors, tracker.AuthorInfo{
					Name:     cp.Author,
					Type:     cp.Type,
					Lines:    change.Lines,
					Metadata: cp.Metadata,
				})
			}

			log.Files[filepath] = fileInfo
		}
	}

	return log, nil
}

// CountLines counts total lines from line ranges
func CountLines(ranges [][]int) int {
	total := 0
	for _, r := range ranges {
		if len(r) == 1 {
			total++
		} else if len(r) == 2 {
			total += r[1] - r[0] + 1
		}
	}
	return total
}
