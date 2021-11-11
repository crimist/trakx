package storage

type DatabaseDriver struct {
	db      Database
	backups map[string]Backup
}

type BackupInfo struct {
	Name string
	Back Backup
}

type DatabaseInfo struct {
	Name    string
	DB      Database
	Backups []BackupInfo
}

var drivers map[string]DatabaseDriver

func init() { drivers = make(map[string]DatabaseDriver) }

// Register appends `databaseinfo` into the map of available drivers.
func Register(dbnfo DatabaseInfo) {
	drivers[dbnfo.Name] = DatabaseDriver{
		db:      dbnfo.DB,
		backups: make(map[string]Backup),
	}

	for _, backnfo := range dbnfo.Backups {
		drivers[dbnfo.Name].backups[backnfo.Name] = backnfo.Back
	}
}
