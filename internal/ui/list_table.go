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
	Detail   string
}

func FormatListTable(entries []ListTableEntry) string {
	if len(entries) == 0 {
		return ""
	}

	branchWidth := listBranchWidth(entries)
	var buf strings.Builder

	buf.WriteString(listTableBorder("┌", "┬", "┐", branchWidth))
	buf.WriteByte('\n')
	buf.WriteString(listTableRow(humanIndexHeader, humanStatusHeader, humanBranchHeader, listABHeader, listChangesHeader, humanPathHeader, branchWidth))
	buf.WriteByte('\n')
	buf.WriteString(listTableBorder("├", "┼", "┤", branchWidth))
	buf.WriteByte('\n')

	for i, entry := range entries {
		for _, row := range listTableRows(entry, branchWidth) {
			buf.WriteString(row)
			buf.WriteByte('\n')
		}
		if i == len(entries)-1 {
			buf.WriteString(listTableBorder("└", "┴", "┘", branchWidth))
		} else {
			buf.WriteString(listTableBorder("├", "┼", "┤", branchWidth))
			buf.WriteByte('\n')
		}
	}

	return buf.String()
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

func listTableRows(entry ListTableEntry, branchWidth int) []string {
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
			status = StatusText(entry.Worktree)
			ab = FormatWorktreeAheadBehind(entry.Worktree)
			changes = FormatFileChanges(entry.Worktree.Staged, entry.Worktree.Unstaged, entry.Worktree.Untracked)
		}
		branch := lineAt(branchLines, i)
		pathLine := lineAt(pathLines, i)
		rows = append(rows, listTableRow(index, status, branch, ab, changes, pathLine, branchWidth))
	}
	return rows
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func listTableRow(index, status, branch, ab, changes, path string, branchWidth int) string {
	return fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %s │ %s │ %-*s │",
		listIndexWidth, index,
		humanStatusWidth, status,
		branchWidth, branch,
		PadRight(ab, listABWidth),
		PadRight(changes, listChangesWidth),
		listPathWidth, path,
	)
}

func listTableBorder(left, mid, right string, branchWidth int) string {
	return left +
		strings.Repeat("─", listIndexWidth+2) + mid +
		strings.Repeat("─", humanStatusWidth+2) + mid +
		strings.Repeat("─", branchWidth+2) + mid +
		strings.Repeat("─", listABWidth+2) + mid +
		strings.Repeat("─", listChangesWidth+2) + mid +
		strings.Repeat("─", listPathWidth+2) +
		right
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
