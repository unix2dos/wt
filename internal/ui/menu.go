package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"ww/internal/worktree"
)

const humanStatusWidth = len("[CURRENT] [MERGED]")
const humanIndexHeader = "INDEX"
const humanStatusHeader = "STATUS"
const humanBranchHeader = "BRANCH"
const humanPathHeader = "PATH"

func normalizedBranchWidth(branchWidth int) int {
	if branchWidth < len(humanBranchHeader) {
		return len(humanBranchHeader)
	}
	return branchWidth
}

func RenderMenu(w io.Writer, items []worktree.Worktree) {
	branchWidth := HumanBranchWidth(items)
	for _, item := range items {
		fmt.Fprintln(w, formatMenuRow(item, branchWidth))
	}
	fmt.Fprint(w, "Select a worktree [number]: ")
}

func HumanBranchWidth(items []worktree.Worktree) int {
	width := 0
	for _, item := range items {
		if len(item.BranchLabel) > width {
			width = len(item.BranchLabel)
		}
	}
	return width
}

func FormatHumanHeader(branchWidth int) string {
	branchWidth = normalizedBranchWidth(branchWidth)
	return fmt.Sprintf("%-5s %-*s %-*s %s", humanIndexHeader, humanStatusWidth, humanStatusHeader, branchWidth, humanBranchHeader, humanPathHeader)
}

func FormatHumanDivider(branchWidth int) string {
	branchWidth = normalizedBranchWidth(branchWidth)
	return fmt.Sprintf("%-5s %-*s %-*s %s", strings.Repeat("-", len(humanIndexHeader)), humanStatusWidth, strings.Repeat("-", humanStatusWidth), branchWidth, strings.Repeat("-", branchWidth), strings.Repeat("-", len(humanPathHeader)))
}

func FormatHumanRow(item worktree.Worktree, branchWidth int) string {
	branchWidth = normalizedBranchWidth(branchWidth)
	return fmt.Sprintf("[%d] %-*s %-*s %s", item.Index, humanStatusWidth, StatusText(item), branchWidth, item.BranchLabel, item.Path)
}

func formatMenuRow(item worktree.Worktree, branchWidth int) string {
	return FormatHumanRow(item, branchWidth)
}

func formatTUIRow(item worktree.Worktree, active bool, branchWidth int) string {
	prefix := " "
	if active {
		prefix = "*"
	}
	return fmt.Sprintf("%s %s", prefix, formatMenuRow(item, branchWidth))
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
