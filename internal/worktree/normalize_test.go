package worktree

import "testing"

func TestNormalizeOrdersCurrentFirstAndAssignsStableIndexes(t *testing.T) {
	items := []Worktree{
		{Path: "/repo/.worktrees/beta", BranchLabel: "beta"},
		{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha"},
		{Path: "/repo", BranchLabel: "main"},
	}

	got := Normalize(items, "/repo/./")

	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if !got[0].IsCurrent || got[0].Path != "/repo" {
		t.Fatalf("expected current worktree first, got %#v", got[0])
	}
	if got[0].Index != 1 || got[1].Index != 2 || got[2].Index != 3 {
		t.Fatalf("expected 1-based sequential indexes, got %#v", got)
	}
	if got[1].Path != "/repo/.worktrees/alpha" || got[2].Path != "/repo/.worktrees/beta" {
		t.Fatalf("expected remaining worktrees sorted by path, got %#v", got)
	}
}
