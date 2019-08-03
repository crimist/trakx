package http

import (
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
)

type announce struct {
	infohash shared.Hash
	peerid   shared.PeerID
	compact  bool
	noPeerID bool
	numwant  int
	peer     shared.Peer

	conn    net.Conn
	tracker *HTTPTracker
}

func (a *announce) SetPeer(postIP, port, event, left string) bool {
	var err error
	var parsedIP net.IP

	if !a.tracker.conf.Trakx.Prod && postIP != "" {
		parsedIP = net.ParseIP(postIP).To4()
	} else {
		ipStr, _, _ := net.SplitHostPort(a.conn.RemoteAddr().String())
		parsedIP = net.ParseIP(ipStr).To4()
	}

	if parsedIP == nil {
		a.tracker.clientError(a.conn, "ipv6 unsupported")
		return false
	}
	copy(a.peer.IP[:], parsedIP)

	portInt, err := strconv.Atoi(port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		a.tracker.clientError(a.conn, "Invalid port")
		return false
	}

	if event == "completed" || left == "0" {
		a.peer.Complete = true
	}

	a.peer.Port = uint16(portInt)
	a.peer.LastSeen = time.Now().Unix()

	return true
}

func (a *announce) SetInfohash(infohash string) bool {
	if len(infohash) != 20 {
		a.tracker.clientError(a.conn, "Invalid infohash")
		return false
	}
	copy(a.infohash[:], infohash)

	return true
}

func (a *announce) SetPeerid(peerid string) bool {
	if len(peerid) != 20 {
		a.tracker.clientError(a.conn, "Invalid peerid")
		return false
	}
	copy(a.peerid[:], peerid)

	return true
}

func (a *announce) SetCompact(compact string) {
	if compact == "1" {
		a.compact = true
	}
}

func (a *announce) SetNumwant(numwant string) bool {
	a.numwant = int(a.tracker.conf.Tracker.Numwant.Default)

	if numwant != "" {
		numwantInt, err := strconv.Atoi(numwant)
		if err != nil {
			a.tracker.clientError(a.conn, "Invalid numwant")
			return false
		}
		if numwantInt > int(a.tracker.conf.Tracker.Numwant.Max) || numwantInt < 1 {
			a.numwant = int(a.tracker.conf.Tracker.Numwant.Default)
		} else {
			a.numwant = numwantInt
		}
	}
	return true
}

func (a *announce) SetNopeerid(nopeerid string) {
	if nopeerid == "1" {
		a.noPeerID = true
	}
}

func (t *HTTPTracker) Announce(c *ctx) {
	atomic.AddInt64(&shared.Expvar.Announces, 1)
	query := c.u.Query()

	event := query.Get("event")
	a := &announce{conn: c.conn, peer: shared.Peer{}, tracker: t}

	// Set up announce
	if ok := a.SetPeer(query.Get("ip"), query.Get("port"), event, query.Get("left")); !ok {
		return
	}
	if ok := a.SetInfohash(query.Get("info_hash")); !ok {
		return
	}
	if ok := a.SetPeerid(query.Get("peer_id")); !ok {
		return
	}
	if ok := a.SetNumwant(query.Get("numwant")); !ok {
		return
	}
	a.SetCompact(query.Get("compact"))
	a.SetNopeerid(query.Get("no_peer_id"))

	// If the peer stopped delete() them and exit
	if event == "stopped" {
		t.peerdb.Drop(&a.peer, &a.infohash, &a.peerid)
		atomic.AddInt64(&shared.Expvar.AnnouncesOK, 1)
		c.WriteHTTP("200", t.conf.Tracker.StoppedMsg)
		return
	}

	t.peerdb.Save(&a.peer, &a.infohash, &a.peerid)

	complete, incomplete := t.peerdb.HashStats(&a.infohash)

	// Bencode response
	d := bencoding.NewDict()
	d.Add("interval", t.conf.Tracker.AnnounceInterval)
	d.Add("complete", complete)
	d.Add("incomplete", incomplete)

	// Add peer list
	if a.compact {
		d.Add("peers", t.peerdb.PeerListBytes(&a.infohash, a.numwant))
	} else {
		d.Add("peers", t.peerdb.PeerList(&a.infohash, a.numwant, a.noPeerID))
	}

	atomic.AddInt64(&shared.Expvar.AnnouncesOK, 1)
	c.WriteHTTP("200", d.Get())
}
