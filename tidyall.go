// Package tidyall runs go mod tidy for every module and pins each go directive
// to the lowest version that mingo computes.
package tidyall

import (
	"context"
	"log/slog"

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
		if err := tidy(ctx, log, dir, cfg.DryRun); err != nil {
			return err
		}

		if err := pinGoDirective(log, dir, cfg.DryRun, cfg.MinGoVersion); err != nil {
			return err
		}
	}

	return nil
}
