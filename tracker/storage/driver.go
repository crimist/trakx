package storage

import "sync"

type DatabaseDriver struct {
	db      Database
	backups map[string]Backup
}

var drivers map[string]DatabaseDriver

type BackupInfo struct {
	Name string
	Back Backup
}

type DatabaseInfo struct {
	Name    string
	DB      Database
	Backups []BackupInfo
}

var initDrivers sync.Once

func Register(dbnfo DatabaseInfo) {
	initDrivers.Do(func() {
		drivers = make(map[string]DatabaseDriver)
	})

	drivers[dbnfo.Name] = DatabaseDriver{
		db:      dbnfo.DB,
		backups: make(map[string]Backup),
	}

	for _, backnfo := range dbnfo.Backups {
		drivers[dbnfo.Name].backups[backnfo.Name] = backnfo.Back
	}
}
