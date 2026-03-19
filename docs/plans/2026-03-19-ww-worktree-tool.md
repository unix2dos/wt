# ww Worktree Tool Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reposition the current `wt` project into a shell-first `ww` worktree tool whose default behavior is interactive switching with direct `cd`, while adding `list` and `new` for a broader worktree workflow.

**Architecture:** Keep Git parsing and worktree normalization in Go, but move the user-facing entrypoint to a shell function named `ww` so the default command can change the current shell directory. Introduce a hidden helper binary for machine-safe subcommands (`switch-path`, `list`, `new-path`) and use shell glue for the final `cd`. Add a small local state file for per-repo MRU ordering and build an internal arrow-key TUI fallback when `fzf` is unavailable.

**Tech Stack:** Go CLI helper, bash/zsh shell function, Git porcelain output, optional `fzf`, local state file under XDG/home config, Go tests + end-to-end shell tests.

---

## Architecture Summary

### Public Commands

- `ww`
  - Default entrypoint.
  - Equivalent to `ww switch`.
  - Chooses a target interactively, then `cd`s current shell into it.
- `ww switch`
  - Interactive selection and shell directory change.
- `ww switch <name>`
  - Direct match by worktree name.
  - Match rule: exact first, then unique prefix.
- `ww list`
  - Human-readable worktree listing only.
- `ww new <name>`
  - Create branch `<name>` from current HEAD into `./.worktrees/<name>`, then `cd`.

### Internal Helper Binary

- Install a helper binary named `ww-helper`.
- Keep it pure CLI with no shell mutation.
- It will expose machine-facing subcommands:
  - `ww-helper switch-path`
  - `ww-helper switch-path <name>`
  - `ww-helper list`
  - `ww-helper new-path <name>`
  - `ww-helper --help`

### State Model

- Persist per-repo MRU ordering in a local state file.
- Suggested location:
  - macOS/Linux: `${XDG_STATE_HOME:-$HOME/.local/state}/ww/state.json`
- Key by canonical repo root path.
- Update MRU after successful switch and successful new.

### Selection Strategy

- `fzf` installed:
  - default interactive selector uses `fzf`
- `fzf` missing:
  - fallback to built-in arrow-key TUI
- Both selectors consume the same ordered worktree list.

### Naming Assumption

- Repo/project/release assets move to `ww`.
- Hidden helper binary remains `ww-helper` to avoid command ambiguity with the shell function.

---

### Task 1: Create the planning/doc scaffold

**Files:**
- Create: `docs/plans/2026-03-19-ww-worktree-tool.md`

**Step 1: Verify plan directory does not exist yet**

Run: `cd /Users/liuwei/workspace/wt && ls docs docs/plans`
Expected: missing path error

**Step 2: Create the plan file**

Create this document with the architecture and tasks below.

**Step 3: Verify file exists**

Run: `cd /Users/liuwei/workspace/wt && sed -n '1,40p' docs/plans/2026-03-19-ww-worktree-tool.md`
Expected: header renders correctly

**Step 4: Commit**

```bash
git add docs/plans/2026-03-19-ww-worktree-tool.md
git commit -m "docs: add ww implementation plan"
```

### Task 2: Rename command surface from `wt` to `ww`

**Files:**
- Modify: `README.md`
- Modify: `install.sh`
- Modify: `uninstall.sh`
- Modify: `.github/workflows/release.yml`
- Modify: `scripts/release.sh`
- Modify: `scripts/install-release.sh`
- Modify: `test/release/release_test.go`
- Modify: `test/install/install_test.go`
- Modify: `test/online_install/online_install_test.go`

**Step 1: Write the failing release/install naming tests**

Update tests to expect:
- release artifacts named `ww-vX.Y.Z-...`
- installed public command named `ww`
- installed wrapper named `ww.sh` or equivalent managed shell entry
- helper binary named `ww-helper`

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/release ./test/install ./test/online_install -v`
Expected: FAIL because current assets and install paths still use `wt`

**Step 3: Write minimal implementation**

Update packaging, README, install, uninstall, and workflow files to consistently publish/install:
- `ww` shell entrypoint
- `ww-helper` helper binary
- `ww-v...tar.gz` release archives

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/release ./test/install ./test/online_install -v`
Expected: PASS

