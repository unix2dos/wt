# wt

`wt` is a small Git worktree switcher for the current repository.

## Install

### Install From Source

```bash
cd /Users/liuwei/workspace/wt
bash install.sh
source ~/.zshrc
```

If you use Bash, reload with `source ~/.bashrc` instead.

The installer builds `wt` into `~/.local/bin/wt` and appends a managed shell block that sources `shell/cwt.sh`.

Source installs require a working Go toolchain.

### Install From A Release Bundle

```bash
tar -xzf wt-v0.1.0-darwin-arm64.tar.gz
cd wt-v0.1.0-darwin-arm64
bash install.sh
source ~/.zshrc
```

Release bundle installs copy the prebuilt `bin/wt` binary and do not require Go.

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

```bash
wt
wt --fzf
wt 2
cwt
cwt --fzf
```

`wt` prints the selected worktree path. `cwt` is the shell wrapper that changes the current shell directory after `wt` returns a path.

## Smoke Test Matrix

```bash
wt --help
wt 1
printf '2\n' | wt
wt --fzf
cwt 1
```

Installer checks:

```bash
bash install.sh
bash install.sh
```

## Release

Build release archives locally:

```bash
bash scripts/release.sh v0.1.0
```

Artifacts are written to `dist/`:

- `wt-v0.1.0-darwin-arm64.tar.gz`
- `wt-v0.1.0-darwin-amd64.tar.gz`
- `wt-v0.1.0-linux-arm64.tar.gz`
- `wt-v0.1.0-linux-amd64.tar.gz`
- `checksums.txt`

GitHub release publishing is wired through `.github/workflows/release.yml` and runs on tags matching `v*`.
