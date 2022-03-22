package config

import (
	"os"
	"strconv"

	"github.com/kkyr/fig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Load attempts to load the config from the disk or environment.
// The config file must be named "trakx.yaml".
// Load searches for the config file in ".", "~/.config/trakx" in order.
// Environment variables overwrite file configuration, see trakx.yaml in ./embedded for examples.
// This function is called when the config package is imported.
func Load() (*Config, error) {
	conf := new(Config)

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
		return nil, errors.Wrap(err, "fig failed to load a config")
	}

	// If $PORT var set override port for appengines (like heroku)
	if appenginePort := os.Getenv("PORT"); appenginePort != "" {
		appPort, err := strconv.Atoi(appenginePort)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse $PORT env variable (not an int)")
		}

		Logger.Info("PORT env variable detected. Overriding config...", zap.Int("$PORT", appPort))
		conf.Tracker.HTTP.Port = appPort
	}

	return conf, conf.Update()
}
