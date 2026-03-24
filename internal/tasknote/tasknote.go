package tasknote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Note struct {
	TaskLabel string
	Branch    string
	CreatedAt time.Time
	Intent    string
	Body      string
}

func Render(note Note) ([]byte, error) {
	if strings.TrimSpace(note.TaskLabel) == "" {
		return nil, fmt.Errorf("task_label is required")
	}
	if strings.TrimSpace(note.Branch) == "" {
		return nil, fmt.Errorf("branch is required")
	}
	if note.CreatedAt.IsZero() {
		return nil, fmt.Errorf("created_at is required")
	}

	lines := []string{
		"---",
		"task_label: " + note.TaskLabel,
		"branch: " + note.Branch,
		"created_at: " + note.CreatedAt.UTC().Format(time.RFC3339),
	}
	if strings.TrimSpace(note.Intent) != "" {
		lines = append(lines, "intent: "+note.Intent)
	}
	lines = append(lines, "---")
	if note.Body != "" {
		lines = append(lines, note.Body)
	}

	return []byte(strings.Join(lines, "\n")), nil
}

func Parse(raw []byte) (Note, error) {
	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return Note{}, fmt.Errorf("missing opening delimiter")
	}

	var note Note
	var closing int = -1
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if line == "---" {
			closing = i
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return Note{}, fmt.Errorf("malformed header line %q", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "task_label":
			note.TaskLabel = value
		case "branch":
			note.Branch = value
		case "created_at":
			parsed, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return Note{}, fmt.Errorf("invalid created_at %q: %w", value, err)
			}
			note.CreatedAt = parsed
		case "intent":
			note.Intent = value
		default:
			return Note{}, fmt.Errorf("unknown header key %q", key)
		}
	}

	if closing == -1 {
		return Note{}, fmt.Errorf("missing closing delimiter")
	}
	if strings.TrimSpace(note.TaskLabel) == "" {
		return Note{}, fmt.Errorf("task_label is required")
	}
	if strings.TrimSpace(note.Branch) == "" {
		return Note{}, fmt.Errorf("branch is required")
	}
	if note.CreatedAt.IsZero() {
		return Note{}, fmt.Errorf("created_at is required")
	}

	if closing+1 < len(lines) {
		note.Body = strings.Join(lines[closing+1:], "\n")
	}
	return note, nil
}

func ReadFile(path string) (Note, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Note{}, err
	}
	return Parse(raw)
}

func WriteFile(path string, note Note) error {
	raw, err := Render(note)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
