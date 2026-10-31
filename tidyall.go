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
		staged, cleanup, err := stageModule(ctx, log, dir)
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

func stageModule(ctx context.Context, log *slog.Logger, dir string) (string, func(), error) {
	staged, err := os.MkdirTemp("", "go-tidy-all-*")
	if err != nil {
		return "", nil, errors.WithStack(fmt.Errorf("create stage dir: %w", err))
	}

	cleanup := func() { _ = os.RemoveAll(staged) }

	if err := copyDir(ctx, log, dir, staged); err != nil {
		cleanup()

		return "", nil, err
	}

	return staged, cleanup, nil
}

func copyDir(ctx context.Context, log *slog.Logger, src, dst string) error {
	cleanSrc := filepath.Clean(src)

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return errors.WithStack(fmt.Errorf("create dst root %s: %w", dst, err))
	}

	dstRoot, err := os.OpenRoot(dst)
	if err != nil {
		return errors.WithStack(fmt.Errorf("open dst root %s: %w", dst, err))
	}
	defer func() { _ = dstRoot.Close() }()

	return copyDirValidated(ctx, log, cleanSrc, dstRoot, ".")
}

func copyDirValidated(ctx context.Context, log *slog.Logger, srcRoot string, dstRoot *os.Root, dir string) error {
	if err := ctx.Err(); err != nil {
		return errors.WithStack(err)
	}

	srcDir := filepath.Join(srcRoot, filepath.FromSlash(dir))

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		log.Warn("skip unreadable dir in dry-run copy", "dir", dir, "error", err)

		return nil
	}

	for _, entry := range entries {
		if err := copyEntry(ctx, log, srcRoot, dstRoot, dir, entry); err != nil {
			return err
		}
	}

	return nil
}

func copyEntry(ctx context.Context, log *slog.Logger, srcRoot string, dstRoot *os.Root, dir string, entry fs.DirEntry) error {
	name := entry.Name()
	if name == "." || name == ".." {
		return errors.WithStack(fmt.Errorf("invalid entry name %q", name))
	}

	rel := name
	if dir != "." {
		rel = filepath.Join(dir, name)
	}

	if entry.Type()&fs.ModeSymlink != 0 {
		return copySymlinkTarget(ctx, log, srcRoot, dstRoot, rel)
	}

	if !entry.IsDir() && !entry.Type().IsRegular() {
		log.Warn("skip non-regular file in dry-run copy", "path", rel)

		return nil
	}

	if entry.IsDir() {
		if name == ".git" {
			return nil
		}

		if err := dstRoot.MkdirAll(rel, 0o755); err != nil {
			return errors.WithStack(fmt.Errorf("create dir %s: %w", rel, err))
		}

		return copyDirValidated(ctx, log, srcRoot, dstRoot, rel)
	}

	return copyFile(ctx, srcRoot, dstRoot, entry, rel)
}

func copyFile(ctx context.Context, srcRoot string, dstRoot *os.Root, entry fs.DirEntry, rel string) error {
	if err := ctx.Err(); err != nil {
		return errors.WithStack(err)
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

	return nil
}

func copySymlinkTarget(ctx context.Context, log *slog.Logger, srcRoot string, dstRoot *os.Root, rel string) error {
	if err := ctx.Err(); err != nil {
		return errors.WithStack(err)
	}

	srcPath := filepath.Join(srcRoot, filepath.FromSlash(rel))

	target, err := filepath.EvalSymlinks(srcPath)
	if err != nil {
		log.Warn("skip dangling symlink in dry-run copy", "path", rel, "error", err)

		return nil
	}

	info, err := os.Stat(target)
	if err != nil {
		log.Warn("skip unreadable symlink target in dry-run copy", "path", rel, "error", err)

		return nil
	}

	if info.IsDir() {
		if err := dstRoot.MkdirAll(rel, 0o755); err != nil {
			return errors.WithStack(fmt.Errorf("create dir %s: %w", rel, err))
		}

		return copyDirValidated(ctx, log, srcRoot, dstRoot, rel)
	}

	if !info.Mode().IsRegular() {
		log.Warn("skip non-regular symlink target in dry-run copy", "path", rel)

		return nil
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return errors.WithStack(fmt.Errorf("read symlink target %s: %w", rel, err))
	}
	//nolint:gosec // dstRoot confines the write; rel is built from validated entry names.
	if err := dstRoot.WriteFile(rel, data, info.Mode().Perm()); err != nil {
		return errors.WithStack(fmt.Errorf("write file %s: %w", rel, err))
	}

	log.Info("copied symlink target in dry-run copy", "path", rel)

	return nil
}
