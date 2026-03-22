package state

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
