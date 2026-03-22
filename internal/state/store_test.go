package state

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestStoreTouchAndLoadPersistsPerRepo(t *testing.T) {
	dir := t.TempDir()
	store := &Store{
		path: filepath.Join(dir, "state.json"),
		now: func() time.Time {
			return time.Unix(100, 0)
		},
	}

	if err := store.Touch("/repo-a/.git", "/repo-a/.worktrees/alpha"); err != nil {
		t.Fatalf("touch alpha: %v", err)
	}
	store.now = func() time.Time {
		return time.Unix(200, 0)
	}
	if err := store.Touch("/repo-a/.git", "/repo-a"); err != nil {
		t.Fatalf("touch current: %v", err)
	}
	store.now = func() time.Time {
		return time.Unix(300, 0)
	}
	if err := store.Touch("/repo-b/.git", "/repo-b/.worktrees/beta"); err != nil {
		t.Fatalf("touch beta: %v", err)
	}

	gotA, err := store.Load("/repo-a/.git")
	if err != nil {
		t.Fatalf("load repo a: %v", err)
	}
	wantA := map[string]int64{
		"/repo-a/.worktrees/alpha": time.Unix(100, 0).UnixNano(),
		"/repo-a":                  time.Unix(200, 0).UnixNano(),
	}
	if !reflect.DeepEqual(gotA, wantA) {
		t.Fatalf("repo a state mismatch: got %#v want %#v", gotA, wantA)
	}

	gotB, err := store.Load("/repo-b/.git")
	if err != nil {
		t.Fatalf("load repo b: %v", err)
	}
	wantB := map[string]int64{
		"/repo-b/.worktrees/beta": time.Unix(300, 0).UnixNano(),
	}
	if !reflect.DeepEqual(gotB, wantB) {
		t.Fatalf("repo b state mismatch: got %#v want %#v", gotB, wantB)
	}
}

func TestStoreLoadMissingRepoReturnsEmptyMap(t *testing.T) {
	dir := t.TempDir()
	store := &Store{path: filepath.Join(dir, "state.json")}

	got, err := store.Load("/repo/.git")
	if err != nil {
		t.Fatalf("load missing repo: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %#v", got)
	}
}

func TestStoreTouchSerializesConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	storeA := &Store{
		path: path,
		now: func() time.Time {
			return time.Unix(100, 0)
		},
	}
	storeB := &Store{
		path: path,
		now: func() time.Time {
			return time.Unix(200, 0)
		},
	}

	errCh := make(chan error, 2)
	go func() {
		errCh <- storeA.Touch("/repo/.git", "/repo/.worktrees/alpha")
	}()
	go func() {
		errCh <- storeB.Touch("/repo/.git", "/repo/.worktrees/beta")
	}()

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("touch %d: %v", i, err)
		}
	}

	got, err := storeA.Load("/repo/.git")
	if err != nil {
		t.Fatalf("load after concurrent touch: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected both entries to persist, got %#v", got)
	}
}

func TestStoreLockBlocksSecondInstance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	storeA := &Store{path: path}
	storeB := &Store{path: path}

	held := make(chan struct{})
	released := make(chan struct{})
	acquired := make(chan struct{})
	errCh := make(chan error, 2)

	go func() {
		errCh <- storeA.withLock(func() error {
			close(held)
			<-released
			return nil
		})
	}()

	<-held
	go func() {
		errCh <- storeB.withLock(func() error {
			close(acquired)
			return nil
		})
	}()

	select {
	case <-acquired:
		t.Fatal("second store acquired lock before first released it")
	case <-time.After(50 * time.Millisecond):
	}

	close(released)

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second store did not acquire lock after release")
	}

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("withLock %d: %v", i, err)
		}
	}
}

