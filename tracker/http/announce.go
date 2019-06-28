package http

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
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

func (a *announce) SetPeer(postIP, port, key, event, left string) bool {
	var err error

	if shared.Env == shared.Dev && postIP != "" {
		a.peer.IP = postIP
	} else {
		a.peer.IP, _, err = net.SplitHostPort(a.req.RemoteAddr)
		if err != nil {
			clientError("Invalid IP address, how the fuck does this happen?", a.writer)
			shared.Logger.Error("net.SplitHostPort failed", zap.Error(err))
			return false
		}
	}
	if strings.Contains(a.peer.IP, ":") {
		clientError("IPv6 unsupported", a.writer, zap.String("ip", a.peer.IP))
		return false
	}

	if port == "" {
		clientError("Invalid port", a.writer)
		return false
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		clientError("Invalid port", a.writer, zap.String("port", port))
		return false
	}
	if portInt > 65535 || portInt < 1 {
		clientError("Invalid port", a.writer, zap.Int("port", portInt))
		return false
	}

	if event == "completed" || left == "0" {
		a.peer.Complete = true
	}

	a.peer.Port = uint16(portInt)
	a.peer.Key = []byte(key)
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
	if numwant != "" {
		numwantInt, err := strconv.ParseInt(numwant, 10, 64)
		if err != nil {
			clientError("Invalid numwant", a.writer, zap.String("numwant", numwant))
			return false
		}
		a.numwant = numwantInt
	} else {
		a.numwant = shared.DefaultNumwant
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
	if ok := a.SetPeer(query.Get("ip"), query.Get("port"), query.Get("key"), event, query.Get("left")); !ok {
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
		fmt.Fprint(w, shared.Bye)
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

	fmt.Fprint(w, d.Get())
}
