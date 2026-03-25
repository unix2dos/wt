package ui

import (
	"testing"

	"ww/internal/worktree"
)

func TestStatusLabelShowsCurrentForCurrentCleanWorktree(t *testing.T) {
	got := StatusLabel(worktree.Worktree{IsCurrent: true})
	if got != "[CURRENT]" {
		t.Fatalf("expected [CURRENT], got %q", got)
	}
}

func TestStatusLabelShowsCurrentAndDirtyForCurrentDirtyWorktree(t *testing.T) {
	got := StatusLabel(worktree.Worktree{IsCurrent: true, IsDirty: true})
	if got != "[CURRENT] [DIRTY]" {
		t.Fatalf("expected [CURRENT] [DIRTY], got %q", got)
	}
}

func TestStatusLabelShowsDirtyForDirtyNonCurrentWorktree(t *testing.T) {
	got := StatusLabel(worktree.Worktree{IsDirty: true})
	if got != "[DIRTY]" {
		t.Fatalf("expected [DIRTY], got %q", got)
	}
}

func TestStatusLabelIsBlankForCleanNonCurrentWorktree(t *testing.T) {
	got := StatusLabel(worktree.Worktree{})
	if got != "" {
		t.Fatalf("expected blank status label, got %q", got)
	}
}
