package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ww/internal/git"
	"ww/internal/state"
	"ww/internal/tasknote"
	"ww/internal/worktree"
)

func TestRunCheckShowsCurrentWorktreeSummary(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	notePath := filepath.Join(t.TempDir(), "git-private", "ww", "task-note.md")
	if err := tasknote.WriteFile(notePath, tasknote.Note{
		TaskLabel: "task:fix-login",
		Branch:    "alpha",
		CreatedAt: time.Date(2026, 3, 24, 12, 34, 56, 0, time.UTC),
		Intent:    "Fix the login redirect loop",
		Body:      "Created by ww.",
	}); err != nil {
		t.Fatalf("write note: %v", err)
	}

	deps := fakeDeps{
		repoKey: "/repo/.git",
		worktrees: []worktree.Worktree{
			{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha", BranchRef: "refs/heads/alpha", IsCurrent: true},
		},
		metadata: map[string]map[string]state.WorktreeMetadata{
			"/repo/.git": {
				"/repo/.worktrees/alpha": {
					Label: "task:fix-login",
				},
			},
		},
		worktreeGitPath: notePath,
	}

	code := Run(context.Background(), []string{"check"}, bytes.NewReader(nil), stdout, stderr, deps)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Path: /repo/.worktrees/alpha") {
		t.Fatalf("expected path in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Branch: alpha") {
		t.Fatalf("expected branch in output, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "Task:") {
		t.Fatalf("expected task wording removed, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Changes: clean") {
		t.Fatalf("expected clean state in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Workspace context: Fix the login redirect loop") {
		t.Fatalf("expected workspace context in output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no warnings, got %q", stderr.String())
	}
}

func TestRunCheckWarnsForDetachedWorktree(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	deps := fakeDeps{
		repoKey: "/repo/.git",
		worktrees: []worktree.Worktree{
			{Path: "/repo/.worktrees/detached", BranchLabel: "HEAD", IsCurrent: true, BranchRef: ""},
		},
	}

	code := Run(context.Background(), []string{"check"}, bytes.NewReader(nil), stdout, stderr, deps)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Branch: DETACHED") {
		t.Fatalf("expected detached branch output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "detached from a branch") {
		t.Fatalf("expected detached warning, got %q", stderr.String())
	}
}

func TestRunCheckShowsSavedContextWithoutIntent(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	notePath := filepath.Join(t.TempDir(), "git-private", "ww", "task-note.md")
	if err := tasknote.WriteFile(notePath, tasknote.Note{
		TaskLabel: "task:chores",
		Branch:    "alpha",
		CreatedAt: time.Date(2026, 3, 24, 12, 34, 56, 0, time.UTC),
		Body:      "Created by ww.",
	}); err != nil {
		t.Fatalf("write note: %v", err)
	}

	deps := fakeDeps{
		repoKey: "/repo/.git",
		worktrees: []worktree.Worktree{
			{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha", BranchRef: "refs/heads/alpha", IsCurrent: true},
		},
		metadata: map[string]map[string]state.WorktreeMetadata{
			"/repo/.git": {
				"/repo/.worktrees/alpha": {
					Label: "task:chores",
				},
			},
		},
		worktreeGitPath: notePath,
	}

	code := Run(context.Background(), []string{"check"}, bytes.NewReader(nil), stdout, stderr, deps)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Workspace context: saved notes available") {
		t.Fatalf("expected saved-context fallback output, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no warnings, got %q", stderr.String())
	}
}

func TestRunCheckWarnsForUnlabeledWorktree(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	deps := fakeDeps{
		repoKey: "/repo/.git",
		worktrees: []worktree.Worktree{
			{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha", BranchRef: "refs/heads/alpha", IsCurrent: true},
		},
	}

	code := Run(context.Background(), []string{"check"}, bytes.NewReader(nil), stdout, stderr, deps)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.Contains(stdout.String(), "Workspace context:") {
		t.Fatalf("expected no workspace context line, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "no saved workspace context") {
		t.Fatalf("expected missing-context warning, got %q", stderr.String())
	}
}

func TestRunCheckWarnsWhenTaskNoteIsMissing(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	deps := fakeDeps{
		repoKey: "/repo/.git",
		worktrees: []worktree.Worktree{
			{Path: "/repo/.worktrees/alpha", BranchLabel: "alpha", BranchRef: "refs/heads/alpha", IsCurrent: true},
		},
		metadata: map[string]map[string]state.WorktreeMetadata{
			"/repo/.git": {
				"/repo/.worktrees/alpha": {
					Label: "task:fix-login",
				},
			},
		},
		worktreeGitPath: filepath.Join(t.TempDir(), "git-private", "ww", "task-note.md"),
	}

	code := Run(context.Background(), []string{"check"}, bytes.NewReader(nil), stdout, stderr, deps)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.Contains(stdout.String(), "Task:") {
		t.Fatalf("expected task wording removed, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "saved workspace context could not be read") {
		t.Fatalf("expected human missing-context warning, got %q", stderr.String())
	}
}

func TestRunCheckReturnsNonRepoError(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(context.Background(), []string{"check"}, bytes.NewReader(nil), stdout, stderr, fakeDeps{
		err: git.ErrNotGitRepository,
	})

	if code != 3 {
		t.Fatalf("expected exit code 3, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("expected no stdout output, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "not a git repository") {
		t.Fatalf("expected non-repo message, got %q", stderr.String())
	}
}
