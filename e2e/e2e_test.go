// Package e2e_test exercises the go-tidy-all binary end to end.
package e2e_test

import (
	"context"
	"errors"
	"fmt"
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

func copyFixture(t *testing.T, fixture, root string) {
	t.Helper()

	src := filepath.Join("testdata", fixture)

	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixture, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(root, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				t.Fatalf("copy fixture dir: %v", err)
			}

			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read fixture file: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			t.Fatalf("create parent dir: %v", err)
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", src, err)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", dst, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}

			continue
		}

		data, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", srcPath, err)
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			return fmt.Errorf("write file %s: %w", dstPath, err)
		}
	}

	return nil
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
		copyFixture(t, "single", root)

		got := runTidy(t, root, "--root", root)
		if got.code != 0 {
			t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", got.code, got.stdout, got.stderr)
		}
	})

	t.Run("tidies multiple modules", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		copyFixture(t, "multi", root)

		got := runTidy(t, root, "--root", root)
		if got.code != 0 {
			t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", got.code, got.stdout, got.stderr)
		}
	})

	t.Run("dry run changes nothing", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		copyFixture(t, "single", root)
		before := readFile(t, filepath.Join(root, "go.mod"))

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
		copyFixture(t, "skip", root)

		got := runTidy(t, root, "--root", root, "--skip-dirs", "skip", "-n")
		if got.code != 0 {
			t.Fatalf("exit code = %d, want 0\nstdout: %s\nstderr: %s", got.code, got.stdout, got.stderr)
		}
	})

	t.Run("no modules fails", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		copyFixture(t, "empty", root)

		got := runTidy(t, root, "--root", root)
		if got.code == 0 {
			t.Errorf("exit code = 0, want non-zero")
		}
	})

	t.Run("invalid min-go-version fails before changes", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		copyFixture(t, "single", root)

		got := runTidy(t, root, "--root", root, "--min-go-version", "bad")
		if got.code == 0 {
			t.Errorf("exit code = 0, want non-zero")
		}
	})
}
