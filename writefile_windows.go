//go:build windows

package tidyall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
)

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	// Best effort crash-safe write: os.Rename is not documented as atomic on
	// Windows, so this only guarantees no partial file remains on error, not
	// atomic replacement.
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return errors.WithStack(fmt.Errorf("create temp file: %w", err))
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("write temp file: %w", err))
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("close temp file: %w", err))
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("chmod temp file: %w", err))
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)

		return errors.WithStack(fmt.Errorf("rename temp file: %w", err))
	}

	return nil
}
