package ui

import (
	"strings"

	"ww/internal/worktree"
)

func StatusTags(item worktree.Worktree) []string {
	tags := make([]string, 0, 2)
	if item.IsCurrent {
		tags = append(tags, "[CURRENT]")
	}
	if item.IsDirty {
		tags = append(tags, "[DIRTY]")
	}
	return tags
}

func StatusText(item worktree.Worktree) string {
	return strings.Join(StatusTags(item), " ")
}
