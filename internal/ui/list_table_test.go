package ui

import (
	"strings"
	"testing"

	"ww/internal/worktree"
)

func TestFormatListTableUsesUnicodeBoxBorders(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{Worktree: worktree.Worktree{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true, Staged: 2, Unstaged: 1}},
		{Worktree: worktree.Worktree{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", IsMerged: true}},
	})

	stripped := StripAnsi(got)
	for _, fragment := range []string{
		"┌",
		"┬",
		"│ INDEX",
		"│ STATUS",
		"│ AHEAD/BEHIND",
		"│ CHANGES",
		"├",
		"┼",
		"└",
		"┴",
		"│ 1",
		"[CURRENT]",
		"[MERGED]",
	} {
		if !strings.Contains(stripped, fragment) {
			t.Fatalf("expected %q in table output, got %q", fragment, stripped)
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
				Unstaged:    1,
				IsDirty:     true,
			},
		},
	})

	stripped := StripAnsi(got)
	if !strings.Contains(stripped, "│ 2") {
		t.Fatalf("expected first row for wrapped item, got %q", stripped)
	}
	if !strings.Contains(stripped, "current-dirty-status") || !strings.Contains(stripped, "very/long/path") {
		t.Fatalf("expected full path content across wrapped lines, got %q", stripped)
	}
}

func TestFormatListTableShowsDetailOutsidePathCell(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{
			Worktree: worktree.Worktree{
				Index:       2,
				BranchLabel: "(detached)",
				Path:        "/repo/.worktrees/scratch",
				IsDetached:  true,
			},
			Detail: "idle scratch",
		},
	})

	stripped := StripAnsi(got)
	for _, line := range strings.Split(stripped, "\n") {
		if strings.Contains(line, "idle scratch") && strings.Contains(line, "/repo/.worktrees/scratch") {
			t.Fatalf("expected detail to be separate from path cell, got line %q in table %q", line, stripped)
		}
	}
	if !strings.Contains(stripped, "(detached)") || !strings.Contains(stripped, "idle scratch") {
		t.Fatalf("expected branch and detail in table output, got %q", stripped)
	}
}

func TestFormatListTableHidesAheadBehindForMergedWorktrees(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{
			Worktree: worktree.Worktree{
				Index:       2,
				BranchLabel: "fix/merged",
				Path:        "/repo/.worktrees/fix-merged",
				IsMerged:    true,
				Ahead:       4,
				Behind:      9,
			},
		},
		{
			Worktree: worktree.Worktree{
				Index:       3,
				BranchLabel: "feat/open",
				Path:        "/repo/.worktrees/feat-open",
				Ahead:       2,
				Behind:      1,
			},
		},
	})

	stripped := StripAnsi(got)
	for _, line := range strings.Split(stripped, "\n") {
		if strings.Contains(line, "fix/merged") && (strings.Contains(line, "↑4") || strings.Contains(line, "↓9")) {
			t.Fatalf("expected merged row to hide ahead/behind, got line %q in table %q", line, stripped)
		}
	}
	if !strings.Contains(stripped, "feat/open") || !strings.Contains(stripped, "↑2 ↓1") {
		t.Fatalf("expected open branch to keep ahead/behind, got %q", stripped)
	}
}
