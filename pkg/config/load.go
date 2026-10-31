package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/k1LoW/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Load parses flags, environment, and config file into a Config.
func Load(args []string) (Config, error) {
	flagSet := flags()
	if err := flagSet.Parse(args); err != nil {
		return Config{}, errors.WithStack(fmt.Errorf("parse flags: %w", err))
	}

	if version, _ := flagSet.GetBool(keyVersion); version {
		return Config{Version: true}, nil
	}

	explicitRoot := resolveExplicitRoot(flagSet)

	conf := newViper(flagSet, explicitRoot)
	if err := readConfig(conf, flagSet); err != nil {
		return Config{}, err
	}

	var loaded Config
	if err := conf.Unmarshal(&loaded); err != nil {
		return Config{}, errors.WithStack(fmt.Errorf("decode config: %w", err))
	}

	root, err := resolveRoot(explicitRoot)
	if err != nil {
		return Config{}, err
	}
	loaded.Root = root

	if err := Validate(loaded); err != nil {
		return Config{}, err
	}

	return loaded, nil
}

func resolveExplicitRoot(flagSet *pflag.FlagSet) string {
	rootFlag, _ := flagSet.GetString(keyRoot)

	return firstNonEmpty(rootFlag, os.Getenv("TIDYALL_ROOT"))
}

func newViper(flagSet *pflag.FlagSet, explicitRoot string) *viper.Viper {
	conf := viper.New()
	conf.SetEnvPrefix("TIDYALL")
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	conf.AutomaticEnv()

	setDefaults(conf)

	rootExplicit := explicitRoot != ""

	searchRoot := explicitRoot
	if searchRoot == "" {
		searchRoot = detectRoot()
	}

	configFile, _ := flagSet.GetString(keyConfig)
	if configFile != "" {
		conf.SetConfigFile(configFile)

		return conf
	}

	conf.SetConfigName(".tidy-all")

	if searchRoot != "" {
		conf.AddConfigPath(searchRoot)
	}
	if !rootExplicit {
		conf.AddConfigPath(".")
	}

	if explicit := preferredConfigFile(searchRoot, rootExplicit); explicit != "" {
		conf.SetConfigFile(explicit)
	}

	return conf
}

func readConfig(conf *viper.Viper, flagSet *pflag.FlagSet) error {
	var notFound viper.ConfigFileNotFoundError
	if err := conf.ReadInConfig(); err != nil {
		if !errors.As(err, &notFound) {
			return errors.WithStack(fmt.Errorf("read config: %w", err))
		}
	}

	if err := conf.BindPFlags(flagSet); err != nil {
		return errors.WithStack(fmt.Errorf("bind flags: %w", err))
	}

	if conf.InConfig(keyRoot) {
		return errors.WithStack(fmt.Errorf("unsupported config key %q: set the root with --root or TIDYALL_ROOT", keyRoot))
	}

	return nil
}
