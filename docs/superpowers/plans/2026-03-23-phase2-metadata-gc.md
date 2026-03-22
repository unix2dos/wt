# WW Phase 2 Metadata And GC Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add structured worktree metadata with rollback-safe `state-v2.json`, then ship explicit `ww gc` cleanup on top of that metadata without breaking the Phase 1 JSON contract.

**Architecture:** Keep the shell UX path-first and human-first. Extend `internal/state` with a v2 metadata store and migration path, then thread metadata through `internal/app/run.go` for `new`, `list`, and `gc`. Reuse `internal/git/remove.go` for destructive work so cleanup rules do not diverge from existing branch-deletion safety checks.

**Tech Stack:** Go, standard library, Git CLI, existing `internal/app` command router, existing `internal/state` file-locking store, repo docs/e2e tests.

---

## File Map

- Create: `internal/state/model.go`
  - v2 schema structs and metadata record types
- Create: `internal/state/duration.go`
  - shared parser and canonical formatter for `m` / `h` / `d` / `w`
- Create: `internal/state/duration_test.go`
  - duration parse/normalize/expiry coverage
- Modify: `internal/state/store.go`
  - v2 path resolution, lazy migration, metadata read/write, touch semantics
- Modify: `internal/state/store_test.go`
  - migration, v2 persistence, concurrency, fallback coverage
- Modify: `internal/app/run.go`
  - metadata-aware deps, `new --label --ttl`, `list --filter --verbose`, `gc`
- Modify: `internal/app/run_test.go`
  - app-layer JSON, filter, verbose, and gc coverage
- Modify: `README.md`
  - document metadata and explicit gc workflow
- Modify: `docs/reference.md`
  - document new flags, filter syntax, gc JSON contract, exit codes
- Modify: `test/docs/docs_test.go`
  - keep docs assertions aligned
- Modify: `test/e2e/e2e_test.go`
  - migrate old state, create labeled TTL worktree, dry-run gc

## Task 1: Add Shared Duration Parsing

**Files:**
- Create: `internal/state/duration.go`
- Create: `internal/state/duration_test.go`

- [ ] **Step 1: Write failing duration tests**

Add table-driven tests covering:

```go
func TestParseHumanDurationAcceptsSupportedUnits(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"15m", "15m"},
		{"24h", "24h"},
		{"7d", "7d"},
		{"2w", "2w"},
	}
	// parse input, assert canonical string
}
```

Also add negative cases for `""`, `"7"`, `"1mo"`, `"abc"`.

- [ ] **Step 2: Run targeted tests to verify failure**

Run: `go test ./internal/state -run 'TestParseHumanDuration'`
Expected: FAIL with missing parser/compiler errors.

- [ ] **Step 3: Implement the minimal parser**

Add a small value type and parser in `internal/state/duration.go`:

```go
type DurationSpec struct {
	Raw   string
	Value time.Duration
}

func ParseHumanDuration(input string) (DurationSpec, error)
func (d DurationSpec) String() string
```

Rules:
- accept one integer + one unit
- support `m`, `h`, `d`, `w`
- reject zero or negative values

- [ ] **Step 4: Add expiry helper coverage**

Extend `internal/state/duration_test.go` with:

```go
func TestDurationSpecExpiry(t *testing.T) {
	createdAt := time.Unix(100, 0)
	spec, _ := ParseHumanDuration("24h")
	got := spec.ExpiresAt(createdAt)
	// assert createdAt + 24h
}
```

- [ ] **Step 5: Run targeted tests to verify pass**

