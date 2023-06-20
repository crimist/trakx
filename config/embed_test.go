package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/pkg/errors"
)

func TestWriteEmbeddedConfig(t *testing.T) {
	const testHomeDir = "test_home"
	const testConfigPath = testHomeDir + "/.config/trakx.yaml"
	const testCachePath = testHomeDir + "/.cache/trakx/"

	homeEnv := "HOME"
	switch runtime.GOOS {
	case "windows":
		homeEnv = "USERPROFILE"
	case "plan9":
		homeEnv = "home"
	}

	originalHomePath := os.Getenv(homeEnv)
	if err := os.Setenv(homeEnv, testHomeDir); err != nil {
		t.Fatal(errors.Wrap(err, "Failed to set home env var"))
	}

	defer func() {
		if err := os.Setenv(homeEnv, originalHomePath); err != nil {
			t.Log(errors.Wrap(err, "failed to restore original home environment variable"))
		}

		if err := os.RemoveAll(testHomeDir + "/"); err != nil {
			t.Log(errors.Wrap(err, "failed to remove test home directory"))
		}
	}()

	_, err := Load()
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

// Assumes that index.html and dmca.html exist
func TestGenerateEmbeddedCache(t *testing.T) {
	cache, err := GenerateEmbeddedCache()

	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to generate embedded cahce"))
	}

	indexBytes, err := os.ReadFile("embedded/index.html")
	if err != nil {
		t.Error(errors.Wrap(err, "failed to read index.html"))
	}
	index := stripNewlineTabs(string(indexBytes))

	if cache["/index.html"] != index {
		t.Error("index.html file and cache differ")
	}

	var cases = []struct {
		filename string
	}{
		{"index.html"},
		{"dmca.html"},
	}

	for _, c := range cases {
		t.Run(c.filename, func(t *testing.T) {
			bytes, err := os.ReadFile("embedded/" + c.filename)
			if err != nil {
				t.Error(errors.Wrap(err, "failed to read index.html"))
			}

			if cache["/"+c.filename] != stripNewlineTabs(string(bytes)) {
				t.Errorf("%v file and cache differ", c.filename)
			}
		})
	}
}
