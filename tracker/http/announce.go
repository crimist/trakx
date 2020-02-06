package http

import (
	"math/rand"
	"net"
	"strconv"
	"time"
	"unsafe"

	"github.com/crimist/trakx/bencoding"
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

func (t *HTTPTracker) announce(conn net.Conn, vals *announceParams, ip storage.PeerIP) {
	storage.AddExpval(&storage.Expvar.Announces, 1)

	// get vars
	var hash storage.Hash
	var peerid storage.PeerID
	peer := storage.Peer{LastSeen: time.Now().Unix(), IP: ip}
	numwant := int(t.conf.Tracker.Numwant.Default)

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

	// Get if stop before continuing
	if vals.event == "stopped" {
		t.peerdb.Drop(hash, peerid)
		storage.AddExpval(&storage.Expvar.AnnouncesOK, 1)
		conn.Write([]byte("HTTP/1.1 200\r\n\r\n"))
		return
	}

	// Port
	portInt, err := strconv.Atoi(vals.port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		t.clientError(conn, "Invalid port")
		return
	}
	peer.Port = uint16(portInt)

	// Complete
	if vals.event == "completed" || vals.noneleft {
		peer.Complete = true
	}

	// numwant
	if vals.numwant != "" {
		numwantInt, err := strconv.Atoi(vals.numwant)
		if err != nil {
			t.clientError(conn, "Invalid numwant")
			return
		}
		if numwantInt < int(t.conf.Tracker.Numwant.Limit) || numwantInt > 0 {
			numwant = numwantInt
		}
	}

	t.peerdb.Save(&peer, hash, peerid)
	complete, incomplete := t.peerdb.HashStats(hash)

	d := bencoding.NewDict()
	d.Int64("interval", int64(t.conf.Tracker.Announce+rand.Int31n(t.conf.Tracker.AnnounceFuzz)))
	d.Int64("complete", int64(complete))
	d.Int64("incomplete", int64(incomplete))
	if vals.compact {
		peerlist := t.peerdb.PeerListBytes(hash, numwant)
		d.String("peers", *(*string)(unsafe.Pointer(&peerlist)))
	} else {
		d.Any("peers", t.peerdb.PeerList(hash, numwant, vals.nopeerid))
	}

	storage.AddExpval(&storage.Expvar.AnnouncesOK, 1)
	conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + d.Get()))
}
