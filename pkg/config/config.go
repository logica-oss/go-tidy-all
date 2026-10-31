// Package config loads go-tidy-all settings.
package config

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds go-tidy-all settings.
type Config struct {
	Root         string   `mapstructure:"-"` // Stays out of the config file to keep relative resolution unambiguous.
	Version      bool     `mapstructure:"-"` // Stays out of the config file because it is a CLI-only flag.
	DryRun       bool     `mapstructure:"dry-run"`
	SkipDirs     []string `mapstructure:"skip-dirs"`
	MinGoVersion string   `mapstructure:"min-go-version"`
}

const (
	keyRoot         = "root"
	keyVersion      = "version"
	keyDryRun       = "dry-run"
	keySkipDirs     = "skip-dirs"
	keyMinGoVersion = "min-go-version"
	keyConfig       = "config"
)

func setDefaults(v *viper.Viper) {
	v.SetDefault(keyDryRun, false)
	v.SetDefault(keySkipDirs, []string{})
	v.SetDefault(keyMinGoVersion, "")
}

func flags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("go-tidy-all", pflag.ContinueOnError)

	flagSet.StringP(keyConfig, "c", "", "config file path (default: <root>/.tidy-all.yaml or .tidy-all.json)")
	flagSet.String(keyRoot, "", "repository root (default: $GITHUB_WORKSPACE or git top level)")
	flagSet.BoolP(keyVersion, "v", false, "print version and exit")
	flagSet.BoolP(keyDryRun, "n", false, "print planned changes without writing")
	flagSet.StringSlice(keySkipDirs, []string{}, "skip modules under these root-relative directories")
	flagSet.String(keyMinGoVersion, "", "pin the go directive to at least this version (e.g. 1.24.0)")

	flagSet.Usage = func() {
		_, _ = fmt.Fprintf(flagSet.Output(), "Usage: go-tidy-all [options]\n\nRun go mod tidy for every module and pin each go directive.\n\nOptions:\n")

		flagSet.PrintDefaults()

		_, _ = fmt.Fprintf(flagSet.Output(),
			"\nConfig file: .tidy-all.yaml (preferred, supports comments) or .tidy-all.json in the repository root.\n"+
				"Environment: TIDYALL_* variables (e.g. TIDYALL_DRY_RUN, TIDYALL_SKIP_DIRS).\n"+
				"Precedence: flag > env > config file > default.\n")
	}

	return flagSet
}
