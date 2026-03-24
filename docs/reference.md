# ww Reference

Use [the README](../README.md) for the product overview and demo. This page keeps the full install, usage, command, and release reference.

## Install

For the best interactive workflow, install `fzf`. If `fzf` is not available, `ww` falls back to the built-in arrow-key selector automatically.

### One-Line Install

Install the latest release for your current platform:

```bash
curl -fsSL https://github.com/unix2dos/ww/releases/latest/download/install-release.sh | bash
source ~/.zshrc
```

For Bash:

```bash
curl -fsSL https://github.com/unix2dos/ww/releases/latest/download/install-release.sh | bash -s -- --shell bash --rc-file ~/.bashrc
source ~/.bashrc
```

Install a specific version:

```bash
curl -fsSL https://github.com/unix2dos/ww/releases/latest/download/install-release.sh | WT_VERSION=vX.Y.Z bash
```

This path does not require Go. It downloads the installer script from the latest GitHub Release, then fetches the matching release archive for your platform and runs the bundled installer.

### Install From Source

```bash
git clone https://github.com/unix2dos/ww.git
cd ww
bash install.sh
source ~/.zshrc
```

If you use Bash, reload with `source ~/.bashrc` instead.

The installer puts `ww-helper` and `ww.sh` into your target bin directory, then appends a managed shell block that exposes `ww`.

Source installs require a working Go toolchain.

### Install From A Release Bundle

```bash
tar -xzf ww-vX.Y.Z-darwin-arm64.tar.gz
cd ww-vX.Y.Z-darwin-arm64
bash install.sh
source ~/.zshrc
```

Release bundle installs copy the prebuilt `bin/ww-helper` binary and `ww.sh`, and do not require Go.

### Installer Options

```bash
bash install.sh --shell zsh
bash install.sh --shell bash --rc-file ~/.bashrc
bash install.sh --bin-dir ~/.local/bin
```

### Uninstall

```bash
bash uninstall.sh
source ~/.zshrc
```

If you installed into Bash, reload `~/.bashrc` instead.

## Usage

`ww` only works for the current repository. Run it inside a Git repository or one of that repository's worktrees.

`ww` is a shell function that switches worktrees and changes your current shell directory.

- `ww` or `ww switch` selects a worktree and switches into it.
- `ww list` prints worktrees without changing directory.
- `ww check` prints the current worktree safety summary without changing directory.
- `ww new <name>` creates a new worktree under `./.worktrees/<name>` and switches into it.
- `ww rm [<name>]` removes a worktree and deletes its branch only when that branch is already merged into the effective base branch.
- `ww help` or `ww --help` prints the command summary.
- `ww` uses `fzf` automatically when available and falls back to the built-in arrow-key selector otherwise.

### For AI Agents

Use `ww-helper` for machine-readable workflows. `ww` remains the human shell entrypoint and still treats `switch` / `new` as directory-changing commands.

Current programmatic commands:

```bash
ww-helper list --json
ww-helper new-path --json --label agent:claude-code --ttl 24h feat-a
ww-helper gc --ttl-expired --idle 7d --dry-run --json
ww-helper rm --json --non-interactive feat-a
```

`ww-helper switch-path` remains a path-printing helper. `ww check` is human-readable only in this release; agents should keep using `switch-path` and the JSON subcommands above for machine-readable flows.

#### JSON Envelope

Successful `--json` responses use:

```json
{
  "ok": true,
  "command": "list",
  "data": { ... }
}
```

Error responses use:

```json
{
  "ok": false,
  "command": "rm",
  "error": {
    "code": "WORKTREE_DIRTY",
    "message": "worktree has uncommitted changes; rerun with --force",
    "exit_code": 1
  }
}
```

#### `ww-helper list --json`

Returns an array of worktrees with:

- `path`
- `branch`
- `dirty`
- `active`
- `created_at`
- `last_used_at`
- `label`
- `ttl`

#### `ww-helper new-path --json --label agent:claude-code --ttl 24h feat-a`

Returns:

- `worktree_path`
- `branch`

`label` is stored as a single free-text string. `ttl` is fixed from creation time; this release does not include a metadata editing command.

When `label` is present, `ww` also generates a private task note for that worktree under Git's per-worktree admin directory via:

```bash
git -C <worktree> rev-parse --git-path ww/task-note.md
```

That task note is not written into tracked files.

#### `ww-helper gc --ttl-expired --idle 7d --dry-run --json`

`gc` requires at least one explicit selector. Supported selectors are:

- `--ttl-expired`
- `--idle <duration>`
- `--merged`

`gc requires at least one explicit selector`; a bare `ww-helper gc --json` returns `GC_RULE_REQUIRED` with exit code `2`.

Dry-run responses use the same envelope and return:

- `summary.matched`
- `summary.removed`
- `summary.skipped`
- `items[].matched_rules`
- `items[].action`
- `items[].reason` when skipped

#### `ww-helper rm --json --non-interactive <target>`

Removes the target without prompting, while still enforcing the normal safety rules:

- dirty worktrees still require `--force`
- the active worktree cannot be removed
- if you omit `<target>` and more than one removable worktree exists, the command returns `AMBIGUOUS_MATCH`

