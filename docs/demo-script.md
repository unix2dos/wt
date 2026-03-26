# ww Demo Script Notes

This file keeps the demo storyline stable when `ww` changes again.

## Storyline

The canonical demo is a compact workflow overview for README and Pages:

1. Start in `main` inside a throwaway repository.
2. Run `ww`, move through the `fzf` picker, and switch into `feat-a`.
3. Run `ww list` to show the current worktree table.
4. Run `ww new feat-demo` and confirm the shell moved into the new worktree.
5. Run `ww check` to show the active workspace summary.
6. Run `ww` again, move through the picker, and switch back to `main`.
7. Run `ww rm feat-demo`, confirm the prompt, and show the happy-path branch deletion output.
8. Run `ww rm --cleanup`, remove one pre-seeded stale workspace, then finish cleanup review.
9. End with `ww-helper list --json`, `ww-helper new-path --json --label agent:codex --ttl 24h feat-agent`, and `ww-helper rm --json --non-interactive feat-agent`.

The generator installs `scripts/demo-fzf.sh` as a deterministic `fzf` shim so the recording stays stable across machines while still exercising the `fzf` code path.

## Pacing Knobs

`bash scripts/generate-demo.sh` accepts these environment overrides:

- `WW_DEMO_KEYSTROKE_DELAY_MS`
- `WW_DEMO_STEP_DELAY_MS`
- `WW_DEMO_FZF_FOCUS_DELAY_MS`
- `WW_DEMO_FZF_KEYSTROKE_DELAY_MS`
- `WW_DEMO_FZF_MOVE_DELAY_MS`
- `WW_DEMO_FZF_REFRESH_DELAY_MS`
- `WW_DEMO_FZF_QUERY_SETTLE_MS`
- `WW_DEMO_CONFIRM_DELAY_MS`
- `WW_DEMO_IDLE_TIME_LIMIT`

The default pacing is tuned for a `60-90s` viewing window on the Pages player.

## Regeneration

```bash
bash scripts/generate-demo.sh
```

That command rebuilds the helper, records `docs/assets/ww-demo.cast`, and regenerates `docs/assets/ww-demo.svg`.