**Step 5: Commit**

```bash
git add README.md install.sh uninstall.sh .github/workflows/release.yml scripts/release.sh scripts/install-release.sh test/release/release_test.go test/install/install_test.go test/online_install/online_install_test.go
git commit -m "refactor: rename public tool surface to ww"
```

### Task 3: Introduce helper-binary-first command architecture

**Files:**
- Create: `cmd/ww-helper/main.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`
- Modify: `go.mod`
- Modify: `test/e2e/e2e_test.go`
- Modify: `test/e2e/testrepo.go`

**Step 1: Write the failing tests for new helper command shape**

Add tests that expect:
- `ww-helper --help` mentions `switch-path`, `list`, `new-path`
- `ww-helper switch-path 2` or `ww-helper switch-path alpha` returns a path, not a `cd`
- old `cmd/wt` path is no longer the primary build target

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/app ./test/e2e -v`
Expected: FAIL because the current entrypoint is still `wt`

**Step 3: Write minimal implementation**

Split current app runner into helper-oriented subcommands:
- `switch-path` for machine-readable path selection
- `list` for human-readable output
- `new-path` for create-and-return-path behavior

Keep stdout machine-clean for `switch-path` and `new-path`.

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/app ./test/e2e -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/ww-helper/main.go internal/app/run.go internal/app/run_test.go test/e2e/e2e_test.go test/e2e/testrepo.go go.mod
git commit -m "refactor: introduce ww helper binary"
```

### Task 4: Add MRU state model for per-repo ordering

**Files:**
- Create: `internal/state/store.go`
- Create: `internal/state/store_test.go`
- Modify: `internal/worktree/model.go`
- Modify: `internal/worktree/normalize.go`
- Modify: `internal/worktree/normalize_test.go`
- Modify: `internal/git/list.go`
- Modify: `internal/app/run.go`

**Step 1: Write the failing tests**

Add tests covering:
- repo root keying
- current worktree always first
- non-current worktrees ordered by MRU timestamps
- missing MRU data falls back to deterministic name ordering
- successful switch/new updates MRU

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/state ./internal/worktree ./internal/app -v`
Expected: FAIL because no state store exists and normalization still sorts by path

**Step 3: Write minimal implementation**

Implement:
- state file load/save
- repo-root-scoped MRU map
- normalize function using:
  - current first
  - then descending last-used
  - then deterministic name fallback

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/state ./internal/worktree ./internal/app -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/state/store.go internal/state/store_test.go internal/worktree/model.go internal/worktree/normalize.go internal/worktree/normalize_test.go internal/git/list.go internal/app/run.go
git commit -m "feat: add repo-scoped MRU ordering"
```

### Task 5: Add name matching for direct `switch`

**Files:**
- Create: `internal/worktree/match.go`
- Create: `internal/worktree/match_test.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`

**Step 1: Write the failing tests**

Cover:
- exact branch match wins
- exact worktree name wins if supported
- unique prefix match succeeds
- ambiguous prefix match errors
- no match errors

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/worktree ./internal/app -run 'TestMatch|TestRun' -v`
Expected: FAIL because direct name matching is missing

**Step 3: Write minimal implementation**

Add a matcher that:
- extracts a worktree name from branch label and/or leaf directory name
- tries exact match first
- falls back to unique prefix match
- returns structured not-found/ambiguous errors

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/worktree ./internal/app -run 'TestMatch|TestRun' -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/worktree/match.go internal/worktree/match_test.go internal/app/run.go internal/app/run_test.go
git commit -m "feat: add direct switch name matching"
```

### Task 6: Replace numeric fallback with built-in arrow-key TUI

**Files:**
- Create: `internal/ui/tui.go`
- Create: `internal/ui/tui_test.go`
- Modify: `internal/ui/menu.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`

**Step 1: Write the failing tests**

