package ui

import (
	"strings"
	"testing"

	"ww/internal/worktree"
)

func TestFormatListTableUsesUnicodeBoxBorders(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{Worktree: worktree.Worktree{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true}},
		{Worktree: worktree.Worktree{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", IsDirty: true}},
	})

	for _, fragment := range []string{
		"┌",
		"┬",
		"│ INDEX ",
		"├",
		"┼",
		"└",
		"┴",
		"│ 1",
		"[CURRENT]",
		"[DIRTY]",
	} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected %q in table output, got %q", fragment, got)
		}
	}
}

func TestFormatListTableWrapsLongPathInsidePathCell(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{
			Worktree: worktree.Worktree{
				Index:       2,
				BranchLabel: "codex/current-dirty-status",
				Path:        "/Users/liuwei/workspace/ww/.worktrees/current-dirty-status/very/long/path/for/wrapping",
				IsDirty:     true,
			},
		},
	})

	if !strings.Contains(got, "│ 2") {
		t.Fatalf("expected first row for wrapped item, got %q", got)
	}
	if !strings.Contains(got, "│       │                   │                            │") {
		t.Fatalf("expected continuation row with blank leading cells, got %q", got)
	}
	if !strings.Contains(got, "current-dirty-status") || !strings.Contains(got, "very/long/path") {
		t.Fatalf("expected full path content across wrapped lines, got %q", got)
	}
}
