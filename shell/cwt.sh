cwt() {
  local target

  target="$(wt "$@")" || return $?
  if [ -z "$target" ]; then
    return 1
  fi

  cd "$target" || return $?
}
