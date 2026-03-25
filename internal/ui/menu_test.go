package ui

import (
	"bytes"
	"strings"
	"testing"

	"ww/internal/worktree"
)

func TestRenderMenuIncludesIndexBranchPathAndActiveStatus(t *testing.T) {
	var buf bytes.Buffer

	RenderMenu(&buf, []worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true, IsDirty: true},
		{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", IsDirty: true},
	})

	got := buf.String()
	if !strings.Contains(got, "[1] [CURRENT] [DIRTY] main   /repo") {
		t.Fatalf("expected current row, got %q", got)
	}
	if !strings.Contains(got, "[2] [DIRTY]           feat-a /repo/.worktrees/feat-a") {
		t.Fatalf("expected non-current row, got %q", got)
	}
	if !strings.Contains(got, "Select a worktree [number]: ") {
		t.Fatalf("expected prompt, got %q", got)
	}
}

func TestRenderMenuAlignsPathColumnAcrossRows(t *testing.T) {
	var buf bytes.Buffer

	RenderMenu(&buf, []worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true},
		{Index: 2, BranchLabel: "codex/current-dirty-status", Path: "/repo/.worktrees/current-dirty-status", IsDirty: true},
	})

	lines := strings.Split(buf.String(), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least two rendered rows, got %q", buf.String())
	}

	mainPathCol := strings.Index(lines[0], "/repo")
	featurePathCol := strings.Index(lines[1], "/repo/.worktrees/current-dirty-status")
	if mainPathCol == -1 || featurePathCol == -1 {
		t.Fatalf("expected both paths in output, got %q", buf.String())
	}
	if mainPathCol != featurePathCol {
		t.Fatalf("expected aligned path columns, got %d and %d in %q", mainPathCol, featurePathCol, buf.String())
	}
}

func TestReadSelectionRetriesAfterInvalidInput(t *testing.T) {
	var stderr bytes.Buffer

	index, err := ReadSelection(strings.NewReader("abc\n2\n"), &stderr, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if index != 2 {
		t.Fatalf("expected selection 2, got %d", index)
	}
	if !strings.Contains(stderr.String(), "invalid worktree selection") {
		t.Fatalf("expected invalid selection message, got %q", stderr.String())
	}
}
