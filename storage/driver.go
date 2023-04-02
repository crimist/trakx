package storage

type creator func(Config) (Database, error)

type Metadata struct {
	Name    string
	Creator func(Config) (Database, error)
}

var drivers = make(map[string]creator)

// RegisterDriver registers a database driver and its backup drivers into the driver list.
func RegisterDriver(metadata Metadata) {
	drivers[metadata.Name] = metadata.Creator
}
