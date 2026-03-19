package worktree

import "testing"

func TestMatchExactBranchWinsOverDirectoryName(t *testing.T) {
	items := []Worktree{
		{Path: "/repo/.worktrees/branch-alias", BranchLabel: "alpha"},
		{Path: "/repo/.worktrees/alpha", BranchLabel: "feature/other"},
	}

	got, err := Match(items, "alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "/repo/.worktrees/branch-alias" {
		t.Fatalf("expected exact branch match, got %#v", got)
	}
}

func TestMatchFallsBackToExactDirectoryName(t *testing.T) {
	items := []Worktree{
		{Path: "/repo/.worktrees/docs", BranchLabel: "feature/docs"},
	}

	got, err := Match(items, "docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "/repo/.worktrees/docs" {
		t.Fatalf("expected exact directory match, got %#v", got)
	}
}

func TestMatchUniquePrefixSucceeds(t *testing.T) {
	items := []Worktree{
		{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha"},
		{Path: "/repo/.worktrees/beta", BranchLabel: "beta"},
	}

	got, err := Match(items, "alp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "/repo/.worktrees/alpha" {
		t.Fatalf("expected unique prefix match, got %#v", got)
	}
}

func TestMatchAmbiguousPrefixFails(t *testing.T) {
	items := []Worktree{
		{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha"},
		{Path: "/repo/.worktrees/alpine", BranchLabel: "alpine"},
	}

	_, err := Match(items, "alp")
	if err == nil {
		t.Fatal("expected ambiguous prefix error")
	}
	if err.Error() != `ambiguous worktree match "alp": alpha, alpine` {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMatchNoMatchFails(t *testing.T) {
	items := []Worktree{
		{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha"},
	}

	_, err := Match(items, "gamma")
	if err == nil {
		t.Fatal("expected no-match error")
	}
	if err.Error() != `no worktree matches "gamma"` {
		t.Fatalf("unexpected error: %v", err)
	}
}
