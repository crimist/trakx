package database

// TOOD

const (
	LoadFail        = -1
	LoadUnknownSize = 0
)

type Database interface {
	Check() bool
	Save()
	Drop()
	Trim()
	Backup() Backup
}

type Backup interface {
	Save() int
	Load() int
}
