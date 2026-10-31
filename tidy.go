package tidyall

import (
	"context"
	"log/slog"
	"os"
	"os/exec"

	"github.com/k1LoW/errors"
)

func tidy(ctx context.Context, log *slog.Logger, dir string, dryRun bool) error {
	args := []string{"mod", "tidy"}
	if dryRun {
		args = append(args, "-diff")
	}

	log.Info("go mod tidy", "dir", dir, "dry_run", dryRun)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if dryRun {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
				log.Info("go mod tidy would change files", "dir", dir)

				return nil
			}
		}

		return errors.WithStack(err)
	}

	return nil
}
