// Package e2e_test exercises the go-tidy-all binary end to end.
package e2e_test

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const (
	e2eEnv = "TIDYALL_E2E"
	binEnv = "TIDYALL_E2E_BIN"
)

// TestMain builds the go-tidy-all binary once for the e2e run.
func TestMain(m *testing.M) {
	if os.Getenv(e2eEnv) != "1" {
		os.Exit(m.Run())
	}

	dir, err := os.MkdirTemp("", "go-tidy-all-e2e-")
	if err != nil {
		log.Fatalf("create temp dir: %v", err)
	}

	bin := filepath.Join(dir, "go-tidy-all")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	//nolint:gosec // test-only build of the module binary into a fresh temp dir
	build := exec.CommandContext(context.Background(), "go", "build", "-o", bin, "github.com/logica-oss/go-tidy-all/cmd/go-tidy-all")
	if out, err := build.CombinedOutput(); err != nil {
		log.Fatalf("build go-tidy-all: %v\n%s", err, out)
	}

	if err := os.Setenv(binEnv, bin); err != nil {
		log.Fatalf("set %s: %v", binEnv, err)
	}

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func tidyBin(t *testing.T) string {
	t.Helper()

	bin := os.Getenv(binEnv)
	if bin == "" {
		t.Fatalf("%s is not set", binEnv)
	}

	return bin
}

type result struct {
	stdout string
	stderr string
	code   int
}

func runTidy(t *testing.T, root string, args ...string) result {
	t.Helper()

	var stdout, stderr strings.Builder
	//nolint:gosec // test-only execution of the binary built in TestMain
	cmd := exec.CommandContext(t.Context(), tidyBin(t), args...)
	cmd.Dir = root
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	code := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		//nolint:modernize // errors.As with explicit target keeps stdlib-only deps.
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run go-tidy-all: %v", err)
		}
	}

	return result{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	return string(data)
}

func TestE2E(t *testing.T) {
	t.Parallel()

	if os.Getenv(e2eEnv) != "1" {
		t.Skip("set TIDYALL_E2E=1 to run e2e")
	}

	t.Run("tidies single module", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, filepath.Join(root, "go.mod"), "module example.com/m\n\ngo 1.21.0\n")
		writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")

		got := runTidy(t, root, "--root", root)
		if got.code != 0 {
			t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", got.code, got.stdout, got.stderr)
		}
	})

	t.Run("dry run changes nothing", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		before := "module example.com/m\n\ngo 1.21.0\n"
		writeFile(t, filepath.Join(root, "go.mod"), before)
		writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")

		got := runTidy(t, root, "--root", root, "-n")
		if got.code != 0 {
			t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", got.code, got.stdout, got.stderr)
		}

		if after := readFile(t, filepath.Join(root, "go.mod")); after != before {
			t.Errorf("go.mod changed in dry run:\n%s", after)
		}
	})

	t.Run("skips dirs", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
		writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")
		writeFile(t, filepath.Join(root, "skipme", "go.mod"), "module example.com/skip\n\ngo 1.21.0\n")
		writeFile(t, filepath.Join(root, "skipme", "main.go"), "package main\n\nfunc main() {}\n")

		got := runTidy(t, root, "--root", root, "--skip-dirs", "skipme", "-n")
		if got.code != 0 {
			t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", got.code, got.stdout, got.stderr)
		}
	})

	t.Run("no modules fails", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, filepath.Join(root, "main.go"), "package main\n")

		got := runTidy(t, root, "--root", root)
		if got.code == 0 {
			t.Errorf("exit code = 0, want non-zero")
		}
	})

	t.Run("invalid min-go-version fails before changes", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		writeFile(t, filepath.Join(root, "go.mod"), "module example.com/m\n\ngo 1.21.0\n")

		got := runTidy(t, root, "--root", root, "--min-go-version", "bad")
		if got.code == 0 {
			t.Errorf("exit code = 0, want non-zero")
		}
	})
}