Cover:
- arrow-up/arrow-down selection changes highlighted row
- enter returns selected item
- escape or ctrl-c cancels
- rendering shows current marker and active row
- selection works without `fzf`

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/ui ./internal/app -v`
Expected: FAIL because only numeric input exists today

**Step 3: Write minimal implementation**

Implement a simple TTY selector using raw terminal input:
- `↑` / `↓`
- `j` / `k` optional convenience
- `Enter`
- cancel handling

Do not remove existing formatting helpers until replacement passes.

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/ui ./internal/app -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/tui.go internal/ui/tui_test.go internal/ui/menu.go internal/app/run.go internal/app/run_test.go
git commit -m "feat: add built-in arrow-key selector"
```

### Task 7: Make selector routing prefer `fzf`, then TUI fallback

**Files:**
- Modify: `internal/ui/fzf.go`
- Modify: `internal/ui/fzf_test.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`

**Step 1: Write the failing tests**

Cover:
- `fzf` present: helper uses `fzf`
- `fzf` missing: helper falls back to TUI automatically
- `--help` explains auto-selection behavior
- cancellation propagates without mutating output

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/ui ./internal/app -v`
Expected: FAIL because current behavior requires explicit `--fzf`

**Step 3: Write minimal implementation**

Remove `--fzf` as the primary UX requirement and route selection by environment:
- detect `fzf`
- use `fzf` if available
- otherwise use built-in TUI

Retain an explicit override flag only if implementation needs one for tests/debugging.

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/ui ./internal/app -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/fzf.go internal/ui/fzf_test.go internal/app/run.go internal/app/run_test.go
git commit -m "feat: auto-select fzf or built-in tui"
```

### Task 8: Implement `list` command for human-readable output

**Files:**
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`
- Modify: `README.md`

**Step 1: Write the failing tests**

Cover:
- `ww-helper list` prints current marker, branch, path
- ordering matches MRU logic
- no shell mutation side effects

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/app -run TestRunList -v`
Expected: FAIL because `list` subcommand does not exist

**Step 3: Write minimal implementation**

Add `list` command in helper binary and document it.

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/app -run TestRunList -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/run.go internal/app/run_test.go README.md
git commit -m "feat: add ww list command"
```

### Task 9: Implement `new-path` helper behavior

**Files:**
- Create: `internal/git/create.go`
- Create: `internal/git/create_test.go`
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`
- Modify: `test/e2e/testrepo.go`
- Modify: `test/e2e/e2e_test.go`

**Step 1: Write the failing tests**

Cover:
- `new-path <name>` creates branch from current HEAD
- worktree path is `./.worktrees/<name>`
- existing branch/path errors cleanly
- stdout returns only the created path
- MRU updates after create

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/git ./internal/app ./test/e2e -v`
Expected: FAIL because create flow does not exist

**Step 3: Write minimal implementation**

Use `git worktree add -b <name> ./.worktrees/<name>` from repo root.

Ensure repo root is canonicalized before path return.

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./internal/git ./internal/app ./test/e2e -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/git/create.go internal/git/create_test.go internal/app/run.go internal/app/run_test.go test/e2e/testrepo.go test/e2e/e2e_test.go
git commit -m "feat: add ww new command"
```

### Task 10: Install shell-first `ww` function

**Files:**
- Create: `shell/ww.sh`
- Modify: `install.sh`
- Modify: `uninstall.sh`
- Modify: `test/install/install_test.go`
- Modify: `test/online_install/online_install_test.go`
- Modify: `README.md`

**Step 1: Write the failing tests**

Cover:
- installed rc sources `ww.sh`
- `ww` function exists after sourcing rc
- bare `ww` calls helper `switch-path`, then `cd`s
- `ww switch <name>` and `ww new <name>` `cd` on success
- `ww list` does not `cd`
- failure/cancel do not change current directory

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/install ./test/online_install -v`
Expected: FAIL because current wrapper shape is still `cwt`

**Step 3: Write minimal implementation**

Implement `shell/ww.sh` with:
- `ww()` shell function dispatch
- call `ww-helper switch-path` and `cd`
- call `ww-helper new-path` and `cd`
- call `ww-helper list` and print only
- `ww switch` aliasing bare `ww`

Install:
- public shell wrapper
- hidden helper binary

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/install ./test/online_install -v`
Expected: PASS

