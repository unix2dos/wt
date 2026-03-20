# WW Dirty Worktree Indicator Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Render dirty-state markers for each worktree across `ww` list and selector views using `git status --porcelain`.

**Architecture:** Extend the shared `worktree.Worktree` model with a dirty flag, annotate it centrally in the Git listing flow, and route all status text through one UI formatting helper. Reuse the existing dirty-check behavior already used by removal flows so the definition stays consistent.

**Tech Stack:** Go, Git CLI, standard library tests

---

### Task 1: Add failing tests for dirty-state annotation and rendering

**Files:**
- Modify: `internal/git/list_test.go`
- Modify: `internal/ui/menu_test.go`
- Modify: `internal/ui/tui_test.go`
- Modify: `internal/ui/fzf_test.go`
- Modify: `internal/app/run_test.go`

**Step 1: Write the failing tests**

Add tests that expect:

- `ListWorktrees` marks a worktree dirty when `git -C <path> status --porcelain` returns output
- menu and list output render `ACTIVE*` for the current dirty worktree
- menu, TUI, and `fzf` output render `DIRTY` for a non-current dirty worktree

**Step 2: Run test to verify it fails**

Run: `go test ./internal/git ./internal/ui ./internal/app`

Expected: FAIL because dirty state is not yet part of listed worktrees or shared status formatting.

**Step 3: Write minimal implementation**

Add `IsDirty` to the worktree model and update rendering/tests only as needed to satisfy the new expectations.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/git ./internal/ui ./internal/app`

Expected: PASS

### Task 2: Reuse the dirty helper in list and removal flows

**Files:**
- Modify: `internal/git/list.go`
- Modify: `internal/git/remove.go`
- Modify: `internal/worktree/model.go`
- Modify: `internal/ui/menu.go`
- Modify: `internal/ui/fzf.go`
- Modify: `internal/app/run.go`

**Step 1: Write the failing test**

Use the tests from Task 1 as the regression target. No extra test file is needed if coverage already fails for the missing behavior.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/git ./internal/ui ./internal/app`

Expected: FAIL until the dirty helper is shared and list rendering is updated.

**Step 3: Write minimal implementation**

- extract the Git dirty check into reusable list/removal helpers
- annotate listed worktrees with dirty state
- centralize status-label formatting for app/UI/fzf rendering

**Step 4: Run test to verify it passes**

Run: `go test ./internal/git ./internal/ui ./internal/app`

Expected: PASS

### Task 3: Verify end-to-end CLI behavior

**Files:**
- Modify: `test/e2e/e2e_test.go`

**Step 1: Write the failing test**

Add an e2e test that creates an untracked file in a linked worktree and expects `ww list` to show `DIRTY` for that worktree.

**Step 2: Run test to verify it fails**

Run: `go test ./test/e2e -run TestCLIListShowsDirtyWorktrees`

Expected: FAIL because CLI list output does not yet expose dirty state.

**Step 3: Write minimal implementation**

No extra production changes should be needed beyond Tasks 1-2; only refine formatting if the e2e output exposes a gap.

**Step 4: Run test to verify it passes**

Run: `go test ./test/e2e -run TestCLIListShowsDirtyWorktrees`

Expected: PASS

### Task 4: Final verification

**Files:**
- Modify: `docs/reference.md` (only if the rendered status text in user-facing docs now needs updating)

**Step 1: Run focused verification**

Run: `go test ./internal/git ./internal/ui ./internal/app ./test/e2e`

Expected: PASS

**Step 2: Run full verification**

Run: `go test ./...`

Expected: PASS

**Step 3: Commit**

```bash
git add docs/plans/2026-03-20-ww-dirty-worktree-indicator-design.md docs/plans/2026-03-20-ww-dirty-worktree-indicator.md internal/git/list.go internal/git/remove.go internal/git/list_test.go internal/worktree/model.go internal/ui/menu.go internal/ui/menu_test.go internal/ui/tui_test.go internal/ui/fzf.go internal/ui/fzf_test.go internal/app/run.go internal/app/run_test.go test/e2e/e2e_test.go docs/reference.md
git commit -m "feat: show dirty worktree status in list views"
```
