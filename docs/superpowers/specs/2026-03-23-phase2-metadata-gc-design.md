# WW Phase 2 Metadata And GC Design

## Goal

Ship Phase 2 in two implementation steps without reopening the Phase 1 contract:

1. add structured worktree metadata via `state-v2.json`
2. add explicit, safe `ww gc` cleanup on top of that metadata

The shell-first human workflow remains unchanged. Programmatic callers continue to use `ww-helper` and the Phase 1 JSON envelope.

## Product Decisions

- State moves to a new file, `state-v2.json`; `state.json` remains untouched for rollback safety.
- `ttl` is a fixed lifetime computed from `created_at + ttl`.
- Idle cleanup is a separate rule driven by `last_used_at`.
- `ww gc` does nothing unless at least one cleanup rule is explicitly provided.
- Metadata is write-on-create in Phase 2: `label` and `ttl` can only be set by `ww new` / `ww-helper new-path`.
- `ww list --json` always returns full metadata fields when available, regardless of `--verbose`.
- `ww gc` reuses existing removal safety rules for active worktrees, dirty worktrees, and merged-branch deletion.

## Scope

### In Scope

- `state-v2.json` format and lazy migration from v1
- `ww new --label --ttl`
- `ww-helper new-path --label --ttl --json`
- `ww list --filter ... --verbose`
- metadata fields in `ww list --json`
- `ww gc --ttl-expired --idle <duration> --merged --dry-run --force --json`

### Out Of Scope

- metadata mutation after creation (`ww set`, `ww update`)
- automatic background cleanup
- default cleanup thresholds
- MCP server design
- changes to the shell contract for `ww switch` / `ww new`

## State Model

Phase 2 introduces a new state file at the same state root as v1:

- v1: `state.json`
- v2: `state-v2.json`

New versions only read and write `state-v2.json`. Old versions continue using `state.json`.

### V2 Schema

```json
{
  "version": 2,
  "repos": {
    "/path/to/repo": {
      "worktrees": {
        "/path/to/repo/.worktrees/feat-a": {
          "last_used_at": 1711100000000000000,
          "created_at": 1711000000000000000,
          "label": "agent:claude-code",
          "ttl": "24h"
        }
      }
    }
  }
}
```

### Field Semantics

- `last_used_at`: Unix nanoseconds. Updated by the same touch points already used for MRU behavior.
- `created_at`: Unix nanoseconds. For new worktrees, set when metadata is first written. For migrated worktrees, backfilled from filesystem creation time when available.
- `label`: optional free-text string. Stored exactly as provided after trimming and non-empty validation.
- `ttl`: optional normalized duration string. Expiry is `created_at + ttl`.

### Why `repos -> worktrees -> path`

The extra `worktrees` object makes repo-level metadata possible later without another breaking schema change. It also avoids mixing repo attributes with per-worktree records.

## Migration Strategy

Migration is lazy and one-way:

1. if `state-v2.json` exists, use it
2. otherwise, if `state.json` exists, migrate it once into `state-v2.json`
3. otherwise, create a fresh empty v2 state on first write

Migration rules:

- copy v1 MRU timestamps into `last_used_at`
- try to backfill `created_at` from the existing worktree path on disk
- leave `label` and `ttl` empty for migrated entries
- never rewrite or delete `state.json`

Rollback behavior:

- old binaries ignore `state-v2.json` and keep using `state.json`
- MRU updates written by the new version are lost after rollback, which is acceptable because they are preference data

## CLI Design

### Create

Human entrypoint:

```sh
ww new <name> [--label <text>] [--ttl <duration>]
```

Programmatic entrypoint:

```sh
ww-helper new-path <name> [--label <text>] [--ttl <duration>] [--json]
```

Rules:

- `--label` accepts a single free-text value
- `--ttl` uses the shared duration parser
- metadata is written only for the newly created worktree
- missing or invalid `--ttl` is a user input error

### List

Human entrypoint:

```sh
ww list [--filter <expr> ...] [--verbose]
```

Programmatic entrypoint:

```sh
ww-helper list [--filter <expr> ...] [--json]
```

Supported filter expressions:

- `dirty`
- `label=<exact>`
- `label~<substring>`
- `stale=<duration>`

Rules:

- multiple `--filter` flags are ANDed
- `--verbose` controls only human-readable output
- `--json` always includes metadata fields when known
- worktrees missing v2 metadata are still listed; metadata fields are omitted or zero-valued

