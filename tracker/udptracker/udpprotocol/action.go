package udpprotocol

type Action int32

func (action Action) IsInvalid() bool {
	return action < 0 || action > 4
}

const (
	ActionConnect   Action = 0
	ActionAnnounce  Action = 1
	ActionScrape    Action = 2
	ActionError     Action = 3
	ActionHeartbeat Action = 4 // not in spec, Trakx specific heartbeet
)
