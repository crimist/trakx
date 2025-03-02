package config

import (
	"embed"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

//go:embed embedded/*
var embeddedFS embed.FS

func writeEmbeddedConfig(path string) error {
	syscall.Umask(0)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		configData, err := embeddedFS.ReadFile("embedded/trakx.yaml")
		if err != nil {
			return errors.Wrap(err, "failed to read embedded FS")
		}

		if err = os.MkdirAll(filepath.Dir(path), folderPerm); err != nil {
			zap.L().Warn("failed to create config directory", zap.Error(err))
		}

		if err = os.WriteFile(path, configData, filePerm); err != nil {
			return errors.Wrap(err, "failed to write configuration file to "+path)
		}
	} else if err != nil {
		return errors.Wrap(err, "failed to stat config file "+path)
	}

	return nil
}
