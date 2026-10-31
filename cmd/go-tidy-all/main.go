// Command go-tidy-all runs go mod tidy for every module in the repository and pins
// each go directive to the lowest version that mingo computes.
package main

import (
	"context"
	standarderrors "errors"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/k1LoW/errors"
	"github.com/spf13/pflag"

	tidyall "github.com/logica-oss/go-tidy-all"
	"github.com/logica-oss/go-tidy-all/pkg/config"
)

// version is injected at build time via ldflags.
//
//nolint:gochecknoglobals // build-time injection requires a package variable.
var version = "dev"

func resolveVersion() string {
	if version != "" && version != "dev" && version != "(devel)" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}

	return version
}

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		if standarderrors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "go-tidy-all: %v\n", err)
		os.Exit(1)
	}

	if cfg.Version {
		if _, err := fmt.Fprintln(os.Stdout, resolveVersion()); err != nil {
			fmt.Fprintf(os.Stderr, "go-tidy-all: print version: %v\n", err)
			os.Exit(1)
		}

		return
	}

	log.Info("starting tidy",
		"root", cfg.Root,
		"dry_run", cfg.DryRun,
		"skip_dirs", cfg.SkipDirs,
		"min_go_version", cfg.MinGoVersion,
	)

	if err := tidyall.Run(context.Background(), log, tidyall.Config{
		Root:         cfg.Root,
		DryRun:       cfg.DryRun,
		SkipDirs:     cfg.SkipDirs,
		MinGoVersion: cfg.MinGoVersion,
	}); err != nil {
		log.Error("tidy failed", "error", err, "stack", errors.StackTraces(err))
		os.Exit(1)
	}
}
