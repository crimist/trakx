package database

type PersistanceStrategy interface {
	write(db *Database, address string) error
	read(db *Database, address string) error
}