Run: `go test ./internal/state -run 'TestParseHumanDuration|TestDurationSpecExpiry'`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/state/duration.go internal/state/duration_test.go
git commit -m "feat(state): add shared metadata duration parser"
```

## Task 2: Add `state-v2.json` Schema And Lazy Migration

**Files:**
- Create: `internal/state/model.go`
- Modify: `internal/state/store.go`
- Modify: `internal/state/store_test.go`

- [ ] **Step 1: Write failing v2 store tests**

Add tests for:

```go
func TestStoreMigratesV1IntoV2State(t *testing.T) {}
func TestStoreTouchWritesV2LastUsedAt(t *testing.T) {}
func TestStoreCreateMetadataPersistsLabelAndTTL(t *testing.T) {}
func TestStoreLoadMissingRepoReturnsEmptyMetadataMap(t *testing.T) {}
```

Each test should assert:
- v1 `state.json` is left untouched
- new writes land in `state-v2.json`
- migrated entries preserve `last_used_at`
- label/ttl/created_at survive reloads

- [ ] **Step 2: Run targeted tests to verify failure**

Run: `go test ./internal/state -run 'TestStore(Migrates|TouchWrites|CreateMetadata|LoadMissingRepo)'`
Expected: FAIL because v2 APIs and schema do not exist.

- [ ] **Step 3: Add v2 schema types**

Create `internal/state/model.go` with focused types:

```go
type WorktreeMetadata struct {
	LastUsedAt int64  `json:"last_used_at,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	Label      string `json:"label,omitempty"`
	TTL        string `json:"ttl,omitempty"`
}

type RepoState struct {
	Worktrees map[string]WorktreeMetadata `json:"worktrees"`
}

type DiskStateV2 struct {
	Version int                  `json:"version"`
	Repos   map[string]RepoState `json:"repos"`
}
```

- [ ] **Step 4: Extend the store API minimally**

Modify `internal/state/store.go` to expose:

```go
func (s *Store) LoadMetadata(repoKey string) (map[string]WorktreeMetadata, error)
func (s *Store) Touch(repoKey, path string) error
func (s *Store) RecordWorktree(repoKey, path string, meta WorktreeMetadata) error
```

Keep `Touch` behavior additive:
- create v2 state if absent
- update only `last_used_at`
- preserve existing `label`, `ttl`, `created_at`

- [ ] **Step 5: Implement lazy migration**

Inside `store.go`:
- derive `state.json` and `state-v2.json` paths from the same state root
- on first v2 read, if only v1 exists, load v1 JSON and materialize v2
- backfill `created_at` from the worktree path on disk when possible
- write only `state-v2.json`

- [ ] **Step 6: Run targeted tests to verify pass**

Run: `go test ./internal/state`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/state/model.go internal/state/store.go internal/state/store_test.go
git commit -m "feat(state): add v2 metadata store and migration"
```

## Task 3: Thread Metadata Through `new` And `list`

**Files:**
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`

- [ ] **Step 1: Write failing app tests for metadata creation**

Add tests covering:

```go
func TestRunNewPathJSONIncludesMetadataInputs(t *testing.T) {}
func TestRunNewPathRejectsInvalidTTL(t *testing.T) {}
func TestRunNewPathRejectsEmptyLabel(t *testing.T) {}
```

Assert:
- `ww-helper new-path alpha --label agent:claude --ttl 24h --json` returns exit `0`
- JSON still uses `{ok, command, data}`
- the state-touch path receives `label`, `ttl`, `created_at`

- [ ] **Step 2: Write failing app tests for list filters**

Add tests covering:

```go
func TestRunListJSONIncludesMetadataFields(t *testing.T) {}
func TestRunListVerboseShowsLabelAndTTL(t *testing.T) {}
func TestRunListFiltersByLabelAndStale(t *testing.T) {}
func TestRunListRejectsInvalidFilter(t *testing.T) {}
```

Assert:
- `--json` always returns `last_used_at`, `label`, `ttl`
- `--verbose` affects only human output
- repeated `--filter` flags are ANDed

- [ ] **Step 3: Run targeted tests to verify failure**

Run: `go test ./internal/app -run 'TestRun(NewPath|List)'`
Expected: FAIL because args, metadata loading, and filters are not implemented.

- [ ] **Step 4: Expand the app deps boundary**

Adjust `internal/app/run.go` to use richer state methods. Introduce an app-local shape if needed:

```go
type WorktreeStateRecord struct {
	LastUsedAt int64
	CreatedAt  int64
	Label      string
	TTL        string
}
```

Update `Deps` and `RealDeps` so `orderedWorktrees` can load metadata without app code reaching into JSON files directly.

- [ ] **Step 5: Implement `new --label --ttl` and `new-path --label --ttl --json`**

Extend `parseNewPathArgs` to accept:
- `--label <text>`
- `--label=<text>`
- `--ttl <duration>`
- `--ttl=<duration>`

Validation rules:
- empty label is invalid
- invalid duration returns `INVALID_DURATION`
- metadata is persisted only after worktree creation succeeds

- [ ] **Step 6: Implement `list --filter ... --verbose --json`**

Extend `parseListArgs` with:
- repeated `--filter`
- `--verbose`

Implement helpers:

```go
func filterWorktrees(items []worktree.Worktree, filters []listFilter, now time.Time) ([]worktree.Worktree, error)
func listJSONItem(item worktree.Worktree) map[string]any
```

Do not change the default terse human output unless `--verbose` is present.

- [ ] **Step 7: Run targeted tests to verify pass**

Run: `go test ./internal/app -run 'TestRun(NewPath|List)'`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/app/run.go internal/app/run_test.go
git commit -m "feat(app): add metadata-aware new and list commands"
```

## Task 4: Add Explicit `ww gc`

**Files:**
- Modify: `internal/app/run.go`
- Modify: `internal/app/run_test.go`

- [ ] **Step 1: Write failing gc tests**

Add tests for:

```go
func TestRunGCRequiresAtLeastOneRule(t *testing.T) {}
func TestRunGCDryRunJSONSummarizesMatches(t *testing.T) {}
func TestRunGCSkipsDirtyAndActiveWorktrees(t *testing.T) {}
func TestRunGCForceAllowsDirtyRemoval(t *testing.T) {}
func TestRunGCMergedUsesBaseBranchResolution(t *testing.T) {}
```

Assert:
- bare `gc` returns `GC_RULE_REQUIRED`, exit `2`
- candidate rules are unioned
- `--dry-run` never calls `RemoveWorktree`
- active worktrees are always skipped

- [ ] **Step 2: Run targeted tests to verify failure**

Run: `go test ./internal/app -run 'TestRunGC'`
Expected: FAIL because the command does not exist.

- [ ] **Step 3: Add the `gc` command entry point**

Update `Run(...)` and helper help text to route:

```go
case "gc":
	return runGC(ctx, args[1:], out, errOut, deps)
```

Add a `gcConfig` parser supporting:
- `--ttl-expired`
- `--idle <duration>`
- `--merged`
- `--dry-run`
- `--force`
- `--json`
- `--base`

- [ ] **Step 4: Implement candidate selection**

In `run.go`, add focused helpers:

```go
func selectGCCandidates(items []worktree.Worktree, cfg gcConfig, now time.Time) ([]gcCandidate, error)
func gcMatchedRules(item worktree.Worktree, cfg gcConfig, now time.Time) []string
```

Rules:
- no selector => `GC_RULE_REQUIRED`
- `ttl_expired`, `idle`, `merged` are unioned
- metadata-less worktrees can still match `merged`

- [ ] **Step 5: Reuse existing removal logic**

For non-dry-run:
- preview with `deps.PreviewRemoval`
- skip active always
- skip dirty unless `--force`
- remove via `deps.RemoveWorktree`

Do not duplicate Git deletion logic inside `gc`.

- [ ] **Step 6: Implement JSON and human output**

Return the Phase 1 envelope shape:

```json
{
  "ok": true,
  "command": "gc",
  "data": {
    "summary": {"matched": 3, "removed": 1, "skipped": 2},
    "items": []
  }
}
```

Human output should:
- say which selectors were active
- list removed worktrees
- list skipped worktrees with reasons

- [ ] **Step 7: Run targeted tests to verify pass**

Run: `go test ./internal/app -run 'TestRunGC'`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/app/run.go internal/app/run_test.go
git commit -m "feat(app): add explicit gc command"
```

## Task 5: Update Docs And End-To-End Verification

**Files:**
- Modify: `README.md`
- Modify: `docs/reference.md`
- Modify: `test/docs/docs_test.go`
- Modify: `test/e2e/e2e_test.go`

- [ ] **Step 1: Write or update failing docs/e2e tests**

Cover:
- metadata flags in command reference
- filter syntax examples
- `gc` explicit-rule requirement
- a migration-oriented e2e path that starts with a v1 `state.json`

- [ ] **Step 2: Run targeted tests to verify failure**

Run: `go test ./test/docs ./test/e2e`
Expected: FAIL until docs and new e2e flows are added.

- [ ] **Step 3: Update docs**

Add concise docs for:
- `ww new --label --ttl`
- `ww-helper new-path --label --ttl --json`
- `ww list --filter ... --verbose`
- `ww-helper list --json` metadata fields
- `ww gc --ttl-expired|--idle|--merged --dry-run --json`

Make sure the docs state clearly:
- `gc` requires explicit selectors
- TTL is fixed from creation time
- metadata is immutable after creation in this release

- [ ] **Step 4: Add e2e coverage**

Add a scenario that:
- writes a v1 `state.json`
- runs the new binary
- confirms list still works after lazy migration
- creates a labeled TTL worktree
- runs `gc --ttl-expired --dry-run --json`

- [ ] **Step 5: Run targeted tests to verify pass**

Run: `go test ./test/docs ./test/e2e`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add README.md docs/reference.md test/docs/docs_test.go test/e2e/e2e_test.go
git commit -m "docs: add phase 2 metadata and gc documentation"
```

## Task 6: Full Verification And Release Readiness

**Files:**
- Verify only

- [ ] **Step 1: Run the full test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 2: Run manual smoke commands**

In a temp repo, verify:

```bash
ww-helper new-path feat-meta --label agent:smoke --ttl 24h --json
ww-helper list --json
ww-helper gc --ttl-expired --dry-run --json
ww-helper gc --merged --json
```

Expected:
- metadata fields appear in JSON
- `gc` without rules returns exit `2`
- `gc --dry-run` performs no deletion

- [ ] **Step 3: Check the contract against the spec**

Confirm all of the following:
- `state-v2.json` is the only new write target
- `state.json` remains intact after migration
- list JSON is additive, not breaking
- `gc` only acts on explicit selectors
- dirty and active protections still hold

- [ ] **Step 4: Prepare release notes**

Call out:
- new metadata flags
- `state-v2.json` migration behavior
- explicit `gc` workflow
- no metadata editing command in this release
