// Package tidyall runs go mod tidy for every module and pins each go directive to the lowest version that mingo computes.
package tidyall

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/k1LoW/errors"
)

// ErrNoModule is returned when no go.mod is found under the root.
var ErrNoModule = errors.New("no go.mod found")

// Config holds tidy-all settings.
type Config struct {
	// Root is the repository root.
	Root string
	// DryRun prints planned changes without writing.
	DryRun bool
	// SkipDirs skips modules under these root-relative directories.
	SkipDirs []string
	// MinGoVersion pins the go directive to at least this version (e.g. "1.24.0").
	// Empty follows the mingo result.
	MinGoVersion string
}

// Run tidies every module under root.
func Run(ctx context.Context, log *slog.Logger, cfg Config) error {
	dirs, err := findModules(cfg.Root, cfg.SkipDirs)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return errors.WithStack(ErrNoModule)
	}

	for _, dir := range dirs {
		if err := processModule(ctx, log, dir, cfg); err != nil {
			return err
		}
	}

	return nil
}

func processModule(ctx context.Context, log *slog.Logger, dir string, cfg Config) error {
	target := dir
	if cfg.DryRun {
		staged, cleanup, err := stageModule(log, dir)
		if err != nil {
			return err
		}
		defer cleanup()

		target = staged
	}

	if err := tidy(ctx, log, target, cfg.DryRun); err != nil {
		return err
	}

	return pinGoDirective(log, target, cfg.DryRun, cfg.MinGoVersion)
}

func stageModule(log *slog.Logger, dir string) (string, func(), error) {
	staged, err := os.MkdirTemp("", "go-tidy-all-*")
	if err != nil {
		return "", nil, errors.WithStack(fmt.Errorf("create stage dir: %w", err))
	}

	cleanup := func() { _ = os.RemoveAll(staged) }

	if err := copyDir(log, dir, staged); err != nil {
		cleanup()

		return "", nil, err
	}

	return staged, cleanup, nil
}

func copyDir(log *slog.Logger, src, dst string) error {
	cleanSrc := filepath.Clean(src)

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return errors.WithStack(fmt.Errorf("create dst root %s: %w", dst, err))
	}

	dstRoot, err := os.OpenRoot(dst)
	if err != nil {
		return errors.WithStack(fmt.Errorf("open dst root %s: %w", dst, err))
	}
	defer func() { _ = dstRoot.Close() }()

	return copyDirValidated(log, cleanSrc, dstRoot, ".")
}

func copyDirValidated(log *slog.Logger, srcRoot string, dstRoot *os.Root, dir string) error {
	srcDir := filepath.Join(srcRoot, filepath.FromSlash(dir))

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		log.Warn("skip unreadable dir in dry-run copy", "dir", dir, "error", err)

		return nil
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." {
			return errors.WithStack(fmt.Errorf("invalid entry name %q", name))
		}

		if entry.Type()&fs.ModeSymlink != 0 {
			log.Warn("skip symlink in dry-run copy", "path", filepath.Join(dir, name))

			continue
		}

		rel := name
		if dir != "." {
			rel = filepath.Join(dir, name)
		}

		if entry.IsDir() {
			if name == ".git" {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				return errors.WithStack(fmt.Errorf("stat dir %s: %w", rel, err))
			}
			if err := dstRoot.MkdirAll(rel, info.Mode().Perm()); err != nil {
				return errors.WithStack(fmt.Errorf("create dir %s: %w", rel, err))
			}
			if err := copyDirValidated(log, srcRoot, dstRoot, rel); err != nil {
				return err
			}

			continue
		}

		data, err := os.ReadFile(filepath.Join(srcRoot, filepath.FromSlash(rel)))
		if err != nil {
			return errors.WithStack(fmt.Errorf("read file %s: %w", rel, err))
		}
		info, err := entry.Info()
		if err != nil {
			return errors.WithStack(fmt.Errorf("stat file %s: %w", rel, err))
		}
		//nolint:gosec // dstRoot confines the write; rel is built from validated entry names.
		if err := dstRoot.WriteFile(rel, data, info.Mode().Perm()); err != nil {
			return errors.WithStack(fmt.Errorf("write file %s: %w", rel, err))
		}
	}

	return nil
}
