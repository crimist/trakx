package http

import (
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type announce struct {
	infohash shared.Hash
	peerid   shared.PeerID
	compact  bool
	noPeerID bool
	numwant  int
	peer     shared.Peer

	writer  http.ResponseWriter
	req     *http.Request
	tracker *HTTPTracker
}

func (a *announce) SetPeer(postIP, port, event, left string) bool {
	var err error
	var parsedIP net.IP

	if !a.tracker.conf.Trakx.Prod && postIP != "" {
		parsedIP = net.ParseIP(postIP).To4()
	} else {
		ipStr, _, _ := net.SplitHostPort(a.req.RemoteAddr)
		parsedIP = net.ParseIP(ipStr).To4()
	}

	if parsedIP == nil {
		a.tracker.clientError("ipv6 unsupported", a.writer)
		return false
	}
	copy(a.peer.IP[:], parsedIP)

	portInt, err := strconv.Atoi(port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		a.tracker.clientError("Invalid port", a.writer, zap.String("port", port), zap.Int("port", portInt))
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
		a.tracker.clientError("Invalid infohash", a.writer, zap.Int("infoHash len", len(infohash)), zap.Any("infohash", infohash))
		return false
	}
	copy(a.infohash[:], infohash)

	return true
}

func (a *announce) SetPeerid(peerid string) bool {
	if len(peerid) != 20 {
		a.tracker.clientError("Invalid peerid", a.writer, zap.Int("peerid len", len(peerid)), zap.Any("peerid", peerid))
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
		numwantInt, err := strconv.ParseInt(numwant, 10, 64)
		if err != nil {
			a.tracker.clientError("Invalid numwant", a.writer, zap.String("numwant", numwant))
			return false
		}
		if numwantInt < int64(a.tracker.conf.Tracker.Numwant.Max) {
			a.numwant = int(numwantInt)
		}
	}
	return true
}

func (a *announce) SetNopeerid(nopeerid string) {
	if nopeerid == "1" {
		a.noPeerID = true
	}
}

func (t *HTTPTracker) AnnounceHandle(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&shared.Expvar.Announces, 1)
	query := r.URL.Query()

	event := query.Get("event")
	a := &announce{writer: w, req: r, peer: shared.Peer{}, tracker: t}

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
		shared.PeerDB.Drop(&a.peer, &a.infohash, &a.peerid)
		atomic.AddInt64(&shared.Expvar.AnnouncesOK, 1)
		w.Write([]byte(t.conf.Tracker.StoppedMsg))
		return
	}

	shared.PeerDB.Save(&a.peer, &a.infohash, &a.peerid)

	complete, incomplete := shared.PeerDB.HashStats(&a.infohash)

	// Bencode response
	d := bencoding.NewDict()
	d.Add("interval", t.conf.Tracker.AnnounceInterval)
	d.Add("complete", complete)
	d.Add("incomplete", incomplete)

	// Add peer list
	if a.compact == true {
		peerList := string(shared.PeerDB.PeerListBytes(&a.infohash, a.numwant))
		d.Add("peers", peerList)
	} else {
		peerList := shared.PeerDB.PeerList(&a.infohash, a.numwant, a.noPeerID)
		d.Add("peers", peerList)
	}

	atomic.AddInt64(&shared.Expvar.AnnouncesOK, 1)
	w.Write([]byte(d.Get()))
}
