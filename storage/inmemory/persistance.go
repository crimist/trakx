package inmemory

type PersistanceStrategy interface {
	write(db *InMemory, address string) error
	read(db *InMemory, address string) error
}
