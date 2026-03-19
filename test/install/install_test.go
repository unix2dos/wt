package install_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallIsIdempotentAndBuildsBinary(t *testing.T) {
	home := t.TempDir()
	rcPath := filepath.Join(home, ".zshrc")
	if err := os.WriteFile(rcPath, []byte(""), 0o644); err != nil {
		t.Fatalf("write rc file: %v", err)
	}

	firstInstall := runInstall(t, home)
	runInstall(t, home)

	if !strings.Contains(firstInstall, "Use `cwt` to switch") {
		t.Fatalf("expected install output to mention cwt, got %q", firstInstall)
	}
	if !strings.Contains(firstInstall, "Use `wt --fzf`") {
		t.Fatalf("expected install output to mention wt --fzf, got %q", firstInstall)
	}

	data, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read rc file: %v", err)
	}

	sourceLine := "source \"" + filepath.Join(home, ".local", "bin", "wt-cwt.sh") + "\""
	if strings.Count(string(data), sourceLine) != 1 {
		t.Fatalf("expected one managed source line, got %q", string(data))
	}
	if strings.Count(string(data), "wt shell wrapper begin") != 1 {
		t.Fatalf("expected one managed block, got %q", string(data))
	}

	binPath := filepath.Join(home, ".local", "bin", "wt")
	if info, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected built binary at %s: %v", binPath, err)
	} else if info.Mode()&0o111 == 0 {
		t.Fatalf("expected built binary to be executable, mode=%v", info.Mode())
	}

	wrapperPath := filepath.Join(home, ".local", "bin", "wt-cwt.sh")
	if info, err := os.Stat(wrapperPath); err != nil {
		t.Fatalf("expected installed wrapper at %s: %v", wrapperPath, err)
	} else if info.Mode()&0o111 == 0 {
		t.Fatalf("expected installed wrapper to be executable, mode=%v", info.Mode())
	}
}

func TestInstallSupportsCustomRcFileAndBinDir(t *testing.T) {
	home := t.TempDir()
	rcPath := filepath.Join(home, ".config", "wt-test.rc")
	binDir := filepath.Join(home, ".bin")

	runInstall(t, home, "--shell", "bash", "--rc-file", rcPath, "--bin-dir", binDir)

	data, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read rc file: %v", err)
	}

	sourceLine := "source \"" + filepath.Join(binDir, "wt-cwt.sh") + "\""
	if !strings.Contains(string(data), sourceLine) {
		t.Fatalf("expected source line in custom rc file, got %q", string(data))
	}

	binPath := filepath.Join(binDir, "wt")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("expected custom binary at %s: %v", binPath, err)
	}

	wrapperPath := filepath.Join(binDir, "wt-cwt.sh")
	if _, err := os.Stat(wrapperPath); err != nil {
		t.Fatalf("expected custom wrapper at %s: %v", wrapperPath, err)
	}
}

func TestUninstallRemovesManagedBlockAndBinary(t *testing.T) {
	home := t.TempDir()
	rcPath := filepath.Join(home, ".zshrc")
	if err := os.WriteFile(rcPath, []byte(""), 0o644); err != nil {
		t.Fatalf("write rc file: %v", err)
	}

	runInstall(t, home)
	runUninstall(t, home)

	if _, err := os.Stat(filepath.Join(home, ".local", "bin", "wt")); !os.IsNotExist(err) {
		t.Fatalf("expected binary to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".local", "bin", "wt-cwt.sh")); !os.IsNotExist(err) {
		t.Fatalf("expected wrapper to be removed, got err=%v", err)
	}

	data, err := os.ReadFile(rcPath)
	if err != nil {
		t.Fatalf("read rc file: %v", err)
	}
	if strings.Contains(string(data), "wt shell wrapper begin") {
		t.Fatalf("expected managed block removed, got %q", string(data))
	}
}

func TestCwtChangesDirectoryOnSuccess(t *testing.T) {
	repoRoot := projectRoot(t)
	origin := t.TempDir()
	target := filepath.Join(t.TempDir(), "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}

	out := runShell(t, repoRoot, fmt.Sprintf(`
		cd %q
		source %q
		wt() { printf '%%s\n' %q; }
		cwt >/dev/null
		pwd
	`, origin, filepath.Join(repoRoot, "shell", "cwt.sh"), target))

	if got := strings.TrimSpace(out); got != target {
		t.Fatalf("expected shell to cd to %q, got %q", target, got)
	}
}

func TestCwtLeavesDirectoryOnFailure(t *testing.T) {
	repoRoot := projectRoot(t)
	origin := t.TempDir()

	out := runShell(t, repoRoot, fmt.Sprintf(`
		cd %q
		source %q
		wt() { return 1; }
		if cwt >/dev/null 2>&1; then
			echo unexpected-success
			exit 1
		fi
		pwd
	`, origin, filepath.Join(repoRoot, "shell", "cwt.sh")))

	if got := strings.TrimSpace(out); got != origin {
		t.Fatalf("expected shell to stay in %q, got %q", origin, got)
	}
}

func TestCwtLeavesDirectoryOnEmptyOutput(t *testing.T) {
	repoRoot := projectRoot(t)
	origin := t.TempDir()

	out := runShell(t, repoRoot, fmt.Sprintf(`
		cd %q
		source %q
		wt() { :; }
		if cwt >/dev/null 2>&1; then
			echo unexpected-success
			exit 1
		fi
		pwd
	`, origin, filepath.Join(repoRoot, "shell", "cwt.sh")))

	if got := strings.TrimSpace(out); got != origin {
		t.Fatalf("expected shell to stay in %q, got %q", origin, got)
	}
}

func runInstall(t *testing.T, home string, args ...string) string {
	t.Helper()

	cmdArgs := append([]string{"install.sh"}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = projectRoot(t)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"SHELL=/bin/zsh",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out)
	}
	t.Fatalf("install failed: %v\n%s", err, out)
	return ""
}

func runUninstall(t *testing.T, home string, args ...string) {
	t.Helper()

	cmdArgs := append([]string{"uninstall.sh"}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = projectRoot(t)
	cmd.Env = append(os.Environ(),
		"HOME="+home,
		"SHELL=/bin/zsh",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return
	}
	t.Fatalf("uninstall failed: %v\n%s", err, out)
}

func runShell(t *testing.T, repoRoot, script string) string {
	t.Helper()

	cmd := exec.Command("bash", "-lc", script)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("shell script failed: %v\n%s", err, out)
	}
	return string(out)
}

func projectRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
