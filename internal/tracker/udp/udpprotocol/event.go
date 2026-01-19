package udpprotocol

type AnnounceEvent int32

const (
	EventNone      AnnounceEvent = 0
	EventCompleted AnnounceEvent = 1
	EventStarted   AnnounceEvent = 2
	EventStopped   AnnounceEvent = 3
)
