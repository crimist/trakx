package http

import (
	"math/rand"
	"net"
	"net/netip"
	"strconv"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/storage"
)

type announceParameters struct {
	compact  bool
	nopeerid bool
	noneleft bool
	event    string
	port     string
	hash     string
	peerid   string
	numwant  string
}

func (tracker *Tracker) announce(conn net.Conn, parameters *announceParameters, addr netip.Addr) {
	tracker.collector.Announce()

	var hash storage.Hash
	var peerid storage.PeerID

	if len(parameters.hash) != 20 {
		tracker.error(conn, "Invalid infohash")
		return
	}
	copy(hash[:], parameters.hash)

	if len(parameters.peerid) != 20 {
		tracker.error(conn, "Invalid peerid")
		return
	}
	copy(peerid[:], parameters.peerid)

	if parameters.event == "stopped" {
		tracker.peerdb.PeerRemove(hash, peerid)
	} else {
		portInt, err := strconv.Atoi(parameters.port)
		if err != nil || (portInt > 65535 || portInt < 1) {
			tracker.error(conn, "Invalid port")
			return
		}

		peerComplete := false
		if parameters.event == "completed" || parameters.noneleft {
			peerComplete = true
		}

		tracker.peerdb.PeerAdd(hash, peerid, addr, uint16(portInt), peerComplete)
	}

	numwant := tracker.config.DefaultNumwant
	if parameters.numwant != "" {
		numwantInt, err := strconv.Atoi(parameters.numwant)
		if err != nil || numwantInt < 0 {
			tracker.error(conn, "Invalid numwant")
			return
		}

		numwant = min(uint(numwantInt), tracker.config.MaximumNumwant)
	}

	seeds, leeches := tracker.peerdb.TorrentStats(hash)

	interval := tracker.config.Interval
	if tracker.config.IntervalVariance > 0 {
		interval += uint(rand.Int31n(int32(tracker.config.IntervalVariance)))
	}

	dictionary := pools.Dictionaries.Get()
	dictionary.Int64("interval", int64(interval))
	dictionary.Int64("complete", int64(seeds))
	dictionary.Int64("incomplete", int64(leeches))
	if parameters.compact {
		peers4, peers6 := tracker.peerdb.TorrentPeersCompact(hash, uint(numwant), storage.IPv4|storage.IPv6)
		dictionary.StringBytes("peers", peers4)
		dictionary.StringBytes("peers6", peers6)

		pools.Peerlists4.Put(peers4)
		pools.Peerlists6.Put(peers6)
	} else {
		dictionary.BytesliceSlice("peers", tracker.peerdb.TorrentPeers(hash, numwant, !parameters.nopeerid))
	}

	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))
	pools.Dictionaries.Put(dictionary)
}
