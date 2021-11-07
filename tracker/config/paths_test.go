package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/pkg/errors"
)

const (
	testHomeVar   = "TEST"
	testConfigDir = testHomeVar + "/.config/trakx/"
	testCacheDir  = testHomeVar + "/.cache/trakx/"
)

func TestInitPaths(t *testing.T) {
	env := "HOME"
	switch runtime.GOOS {
	case "windows":
		env = "USERPROFILE"
	case "plan9":
		env = "home"
	}

	// setup testing home var
	realHome := os.Getenv(env)
	err := os.Setenv(env, testHomeVar)
	if err != nil {
		t.Fatal(errors.Wrap(err, "Failed to set home env var"))
	}

	// check directory are correctly set
	initPaths()

	if ConfigDir != testConfigDir {
		t.Fatal("Invalid config directory set")
	}
	if CacheDir != testCacheDir {
		t.Fatal("Invalid cache directory set")
	}

	if _, err := os.Stat(testConfigDir); os.IsNotExist(err) {
		t.Fatal("failed to create config directory")
	}
	if _, err := os.Stat(testCacheDir); os.IsNotExist(err) {
		t.Fatal("failed to create cache directory")
	}

	// restore real home variable
	err = os.Setenv(env, realHome)
	if err != nil {
		t.Log(errors.Wrap(err, "failed to restore home env var"))
	}

	// clean up directories
	err = os.RemoveAll("./" + testHomeVar + "/")
	if err != nil {
		t.Log(errors.Wrap(err, "failed to remove test directory"))
	}
}