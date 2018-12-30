package tracker

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	numwant    int64
	key        string
	trackerID  string // ignored

	complete bool

	peer Peer

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
	IP := strings.Split(r.RemoteAddr, ":")[0]
	if env == Dev && providedIP != "" {
		IP = providedIP
	}
	if net.ParseIP(IP) == nil {
		a.ClientError("Invalid IP address")
		return nil
	}
	if strings.Contains(IP, ":") {
		// We don't support ipv6
		a.ClientError("IPv6 unsupported")
		return nil
	}

	// InfoHash
	if len(infoHash) != 20 {
		a.ClientError("Invalid infohash")
		return nil
	}

	// PeerID
	if len(peerID) != 20 {
		a.ClientError("Invalid peer ID")
		return nil
	}
	if IsBanned(infoHash) == Banned {
		a.ClientError("Banned hash")
		return nil
	}

	// Port
	if port == "" {
		a.ClientError("provide a port")
		return nil
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		a.ClientError("port not valid number")
		return nil
	}
	if portInt > 65535 || portInt < 1 {
		a.ClientError("invalid port number")
		return nil
	}

	// Left
	leftInt, err := strconv.ParseUint(left, 10, 64)
	if err != nil {
		a.ClientError("Invalid left")
		return nil
	}
	a.left = leftInt

	// Uploaded
	if uploaded != "" {
		uploadedInt, err := strconv.ParseUint(uploaded, 10, 64)
		if err != nil {
			a.ClientError("Invalid uploaded")
			return nil
		}
		a.uploaded = uploadedInt
	}

	// Downloaded
	if downloaded != "" {
		downloadedInt, err := strconv.ParseUint(downloaded, 10, 64)
		if err != nil {
			a.ClientError("Invalid downloaded")
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
			a.ClientError("Invalid numwant")
			return nil
		}
		a.numwant = numwantInt
	}

	complete := false
	if a.event == "completed" || (event == "started" && left == "0") {
		complete = true
	}

	a.event = event
	a.noPeerID = noPeerID
	a.key = key
	a.trackerID = trackerID

	a.peer = Peer{
		ID:       peerID,
		PeerKey:  key,
		Hash:     infoHash,
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

func (a *announce) ClientError(reason string) {
	a.error(reason)
	logger.Info("Client Error",
		zap.String("ip", a.peer.IP),
		zap.String("reason", reason),
	)
}

func (a *announce) ClientWarn(reason string) {
	a.warn(reason)
	logger.Info("Client Warn",
		zap.String("ip", a.peer.IP),
		zap.String("reason", reason),
	)
}

// InternalError is a wrapper to tell the client I fucked up
func (a *announce) InternalError(err error) {
	a.error("Internal Server Error")
	logger.Info("Internal Server Error",
		zap.Error(err),
	)
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

	c, err := Complete()
	if err != nil {
		a.InternalError(err)
		return
	}

	i, err := Incomplete()
	if err != nil {
		a.InternalError(err)
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
		peerList, err := PeerListCompact(a.numwant)
		if err != nil {
			a.InternalError(err)
			return
		}
		d.Add("peers", peerList)
	} else {
		peerList, err := PeerList(a.numwant)
		if err != nil {
			a.InternalError(err)
			return
		}
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}
