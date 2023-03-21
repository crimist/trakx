package http

import (
	"math/rand"
	"net"
	"net/netip"
	"strconv"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/storage"
)

type announceParams struct {
	compact  bool
	nopeerid bool
	noneleft bool
	event    string
	port     string
	hash     string
	peerid   string
	numwant  string
}

func (t *HTTPTracker) announce(conn net.Conn, vals *announceParams, ip netip.Addr) {
	stats.Announces.Add(1)

	// get vars
	var hash storage.Hash
	var peerid storage.PeerID

	// hash
	if len(vals.hash) != 20 {
		t.clientError(conn, "Invalid infohash")
		return
	}
	copy(hash[:], vals.hash)

	// peerid
	if len(vals.peerid) != 20 {
		t.clientError(conn, "Invalid peerid")
		return
	}
	copy(peerid[:], vals.peerid)

	// get if stop before continuing
	if vals.event == "stopped" {
		t.peerdb.Drop(hash, peerid)
		conn.Write(httpSuccessBytes)
		return
	}

	// port
	portInt, err := strconv.Atoi(vals.port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		t.clientError(conn, "Invalid port")
		return
	}

	// numwant
	numwant := config.Config.Numwant.Default

	if vals.numwant != "" {
		numwantInt, err := strconv.Atoi(vals.numwant)
		if err != nil || numwantInt < 0 {
			t.clientError(conn, "Invalid numwant")
			return
		}
		numwantUint := uint(numwantInt)

		// if numwant is within our limit than listen to the client
		if numwantUint <= config.Config.Numwant.Limit {
			numwant = numwantUint
		} else {
			numwant = config.Config.Numwant.Limit
		}
	}

	peerComplete := false
	if vals.event == "completed" || vals.noneleft {
		peerComplete = true
	}

	t.peerdb.Save(ip, uint16(portInt), peerComplete, hash, peerid)
	complete, incomplete := t.peerdb.HashStats(hash)

	interval := int64(config.Config.Announce.Base.Seconds())
	if int32(config.Config.Announce.Fuzz.Seconds()) > 0 {
		interval += rand.Int63n(int64(config.Config.Announce.Fuzz.Seconds()))
	}

	dictionary := pools.Dictionaries.Get()
	dictionary.Int64("interval", interval)
	dictionary.Int64("complete", int64(complete))
	dictionary.Int64("incomplete", int64(incomplete))
	if vals.compact {
		peers4, peers6 := t.peerdb.PeerListBytes(hash, numwant)
		dictionary.StringBytes("peers", peers4)
		dictionary.StringBytes("peers6", peers6)

		pools.Peerlists4.Put(peers4)
		pools.Peerlists6.Put(peers6)
	} else {
		dictionary.BytesliceSlice("peers", t.peerdb.PeerList(hash, numwant, vals.nopeerid))
	}

	// double write no append is more efficient when > ~250 peers in response
	// conn.Write(httpSuccessBytes)
	// conn.Write(d.GetBytes())

	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))
	pools.Dictionaries.Put(dictionary)
}
