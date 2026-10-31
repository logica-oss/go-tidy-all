package tidyall_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
)

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
