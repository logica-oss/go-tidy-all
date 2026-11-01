package tidyall_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/errors"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := mkParent(path); err != nil {
		t.Fatal(err)
	}
	if err := writeTestFile(path, content); err != nil {
		t.Fatal(err)
	}
}

func resolveRoot(t *testing.T, root string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}

	return resolved
}

func mkParent(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return errors.WithStack(fmt.Errorf("create parent dir: %w", err))
	}

	return nil
}

func writeTestFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return errors.WithStack(fmt.Errorf("write test file: %w", err))
	}

	return nil
}
