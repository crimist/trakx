package gomap

import (
	"github.com/crimist/trakx/storage"
)

func init() {
	storage.RegisterDriver(storage.DatabaseMetadata{
		Name:     "memory",
		Database: &Memory{},
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
