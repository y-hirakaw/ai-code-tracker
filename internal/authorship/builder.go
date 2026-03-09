package authorship

import (
	"time"

	"github.com/y-hirakaw/ai-code-tracker/internal/tracker"
)

// BuildAuthorshipMap はチェックポイントから filepath -> author のマップを構築します。
// 照合戦略:
//
//	Phase 1: cp.Changes のファイルパスと changedFiles の完全一致（既存動作）
//	Phase 2: cp.Snapshot のハッシュと commitParentSnapshot の比較（stash/restore対応）
//
// commitParentSnapshot が nil の場合は Phase 1 のみ実行します（後方互換）。
func BuildAuthorshipMap(checkpoints []*tracker.CheckpointV2, changedFiles map[string]bool, commitParentSnapshot map[string]string) map[string]*tracker.CheckpointV2 {
	authorMap := make(map[string]*tracker.CheckpointV2)

	// Phase 1: ファイルパス完全一致（既存ロジック）
	for _, cp := range checkpoints {
		for fpath := range cp.Changes {
			if changedFiles[fpath] {
				authorMap[fpath] = cp
			}
		}
	}

	// Phase 2: Snapshot ハッシュベース照合（stash/restore対応）
	if commitParentSnapshot != nil {
		for fpath := range changedFiles {
			if _, alreadyMatched := authorMap[fpath]; alreadyMatched {
				continue
			}
			if bestCP := findCheckpointBySnapshot(checkpoints, fpath, commitParentSnapshot); bestCP != nil {
				authorMap[fpath] = bestCP
			}
		}
	}

	return authorMap
}

// findCheckpointBySnapshot は Snapshot のハッシュを使ってファイルを変更したチェックポイントを逆順に探索します。
// 照合条件:
//  1. チェックポイントの Snapshot に対象ファイルが存在する
//  2. そのハッシュが commitParentSnapshot のハッシュと異なる（= 変更の証拠）
//  3. commitParentSnapshot に対象ファイルが存在しない場合は新規ファイルと見なす
func findCheckpointBySnapshot(checkpoints []*tracker.CheckpointV2, targetFile string, commitParentSnapshot map[string]string) *tracker.CheckpointV2 {
	parentHash, parentExists := commitParentSnapshot[targetFile]

	for i := len(checkpoints) - 1; i >= 0; i-- {
		cp := checkpoints[i]
		snap, snapExists := cp.Snapshot[targetFile]
		if !snapExists {
			continue
		}

		if !parentExists {
			// 親コミットにファイルが存在しない = 新規ファイル
			return cp
		}

		if snap.Hash != parentHash {
			// このチェックポイント時点のファイル状態が親コミットと異なる
			return cp
		}
	}

	return nil
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
