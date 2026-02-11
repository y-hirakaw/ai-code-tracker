package authorship

import (
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// BuildAuthorshipMap はチェックポイントから filepath -> author のマップを構築します。
// 各ファイルについて最後に変更したチェックポイントが優先されます。
func BuildAuthorshipMap(checkpoints []*tracker.CheckpointV2, changedFiles map[string]bool) map[string]*tracker.CheckpointV2 {
	authorMap := make(map[string]*tracker.CheckpointV2)

	for _, cp := range checkpoints {
		for fpath := range cp.Changes {
			if changedFiles[fpath] {
				authorMap[fpath] = cp
			}
		}
	}

	return authorMap
}

// BuildAuthorshipLogFromDiff はdiffとauthorshipマッピングからAuthorship Logを作成します。
// changedFiles内のファイルのうち、追跡対象の拡張子かつ除外パターンに該当しないもののみ含めます。
func BuildAuthorshipLogFromDiff(
	diffMap map[string]tracker.Change,
	authorMap map[string]*tracker.CheckpointV2,
	commitHash string,
	changedFiles map[string]bool,
	cfg *tracker.Config,
) (*tracker.AuthorshipLog, error) {
	log := &tracker.AuthorshipLog{
		Version:   AuthorshipLogVersion,
		Commit:    commitHash,
		Timestamp: time.Now(),
		Files:     make(map[string]tracker.FileInfo),
	}

	for fpath, change := range diffMap {
		if !changedFiles[fpath] {
			continue
		}

		if !tracker.IsTrackedFile(fpath, cfg) {
			continue
		}

		var authorName string
		var authorType tracker.AuthorType
		var metadata map[string]string

		if cp, exists := authorMap[fpath]; exists {
			authorName = cp.Author
			authorType = cp.Type
			metadata = cp.Metadata
		} else {
			authorName = cfg.DefaultAuthor
			authorType = tracker.AuthorTypeHuman
			metadata = map[string]string{"message": "No checkpoint found, assigned to default author"}
		}

		fileInfo := tracker.FileInfo{
			Authors: []tracker.AuthorInfo{
				{
					Name:     authorName,
					Type:     authorType,
					Lines:    change.Lines,
					Metadata: metadata,
				},
			},
		}

		log.Files[fpath] = fileInfo
	}

	return log, nil
}

// BuildAuthorshipLog converts checkpoints to AuthorshipLog
// SPEC.md § チェックポイント → Authorship Log変換
// changedFiles: numstatで実際に変更されたファイルのリスト（nil の場合はフィルタリングなし）
func BuildAuthorshipLog(checkpoints []*tracker.CheckpointV2, commitHash string, changedFiles map[string]bool) (*tracker.AuthorshipLog, error) {
	log := &tracker.AuthorshipLog{
		Version:   AuthorshipLogVersion,
		Commit:    commitHash,
		Timestamp: time.Now(),
		Files:     make(map[string]tracker.FileInfo),
	}

	// ファイルごとに作成者情報を集約
	for _, cp := range checkpoints {
		for filepath, change := range cp.Changes {
			// numstatフィルタリング: 実際に変更されたファイルのみ含める
			if changedFiles != nil && !changedFiles[filepath] {
				continue // このファイルは実際には変更されていないのでスキップ
			}

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
