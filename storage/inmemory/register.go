package inmemory

import (
	"github.com/crimist/trakx/storage"
)

func init() {
	storage.RegisterDriver(storage.Metadata{
		Name:    "inmemory",
		Creator: NewInMemory,
		PersistanceDrivers: []storage.PersistanceDriver{
			{
				Name:   "file",
				Backup: &FileBackup{},
			},
			{
				Name:   "pg",
				Backup: &PgBackup{},
			},
			{
				Name:   "none",
				Backup: &NoneBackup{},
			},
		},
	})
}
