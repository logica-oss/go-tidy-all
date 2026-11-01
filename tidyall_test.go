package tidyall_test

import (
	"context"
	"log/slog"
	"path/filepath"
	"testing"

	tidyall "github.com/logica-oss/go-tidy-all"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRunNoModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")

	err := tidyall.Run(context.Background(), newTestLogger(), tidyall.Config{Root: root})
	if err == nil {
		t.Fatal("Run = nil, want tidyall.ErrNoModule")
	}
}

func TestRunDryRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/m\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")

	cfg := tidyall.Config{Root: root, DryRun: true}
	if err := tidyall.Run(context.Background(), newTestLogger(), cfg); err != nil {
		t.Fatal(err)
	}

	// DryRun must not modify go.mod.
	got, err := tidyall.ReadGoVersion(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.24.0" {
		t.Errorf("go directive = %q, want unchanged 1.24.0", got)
	}
}

func TestRunSkipsDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/root\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")
	writeFile(t, filepath.Join(root, "skipme", "go.mod"), "module example.com/skip\n\ngo 1.24.0\n")
	writeFile(t, filepath.Join(root, "skipme", "main.go"), "package main\n\nfunc main() {}\n")

	cfg := tidyall.Config{Root: root, DryRun: true, SkipDirs: []string{"skipme"}}
	if err := tidyall.Run(context.Background(), newTestLogger(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestTidyDryRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/m\n\ngo 1.24.0\n")

	if err := tidyall.Tidy(context.Background(), newTestLogger(), root, true); err != nil {
		t.Fatal(err)
	}
}

func TestPinGoDirectiveDryRun(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/m\n\ngo 1.21.0\n")
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")

	if err := tidyall.PinGoDirective(newTestLogger(), dir, true, ""); err != nil {
		t.Fatal(err)
	}

	// DryRun must not modify go.mod.
	got, err := tidyall.ReadGoVersion(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.21.0" {
		t.Errorf("go directive = %q, want unchanged 1.21.0", got)
	}
}

func TestPinGoDirectiveMinVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/m\n\ngo 1.21.0\n")
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")

	if err := tidyall.PinGoDirective(newTestLogger(), dir, false, "1.99.0"); err != nil {
		t.Fatal(err)
	}

	got, err := tidyall.ReadGoVersion(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.99.0" {
		t.Errorf("go directive = %q, want 1.99.0", got)
	}
}
