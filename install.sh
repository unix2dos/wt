#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"
WRAPPER_SOURCE_PATH="$REPO_ROOT/shell/cwt.sh"
WRAPPER_NAME="wt-cwt.sh"
RC_MARKER_BEGIN="# wt shell wrapper begin"
RC_MARKER_END="# wt shell wrapper end"
INSTALL_SHELL=""
RC_FILE=""
BIN_DIR="$HOME/.local/bin"

usage() {
  cat <<'EOF'
Usage: bash install.sh [--shell zsh|bash] [--rc-file PATH] [--bin-dir PATH]

Installs `wt` to the target bin directory and appends a managed block that
sources the installed `wt-cwt.sh` wrapper from the chosen shell rc file.
EOF
}

installed_wrapper_path() {
  printf '%s\n' "$BIN_DIR/$WRAPPER_NAME"
}

strip_managed_block() {
  local rc_file="$1"
  local tmp

  [ -f "$rc_file" ] || return 0

  tmp="$(mktemp)"
  awk -v begin="$RC_MARKER_BEGIN" -v end="$RC_MARKER_END" '
    $0 == begin { skip = 1; next }
    $0 == end { skip = 0; next }
    skip != 1 { print }
  ' "$rc_file" >"$tmp"
  mv "$tmp" "$rc_file"
}

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --shell)
        [ "$#" -ge 2 ] || { echo "missing value for --shell" >&2; exit 2; }
        INSTALL_SHELL="$2"
        shift 2
        ;;
      --rc-file)
        [ "$#" -ge 2 ] || { echo "missing value for --rc-file" >&2; exit 2; }
        RC_FILE="$2"
        shift 2
        ;;
      --bin-dir)
        [ "$#" -ge 2 ] || { echo "missing value for --bin-dir" >&2; exit 2; }
        BIN_DIR="$2"
        shift 2
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        echo "unknown argument: $1" >&2
        exit 2
        ;;
    esac
  done
}

choose_rc_file() {
  if [ -n "$RC_FILE" ]; then
    printf '%s\n' "$RC_FILE"
    return
  fi

  case "$INSTALL_SHELL" in
    zsh) printf '%s\n' "$HOME/.zshrc"; return ;;
    bash) printf '%s\n' "$HOME/.bashrc"; return ;;
    "") ;;
    *)
      echo "unsupported shell: $INSTALL_SHELL" >&2
      exit 2
      ;;
  esac

  if [ -n "${ZDOTDIR:-}" ] && [ -f "${ZDOTDIR}/.zshrc" ]; then
    printf '%s\n' "${ZDOTDIR}/.zshrc"
    return
  fi

  case "${SHELL:-}" in
    */zsh)
      printf '%s\n' "$HOME/.zshrc"
      return
      ;;
    */bash)
      printf '%s\n' "$HOME/.bashrc"
      return
      ;;
  esac

  if [ -f "$HOME/.zshrc" ]; then
    printf '%s\n' "$HOME/.zshrc"
    return
  fi

  if [ -f "$HOME/.bashrc" ]; then
    printf '%s\n' "$HOME/.bashrc"
    return
  fi

  printf '%s\n' "$HOME/.zshrc"
}

append_shell_wrapper() {
  local rc_file="$1"
  local wrapper_path="$2"

  mkdir -p "$(dirname "$rc_file")"
  touch "$rc_file"
  strip_managed_block "$rc_file"

  {
    printf '%s\n' "$RC_MARKER_BEGIN"
    printf '%s\n' "if [ -f \"$wrapper_path\" ]; then"
    printf '%s\n' "  source \"$wrapper_path\""
    printf '%s\n' "fi"
    printf '%s\n' "$RC_MARKER_END"
  } >>"$rc_file"
}

install_binary() {
  local bin_path="$BIN_DIR/wt"

  mkdir -p "$BIN_DIR"
  if [ -x "$REPO_ROOT/bin/wt" ]; then
    cp "$REPO_ROOT/bin/wt" "$bin_path"
    chmod +x "$bin_path"
    return
  fi

  cd "$REPO_ROOT"
  go build -o "$bin_path" ./cmd/wt
}

install_wrapper() {
  local wrapper_path

  wrapper_path="$(installed_wrapper_path)"
  mkdir -p "$BIN_DIR"
  cp "$WRAPPER_SOURCE_PATH" "$wrapper_path"
  chmod +x "$wrapper_path"
}

parse_args "$@"
install_binary
install_wrapper

RC_TARGET="$(choose_rc_file)"
append_shell_wrapper "$RC_TARGET" "$(installed_wrapper_path)"

printf 'Installed wt to %s\n' "$BIN_DIR/wt"
printf 'Installed shell wrapper to %s\n' "$(installed_wrapper_path)"
printf 'Updated shell rc: %s\n' "$RC_TARGET"