func TestStoreMigratesV1IntoV2State(t *testing.T) {
	dir := t.TempDir()
	repoKey := filepath.Join(dir, "repo", ".git")
	worktreePath := filepath.Join(dir, "repo", ".worktrees", "alpha")
	if err := os.MkdirAll(worktreePath, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	v1Path := filepath.Join(dir, "state.json")
	original := []byte("{\n  \"repos\": {\n    \"" + repoKey + "\": {\n      \"" + worktreePath + "\": 100\n    }\n  }\n}\n")
	if err := os.WriteFile(v1Path, original, 0o644); err != nil {
		t.Fatalf("write v1 state: %v", err)
	}

	store := NewStoreAt(v1Path)
	got, err := store.LoadMetadata(repoKey)
	if err != nil {
		t.Fatalf("LoadMetadata() error: %v", err)
	}

	meta, ok := got[worktreePath]
	if !ok {
		t.Fatalf("expected migrated metadata for %q, got %#v", worktreePath, got)
	}
	if meta.LastUsedAt != 100 {
		t.Fatalf("LastUsedAt = %d, want 100", meta.LastUsedAt)
	}
	if meta.CreatedAt == 0 {
		t.Fatalf("CreatedAt = 0, want backfilled timestamp")
	}

	after, err := os.ReadFile(v1Path)
	if err != nil {
		t.Fatalf("read v1 state: %v", err)
	}
	if string(after) != string(original) {
		t.Fatalf("v1 state changed:\n%s", string(after))
	}

	if _, err := os.Stat(filepath.Join(dir, "state-v2.json")); err != nil {
		t.Fatalf("expected state-v2.json to exist: %v", err)
	}
}

func TestStoreTouchWritesV2LastUsedAt(t *testing.T) {
	dir := t.TempDir()
	store := &Store{
		path: filepath.Join(dir, "state.json"),
		now: func() time.Time {
			return time.Unix(100, 0)
		},
	}

	repoKey := filepath.Join(dir, "repo", ".git")
	worktreePath := filepath.Join(dir, "repo", ".worktrees", "alpha")
	if err := store.Touch(repoKey, worktreePath); err != nil {
		t.Fatalf("Touch() error: %v", err)
	}

	got, err := store.LoadMetadata(repoKey)
	if err != nil {
		t.Fatalf("LoadMetadata() error: %v", err)
	}
	if got[worktreePath].LastUsedAt != time.Unix(100, 0).UnixNano() {
		t.Fatalf("LastUsedAt = %d, want %d", got[worktreePath].LastUsedAt, time.Unix(100, 0).UnixNano())
	}
	if _, err := os.Stat(filepath.Join(dir, "state-v2.json")); err != nil {
		t.Fatalf("expected v2 file: %v", err)
	}
}

func TestStoreCreateMetadataPersistsLabelAndTTL(t *testing.T) {
	dir := t.TempDir()
	store := &Store{
		path: filepath.Join(dir, "state.json"),
		now: func() time.Time {
			return time.Unix(300, 0)
		},
	}

	repoKey := filepath.Join(dir, "repo", ".git")
	worktreePath := filepath.Join(dir, "repo", ".worktrees", "alpha")
	meta := WorktreeMetadata{
		CreatedAt: 200,
		Label:     "agent:claude-code",
		TTL:       "24h",
	}
	if err := store.RecordWorktree(repoKey, worktreePath, meta); err != nil {
		t.Fatalf("RecordWorktree() error: %v", err)
	}
	if err := store.Touch(repoKey, worktreePath); err != nil {
		t.Fatalf("Touch() error: %v", err)
	}

	got, err := store.LoadMetadata(repoKey)
	if err != nil {
		t.Fatalf("LoadMetadata() error: %v", err)
	}
	if got[worktreePath].CreatedAt != meta.CreatedAt {
		t.Fatalf("CreatedAt = %d, want %d", got[worktreePath].CreatedAt, meta.CreatedAt)
	}
	if got[worktreePath].Label != meta.Label {
		t.Fatalf("Label = %q, want %q", got[worktreePath].Label, meta.Label)
	}
	if got[worktreePath].TTL != meta.TTL {
		t.Fatalf("TTL = %q, want %q", got[worktreePath].TTL, meta.TTL)
	}
	if got[worktreePath].LastUsedAt != time.Unix(300, 0).UnixNano() {
		t.Fatalf("LastUsedAt = %d, want %d", got[worktreePath].LastUsedAt, time.Unix(300, 0).UnixNano())
	}
}

func TestStoreLoadMissingRepoReturnsEmptyMetadataMap(t *testing.T) {
	dir := t.TempDir()
	store := &Store{path: filepath.Join(dir, "state.json")}

	got, err := store.LoadMetadata(filepath.Join(dir, "repo", ".git"))
	if err != nil {
		t.Fatalf("LoadMetadata() error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty metadata map, got %#v", got)
	}
}