Recommended JSON item shape:

```json
{
  "path": "/repo/.worktrees/feat-a",
  "branch": "feat-a",
  "dirty": false,
  "active": false,
  "created_at": 1711000000000000000,
  "last_used_at": 1711100000000000000,
  "label": "agent:claude-code",
  "ttl": "24h"
}
```

### GC

Entry points:

```sh
ww gc [--ttl-expired] [--idle <duration>] [--merged] [--dry-run] [--force]
ww-helper gc [--ttl-expired] [--idle <duration>] [--merged] [--dry-run] [--force] [--json]
```

Rules:

- at least one cleanup selector is required
- selectors form a union of candidates, not an intersection
- `--ttl-expired` selects worktrees where `created_at + ttl <= now`
- `--idle <duration>` selects worktrees where `last_used_at <= now - idle`
- `--merged` reuses the existing effective base branch logic
- `--dry-run` performs matching and safety evaluation only
- actual deletion reuses the current removal path

Safety behavior:

- active worktrees are always skipped
- dirty worktrees are skipped unless `--force` is set
- worktrees without TTL are never selected by `--ttl-expired`
- worktrees without metadata can still be selected by `--merged`

Recommended JSON envelope:

```json
{
  "ok": true,
  "command": "gc",
  "data": {
    "summary": {
      "matched": 3,
      "removed": 1,
      "skipped": 2
    },
    "items": [
      {
        "path": "/repo/.worktrees/feat-a",
        "branch": "feat-a",
        "matched_rules": ["ttl_expired"],
        "action": "removed"
      },
      {
        "path": "/repo/.worktrees/feat-b",
        "branch": "feat-b",
        "matched_rules": ["idle"],
        "action": "skipped",
        "reason": "dirty"
      }
    ]
  }
}
```

## Duration Rules

One parser should back:

- `--ttl`
- `--idle`
- `stale=<duration>`

Supported units:

- `m`
- `h`
- `d`
- `w`

Stored values should be normalized to a canonical string so state and JSON output are stable.

## Error Model

Phase 2 should extend the Phase 1 app-error model with stable codes:

- `INVALID_DURATION`
- `INVALID_FILTER`
- `GC_RULE_REQUIRED`
- `STATE_MIGRATION_FAILED`
- `STATE_SCHEMA_INVALID`

Exit code policy remains:

- `0` success
- `1` runtime failure
- `2` user input error
- `3` not in a Git repository
- `130` cancelled interaction

## Implementation Order

### Step 1: State And Metadata

- extend `internal/state` with v2 load, write, and migration
- teach `new` / `new-path` to persist `label`, `ttl`, `created_at`
- teach `list` to read metadata and expose it in JSON output
- add `--filter` and `--verbose`

### Step 2: Explicit GC

- add candidate selection for `--ttl-expired`, `--idle`, and `--merged`
- add `--dry-run` and JSON reporting
- execute removals through existing remove logic
- keep all existing safety rules intact

This split keeps the data model stable before cleanup logic depends on it.

## Testing Strategy

### State Tests

- v1 to v2 migration
- fresh v2 write and read
- missing metadata fallback
- duration normalization and expiry checks

### App Tests

- `new --label --ttl`
- `new-path --json --label --ttl`
- `list --filter ... --verbose`
- `list --json` metadata fields
- `gc` rule validation and dry-run JSON output

### Git And Removal Tests

- merged detection still follows effective base branch resolution
- dirty and active skip rules do not regress
- `gc` reuses existing branch deletion rules

### End-To-End Verification

- migrate an existing v1 state and confirm `ww list` still works
- create a labeled, TTL-bound worktree and observe it via `ww-helper list --json`
- run `ww-helper gc --ttl-expired --dry-run --json`

## Risks And Guardrails

- Filesystem creation time is not portable; code must tolerate `created_at == 0` for migrated entries.
- `gc` must not invent its own deletion path. Reuse the current removal implementation to avoid divergence.
- JSON output should stay envelope-based and additive so Phase 1 agents do not break on unrelated fields.
- Human shell UX must remain path-first; any machine-readable behavior belongs in `ww-helper`.

## Open Follow-Ups

- whether to surface metadata columns in non-verbose human list output later
- whether post-creation metadata mutation is worth adding after observing real usage
- whether a future MCP wrapper should expose `gc` directly or continue delegating to CLI JSON
