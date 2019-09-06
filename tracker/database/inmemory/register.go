package inmemory

import "github.com/syc0x00/trakx/tracker/database"

func init() {
	database.Register(database.DatabaseInfo{
		Name: "mem",
		DB:   &Memory{},
		Backups: []database.BackupInfo{
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
