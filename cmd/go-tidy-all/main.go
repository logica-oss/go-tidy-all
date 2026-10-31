// Command go-tidy-all runs go mod tidy for every module in the repository and pins
// each go directive to the lowest version that mingo computes.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/spf13/pflag"

	tidyall "github.com/logica-oss/go-tidy-all"
)

// version is injected at build time via ldflags.
//
//nolint:gochecknoglobals // build-time injection requires a package variable.
var version = "dev"

const exitUsage = 2

func parseFlags(args []string) (tidyall.Config, bool, error) {
	var (
		cfg         tidyall.Config
		skipDirs    string
		showVersion bool
	)

	flagSet := pflag.NewFlagSet("go-tidy-all", pflag.ContinueOnError)
	flagSet.StringVar(&cfg.Root, "root", "", "repository root (default: current directory)")
	flagSet.BoolVarP(&cfg.DryRun, "dry-run", "n", false, "print planned changes without writing")
	flagSet.StringVar(&skipDirs, "skip-dirs", "", "comma-separated root-relative directories to skip")
	flagSet.StringVar(&cfg.MinGoVersion, "min-go-version", "", "pin the go directive to at least this version (e.g. 1.24.0)")
	flagSet.BoolVarP(&showVersion, "version", "v", false, "print version and exit")

	flagSet.Usage = func() {
		_, _ = fmt.Fprintf(flagSet.Output(),
			"Usage: go-tidy-all [options]\n\n"+
				"Run go mod tidy for every module and pin each go directive.\n\n"+
				"Options:\n")

		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return tidyall.Config{}, false, pflag.ErrHelp
		}

		return tidyall.Config{}, false, errors.WithStack(err)
	}

	if skipDirs != "" {
		cfg.SkipDirs = strings.Split(skipDirs, ",")
	}

	if cfg.Root == "" {
		root, err := os.Getwd()
		if err != nil {
			return tidyall.Config{}, false, errors.WithStack(fmt.Errorf("get working directory: %w", err))
		}

		cfg.Root = root
	}

	return cfg, showVersion, nil
}

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

	cfg, showVersion, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "go-tidy-all: %v\n", err)
		os.Exit(exitUsage)
	}

	if showVersion {
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

	if err := tidyall.Run(context.Background(), log, cfg); err != nil {
		log.Error("tidy failed", "error", err, "stack", errors.StackTraces(err))
		os.Exit(1)
	}
}
