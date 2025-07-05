/*
Config holds configuration information for trakx.
*/
package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkyr/fig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TrackerModeEnabled  = "enabled"  // http tracker enabled
	TrackerModeInfo     = "info"     // http information server, no tracker
	TrackerModeDisabled = "disabled" // http disabled

	// defaultFolderPermission holds the default permission mask for folders
	defaultFolderPermission = 0700
	// defaultFilePermission holds the default permission mask for files
	defaultFilePermission = 0644
)

var (
	loggerAtom     zap.AtomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	configPathFlag                 = flag.String("config", "", "optional path to configuration file")
)

func Load() (*Configuration, error) {
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), zapcore.Lock(os.Stdout), loggerAtom))
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
	}

	var conf Configuration

	if err := fig.Load(&conf,
		fig.Dirs(filepath.Dir(configPath)),
		fig.File(filepath.Base(configPath)),
		fig.UseEnv("trakx"),
	); err != nil {
		return nil, errors.Wrap(err, "failed to load configuration")
	}

	switch strings.ToLower(string(conf.LogLevel)) {
	case "debug":
		loggerAtom.SetLevel(zap.DebugLevel)
		zap.L().Debug("Debug loglevel set, debug panics enabled")
	case "info":
		loggerAtom.SetLevel(zap.InfoLevel)
	case "warn":
		loggerAtom.SetLevel(zap.WarnLevel)
	case "error":
		loggerAtom.SetLevel(zap.ErrorLevel)
	case "fatal":
		loggerAtom.SetLevel(zap.FatalLevel)
	default:
		zap.L().Warn("Invalid log level was specified, defaulting to warn")
		loggerAtom.SetLevel(zap.WarnLevel)
	}

	zap.L().Debug("Configuration loaded", zap.String("config_path", configPath))

	if conf.Cache == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user cache directory")
		}
		defaultCacheDir := filepath.Join(cacheDir, "trakx")
		conf.Cache = defaultCacheDir
	}

	// remove after config refactor
	conf.HTTP.Mode = strings.ToLower(conf.HTTP.Mode)

	if err = conf.validate(); err != nil {
		return nil, errors.Wrap(err, "configuration validation failed")
	}

	return &conf, nil
}
