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

// HumanBranchWidth returns the max branch label length across items.
func HumanBranchWidth(items []worktree.Worktree) int {
	width := 0
	for _, item := range items {
		if len(item.BranchLabel) > width {
			width = len(item.BranchLabel)
		}
	}
	return width
}

// FormatHumanHeader returns the header row for list-table output.
func FormatHumanHeader(branchWidth int) string {
	branchWidth = normalizedBranchWidth(branchWidth)
	return fmt.Sprintf("%-5s %-*s %-*s %s", humanIndexHeader, humanStatusWidth, humanStatusHeader, branchWidth, humanBranchHeader, humanPathHeader)
}

// FormatHumanDivider returns the divider row for list-table output.
func FormatHumanDivider(branchWidth int) string {
	branchWidth = normalizedBranchWidth(branchWidth)
	return fmt.Sprintf("%-5s %-*s %-*s %s", strings.Repeat("-", len(humanIndexHeader)), humanStatusWidth, strings.Repeat("-", humanStatusWidth), branchWidth, strings.Repeat("-", branchWidth), strings.Repeat("-", len(humanPathHeader)))
}

// FormatHumanRow returns a single row for list-table output.
func FormatHumanRow(item worktree.Worktree, branchWidth int) string {
	branchWidth = normalizedBranchWidth(branchWidth)
	return fmt.Sprintf("[%d] %-*s %-*s %s", item.Index, humanStatusWidth, StatusText(item), branchWidth, item.BranchLabel, item.Path)
}

// formatMenuRow is kept for tui.go backward compat.
func formatMenuRow(item worktree.Worktree, branchWidth int) string {
	return FormatHumanRow(item, branchWidth)
}

// formatTUIRow is used by tui.go.
func formatTUIRow(item worktree.Worktree, active bool, branchWidth int) string {
	prefix := " "
	if active {
		prefix = "*"
	}
	return fmt.Sprintf("%s %s", prefix, formatMenuRow(item, branchWidth))
}

// isSafeToRemove returns true if the worktree can be safely removed.
func isSafeToRemove(item worktree.Worktree) bool {
	return item.IsMerged && !item.IsCurrent && !item.IsDirty
}

// colorizeStatus applies colors to status tags.
func colorizeStatus(item worktree.Worktree) string {
	tags := StatusTags(item)
	colored := make([]string, 0, len(tags))
	for _, tag := range tags {
		switch tag {
		case "[CURRENT]":
			colored = append(colored, Bold(tag))
		case "[MERGED]":
			colored = append(colored, Green(tag))
		default:
			colored = append(colored, tag)
		}
	}
	return strings.Join(colored, " ")
}

// aheadBehindWidth computes the max visual width of ahead/behind column across items.
func aheadBehindWidth(items []worktree.Worktree) int {
	max := 0
	for _, item := range items {
		w := VisualLen(FormatAheadBehind(item.Ahead, item.Behind))
		if w > max {
			max = w
		}
	}
	return max
}

// fileChangesWidth computes the max visual width of file changes column across items.
func fileChangesWidth(items []worktree.Worktree) int {
	max := 0
	for _, item := range items {
		w := VisualLen(FormatFileChanges(item.Staged, item.Unstaged, item.Untracked))
		if w > max {
			max = w
		}
	}
	return max
}

// formatSummary builds the summary line shown below the menu.
func formatSummary(items []worktree.Worktree) string {
	n := len(items)
	var noun string
	if n == 1 {
		noun = "worktree"
	} else {
		noun = "worktrees"
	}

	// Collect safe-to-remove indices
	safeIndices := make([]string, 0)
	for _, item := range items {
		if isSafeToRemove(item) {
			safeIndices = append(safeIndices, strconv.Itoa(item.Index))
		}
	}

	base := fmt.Sprintf("%d %s", n, noun)
	if len(safeIndices) == 0 {
		return base
	}

	k := len(safeIndices)
	var safeNoun string
	if k == 1 {
		safeNoun = "safe to remove"
	} else {
		safeNoun = "safe to remove"
	}
	hint := fmt.Sprintf("ww rm %s", strings.Join(safeIndices, " "))
	return fmt.Sprintf("%s · %d %s (%s)", base, k, safeNoun, hint)
}

// formatEnhancedMenuRow builds one row of the enhanced menu with visual-width-aware padding.
func formatEnhancedMenuRow(item worktree.Worktree, branchWidth, abWidth, fcWidth int) string {
	// Marker column: ★ for current, space otherwise
	var marker string
	if item.IsCurrent {
		marker = "★"
	} else {
		marker = " "
	}

	// Index (not in brackets)
	index := fmt.Sprintf("%d", item.Index)

	// Status column (colored, fixed visual width = len("[CURRENT] [MERGED]"))
	statusStr := colorizeStatus(item)
	statusCol := PadRight(statusStr, humanStatusWidth)

	// Branch column (dimmed if merged)
	branchStr := item.BranchLabel
	if item.IsMerged && !item.IsCurrent {
		branchStr = Dim(branchStr)
	}
	branchCol := PadRight(branchStr, branchWidth)

	// Ahead/behind column
	abStr := FormatAheadBehind(item.Ahead, item.Behind)
	abCol := PadRight(abStr, abWidth)

	// File changes column
	fcStr := FormatFileChanges(item.Staged, item.Unstaged, item.Untracked)
	fcCol := PadRight(fcStr, fcWidth)

	// Path (dimmed if merged)
	pathStr := item.Path
	if item.IsMerged && !item.IsCurrent {
		pathStr = Dim(pathStr)
	}

	// Build with consistent spacing
	// "★ 1  [CURRENT]  main          ↑2  +3 ~1 ?2  /repo"
	return fmt.Sprintf("%s %-2s %s  %s  %s  %s  %s",
		marker,
		index,
		statusCol,
		branchCol,
		abCol,
		fcCol,
		pathStr,
	)
}

// RenderMenu writes the enhanced interactive menu to w.
func RenderMenu(w io.Writer, items []worktree.Worktree) {
	branchWidth := HumanBranchWidth(items)
	abWidth := aheadBehindWidth(items)
	fcWidth := fileChangesWidth(items)

	for _, item := range items {
		fmt.Fprintln(w, formatEnhancedMenuRow(item, branchWidth, abWidth, fcWidth))
	}

	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  "+formatSummary(items))
	fmt.Fprintf(w, "Select [1-%d]> ", len(items))
}

// ReadSelection reads a number in [1,max] from in, retrying on invalid input.
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
