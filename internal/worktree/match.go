package worktree

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func Match(items []Worktree, query string) (Worktree, error) {
	if len(items) == 0 {
		return Worktree{}, fmt.Errorf("no worktree matches %q", query)
	}

	if matched, ok, err := matchByExactBranch(items, query); ok || err != nil {
		return matched, err
	}
	if matched, ok, err := matchByExactName(items, query); ok || err != nil {
		return matched, err
	}

	matches := collectMatches(items, func(item Worktree) bool {
		return strings.HasPrefix(item.BranchLabel, query) || strings.HasPrefix(worktreeName(item), query)
	})
	return finalizeMatches(query, matches)
}

func matchByExactBranch(items []Worktree, query string) (Worktree, bool, error) {
	matches := collectMatches(items, func(item Worktree) bool {
		return item.BranchLabel == query
	})
	if len(matches) == 0 {
		return Worktree{}, false, nil
	}
	matched, err := finalizeMatches(query, matches)
	return matched, true, err
}

func matchByExactName(items []Worktree, query string) (Worktree, bool, error) {
	matches := collectMatches(items, func(item Worktree) bool {
		return worktreeName(item) == query
	})
	if len(matches) == 0 {
		return Worktree{}, false, nil
	}
	matched, err := finalizeMatches(query, matches)
	return matched, true, err
}

func collectMatches(items []Worktree, include func(Worktree) bool) []Worktree {
	matches := make([]Worktree, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if !include(item) {
			continue
		}
		if _, ok := seen[item.Path]; ok {
			continue
		}
		seen[item.Path] = struct{}{}
		matches = append(matches, item)
	}
	return matches
}

func finalizeMatches(query string, matches []Worktree) (Worktree, error) {
	switch len(matches) {
	case 0:
		return Worktree{}, fmt.Errorf("no worktree matches %q", query)
	case 1:
		return matches[0], nil
	default:
		names := make([]string, 0, len(matches))
		for _, item := range matches {
			names = append(names, matchLabel(item))
		}
		sort.Strings(names)
		return Worktree{}, fmt.Errorf("ambiguous worktree match %q: %s", query, strings.Join(names, ", "))
	}
}

func worktreeName(item Worktree) string {
	return filepath.Base(item.Path)
}

func matchLabel(item Worktree) string {
	if item.BranchLabel != "" {
		return item.BranchLabel
	}
	return worktreeName(item)
}
