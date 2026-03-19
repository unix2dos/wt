#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"
BIN_DIR="$HOME/.local/bin"
BIN_PATH="$BIN_DIR/wt"
WRAPPER_PATH="$REPO_ROOT/shell/cwt.sh"
RC_MARKER_BEGIN="# wt shell wrapper begin"
RC_MARKER_END="# wt shell wrapper end"

choose_rc_file() {
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

  mkdir -p "$(dirname "$rc_file")"
  touch "$rc_file"

  if grep -Fq "$RC_MARKER_BEGIN" "$rc_file"; then
    return
  fi

  {
    printf '%s\n' "$RC_MARKER_BEGIN"
    printf '%s\n' "if [ -f \"$WRAPPER_PATH\" ]; then"
    printf '%s\n' "  source \"$WRAPPER_PATH\""
    printf '%s\n' "fi"
    printf '%s\n' "$RC_MARKER_END"
  } >>"$rc_file"
}

mkdir -p "$BIN_DIR"
cd "$REPO_ROOT"
go build -o "$BIN_PATH" ./cmd/wt

append_shell_wrapper "$(choose_rc_file)"
