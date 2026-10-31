// Command go-tidy-all runs go mod tidy for every module in the repository and pins
// each go directive to the lowest version that mingo computes.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/k1LoW/errors"

	tidyall "github.com/logica-oss/go-tidy-all"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()

	root, err := os.Getwd()
	if err != nil {
		log.Error("tidy failed", "error", err)
		os.Exit(1)
	}

	if err := tidyall.Run(ctx, log, tidyall.Config{Root: root}); err != nil {
		log.Error("tidy failed", "error", err, "stack", errors.StackTraces(err))
		os.Exit(1)
	}
}
