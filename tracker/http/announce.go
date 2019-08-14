package http

import (
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
)

func (t *HTTPTracker) announce(conn net.Conn, vals url.Values) {
	shared.AddExpval(&shared.Expvar.Announces, 1)

	// get vars
	var hash shared.Hash
	var peerid shared.PeerID
	peer := shared.Peer{LastSeen: time.Now().Unix()}
	numwant := int(t.conf.Tracker.Numwant.Default)
	compact := vals.Get("compact") == "1"
	nopeerid := vals.Get("no_peer_id") == "1"
	noneleft := vals.Get("left") == "0"
	event := vals.Get("event")
	port := vals.Get("port")
	hashStr := vals.Get("info_hash")
	peeridStr := vals.Get("peer_id")
	numwantStr := vals.Get("numwant")

	// IP
	ipStr, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	parsedIP := net.ParseIP(ipStr).To4()
	if parsedIP == nil {
		t.clientError(conn, "ipv6 unsupported")
		return
	}
	copy(peer.IP[:], parsedIP)

	// Port
	portInt, err := strconv.Atoi(port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		t.clientError(conn, "Invalid port")
		return
	}
	peer.Port = uint16(portInt)

	// Complete
	if event == "completed" || noneleft {
		peer.Complete = true
	}

	// hash
	if len(hashStr) != 20 {
		t.clientError(conn, "Invalid infohash")
		return
	}
	copy(hash[:], hashStr)

	// peerid
	if len(peeridStr) != 20 {
		t.clientError(conn, "Invalid peerid")
		return
	}
	copy(peerid[:], peeridStr)

	// numwant
	if numwantStr != "" {
		numwantInt, err := strconv.Atoi(numwantStr)
		if err != nil {
			t.clientError(conn, "Invalid numwant")
			return
		}
		if numwantInt < int(t.conf.Tracker.Numwant.Max) || numwantInt > 0 {
			numwant = numwantInt
		}
	}

	// Processing
	if event == "stopped" {
		t.peerdb.Drop(&peer, &hash, &peerid)
		shared.AddExpval(&shared.Expvar.AnnouncesOK, 1)
		conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + t.conf.Tracker.StoppedMsg))
		return
	}

	t.peerdb.Save(&peer, &hash, &peerid)
	complete, incomplete := t.peerdb.HashStats(&hash)

	d := bencoding.NewDict()
	d.Int64("interval", int64(t.conf.Tracker.AnnounceInterval))
	d.Int64("complete", int64(complete))
	d.Int64("incomplete", int64(incomplete))
	if compact {
		d.Any("peers", t.peerdb.PeerListBytes(&hash, numwant))
	} else {
		d.Any("peers", t.peerdb.PeerList(&hash, numwant, nopeerid))
	}

	shared.AddExpval(&shared.Expvar.AnnouncesOK, 1)
	conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + d.Get()))
}
