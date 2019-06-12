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
	infoHash   string
	peerID     string
	port       uint16
	uploaded   uint64 // ignored
	downloaded uint64 // ignored
	left       uint64
	compact    bool
	noPeerID   bool
	event      string
	numwant    int64
	key        string
	trackerID  string // ignored

	IP string

	peer Peer
	w    http.ResponseWriter
	r    *http.Request
}

// NewAnnounce does something
func NewAnnounce(
	infoHash string,
	peerID string,
	port string,
	uploaded string,
	downloaded string,
	left string,
	compact string,
	noPeerID string,
	event string,
	numwant string,
	key string,
	trackerID string,
	providedIP string,
	w http.ResponseWriter,
	r *http.Request,
) *announce {
	a := announce{
		w: w,
		r: r,
	}

	// IP
	IP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		a.ClientError("Invalid IP address, how the fuck?")
		logger.Error("net.SplitHostPort failed", zap.Error(err))
		return nil
	}
	a.IP = IP
	if env == Dev && providedIP != "" {
		IP = providedIP
	}
	if strings.Contains(IP, ":") {
		a.ClientError("IPv6 unsupported")
		return nil
	}

	// InfoHash
	if len(infoHash) != 20 {
		a.ClientError("Invalid infohash", zap.Int("infoHash len", len(infoHash)))
		return nil
	}

	// PeerID
	if len(peerID) != 20 {
		a.ClientError("Invalid peer ID", zap.Int("peerID len", len(peerID)))
		return nil
	}

	// Port
	if port == "" {
		a.ClientError("provide a port")
		return nil
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		a.ClientError("port not valid number", zap.String("port", port))
		return nil
	}
	if portInt > 65535 || portInt < 1 {
		a.ClientError("invalid port number", zap.Int("port", portInt))
		return nil
	}

	// Left
	leftInt, err := strconv.ParseUint(left, 10, 64)
	if err != nil {
		a.ClientError("Invalid left", zap.String("left", left))
		return nil
	}
	a.left = leftInt

	// Uploaded
	if uploaded != "" {
		uploadedInt, err := strconv.ParseUint(uploaded, 10, 64)
		if err != nil {
			a.ClientError("Invalid uploaded", zap.String("uploaded", uploaded))
			return nil
		}
		a.uploaded = uploadedInt
	}

	// Downloaded
	if downloaded != "" {
		downloadedInt, err := strconv.ParseUint(downloaded, 10, 64)
		if err != nil {
			a.ClientError("Invalid downloaded", zap.String("downloaded", downloaded))
			return nil
		}
		a.downloaded = downloadedInt
	}

	if compact == "1" {
		a.compact = true
	}

	// Numwant
	if numwant != "" {
		numwantInt, err := strconv.ParseInt(numwant, 10, 64)
		if err != nil {
			a.ClientError("Invalid numwant", zap.String("numwant", numwant))
			return nil
		}
		a.numwant = numwantInt
	} else {
		// default to 200
		a.numwant = 200
	}

	complete := false
	if event == "completed" || a.left == 0 {
		complete = true
	}

	if noPeerID == "1" {
		a.noPeerID = true
	}

	a.event = event
	a.key = key
	a.trackerID = trackerID

	a.peer = Peer{
		ID:       []byte(peerID),
		Key:      []byte(key),
		Hash:     []byte(infoHash),
		IP:       IP,
		Port:     uint16(portInt),
		Complete: complete,
		LastSeen: time.Now().Unix(),
	}

	return &a
}

func (a *announce) error(reason string) {
	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	fmt.Fprint(a.w, d.Get())
}

func (a *announce) warn(reason string) {
	d := bencoding.NewDict()
	d.Add("warning message", reason)
	fmt.Fprint(a.w, d.Get())
}

func (a *announce) ClientError(reason string, fields ...zap.Field) {
	a.error(reason)
	logger.Info("Client Error",
		zap.String("ip", a.IP),
		zap.String("reason", reason),
		fields...,
	)
}

func (a *announce) ClientWarn(reason string) {
	a.warn(reason)
	logger.Info("Client Warn",
		zap.String("ip", a.IP),
		zap.String("reason", reason),
	)
}

// InternalError is a wrapper to tell the client I fucked up
func (a *announce) InternalError(err error) {
	a.error("Internal Server Error")
	logger.Error("Internal Server Error", zap.Error(err))
}

// Announce x
func Announce(w http.ResponseWriter, r *http.Request) {
	a := NewAnnounce(
		r.URL.Query().Get("info_hash"),
		r.URL.Query().Get("peer_id"),
		r.URL.Query().Get("port"),
		r.URL.Query().Get("uploaded"),
		r.URL.Query().Get("downloaded"),
		r.URL.Query().Get("left"),
		r.URL.Query().Get("compact"),
		r.URL.Query().Get("no_peer_id"),
		r.URL.Query().Get("event"),
		r.URL.Query().Get("numwant"),
		r.URL.Query().Get("key"),
		r.URL.Query().Get("trackerid"),
		r.URL.Query().Get("ip"),
		w,
		r,
	)
	if a == nil {
		return
	}

	// Check for banned hash
	if a.peer.Hash.Banned() {
		a.ClientError("Banned hash")
		return
	}

	// If stopped remove the peer and return
	if a.event == "stopped" {
		if err := a.peer.Delete(); err != nil {
			a.InternalError(err)
			return
		}
		fmt.Fprint(w, "Goodbye")
		return
	}

	if err := a.peer.Save(); err != nil {
		a.InternalError(err)
		return
	}

	c, err := a.peer.Hash.Complete()
	if err != nil {
		a.InternalError(err)
		return
	}

	i, err := a.peer.Hash.Incomplete()
	if err != nil {
		a.InternalError(err)
		return
	}

	// Bencode response
	d := bencoding.NewDict()
	d.Add("interval", trackerInterval) // Announce interval
	d.Add("complete", c)               // Seeders
	d.Add("incomplete", i)             // Leeches

	// Add peer list
	if a.compact == true {
		peerList, err := a.peer.Hash.PeerListCompact(a.numwant)
		if err != nil {
			a.InternalError(err)
			return
		}
		d.Add("peers", peerList)
	} else {
		peerList, err := a.peer.Hash.PeerList(a.numwant, a.noPeerID)
		if err != nil {
			a.InternalError(err)
			return
		}
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}
