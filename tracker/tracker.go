package tracker

type Tracker interface {
	Serve() error
	Shutdown()
	ConnectionCount() int
}
