package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trakx", "trakx.yaml")

	if err := installDefaultConfig(path); err != nil {
		t.Fatalf("installDefaultConfig failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	embedded, err := embeddedFS.ReadFile("embedded/trakx.yaml")
	if err != nil {
		t.Fatalf("failed to read embedded config: %v", err)
	}

	if string(content) != string(embedded) {
		t.Error("installed config doesn't match embedded config")
	}
}

func TestInstallDefaultConfigSkipsExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trakx.yaml")
	existing := []byte("existing: config")

	if err := os.WriteFile(path, existing, 0644); err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}

	if err := installDefaultConfig(path); err != nil {
		t.Fatalf("installDefaultConfig failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if string(content) != string(existing) {
		t.Error("installDefaultConfig overwrote existing config")
	}
}
