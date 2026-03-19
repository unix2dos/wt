# wt

`wt` is a small Git worktree switcher.

## Install

```bash
cd /Users/liuwei/workspace/wt
bash install.sh
source ~/.zshrc
```

If you use Bash, reload with `source ~/.bashrc` instead.

The installer builds `wt` into `~/.local/bin/wt` and appends a managed shell block that sources `shell/cwt.sh`.

## Usage

```bash
wt
wt --fzf
wt 2
cwt
cwt --fzf
```

`wt` prints the selected worktree path. `cwt` is the shell wrapper that changes the current shell directory after `wt` returns a path.
