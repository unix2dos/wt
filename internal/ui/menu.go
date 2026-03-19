package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"wt/internal/worktree"
)

func RenderMenu(w io.Writer, items []worktree.Worktree) {
	for _, item := range items {
		marker := " "
		if item.IsCurrent {
			marker = "*"
		}
		fmt.Fprintf(w, "[%d] %s %s %s\n", item.Index, marker, item.BranchLabel, item.Path)
	}
	fmt.Fprint(w, "Select a worktree [number]: ")
}

func ReadSelection(in io.Reader, errOut io.Writer, max int) (int, error) {
	reader := bufio.NewReader(in)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, err
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if errors.Is(err, io.EOF) {
				return 0, io.EOF
			}
			fmt.Fprintln(errOut, "empty selection")
			continue
		}

		index, convErr := strconv.Atoi(trimmed)
		if convErr != nil || index <= 0 || index > max {
			fmt.Fprintf(errOut, "invalid worktree selection: %q\n", trimmed)
			if errors.Is(err, io.EOF) {
				return 0, io.EOF
			}
			continue
		}

		return index, nil
	}
}
