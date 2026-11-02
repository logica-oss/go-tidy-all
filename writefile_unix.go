//go:build !windows

package tidyall

import (
	"fmt"
	"os"

	"github.com/google/renameio/v2"
	"github.com/k1LoW/errors"
)

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if err := renameio.WriteFile(path, data, perm, renameio.IgnoreUmask()); err != nil {
		return errors.WithStack(fmt.Errorf("write file: %w", err))
	}

	return nil
}
