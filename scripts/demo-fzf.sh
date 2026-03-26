#!/usr/bin/env bash

set -euo pipefail

tty_path="${FZF_TTY:-/dev/tty}"
exec 3<>"$tty_path"
refresh_delay_ms="${WW_DEMO_FZF_REFRESH_DELAY_MS:-0}"
prompt="Select a worktree> "
use_tac=0
load_pos=1

declare -a candidates=()
declare -a filtered=()

for arg in "$@"; do
  if [[ "$arg" == "--tac" ]]; then
    use_tac=1
    continue
  fi
  if [[ "$arg" == --prompt=* ]]; then
    prompt="${arg#--prompt=}"
    continue
  fi
  if [[ "$arg" =~ ^--bind=load:pos\(([0-9]+)\)$ ]]; then
    load_pos="${BASH_REMATCH[1]}"
  fi
done

while IFS= read -r line; do
  candidates+=("$line")
done

if (( use_tac == 1 )); then
  declare -a reversed=()
  for ((i=${#candidates[@]}-1; i>=0; i--)); do
    reversed+=("${candidates[$i]}")
  done
  candidates=("${reversed[@]}")
fi

old_stty="$(stty -g <&3)"
query=""
selected=0

cleanup() {
  stty "$old_stty" <&3 >/dev/null 2>&1 || true
  printf '\033[?25h\033[?1049l' >&3 || true
}
trap cleanup EXIT

lower() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]'
}

filter_candidates() {
  filtered=()
  local lowered_query
  local candidate lowered_candidate

  lowered_query="$(lower "$query")"

  for candidate in "${candidates[@]}"; do
    lowered_candidate="$(lower "$candidate")"
    if [[ -z "$lowered_query" || "$lowered_candidate" == *"$lowered_query"* ]]; then
      filtered+=("$candidate")
    fi
  done

  if (( selected >= ${#filtered[@]} )); then
    selected=0
  fi
}

print_candidate() {
  local candidate="$1"
  local index status branch path

  IFS=$'\t' read -r index status branch path <<<"$candidate"
  printf "%s  %-18s %-10s %s" "$index" "${status:-}" "$branch" "$path"
}

render() {
  local screen
  local row
  local prefix

  screen=$'\033[?1049h\033[H\033[2J\033[?25l'
  screen+="${prompt}${query}"$'\n\n'

  if (( ${#filtered[@]} == 0 )); then
    screen+=$'  no matches\n'
    printf '%s' "$screen" >&3
    return
  fi

  local i
  for i in "${!filtered[@]}"; do
    if (( i == selected )); then
      prefix="> "
    else
      prefix="  "
    fi
    row="$(print_candidate "${filtered[$i]}")"
    screen+="${prefix}${row}"$'\n'
  done

  printf '%s' "$screen" >&3
}

stty -echo -icanon min 1 time 0 <&3
filter_candidates
if (( load_pos > 0 && load_pos <= ${#filtered[@]} )); then
  selected=$((load_pos - 1))
fi
render

if [[ "$refresh_delay_ms" =~ ^[0-9]+$ ]] && (( refresh_delay_ms > 0 )); then
  sleep "$(awk "BEGIN { printf \"%.3f\", $refresh_delay_ms / 1000 }")"
  render
fi

while IFS= read -r -s -n1 key <&3; do
  case "$key" in
    "")
      break
      ;;
    $'\r'|$'\n')
      break
      ;;
    $'\177'|$'\b')
      query="${query%?}"
      selected=0
      ;;
    $'\003')
      exit 130
      ;;
    "k")
      if (( selected > 0 )); then
        ((selected--))
      fi
      ;;
    "j")
      if (( selected + 1 < ${#filtered[@]} )); then
        ((selected++))
      fi
      ;;
    $'\033')
      rest=""
      IFS= read -r -s -n2 -t 0.01 rest <&3 || true
      case "$rest" in
        "[A")
          if (( selected > 0 )); then
            ((selected--))
          fi
          ;;
        "[B")
          if (( selected + 1 < ${#filtered[@]} )); then
            ((selected++))
          fi
          ;;
        *)
          exit 130
          ;;
      esac
      ;;
    *)
      query+="$key"
      selected=0
      ;;
  esac

  filter_candidates
  render
done

if (( ${#filtered[@]} == 0 )); then
  exit 130
fi

printf '%s\n' "${filtered[$selected]}"
