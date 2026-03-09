package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func writeTestBytes(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func setTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("OPENCLAW_STATE_DIR", "")
	t.Setenv("OPENCLAW_CONFIG_PATH", "")
	t.Setenv("IRONCLAW_BASE_DIR", "")
	t.Setenv("OLLAMA_MODELS", "")
	return home
}

func withFakeStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(file, []byte(input), 0o600); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	fh, err := os.Open(file)
	if err != nil {
		t.Fatalf("open stdin: %v", err)
	}
	defer fh.Close()

	old := os.Stdin
	os.Stdin = fh
	defer func() {
		os.Stdin = old
	}()

	fn()
}

func withFakeArgs(t *testing.T, args []string, fn func()) {
	t.Helper()
	old := os.Args
	os.Args = args
	defer func() {
		os.Args = old
	}()
	fn()
}

func withStdoutFile(t *testing.T, file *os.File, fn func()) {
	t.Helper()
	old := os.Stdout
	os.Stdout = file
	defer func() {
		os.Stdout = old
	}()
	fn()
}
