package ui

import (
	"fmt"
	"strings"

	"ww/internal/worktree"
)

const listIndexWidth = len(humanIndexHeader)
const listPathWidth = 48
const listABHeader = "AHEAD/BEHIND"
const listChangesHeader = "CHANGES"
const listABWidth = len(listABHeader)           // 12
const listChangesWidth = len(listChangesHeader) // 7

type ListTableEntry struct {
	Worktree worktree.Worktree
	Status   string
	Detail   string
}

type ListTableOptions struct {
	ShowEmptyOptionalColumns bool
}

func FormatListTable(entries []ListTableEntry) string {
	return FormatListTableWithOptions(entries, ListTableOptions{})
}

func FormatListTableWithOptions(entries []ListTableEntry, opts ListTableOptions) string {
	if len(entries) == 0 {
		return ""
	}

	branchWidth := listBranchWidth(entries)
	columns := listTableColumnsFor(entries, opts)
	var buf strings.Builder

	buf.WriteString(listTableBorder("┌", "┬", "┐", branchWidth, columns))
	buf.WriteByte('\n')
	buf.WriteString(listTableRow(humanIndexHeader, humanStatusHeader, humanBranchHeader, listABHeader, listChangesHeader, humanPathHeader, branchWidth, columns))
	buf.WriteByte('\n')
	buf.WriteString(listTableBorder("├", "┼", "┤", branchWidth, columns))
	buf.WriteByte('\n')

	for i, entry := range entries {
		for _, row := range listTableRows(entry, branchWidth, columns) {
			buf.WriteString(row)
			buf.WriteByte('\n')
		}
		if i == len(entries)-1 {
			buf.WriteString(listTableBorder("└", "┴", "┘", branchWidth, columns))
		} else {
			buf.WriteString(listTableBorder("├", "┼", "┤", branchWidth, columns))
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

type listTableColumns struct {
	aheadBehind bool
	changes     bool
}

func listTableColumnsFor(entries []ListTableEntry, opts ListTableOptions) listTableColumns {
	columns := listTableColumns{}
	if opts.ShowEmptyOptionalColumns {
		columns.aheadBehind = true
		columns.changes = true
		return columns
	}

	for _, entry := range entries {
		if FormatWorktreeAheadBehind(entry.Worktree) != "" {
			columns.aheadBehind = true
		}
		if FormatFileChanges(entry.Worktree.Staged, entry.Worktree.Unstaged, entry.Worktree.Untracked) != "" {
			columns.changes = true
		}
	}
	return columns
}

func listBranchWidth(entries []ListTableEntry) int {
	items := make([]worktree.Worktree, 0, len(entries))
	for _, entry := range entries {
		items = append(items, entry.Worktree)
	}
	width := normalizedBranchWidth(HumanBranchWidth(items))
	for _, entry := range entries {
		for _, line := range strings.Split(entry.Detail, "\n") {
			if len(line) > width {
				width = len(line)
			}
		}
	}
	return width
}

func listTableRows(entry ListTableEntry, branchWidth int, columns listTableColumns) []string {
	branchContent := entry.Worktree.BranchLabel
	if entry.Detail != "" {
		branchContent += "\n" + entry.Detail
	}

	branchLines := wrapCell(branchContent, branchWidth)
	pathLines := wrapCell(entry.Worktree.Path, listPathWidth)
	rowCount := max(len(branchLines), len(pathLines))
	rows := make([]string, 0, rowCount)
	for i := 0; i < rowCount; i++ {
		index := ""
		status := ""
		ab := ""
		changes := ""
		if i == 0 {
			index = fmt.Sprintf("%d", entry.Worktree.Index)
			status = entry.Status
			if status == "" {
				status = StatusText(entry.Worktree)
			}
			ab = FormatWorktreeAheadBehind(entry.Worktree)
			changes = FormatFileChanges(entry.Worktree.Staged, entry.Worktree.Unstaged, entry.Worktree.Untracked)
		}
		branch := lineAt(branchLines, i)
		pathLine := lineAt(pathLines, i)
		rows = append(rows, listTableRow(index, status, branch, ab, changes, pathLine, branchWidth, columns))
	}
	return rows
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func listTableRow(index, status, branch, ab, changes, path string, branchWidth int, columns listTableColumns) string {
	cells := []string{
		fmt.Sprintf("%-*s", listIndexWidth, index),
		fmt.Sprintf("%-*s", humanStatusWidth, status),
		fmt.Sprintf("%-*s", branchWidth, branch),
	}
	if columns.aheadBehind {
		cells = append(cells, PadRight(ab, listABWidth))
	}
	if columns.changes {
		cells = append(cells, PadRight(changes, listChangesWidth))
	}
	cells = append(cells, fmt.Sprintf("%-*s", listPathWidth, path))
	return "│ " + strings.Join(cells, " │ ") + " │"
}

func listTableBorder(left, mid, right string, branchWidth int, columns listTableColumns) string {
	widths := []int{listIndexWidth, humanStatusWidth, branchWidth}
	if columns.aheadBehind {
		widths = append(widths, listABWidth)
	}
	if columns.changes {
		widths = append(widths, listChangesWidth)
	}
	widths = append(widths, listPathWidth)

	segments := make([]string, 0, len(widths))
	for _, width := range widths {
		segments = append(segments, strings.Repeat("─", width+2))
	}
	return left + strings.Join(segments, mid) + right
}

func wrapCell(text string, width int) []string {
	if text == "" {
		return []string{""}
	}

	var lines []string
	for _, rawLine := range strings.Split(text, "\n") {
		if rawLine == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, wrapLine(rawLine, width)...)
	}
	return lines
}

func wrapLine(text string, width int) []string {
	if width <= 0 || len(text) <= width {
		return []string{text}
	}

	var lines []string
	remaining := text
	for len(remaining) > width {
		cut, trimLeading := findWrapPoint(remaining, width)
		lines = append(lines, remaining[:cut])
		remaining = remaining[cut:]
		if trimLeading {
			remaining = strings.TrimLeft(remaining, " ")
		}
	}
	lines = append(lines, remaining)
	return lines
}

func findWrapPoint(text string, width int) (int, bool) {
	if len(text) <= width {
		return len(text), false
	}

	for i := width - 1; i >= 0; i-- {
		switch text[i] {
		case '/':
			return i + 1, false
		case '-', '_':
			return i + 1, false
		case ' ':
			return i, true
		}
	}

	return width, false
}