#### Breaking Change

`ww-helper rm --json` used to return a flat JSON object. It now returns the same JSON envelope format as the other Phase 1 machine-readable commands.

### Interactive Pick

```bash
ww
```

Without `fzf`, this opens the built-in selector like:

```text
* [1] ACTIVE main /path/to/repo
  [2]        feat-a /path/to/repo/.worktrees/feat-a

Use Up/Down (or j/k). Enter to confirm. Esc/Ctrl-C to cancel.
```

Move with arrow keys and press Enter to switch. The selector starts on the active shell worktree by default.

The status column can show:

- `ACTIVE` for the current clean worktree
- `ACTIVE*` for the current dirty worktree
- `DIRTY` for a non-current dirty worktree

`ww` ignores its own `.worktrees/` management directory when computing this status so the main worktree is not marked dirty just because linked worktrees exist.

### Direct Index Or Name

```bash
ww 2
ww switch feat-a
ww switch fea
```

Exact name matches win. If no exact match exists, `ww` falls back to a unique prefix match.

### List

```bash
ww list
ww list --filter label=agent:claude-code --verbose
```

This prints the current worktree table without changing your shell directory.

Worktrees are shown from oldest to newest by worktree creation time. Smaller indices refer to older worktrees, the status column uses the same `ACTIVE` / `ACTIVE*` / `DIRTY` markers as the interactive selector, and the default human-readable output includes `task=<label>` or `task=unlabeled`.

Available Phase 2 filters:

- `--filter dirty`
- `--filter label=agent:claude-code`
- `--filter label~agent`
- `--filter stale=7d`

`--verbose` appends metadata such as `label`, `ttl`, and `last_used_at` to the human-readable output.

### Check

```bash
ww check
```

`ww check` prints the current worktree path, branch, task label, dirty state, and task intent when a task note is available.

Warnings stay human-readable and conservative:

- detached worktrees are called out explicitly
- unlabeled worktrees are called out explicitly
- missing task notes warn instead of failing

### New

```bash
ww new feat-a
ww new feat-a --label agent:claude-code --ttl 24h
```

This creates branch `feat-a` from the current `HEAD` in `./.worktrees/feat-a`, then switches into it.

This release stores `created_at` for new worktrees in `state-v2.json`. If you pass `--ttl`, expiry is computed as `created_at + ttl` and does not slide on later access.

When you also pass `--label`, `ww` creates a private task note scaffold for that worktree so later `ww check` and removal summaries can restore task context.

### GC

```bash
ww gc --ttl-expired --dry-run
ww gc --idle 7d
ww gc --merged
ww gc --ttl-expired --idle 7d --dry-run --json
```

`gc` evaluates the union of the selected rules:

- TTL-expired worktrees
- idle worktrees older than the requested threshold
- worktrees whose branch is already merged into the effective base branch

Safety rules:

- active worktrees are always skipped
- dirty worktrees are skipped unless you pass `--force`
- worktrees without TTL are ignored by `--ttl-expired`
- `gc` is manual only; there is no automatic background cleanup

### Remove

```bash
ww rm
ww rm feat-a
ww rm --force feat-a
ww rm --base release/1.0 feat-a
```

`ww rm` groups removable worktrees by deletion risk, prints a plain-language summary card after selection, removes the worktree, and only deletes the branch when it is already merged into the effective base branch. Dirty worktrees stop before confirmation unless you explicitly rerun with `--force`.

When task metadata exists, the summary card also includes task label, task intent, and weak-boundary warnings such as detached or unlabeled state.

### Typical Flow

```bash
cd /path/to/repo
ww
ww switch feat-a
ww list
ww new feat-b
ww rm feat-a
```

`ww`, `ww 2`, and `ww switch feat-a` all switch the current shell into the target worktree.

### Exit Behavior

- `0`: success
- `2`: invalid user input such as a bad index, bad name match, or extra args
- `3`: environment problem such as not being in a Git repo
- `130`: interactive selection canceled

For `ww-helper ... --json`, the envelope `error.exit_code` matches the process exit code.

## Smoke Test Matrix

```bash
ww --help
ww help
ww 1
ww switch feat-a
ww list
ww new feat-b
ww rm feat-a
```

Installer checks:

```bash
bash install.sh
bash install.sh
```

## Release

Build release archives locally:

```bash
bash scripts/release.sh vX.Y.Z
```

Artifacts are written to `dist/`:

- `ww-vX.Y.Z-darwin-arm64.tar.gz`
- `ww-vX.Y.Z-darwin-amd64.tar.gz`
- `ww-vX.Y.Z-linux-arm64.tar.gz`
- `ww-vX.Y.Z-linux-amd64.tar.gz`
- `checksums.txt`
- `install-release.sh`

To publish a GitHub Release, create and push a tag matching `v*`:

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

GitHub release publishing is wired through `.github/workflows/release.yml` and only publishes when the workflow runs for `refs/tags/v*`.

Manual `workflow_dispatch` runs still build the `dist/` artifacts, but they do not publish a GitHub Release.
