package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"wt/internal/worktree"
)

var ErrNotGitRepository = errors.New("not a git repository")

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (stdout []byte, stderr []byte, err error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func ListWorktrees(ctx context.Context, runner Runner) ([]worktree.Worktree, error) {
	rootOut, rootErr, err := runner.Run(ctx, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		if isNotGitRepository(err, rootOut, rootErr) {
			return nil, ErrNotGitRepository
		}
		return nil, fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}

	currentPath := strings.TrimSpace(string(rootOut))
	worktreeOut, worktreeErr, err := runner.Run(ctx, "git", "-C", currentPath, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		if isNotGitRepository(err, worktreeOut, worktreeErr) {
			return nil, ErrNotGitRepository
		}
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	items, err := worktree.ParsePorcelainZ(string(worktreeOut))
	if err != nil {
		return nil, err
	}

	return worktree.Normalize(items, currentPath), nil
}

func isNotGitRepository(err error, stdout []byte, stderr []byte) bool {
	combined := strings.ToLower(string(stdout) + " " + string(stderr) + " " + err.Error())
	return strings.Contains(combined, "not a git repository")
}
