package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteEmbeddedConfig(t *testing.T) {
	testHomeDir := t.TempDir()

	homeEnv := "HOME"
	switch runtime.GOOS {
	case "windows":
		homeEnv = "USERPROFILE"
	case "plan9":
		homeEnv = "home"
	}

	t.Setenv(homeEnv, testHomeDir)

	// macOS: ~/Library/Application Support and ~/Library/Caches
	// Linux: $XDG_CONFIG_HOME or ~/.config, $XDG_CACHE_HOME or ~/.cache
	var testConfigPath, testCachePath string
	switch runtime.GOOS {
	case "darwin":
		testConfigPath = filepath.Join(testHomeDir, "Library", "Application Support", "trakx", "trakx.yaml")
		testCachePath = filepath.Join(testHomeDir, "Library", "Caches", "trakx")
	default:
		xdgConfigHome := filepath.Join(testHomeDir, ".config")
		xdgCacheHome := filepath.Join(testHomeDir, ".cache")
		testConfigPath = filepath.Join(xdgConfigHome, "trakx", "trakx.yaml")
		testCachePath = filepath.Join(xdgCacheHome, "trakx")
	}

	_, err := Load(LoadOptions{})
	if err != nil {
		t.Fatal("failed to load config")
	}

	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Error("load failed to write configuration to default path")
	}
	if _, err := os.Stat(testCachePath); os.IsNotExist(err) {
		t.Error("load failed to create cache directory")
	}
}
