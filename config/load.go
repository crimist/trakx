/*
Config holds configuration information for trakx.
*/
package config

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
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

	defaultConfigPath = "~/.config/trakx.yaml"

	// folderPerm holds the default permission mask for folders
	folderPerm = 0700
	// filePerm holds the default permission mask for files
	filePerm = 0644
)

var (
	loggerAtom zap.AtomicLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
)

func Load() (*Configuration, error) {
	var configPath = flag.String("conf", defaultConfigPath, "path to configuration file")
	flag.Parse()

	if (*configPath)[0] == '~' {
		isDefault := false
		if *configPath == defaultConfigPath {
			isDefault = true
		}

		home, err := os.UserHomeDir()
		if err != nil {
			panic("failed to get users home directory: " + err.Error())
		}
		*configPath = filepath.Join(home, (*configPath)[1:])

		if isDefault {
			if err := writeEmbeddedConfig(*configPath); err != nil {
				return nil, err
			}
		}
	}

	conf := new(Configuration)

	if err := fig.Load(conf,
		fig.File(filepath.Base(*configPath)),
		fig.UseEnv("trakx"),
		fig.Dirs(filepath.Dir(*configPath)),
	); err != nil {
		return nil, errors.Wrap(err, "fig failed to load config")
	}

	if conf.CachePath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			panic("failed to get users home directory: " + err.Error())
		}
		conf.CachePath = filepath.Join(home, conf.CachePath[1:])
	}

	stat, err := os.Stat(conf.CachePath)
	if os.IsNotExist(err) {
		zap.L().Info("configuration CachePath directory does not exist, creating it", zap.String("CachePath", conf.CachePath))
		if err = os.MkdirAll(conf.CachePath, folderPerm); err != nil {
			zap.L().Error("failed to create CachePath directory", zap.Error(err))
		}
	} else if err != nil {
		zap.L().Warn("error on stat CachePath", zap.Error(err))
	} else {
		if !stat.IsDir() {
			zap.L().Warn("configuration CachePath is file, hope you're running an appengine build")
		}
	}

	cfg := zap.NewDevelopmentConfig()

	conf.LogLevel = LogLevel(strings.ToLower(string(conf.LogLevel)))
	conf.HTTP.Mode = strings.ToLower(conf.HTTP.Mode)

	if conf.LogLevel.Debug() {
		cfg.Development = true
	} else {
		cfg.Development = false
	}

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.Lock(os.Stdout), loggerAtom))
	conf.setLogLevel(conf.LogLevel)
	zap.ReplaceGlobals(logger)

	// TODO: create docs for this
	if strings.HasPrefix(conf.DB.Backup.Path, "ENV:") {
		conf.DB.Backup.Path = os.Getenv(strings.TrimPrefix(conf.DB.Backup.Path, "ENV:"))
	}

	// If $PORT environment variable is set override port
	// this is for app engines
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		port, err := strconv.Atoi(portEnv)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse $PORT environment variable")
		}

		zap.L().Info("PORT environment variable detected - writing to config", zap.Int("port", port))
		conf.HTTP.Port = port
	}

	return conf, nil
}
