package storage

type DatabaseDriver struct {
	database      Database
	backupDrivers map[string]Persistance
}

type PersistanceDriver struct {
	Name   string
	Backup Persistance
}

type DatabaseMetadata struct {
	Name               string
	Database           Database
	PersistanceDrivers []PersistanceDriver
}

var databaseDrivers map[string]DatabaseDriver

func init() {
	databaseDrivers = make(map[string]DatabaseDriver)
}

// RegisterDriver registers a database driver and its backup drivers into the driver list.
func RegisterDriver(meta DatabaseMetadata) {
	databaseDrivers[meta.Name] = DatabaseDriver{
		database:      meta.Database,
		backupDrivers: make(map[string]Persistance),
	}

	for _, persistanceDriver := range meta.PersistanceDrivers {
		databaseDrivers[meta.Name].backupDrivers[persistanceDriver.Name] = persistanceDriver.Backup
	}
}
