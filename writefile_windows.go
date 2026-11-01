//go:build windows

package tidyall

import (
	"fmt"
	"os"

	"github.com/k1LoW/errors"
)

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return errors.WithStack(fmt.Errorf("write file: %w", err))
	}

	return nil
}
