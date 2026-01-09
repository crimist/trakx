/*
Config holds configuration information for trakx.
*/
package config

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kkyr/fig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerAtom     zap.AtomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	configPathFlag                 = flag.String("config", "", "optional path to configuration file")
)

func Load() (*Configuration, error) {
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.Lock(os.Stderr), loggerAtom))
	zap.ReplaceGlobals(logger)

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user config directory")
	}
	defaultConfigPath := filepath.Join(configDir, "trakx", "trakx.yaml")
	installDefaultConfig(defaultConfigPath)

	configPath := defaultConfigPath
	if *configPathFlag != "" {
		configPath = *configPathFlag
		zap.L().Debug("Using default config path", zap.String("path", configPath))
	}

	var conf Configuration

	if err := fig.Load(&conf,
		fig.Dirs(filepath.Dir(configPath)),
		fig.File(filepath.Base(configPath)),
		fig.UseEnv("trakx"),
	); err != nil {
		return nil, errors.Wrap(err, "failed to load configuration")
	}

	err = loggerAtom.UnmarshalText([]byte(conf.LogLevel))
	if err != nil {
		return nil, errors.Wrap(err, "Invalid log level")
	}

	zap.L().Debug("Configuration loaded", zap.String("path", configPath))

	if conf.Cache == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user cache directory")
		}
		defaultCacheDir := filepath.Join(cacheDir, "trakx")
		conf.Cache = defaultCacheDir
		zap.L().Debug("Set default cache", zap.String("cache", conf.Cache))
	}

	if conf.DB.Expiry == 0 {
		conf.DB.Expiry = conf.Announce.Base + conf.Announce.Fuzz + 5*time.Minute
		zap.L().Debug("Calculated DB expiry", zap.Duration("expiry", conf.DB.Expiry))
	}

	if conf.DB.Backup.Path == "" {
		conf.DB.Backup.Path = filepath.Join(conf.Cache, "db")
		zap.L().Debug("Set default DB backup path", zap.String("path", conf.DB.Backup.Path))
	}

	// TODO: needs benchmarking to find ideal number of routines
	if conf.HTTP.Routines == 0 {
		conf.HTTP.Routines = runtime.NumCPU() * 2
	}
	if conf.UDP.Routines == 0 {
		conf.UDP.Routines = runtime.NumCPU() * 2
	}

	if err = conf.validate(); err != nil {
		return nil, errors.Wrap(err, "configuration validation failed")
	}

	return &conf, nil
}
