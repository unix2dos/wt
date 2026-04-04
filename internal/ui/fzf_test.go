package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ww/internal/worktree"
)

func TestFormatFzfCandidatesIncludesAllColumns(t *testing.T) {
	got := string(formatFzfCandidates([]worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true, IsDirty: true},
		{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", IsDirty: true},
	}))

	stripped := StripAnsi(got)
	// Current row has ★ marker and [CURRENT] tag
	if !strings.Contains(stripped, "★") {
		t.Fatalf("expected ★ marker for current worktree, got %q", stripped)
	}
	if !strings.Contains(stripped, "[CURRENT]") {
		t.Fatalf("expected [CURRENT] tag, got %q", stripped)
	}
	// Non-current row has space marker
	lines := strings.Split(strings.TrimSpace(stripped), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two candidates, got %q", stripped)
	}
	secondFields := strings.SplitN(lines[1], "\t", 2)
	if len(secondFields) < 2 || !strings.HasPrefix(secondFields[1], " ") {
		t.Fatalf("expected space marker for non-current worktree, got %q", lines[1])
	}
}

func TestFormatFzfCandidatesPadsFields(t *testing.T) {
	got := string(formatFzfCandidates([]worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true},
		{Index: 2, BranchLabel: "codex/current-dirty-status", Path: "/repo/.worktrees/current-dirty-status", IsDirty: true},
	}))

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected two candidates, got %q", got)
	}

	first := strings.Split(lines[0], "\t")
	second := strings.Split(lines[1], "\t")
	if len(first) != 6 || len(second) != 6 {
		t.Fatalf("expected six tab-separated fields, got first=%d second=%d: %q", len(first), len(second), got)
	}

	stripped := StripAnsi(got)
	if !strings.Contains(stripped, "main") || !strings.Contains(stripped, "codex/current-dirty-status") {
		t.Fatalf("expected branch names to survive padding, got %q", stripped)
	}
}

func TestSelectWorktreeWithFzfReturnsSelectedWorktree(t *testing.T) {
	runner := &fakeFzfRunner{
		lookPath: "/usr/bin/fzf",
		stdout:   []byte("2\t  \tfeat-a\t\t\t/repo/.worktrees/feat-a\n"),
	}

	got, err := SelectWorktreeWithFzf(context.Background(), []worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true},
		{Index: 2, BranchLabel: "feat-a", Path: "/repo/.worktrees/feat-a", IsDirty: true},
	}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "/repo/.worktrees/feat-a" {
		t.Fatalf("expected selected worktree, got %#v", got)
	}
	if !strings.Contains(strings.Join(runner.gotArgs, " "), "--nth=2..") {
		t.Fatalf("expected fzf to search non-index fields without rewriting output, args=%q", runner.gotArgs)
	}
	if !strings.Contains(strings.Join(runner.gotArgs, " "), "--pointer=*") {
		t.Fatalf("expected fzf pointer marker to follow active selection, args=%q", runner.gotArgs)
	}
	if !strings.Contains(strings.Join(runner.gotArgs, " "), "--tac") {
		t.Fatalf("expected fzf to keep the list near the prompt while rendering top-down, args=%q", runner.gotArgs)
	}
	if !strings.Contains(strings.Join(runner.gotArgs, " "), "--bind=load:pos(2)") {
		t.Fatalf("expected fzf to focus current worktree by default, args=%q", runner.gotArgs)
	}
}

func TestSelectWorktreeWithFzfFocusesCurrentWorktreeByDefault(t *testing.T) {
	runner := &fakeFzfRunner{
		lookPath: "/usr/bin/fzf",
		stdout:   []byte("2\t[CURRENT]          \tmain\t/repo\n"),
	}

	_, err := SelectWorktreeWithFzf(context.Background(), []worktree.Worktree{
		{Index: 1, BranchLabel: "alpha", Path: "/repo/.worktrees/alpha"},
		{Index: 2, BranchLabel: "main", Path: "/repo", IsCurrent: true, IsDirty: true},
	}, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.Join(runner.gotArgs, " "), "--bind=load:pos(1)") {
		t.Fatalf("expected fzf to position cursor on current worktree, args=%q", runner.gotArgs)
	}
}

