package storage

// PeerChan holds the buffered peers like a `sync.Pool`
var PeerChan peerChan

func init() {
	PeerChan.create()
}

type peerChan struct {
	channel chan *Peer
}

func (pc *peerChan) create() {
	// worth the 8MB cost as it will stabilize @ the maximum number of peers
	// once the tracker has experience a 24hr cycle
	const chanSize = 1e6

	pc.channel = make(chan *Peer, chanSize)
}

func (pc *peerChan) buffer() {
	const amount = 5000

	for i := 0; i < amount; i++ {
		pc.channel <- new(Peer)
	}
	Expvar.Pools.Peer.Add(amount)
}

func (pc *peerChan) Get() *Peer {
	if len(pc.channel) == 0 {
		go pc.buffer()
	}

	return <-pc.channel
}

func (pc *peerChan) Put(peer *Peer) {
	// if the channel is full than drop the peer
	if len(pc.channel) == cap(pc.channel) {
		return
	}

	pc.channel <- peer
}
