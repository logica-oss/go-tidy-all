package tidyall

import (
	"context"
	"log/slog"
	"os"
	"os/exec"

	"github.com/k1LoW/errors"
)

func tidy(ctx context.Context, log *slog.Logger, dir string, dryRun bool) error {
	if dryRun {
		log.Info("dry-run: would run go mod tidy", "dir", dir)

		return nil
	}

	log.Info("go mod tidy", "dir", dir)

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return errors.WithStack(cmd.Run())
}
