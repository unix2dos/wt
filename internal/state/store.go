package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Store struct {
	path string
	now  func() time.Time
	mu   sync.Mutex
}

type diskStateV1 struct {
	Repos map[string]map[string]int64 `json:"repos"`
}

func NewStore() (*Store, error) {
	path, err := defaultPath()
	if err != nil {
		return nil, err
	}
	return &Store{path: path, now: time.Now}, nil
}

func NewStoreAt(path string) *Store {
	return &Store{path: path, now: time.Now}
}

func (s *Store) Load(repoKey string) (map[string]int64, error) {
	meta, err := s.LoadMetadata(repoKey)
	if err != nil {
		return nil, err
	}

	out := make(map[string]int64, len(meta))
	for path, item := range meta {
		out[path] = item.LastUsedAt
	}
	return out, nil
}

func (s *Store) LoadMetadata(repoKey string) (map[string]WorktreeMetadata, error) {
	var out map[string]WorktreeMetadata
	err := s.withLock(func() error {
		state, err := s.readV2Locked()
		if err != nil {
			return err
		}
		out = cloneRepoMetadata(state.Repos[repoKey].Worktrees)
		return nil
	})
	return out, err
}

func (s *Store) Touch(repoKey, path string) error {
	return s.withLock(func() error {
		state, err := s.readV2Locked()
		if err != nil {
			return err
		}
		if state.Repos == nil {
			state.Repos = make(map[string]RepoState)
		}
		repo := state.Repos[repoKey]
		if repo.Worktrees == nil {
			repo.Worktrees = make(map[string]WorktreeMetadata)
		}
		meta := repo.Worktrees[path]
		meta.LastUsedAt = s.nowFunc().UnixNano()
		if meta.CreatedAt == 0 {
			meta.CreatedAt = metadataCreatedAt(path)
		}
		repo.Worktrees[path] = meta
		state.Repos[repoKey] = repo
		return s.writeV2Locked(state)
	})
}

func (s *Store) RecordWorktree(repoKey, path string, meta WorktreeMetadata) error {
	return s.withLock(func() error {
		state, err := s.readV2Locked()
		if err != nil {
			return err
		}
		if state.Repos == nil {
			state.Repos = make(map[string]RepoState)
		}
		repo := state.Repos[repoKey]
		if repo.Worktrees == nil {
			repo.Worktrees = make(map[string]WorktreeMetadata)
		}

		current := repo.Worktrees[path]
		if meta.LastUsedAt == 0 {
			meta.LastUsedAt = current.LastUsedAt
		}
		if meta.CreatedAt == 0 {
			if current.CreatedAt != 0 {
				meta.CreatedAt = current.CreatedAt
			} else {
				meta.CreatedAt = metadataCreatedAt(path)
			}
		}
		if meta.Label == "" {
			meta.Label = current.Label
		}
		if meta.TTL == "" {
			meta.TTL = current.TTL
		}

		repo.Worktrees[path] = meta
		state.Repos[repoKey] = repo
		return s.writeV2Locked(state)
	})
}

func (s *Store) withLock(fn func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.path == "" {
		return errors.New("state path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	lockPath := s.lockPath()
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	return fn()
}

func (s *Store) readV2Locked() (DiskStateV2, error) {
	if s.path == "" {
		return DiskStateV2{}, errors.New("state path is empty")
	}

	raw, err := os.ReadFile(s.v2Path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return s.migrateV1Locked()
		}
		return DiskStateV2{}, err
	}

	if len(raw) == 0 {
		return emptyDiskStateV2(), nil
	}

	var state DiskStateV2
	if err := json.Unmarshal(raw, &state); err != nil {
		return DiskStateV2{}, err
	}
	if state.Repos == nil {
		state.Repos = make(map[string]RepoState)
	}
	if state.Version == 0 {
		state.Version = 2
	}
	return state, nil
}

func (s *Store) migrateV1Locked() (DiskStateV2, error) {
	v1State, exists, err := s.readV1Locked()
	if err != nil {
		return DiskStateV2{}, err
	}
	if !exists {
		return emptyDiskStateV2(), nil
	}

	state := emptyDiskStateV2()
	for repoKey, repo := range v1State.Repos {
		worktrees := make(map[string]WorktreeMetadata, len(repo))
		for path, lastUsedAt := range repo {
			worktrees[path] = WorktreeMetadata{
				LastUsedAt: lastUsedAt,
				CreatedAt:  metadataCreatedAt(path),
			}
		}
		state.Repos[repoKey] = RepoState{Worktrees: worktrees}
	}

	if err := s.writeV2Locked(state); err != nil {
		return DiskStateV2{}, err
	}
	return state, nil
}

func (s *Store) readV1Locked() (diskStateV1, bool, error) {
	raw, err := os.ReadFile(s.v1Path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return diskStateV1{}, false, nil
		}
		return diskStateV1{}, false, err
	}
	if len(raw) == 0 {
		return diskStateV1{Repos: make(map[string]map[string]int64)}, true, nil
	}

	var state diskStateV1
	if err := json.Unmarshal(raw, &state); err != nil {
		return diskStateV1{}, false, err
	}
	if state.Repos == nil {
		state.Repos = make(map[string]map[string]int64)
	}
	return state, true, nil
}

func (s *Store) writeV2Locked(state DiskStateV2) error {
	if s.path == "" {
		return errors.New("state path is empty")
	}

	path := s.v2Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if state.Version == 0 {
		state.Version = 2
	}
	if state.Repos == nil {
		state.Repos = make(map[string]RepoState)
	}

	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(encoded); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func (s *Store) lockPath() string {
	return s.v1Path() + ".lock"
}

func (s *Store) nowFunc() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

func cloneRepoState(src map[string]int64) map[string]int64 {
	if len(src) == 0 {
		return map[string]int64{}
	}
	out := make(map[string]int64, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func cloneRepoMetadata(src map[string]WorktreeMetadata) map[string]WorktreeMetadata {
	if len(src) == 0 {
		return map[string]WorktreeMetadata{}
	}
	out := make(map[string]WorktreeMetadata, len(src))
	for path, meta := range src {
		out[path] = meta
	}
	return out
}

func emptyDiskStateV2() DiskStateV2 {
	return DiskStateV2{
		Version: 2,
		Repos:   make(map[string]RepoState),
	}
}

func metadataCreatedAt(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.ModTime().UnixNano()
}

func (s *Store) v1Path() string {
	if filepath.Base(s.path) == "state-v2.json" {
		return filepath.Join(filepath.Dir(s.path), "state.json")
	}
	return s.path
}

func (s *Store) v2Path() string {
	if filepath.Base(s.path) == "state-v2.json" {
		return s.path
	}
	return filepath.Join(filepath.Dir(s.path), "state-v2.json")
}

func defaultPath() (string, error) {
	if base := os.Getenv("XDG_STATE_HOME"); base != "" {
		return filepath.Join(base, "ww", "state.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", "ww", "state.json"), nil
}
