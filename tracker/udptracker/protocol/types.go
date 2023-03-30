package protocol

type event int32

type Action int32

const (
	EventNone      event = 0
	EventCompleted event = 1
	EventStarted   event = 2
	EventStopped   event = 3

	ActionConnect   Action = 0
	ActionAnnounce  Action = 1
	ActionScrape    Action = 2
	ActionError     Action = 3
	ActionHeartbeat Action = 4 // not in spec, Trakx specific heartbeet
	ActionInvalid   Action = 5
)

var (
	HeartbeatRequest = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(ActionHeartbeat), 0, 0, 0, 0}
	HeartbeatOk      = []byte{0xFF}
)
