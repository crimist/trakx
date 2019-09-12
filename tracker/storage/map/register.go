package gomap

import (
	"github.com/syc0x00/trakx/tracker/storage"
)

func init() {
	storage.Register(storage.DatabaseInfo{
		Name: "gomap",
		DB:   &Memory{},
		Backups: []storage.BackupInfo{
			{
				Name: "file",
				Back: &FileBackup{},
			},
			{
				Name: "pg",
				Back: &PgBackup{},
			},
		},
	})
}
