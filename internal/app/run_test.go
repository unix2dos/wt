package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunHelpPrintsUsageAndExitsZero(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	code := Run(context.Background(), []string{"--help"}, strings.NewReader(""), stdout, stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := stdout.String(); !strings.Contains(got, "wt [--fzf] [index]") {
		t.Fatalf("expected usage to mention wt [--fzf] [index], got %q", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no stderr output, got %q", stderr.String())
	}
}
