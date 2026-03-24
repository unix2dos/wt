package tasknote

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteFileReadFileRoundTrip(t *testing.T) {
	note := Note{
		TaskLabel: "task:fix-login",
		Branch:    "fix-login",
		CreatedAt: time.Date(2026, 3, 24, 12, 34, 56, 0, time.UTC),
		Intent:    "Fix the login redirect loop",
		Body: strings.Join([]string{
			"Created by ww.",
			"",
			"Why this exists:",
			"- Keep login-task changes isolated from docs or release work.",
		}, "\n"),
	}

	path := filepath.Join(t.TempDir(), "notes", "task-note.md")
	if err := WriteFile(path, note); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	if got.TaskLabel != note.TaskLabel {
		t.Fatalf("expected task label %q, got %q", note.TaskLabel, got.TaskLabel)
	}
	if got.Branch != note.Branch {
		t.Fatalf("expected branch %q, got %q", note.Branch, got.Branch)
	}
	if !got.CreatedAt.Equal(note.CreatedAt) {
		t.Fatalf("expected created_at %s, got %s", note.CreatedAt, got.CreatedAt)
	}
	if got.Intent != note.Intent {
		t.Fatalf("expected intent %q, got %q", note.Intent, got.Intent)
	}
	if got.Body != note.Body {
		t.Fatalf("expected body %q, got %q", note.Body, got.Body)
	}
}

func TestParseRejectsMissingOpeningDelimiter(t *testing.T) {
	_, err := Parse([]byte("task_label: task:fix-login\nbranch: fix-login\n"))
	if err == nil {
		t.Fatal("expected error for missing opening delimiter")
	}
}

func TestParseRejectsMalformedHeaderLine(t *testing.T) {
	raw := strings.Join([]string{
		"---",
		"task_label task:fix-login",
		"branch: fix-login",
		"created_at: 2026-03-24T12:34:56Z",
		"---",
		"Created by ww.",
	}, "\n")

	_, err := Parse([]byte(raw))
	if err == nil {
		t.Fatal("expected error for malformed header line")
	}
}

func TestParseRejectsMissingRequiredFields(t *testing.T) {
	raw := strings.Join([]string{
		"---",
		"created_at: 2026-03-24T12:34:56Z",
		"---",
		"Created by ww.",
	}, "\n")

	_, err := Parse([]byte(raw))
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestParsePreservesTrailingBodyText(t *testing.T) {
	raw := strings.Join([]string{
		"---",
		"task_label: task:fix-login",
		"branch: fix-login",
		"created_at: 2026-03-24T12:34:56Z",
		"intent: Fix the login redirect loop",
		"---",
		"Created by ww.",
		"",
		"Next steps:",
		"- Reproduce the redirect loop.",
	}, "\n")

	got, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	wantBody := strings.Join([]string{
		"Created by ww.",
		"",
		"Next steps:",
		"- Reproduce the redirect loop.",
	}, "\n")
	if got.Body != wantBody {
		t.Fatalf("expected body %q, got %q", wantBody, got.Body)
	}
}
