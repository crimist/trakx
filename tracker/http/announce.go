package http

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

type announce struct {
	infohash shared.Hash
	peerid   shared.PeerID
	compact  bool
	noPeerID bool
	numwant  int64
	peer     shared.Peer

	writer http.ResponseWriter
	req    *http.Request
}

func (a *announce) SetPeer(postIP, port, event, left string) bool {
	var err error

	if shared.Env == shared.Dev && postIP != "" {
		if parsedIP := net.ParseIP(postIP); parsedIP == nil {
			clientError("Invalid provided IP", a.writer)
			return false
		} else {
			copy(a.peer.IP[:], parsedIP)
		}
	} else {
		tmpIP, _, _ := net.SplitHostPort(a.req.RemoteAddr)
		copy(a.peer.IP[:], net.ParseIP(tmpIP))
	}

	portInt, err := strconv.Atoi(port)
	if err != nil || (portInt > 65535 || portInt < 1) {
		clientError("Invalid port", a.writer, zap.String("port", port), zap.Int("port", portInt))
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
		clientError("Invalid infohash", a.writer, zap.Int("infoHash len", len(infohash)), zap.Any("infohash", infohash))
		return false
	}
	copy(a.infohash[:], infohash)

	return true
}

func (a *announce) SetPeerid(peerid string) bool {
	if len(peerid) != 20 {
		clientError("Invalid peerid", a.writer, zap.Int("peerid len", len(peerid)), zap.Any("peerid", peerid))
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
	a.numwant = shared.DefaultNumwant

	if numwant != "" {
		numwantInt, err := strconv.ParseInt(numwant, 10, 64)
		if err != nil {
			clientError("Invalid numwant", a.writer, zap.String("numwant", numwant))
			return false
		}
		if numwantInt < shared.MaxNumwant {
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

// AnnounceHandle processes an announce http request
func AnnounceHandle(w http.ResponseWriter, r *http.Request) {
	shared.ExpvarAnnounces++
	query := r.URL.Query()

	event := query.Get("event")
	a := &announce{writer: w, req: r, peer: shared.Peer{}}

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
		a.peer.Delete(a.infohash, a.peerid)
		w.Write([]byte(shared.Bye))
		return
	}

	a.peer.Save(a.infohash, a.peerid)

	complete, incomplete := a.infohash.Complete()

	// Bencode response
	d := bencoding.NewDict()
	d.Add("interval", shared.AnnounceInterval)
	d.Add("complete", complete)
	d.Add("incomplete", incomplete)

	// Add peer list
	if a.compact == true {
		peerList := a.infohash.PeerListBytes(a.numwant)
		d.Add("peers", peerList)
	} else {
		peerList := a.infohash.PeerList(a.numwant, a.noPeerID)
		d.Add("peers", peerList)
	}

	w.Write([]byte(d.Get()))
}
