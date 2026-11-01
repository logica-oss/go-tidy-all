package tidyall

import (
	"bytes"
	"context"
	"io"
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

	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if dryRun {
			if eErr, ok := errors.AsType[*exec.ExitError](err); ok && eErr.ExitCode() == 1 && stdout.Len() > 0 {
				log.Info("go mod tidy would change files", "dir", dir, "diff", stdout.String())

				return nil
			}
		}

		return errors.WithStack(err)
	}

	return nil
}
