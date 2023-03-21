package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/pkg/errors"
)

const (
	testHomePath   = "test_home"
	testConfigPath = testHomePath + "/.config/trakx/"
	testCachePath  = testHomePath + "/.cache/trakx/"
)

func TestPathsAndGenerate(t *testing.T) {
	env := "HOME"
	switch runtime.GOOS {
	case "windows":
		env = "USERPROFILE"
	case "plan9":
		env = "home"
	}

	trueHomePath := os.Getenv(env)
	if err := os.Setenv(env, testHomePath); err != nil {
		t.Fatal(errors.Wrap(err, "Failed to set home env var"))
	}

	defer func() {
		if err := os.Setenv(env, trueHomePath); err != nil {
			t.Log(errors.Wrap(err, "failed to restore home env var"))
		}

		if err := os.RemoveAll("./" + testHomePath + "/"); err != nil {
			t.Log(errors.Wrap(err, "failed to remove test directory"))
		}
	}()

	initDirectories()
	generateConfig()

	if configPath != testConfigPath {
		t.Error("Invalid config directory set")
	}
	if CachePath != testCachePath {
		t.Error("Invalid cache directory set")
	}

	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Error("failed to create config directory")
	}
	if _, err := os.Stat(testCachePath); os.IsNotExist(err) {
		t.Error("failed to create cache directory")
	}

	if _, err := os.Stat(testConfigPath + "trakx.yaml"); err != nil {
		if os.IsNotExist(err) {
			t.Error("failed to create configuration file")
		}
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