func TestSelectWorktreeWithFzfReturnsErrFzfNotInstalled(t *testing.T) {
	_, err := SelectWorktreeWithFzf(context.Background(), nil, &fakeFzfRunner{
		lookPathErr: errors.New("missing"),
	})
	if !errors.Is(err, ErrFzfNotInstalled) {
		t.Fatalf("expected ErrFzfNotInstalled, got %v", err)
	}
}

func TestSelectWorktreeWithFzfReturnsErrSelectionCanceled(t *testing.T) {
	_, err := SelectWorktreeWithFzf(context.Background(), []worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true},
	}, &fakeFzfRunner{
		lookPath: "/usr/bin/fzf",
		err:      exitError{code: 130},
	})
	if !errors.Is(err, ErrSelectionCanceled) {
		t.Fatalf("expected ErrSelectionCanceled, got %v", err)
	}
}

func TestFormatFzfCandidatesShowsFileChanges(t *testing.T) {
	got := string(formatFzfCandidates([]worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true, Staged: 2, Unstaged: 1},
		{Index: 2, BranchLabel: "feat-a", Path: "/wt/feat-a", Untracked: 3},
	}))

	stripped := StripAnsi(got)
	if !strings.Contains(stripped, "+2 ~1") {
		t.Fatalf("expected staged/unstaged counts in fzf output, got %q", stripped)
	}
	if !strings.Contains(stripped, "?3") {
		t.Fatalf("expected untracked count in fzf output, got %q", stripped)
	}
}

func TestFormatFzfCandidatesShowsMergedTagAndDimsBranchPath(t *testing.T) {
	got := string(formatFzfCandidates([]worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true},
		{Index: 2, BranchLabel: "fix/typo", Path: "/wt/fix-typo", IsMerged: true},
	}))

	if !strings.Contains(got, "[MERGED]") {
		t.Fatalf("expected [MERGED] in fzf output, got %q", got)
	}
	if !strings.Contains(got, "[CURRENT]") {
		t.Fatalf("expected [CURRENT] in fzf output, got %q", got)
	}
	// Merged row branch and path should be dimmed (ANSI dim = \x1b[2m)
	if !strings.Contains(got, "\x1b[2mfix/typo\x1b[0m") {
		t.Fatalf("expected dimmed branch for merged worktree, got %q", got)
	}
	if !strings.Contains(got, "\x1b[2m/wt/fix-typo\x1b[0m") {
		t.Fatalf("expected dimmed path for merged worktree, got %q", got)
	}
}

func TestFormatFzfCandidatesShowsAheadBehind(t *testing.T) {
	got := string(formatFzfCandidates([]worktree.Worktree{
		{Index: 1, BranchLabel: "main", Path: "/repo", IsCurrent: true},
		{Index: 2, BranchLabel: "feat-a", Path: "/wt/feat-a", Ahead: 3, Behind: 1},
	}))

	stripped := StripAnsi(got)
	if !strings.Contains(stripped, "↑3") {
		t.Fatalf("expected ahead count in fzf output, got %q", stripped)
	}
	if !strings.Contains(stripped, "↓1") {
		t.Fatalf("expected behind count in fzf output, got %q", stripped)
	}
}

type fakeFzfRunner struct {
	lookPath    string
	lookPathErr error
	stdout      []byte
	stderr      []byte
	err         error
	gotArgs     []string
}

func (f fakeFzfRunner) LookPath(string) (string, error) {
	return f.lookPath, f.lookPathErr
}

func (f *fakeFzfRunner) Run(_ context.Context, _ string, stdin []byte, args ...string) ([]byte, []byte, error) {
	f.gotArgs = append([]string(nil), args...)
	return append([]byte(nil), f.stdout...), append([]byte(nil), f.stderr...), f.err
}

type exitError struct {
	code int
}

func (e exitError) Error() string {
	return "exit status"
}

func (e exitError) ExitCode() int {
	return e.code
}
