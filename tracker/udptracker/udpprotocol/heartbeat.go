package udpprotocol

var (
	// HeartbeatRequest is not part of the official standard, but is used to check if the tracker is still alive.
	HeartbeatRequest = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(ActionHeartbeat), 0, 0, 0, 0}
	HeartbeatOk      = []byte{0xFF}
)
