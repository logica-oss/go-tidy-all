package tidyall

import (
	"bytes"
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

	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if dryRun {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && stdout.Len() > 0 {
				log.Info("go mod tidy would change files", "dir", dir, "diff", stdout.String())

				return nil
			}
		}

		_, _ = os.Stdout.Write(stdout.Bytes())
		_, _ = os.Stderr.Write(stderr.Bytes())

		return errors.WithStack(err)
	}

	if _, err := os.Stdout.Write(stdout.Bytes()); err != nil {
		return errors.WithStack(err)
	}
	if _, err := os.Stderr.Write(stderr.Bytes()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
