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
	// if the length is approaching the cap drop the peer
	if len(pc.channel)+500 > cap(pc.channel) {
		return
	}
	pc.channel <- peer
}
