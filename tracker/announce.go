package tracker

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/Syc0x00/Trakx/bencoding"
)

type event int

const (
	// if event != "started" && event != "stopped" && event != "completed" {
	started event = iota
	stopped
	completed
)

type announce struct {
	infoHash   string
	peerID     string
	port       uint16
	uploaded   uint64 // ignored
	downloaded uint64 // ignored
	left       uint64
	compact    bool
	noPeerID   string // ignored
	event      string
	numwant    uint64
	key        string
	trackerID  string // ignored

	ip       string
	complete bool

	torrent Torrent

	w http.ResponseWriter
	r *http.Request
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
	a.ip = strings.Split(r.RemoteAddr, ":")[0]
	if env == Dev && providedIP != "" {
		a.ip = providedIP
	}
	if net.ParseIP(a.ip) == nil {
		a.ThrowErr("Invalid IP address")
		return nil
	}
	if strings.Contains(a.ip, ":") {
		// We don't support ipv6
		a.ThrowErr("IPv6 unsupported")
		return nil
	}

	// InfoHash
	if len(infoHash) != 20 {
		a.ThrowErr("Invalid infohash")
		return nil
	}
	a.infoHash = infoHash

	// PeerID
	if len(peerID) != 20 {
		a.ThrowErr("Invalid peer ID")
		return nil
	}
	a.peerID = peerID

	// Port
	if port == "" {
		a.ThrowErr("provide a port")
		return nil
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		a.ThrowErr("port not valid number")
		return nil
	}
	if portInt > 65535 || portInt < 1 {
		a.ThrowErr("invalid port number")
		return nil
	}
	a.port = uint16(portInt)

	// Left
	leftInt, err := strconv.ParseUint(left, 10, 64)
	if err != nil {
		a.ThrowErr("Invalid left")
		return nil
	}
	a.left = leftInt

	// Uploaded
	if uploaded != "" {
		uploadedInt, err := strconv.ParseUint(uploaded, 10, 64)
		if err != nil {
			a.ThrowErr("Invalid uploaded")
			return nil
		}
		a.uploaded = uploadedInt
	}

	// Downloaded
	if downloaded != "" {
		downloadedInt, err := strconv.ParseUint(downloaded, 10, 64)
		if err != nil {
			a.ThrowErr("Invalid downloaded")
			return nil
		}
		a.downloaded = downloadedInt
	}

	if compact == "1" {
		a.compact = true
	}

	// Numwant
	if numwant != "" {
		numwantInt, err := strconv.ParseUint(numwant, 10, 64)
		if err != nil {
			a.ThrowErr("Invalid numwant")
			return nil
		}
		a.numwant = numwantInt
	}

	// Complete
	// qBittorrent doesn't send a completed on completion
	// Instead it sends a started but with left being 0
	if a.event == "completed" || (event == "started" && left == "0") {
		a.complete = true
	}

	a.event = event
	a.noPeerID = noPeerID
	a.key = key
	a.trackerID = trackerID

	t, tError := NewTorrent(infoHash)
	if tError == Error {
		a.InternalError()
		return nil
	} else if tError == Banned {
		a.ThrowErr("Banned hash (torrent)")
		return nil
	}
	a.torrent = t

	return &a
}

// ThrowErr throws a tracker bencoded error to the client
func (a *announce) ThrowErr(reason string) {
	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	fmt.Fprint(a.w, d.Get())

	logger.Info("Client Error",
		zap.String("ip", a.ip),
		zap.String("reason", reason),
	)
}

// ThrowWarn throws a tracker bencoded warning to the client
func (a *announce) ThrowWarn(reason string) {
	d := bencoding.NewDict()
	d.Add("warning message", reason)
	fmt.Fprint(a.w, d.Get())

	logger.Info("Client Warn",
		zap.String("ip", a.ip),
		zap.String("reason", reason),
	)
}

// InternalError is a wrapper to tell the client I fucked up
func (a *announce) InternalError() {
	a.ThrowErr("Internal Server Error")
}

func (a *announce) RemovePeer() TrackErr {
	return a.torrent.RemovePeer(a.peerID, a.key)
}

func (a *announce) Peer() TrackErr {
	return a.torrent.Peer(a.peerID, a.key, a.ip, a.port, a.complete)
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

	// If stopped remove the peer and return
	if a.event == "stopped" {
		if a.RemovePeer() != OK {
			a.InternalError()
			return
		}
		fmt.Fprint(w, "Goodbye")
		return
	}

	if a.Peer() != OK {
		a.InternalError()
		return
	}

	// Get number complete and incomplete
	c := a.torrent.Complete()
	if c == -1 {
		a.InternalError()
		return
	}
	i := a.torrent.Incomplete()
	if i == -1 {
		a.InternalError()
		return
	}

	// Encode and send the data
	d := bencoding.NewDict()
	// d.Add("tracker id", "ayy lmao") // Tracker id
	d.Add("interval", 60*5) // How often they should GET this
	d.Add("complete", c)    // Number of seeders
	d.Add("incomplete", i)  // Number of leeches

	// Get the peer list
	if a.compact == true {
		peerList, tErr := a.torrent.GetPeerListCompact(a.numwant)
		if tErr != OK {
			a.InternalError()
			return
		}
		d.Add("peers", peerList)
	} else {
		peerList, tErr := a.torrent.GetPeerList(a.numwant)
		if tErr != OK {
			a.InternalError()
			return
		}
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}
