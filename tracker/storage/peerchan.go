package storage

// worth the 4MB cost as it will stabilize @ the maximum number of peers
// once the tracker has experience a 24hr cycle
const maxChanSize = 5e5

// PeerChan holds buffered peers similar to `sync.Pool`
var PeerChan peerChan

func init() {
	PeerChan.create()
}

type peerChan struct {
	channel chan *Peer
}

func (pc *peerChan) create() {
	pc.channel = make(chan *Peer, maxChanSize)
}

// Add adds n number of peers into the channel
func (pc *peerChan) Add(n uint64) {
	// don't go over maxChanSize
	if n > maxChanSize {
		n = maxChanSize
	}

	for i := uint64(0); i < n; i++ {
		pc.channel <- new(Peer)
	}
	Expvar.Pools.Peer.Add(int64(n))
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
