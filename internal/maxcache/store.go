package maxcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	fileSuffix = ".maximum"
	dirName    = "maximums"

	UpdateFrequency = time.Minute
)

var ErrInvalidKey = errors.New("invalid maximums key")

type Store struct {
	dir    string
	mu     sync.Mutex
	values map[string]int
}

type record struct {
	Max int `json:"max"`
}

// New creates a maximums store rooted at cacheDir/maximums.
// It loads existing maxima if present.
func New(cacheDir string) (*Store, error) {
	dir := filepath.Join(cacheDir, dirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	store := &Store{
		dir:    dir,
		values: make(map[string]int),
	}

	if err := store.load(); err != nil {
		return store, err
	}

	return store, nil
}

// ReadMax returns the cached maximum for key, or 0 if missing/invalid.
func (s *Store) ReadMax(key string) int {
	if err := validateKey(key); err != nil {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.values[key]
}

// Update records value if it exceeds the current maximum for key.
// It writes immediately to disk on change.
func (s *Store) Update(key string, value int) error {
	if err := validateKey(key); err != nil {
		return err
	}
	if value < 0 {
		return fmt.Errorf("invalid maximum value: %d", value)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if value <= s.values[key] {
		return nil
	}

	s.values[key] = value
	return s.writeLocked(key, value)
}

func (s *Store) load() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}

	var errs error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, fileSuffix) {
			continue
		}

		key := strings.TrimSuffix(name, fileSuffix)
		if err := validateKey(key); err != nil {
			continue
		}

		path := filepath.Join(s.dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		value, err := parseValue(data)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if value > s.values[key] {
			s.values[key] = value
		}
	}

	return errs
}

func (s *Store) writeLocked(key string, value int) error {
	path := filepath.Join(s.dir, key+fileSuffix)

	tmp, err := os.CreateTemp(s.dir, "."+key+".tmp-*")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpName)
		}
	}()

	data, err := json.Marshal(record{Max: value})
	if err != nil {
		_ = tmp.Close()
		return err
	}

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}

	removeTmp = false
	return nil
}

func parseValue(data []byte) (int, error) {
	var rec record
	if err := json.Unmarshal(data, &rec); err == nil {
		return rec.Max, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return value, nil
}

func validateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	for _, r := range key {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		switch r {
		case '.', '-', '_':
			continue
		default:
			return ErrInvalidKey
		}
	}

	return nil
}
