package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
	"github.com/pkg/errors"
)

// NoneBackup is an empty backup driver. It doesn't back anything up.
type NoneBackup struct{}

func (bck *NoneBackup) Init(db storage.Database) error {
	memdb := db.(*Memory)
	if memdb == nil {
		return errors.New("database is not type `Memory` for `NoneBackup`")
	}

	memdb.make()
	return nil
}

func (bck NoneBackup) Save() error          { return nil }
func (bck NoneBackup) Load() error          { return nil }
func (bck NoneBackup) trim() (int64, error) { return 0, nil }
