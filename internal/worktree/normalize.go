package worktree

import (
	"path/filepath"
	"sort"
)

func Normalize(items []Worktree, currentPath string) []Worktree {
	if len(items) == 0 {
		return nil
	}

	currentPath = filepath.Clean(currentPath)
	out := make([]Worktree, len(items))
	copy(out, items)

	for i := range out {
		out[i].IsCurrent = filepath.Clean(out[i].Path) == currentPath
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsCurrent != out[j].IsCurrent {
			return out[i].IsCurrent
		}
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].BranchLabel < out[j].BranchLabel
	})

	for i := range out {
		out[i].Index = i + 1
	}

	return out
}
