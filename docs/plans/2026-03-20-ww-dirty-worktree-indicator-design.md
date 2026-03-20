# WW Dirty Worktree Indicator Design

**Goal:** Show whether each listed worktree has uncommitted changes when `ww` renders worktree selections and list output.

## Product Decisions

- Dirty detection uses `git status --porcelain`.
- Any porcelain output counts as dirty, including staged changes, unstaged changes, and untracked files.
- `ww` excludes its own repository-root `.worktrees/` management directory from dirty detection so linked worktree storage does not make the main worktree look dirty by default.
- Dirty state appears anywhere worktrees are listed: `ww list`, fallback menu, built-in TUI, and `fzf`.
- The existing single status column remains, but now supports combined labels:
  - `ACTIVE` for the current clean worktree
  - `ACTIVE*` for the current dirty worktree
  - `DIRTY` for a non-current dirty worktree
  - empty for a non-current clean worktree

## Architecture

### Worktree Model

Add `IsDirty bool` to `internal/worktree.Worktree` so dirty state travels with the same data structure already used by list and selector flows.

### Git Layer

Move dirty detection into reusable Git-layer helpers instead of leaving it embedded in removal-only code. `ListWorktrees` will annotate every parsed worktree with dirty state after parsing and current-path marking.

This keeps one definition of cleanliness across listing and removal behavior.

### UI Layer

List and selector views should reuse one status-formatting helper instead of duplicating `ACTIVE`-only rendering logic in multiple files. The layout stays stable because the status still occupies one column.

## Error Handling

- If a dirty-state probe fails for a worktree, listing should return the Git error instead of silently hiding state.
- Non-repository errors still map to `ErrNotGitRepository` through the existing Git helpers.

## Testing Strategy

- Add Git-layer tests to verify `ListWorktrees` annotates dirty state from `git status --porcelain`.
- Update app/UI tests to verify `ACTIVE*` and `DIRTY` render correctly.
- Keep removal tests aligned with the shared dirty helper so both list and remove use the same definition.
