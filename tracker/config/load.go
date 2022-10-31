package config

import (
	"os"

	"github.com/kkyr/fig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Load attempts to load the config from the disk or environment.
// The config file must be named "trakx.yaml".
// Load searches for the config file in ".", "~/.config/trakx" in order.
// Environment variables overwrite file configuration, see ./embedded/trakx.yaml for examples.
// This function is automatically called when the config package is imported.
func Load() (*Configuration, error) {
	conf := new(Configuration)

	home, err := os.UserHomeDir()
	if err != nil {
		Logger.Error("Failed to get user home dir, attempting to continue config load", zap.Error(err))
	}

	err = fig.Load(conf,
		fig.File("trakx.yaml"),
		fig.UseEnv("trakx"),
		fig.Dirs(".", home+"/.config/trakx"),
	)
	if err != nil {
		return nil, errors.Wrap(err, "fig failed to load config")
	}

	return conf, conf.Parse()
}
