package config

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/k1LoW/errors"
)

// Validate checks the loaded config.
func Validate(cfg Config) error {
	if cfg.MinGoVersion != "" && compareGoVersion(cfg.MinGoVersion, "1.0.0") < 0 {
		return errors.WithStack(fmt.Errorf("invalid min-go-version %q: must be like 1.24.0", cfg.MinGoVersion))
	}

	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

func preferredConfigFile(root string, rootOnly bool) string {
	dirs := []string{"."}
	if root != "" {
		dirs = []string{root, "."}
		if rootOnly {
			dirs = []string{root}
		}
	}

	for _, dir := range dirs {
		for _, name := range []string{".tidy-all.yaml", ".tidy-all.json"} {
			p := filepath.Join(dir, name)

			//nolint:gosec // dir is "." or the resolved root, name is a fixed config filename
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p
			}
		}
	}

	return ""
}

func resolveRoot(root string) (string, error) {
	if root == "" {
		root = detectRoot()
	}

	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.WithStack(fmt.Errorf("resolve root: %w", err))
		}

		root = cwd
	}

	abs, err := filepath.Abs(root)
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("resolve root: %w", err))
	}

	st, err := os.Stat(abs)
	if err != nil {
		return "", errors.WithStack(fmt.Errorf("resolve root: %w", err))
	}
	if !st.IsDir() {
		return "", errors.WithStack(fmt.Errorf("resolve root %q: not a directory", root))
	}

	return abs, nil
}

func detectRoot() string {
	if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
		return ws
	}

	out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func compareGoVersion(a, b string) int {
	const maxParts = 3

	aParts := strings.SplitN(strings.TrimPrefix(a, "go"), ".", maxParts)
	bParts := strings.SplitN(strings.TrimPrefix(b, "go"), ".", maxParts)

	for i := range maxParts {
		aNum := goVersionPart(aParts, i)
		bNum := goVersionPart(bParts, i)
		if aNum != bNum {
			if aNum < bNum {
				return -1
			}

			return 1
		}
	}

	return 0
}

func goVersionPart(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}

	var num int
	_, _ = fmt.Sscanf(parts[i], "%d", &num)

	return num
}
