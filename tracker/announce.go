package tracker

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Syc0x00/Trakx/bencoding"
	"go.uber.org/zap"
)

type announce struct {
	infohash Hash
	peerid   PeerID
	compact  bool
	noPeerID bool
	numwant  int64

	peer Peer

	writer http.ResponseWriter
	req    *http.Request
}

func (a *announce) SetPeer(postIP, port, key, event, left string) bool {
	var err error

	if env == Dev && postIP != "" {
		a.peer.IP = postIP
	} else {
		a.peer.IP, _, err = net.SplitHostPort(a.req.RemoteAddr)
		if err != nil {
			a.ClientError("Invalid IP address, how the fuck does this happen?")
			logger.Error("net.SplitHostPort failed", zap.Error(err))
			return false
		}
	}
	if strings.Contains(a.peer.IP, ":") {
		a.ClientError("IPv6 unsupported", zap.String("ip", a.peer.IP))
		return false
	}

	if port == "" {
		a.ClientError("Invalid port")
		return false
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		a.ClientError("Invalid port", zap.String("port", port))
		return false
	}
	if portInt > 65535 || portInt < 1 {
		a.ClientError("Invalid port", zap.Int("port", portInt))
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
		a.ClientError("Invalid infohash", zap.Int("infoHash len", len(infohash)), zap.Any("infohash", infohash))
		return false
	}
	copy(a.infohash[:], infohash)

	return true
}

func (a *announce) SetPeerid(peerid string) bool {
	if len(peerid) != 20 {
		a.ClientError("Invalid peerid", zap.Int("peerid len", len(peerid)), zap.Any("peerid", peerid))
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
			a.ClientError("Invalid numwant", zap.String("numwant", numwant))
			return false
		}
		a.numwant = numwantInt
	} else {
		a.numwant = trackerDefaultNumwant
	}

	return true
}

func (a *announce) SetNopeerid(nopeerid string) {
	if nopeerid == "1" {
		a.noPeerID = true
	}
}

func (a *announce) error(reason string) {
	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	fmt.Fprint(a.writer, d.Get())
}

func (a *announce) warn(reason string) {
	d := bencoding.NewDict()
	d.Add("warning message", reason)
	fmt.Fprint(a.writer, d.Get())
}

func (a *announce) ClientError(reason string, fields ...zap.Field) {
	a.error(reason)
	if env == Dev {
		fields = append(fields, zap.String("ip", a.peer.IP))
		fields = append(fields, zap.String("reason", reason))
		logger.Info("Client Error", fields...)
	}
}

func (a *announce) ClientWarn(reason string) {
	a.warn(reason)
	if env == Dev {
		logger.Info("Client Warn",
			zap.String("ip", a.peer.IP),
			zap.String("reason", reason),
		)
	}
}

// InternalError is a wrapper to tell the client I fucked up
func (a *announce) InternalError(err error) {
	expvarErrs++
	a.error("Internal Server Error")
	logger.Error("Internal Server Error", zap.Error(err))
}

// AnnounceHandle processes an announce http request
func AnnounceHandle(w http.ResponseWriter, r *http.Request) {
	expvarHits++

	event := r.URL.Query().Get("event")
	a := &announce{writer: w, req: r, peer: Peer{}}

	// Set up announce
	if ok := a.SetPeer(r.URL.Query().Get("ip"), r.URL.Query().Get("port"), r.URL.Query().Get("key"), event, r.URL.Query().Get("left")); !ok {
		return
	}
	if ok := a.SetInfohash(r.URL.Query().Get("info_hash")); !ok {
		return
	}
	if ok := a.SetPeerid(r.URL.Query().Get("peer_id")); !ok {
		return
	}
	if ok := a.SetNumwant(r.URL.Query().Get("numwant")); !ok {
		return
	}
	a.SetCompact(r.URL.Query().Get("compact"))
	a.SetNopeerid(r.URL.Query().Get("no_peer_id"))

	// If the peer stopped seeding remove them
	if event == "stopped" {
		if err := a.peer.Delete(a.infohash, a.peerid); err != nil {
			a.InternalError(err)
		} else {
			fmt.Fprint(w, "See you space cowboy...")
		}
		return
	}

	if err := a.peer.Save(a.infohash, a.peerid); err != nil {
		a.InternalError(err)
		return
	}

	complete, incomplete := a.infohash.Complete()

	// Bencode response
	d := bencoding.NewDict()
	d.Add("interval", trackerAnnounceInterval) // Announce interval
	d.Add("complete", complete)                // Seeders
	d.Add("incomplete", incomplete)            // Leeches

	// Add peer list
	if a.compact == true {
		peerList := a.infohash.PeerListCompact(a.numwant)
		d.Add("peers", peerList)
	} else {
		peerList := a.infohash.PeerList(a.numwant, a.noPeerID)
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}
