/*
Config holds configuration information for trakx.
*/
package config

import "go.uber.org/zap"

const (
	nofileIgnore        = 0
	TrackerModeEnabled  = "enabled"  // http tracker enabled
	TrackerModeInfo     = "info"     // http information server, no tracker
	TrackerModeDisabled = "disabled" // http disabled
)

var (
	// Global instance of config and logger
	Conf   *Config
	Logger *zap.Logger

	loggerAtom zap.AtomicLevel
)

func init() {
	// create temporary logger
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic("failed to create logger")
	}

	// initialize directories
	initDirectories()

	generateConfig()

	// load config
	Conf, err = Load()
	if err != nil {
		Logger.Error("Failed to load a config", zap.Any("config", Conf), zap.Error(err))
	} else {
		Logger.Debug("Loaded config", zap.Any("config", Conf))
	}

	Logger.Debug("initialized paths", zap.String("config", configPath), zap.String("cache", CachePath))
}
