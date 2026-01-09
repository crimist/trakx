package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWriteEmbeddedConfig(t *testing.T) {
	testHomeDir := t.TempDir()
	xdgConfigHome := filepath.Join(testHomeDir, "config")
	xdgCacheHome := filepath.Join(testHomeDir, "cache")
	testConfigPath := filepath.Join(xdgConfigHome, "trakx", "trakx.yaml")
	testCachePath := filepath.Join(xdgCacheHome, "trakx")

	homeEnv := "HOME"
	switch runtime.GOOS {
	case "windows":
		homeEnv = "USERPROFILE"
	case "plan9":
		homeEnv = "home"
	}

	t.Setenv(homeEnv, testHomeDir)
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	t.Setenv("XDG_CACHE_HOME", xdgCacheHome)

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
