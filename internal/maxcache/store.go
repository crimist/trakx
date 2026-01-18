package maxcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	fileSuffix = ".json"
	dirName    = "maximums"
)

var ErrInvalidKey = errors.New("invalid maximums key")

type Store struct {
	dir         string
	decayFactor float64
	mu          sync.Mutex
	values      map[string]int
}

type record struct {
	Max int `json:"max"`
}

type Options struct {
	DecayHalfLifeUpdates int
}

// New creates a maximums store rooted at cacheDir/maximums.
// It loads existing maxima if present.
func New(cacheDir string, opts Options) (*Store, error) {
	dir := filepath.Join(cacheDir, dirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	decayFactor := 1.0
	if opts.DecayHalfLifeUpdates > 0 {
		decayFactor = math.Pow(0.5, 1.0/float64(opts.DecayHalfLifeUpdates))
	}

	store := &Store{
		dir:         dir,
		decayFactor: decayFactor,
		values:      make(map[string]int),
	}

	if err := store.load(); err != nil {
		return store, err
	}

	return store, nil
}

// Get returns the cached maximum for key, or 0 if missing/invalid.
func (s *Store) Get(key string) int {
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

	current := s.values[key]
	next := current
	if s.decayFactor < 1 && current > 0 {
		next = int(math.Floor(float64(current) * s.decayFactor))
	}
	if value > next {
		next = value
	}
	if next == current {
		return nil
	}

	s.values[key] = next
	return s.write(key, next)
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

		var rec record
		if err := json.Unmarshal(data, &rec); err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		s.values[key] = rec.Max
	}

	return errs
}

func (s *Store) write(key string, value int) error {
	path := filepath.Join(s.dir, key+fileSuffix)

	tmp, err := os.CreateTemp(s.dir, "."+key+".tmp-*")
	if err != nil {
		return err
	}
	defer tmp.Close()

	tmpName := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpName)
		}
	}()

	data, err := json.Marshal(record{Max: value})
	if err != nil {
		return err
	}

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
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

func validateKey(key string) error {
	if key == "" {
		return ErrInvalidKey
	}
	if strings.ContainsFunc(key, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_')
	}) {
		return ErrInvalidKey
	}
	return nil
}