**Step 5: Commit**

```bash
git add shell/ww.sh install.sh uninstall.sh test/install/install_test.go test/online_install/online_install_test.go README.md
git commit -m "feat: install shell-first ww command"
```

### Task 11: Refresh end-to-end coverage for real workflows

**Files:**
- Modify: `test/e2e/e2e_test.go`
- Modify: `test/e2e/testrepo.go`

**Step 1: Write the failing tests**

Add e2e scenarios for:
- direct switch by exact name
- direct switch by unique prefix
- ambiguous prefix returns error
- list output order reflects MRU updates
- new creates `.worktrees/<name>`

**Step 2: Run test to verify it fails**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/e2e -v`
Expected: FAIL until helper and state updates are wired together

**Step 3: Write minimal implementation**

Only fill missing glue revealed by e2e failures.

**Step 4: Run test to verify it passes**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/e2e -v`
Expected: PASS

**Step 5: Commit**

```bash
git add test/e2e/e2e_test.go test/e2e/testrepo.go
git commit -m "test: cover ww switch list and new workflows"
```

### Task 12: Final docs and release pass

**Files:**
- Modify: `README.md`
- Modify: `.github/workflows/release.yml`
- Modify: `scripts/release.sh`
- Modify: `scripts/install-release.sh`
- Test: `test/release/release_test.go`
- Test: `test/install/install_test.go`
- Test: `test/online_install/online_install_test.go`

**Step 1: Write or update the failing doc/release tests**

Verify:
- release assets are `ww-*`
- installer instructions mention `ww`
- help text mentions auto `fzf` fallback and shell-first behavior

**Step 2: Run verification**

Run: `cd /Users/liuwei/workspace/wt && go test ./test/release ./test/install ./test/online_install -v`
Expected: PASS

**Step 3: Run full project verification**

Run: `cd /Users/liuwei/workspace/wt && go test ./...`
Expected: PASS

**Step 4: Run release smoke test**

Run: `cd /Users/liuwei/workspace/wt && bash scripts/release.sh v0.2.0`
Expected: `dist/` contains `ww-v0.2.0-...`, `checksums.txt`, and installer assets

**Step 5: Commit**

```bash
git add README.md .github/workflows/release.yml scripts/release.sh scripts/install-release.sh test/release/release_test.go test/install/install_test.go test/online_install/online_install_test.go
git commit -m "docs: finalize ww release flow"
```

---

## ADR Notes

### ADR-001: Shell-first public command

- **Status:** Accepted
- **Decision:** `ww` is a shell function, not the helper binary.
- **Why:** default behavior must `cd` the current shell.
- **Trade-off:** adds shell integration complexity, but matches the core product promise.

### ADR-002: Hidden helper binary

- **Status:** Accepted
- **Decision:** keep a pure Go helper binary behind the shell command.
- **Why:** preserves testability, machine-clean outputs, and release packaging sanity.
- **Trade-off:** two installed artifacts instead of one.

### ADR-003: `fzf` optional, not required

- **Status:** Accepted
- **Decision:** prefer `fzf` when present, otherwise use built-in TUI.
- **Why:** preserves a fast path for `fzf` users without making it a hard dependency.
- **Trade-off:** more UI code to own and test.

### ADR-004: Per-repo MRU persistence

- **Status:** Accepted
- **Decision:** maintain local state keyed by canonical repo root.
- **Why:** required by the product’s ordering rule.
- **Trade-off:** adds state migration/error-handling surface.

---

## Risks

- Raw terminal TUI behavior differs between macOS and Linux shells.
- MRU persistence can corrupt or drift if state writes are not atomic.
- Repo rename to `ww` touches release assets, installer expectations, and user docs in one batch.
- Shell-first architecture increases the number of integration tests needed to trust behavior.

## Mitigations

- Keep helper binary pure and push shell mutation into a thin wrapper.
- Use atomic state-file writes via temp file + rename.
- Land rename early so later tasks do not keep carrying `wt` assumptions.
- Add end-to-end shell tests before final release.
