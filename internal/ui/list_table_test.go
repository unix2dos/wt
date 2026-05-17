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
		{Worktree: worktree.Worktree{Index: 3, BranchLabel: "feat-b", Path: "/repo/.worktrees/feat-b", Ahead: 1}},
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

func TestFormatListTableOmitsEmptyOptionalColumns(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{Worktree: worktree.Worktree{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true}},
		{Worktree: worktree.Worktree{Index: 2, BranchLabel: "scratch", Path: "/repo/.worktrees/scratch"}, Detail: "idle"},
	})

	stripped := StripAnsi(got)
	for _, unexpected := range []string{"AHEAD/BEHIND", "CHANGES"} {
		if strings.Contains(stripped, unexpected) {
			t.Fatalf("expected compact table to omit %q, got %q", unexpected, stripped)
		}
	}
	for _, want := range []string{"│ INDEX", "│ STATUS", "│ BRANCH", "│ PATH", "scratch", "idle"} {
		if !strings.Contains(stripped, want) {
			t.Fatalf("expected %q in compact table, got %q", want, stripped)
		}
	}
}

func TestFormatListTableShowsAheadBehindColumnWhenNeeded(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{Worktree: worktree.Worktree{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true}},
		{Worktree: worktree.Worktree{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", Ahead: 1, Behind: 2}},
	})

	stripped := StripAnsi(got)
	if !strings.Contains(stripped, "AHEAD/BEHIND") || !strings.Contains(stripped, "↑1 ↓2") {
		t.Fatalf("expected ahead/behind column when needed, got %q", stripped)
	}
	if strings.Contains(stripped, "CHANGES") {
		t.Fatalf("expected changes column to stay hidden when empty, got %q", stripped)
	}
}

func TestFormatListTableShowsChangesColumnWhenNeeded(t *testing.T) {
	got := FormatListTable([]ListTableEntry{
		{Worktree: worktree.Worktree{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true}},
		{Worktree: worktree.Worktree{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", Unstaged: 1}},
	})

	stripped := StripAnsi(got)
	if !strings.Contains(stripped, "CHANGES") || !strings.Contains(stripped, "~1") {
		t.Fatalf("expected changes column when needed, got %q", stripped)
	}
	if strings.Contains(stripped, "AHEAD/BEHIND") {
		t.Fatalf("expected ahead/behind column to stay hidden when empty, got %q", stripped)
	}
}

func TestFormatListTableVerboseKeepsOptionalColumns(t *testing.T) {
	got := FormatListTableWithOptions([]ListTableEntry{
		{Worktree: worktree.Worktree{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true}},
	}, ListTableOptions{ShowEmptyOptionalColumns: true})

	stripped := StripAnsi(got)
	for _, want := range []string{"AHEAD/BEHIND", "CHANGES"} {
		if !strings.Contains(stripped, want) {
			t.Fatalf("expected verbose table to keep %q, got %q", want, stripped)
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
