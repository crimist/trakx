package tracker

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Syc0x00/Trakx/bencoding"
)

// Announce x
func Announce(w http.ResponseWriter, r *http.Request) {
	infoHash := r.URL.Query().Get("info_hash")
	peerID := r.URL.Query().Get("peer_id")
	port := r.URL.Query().Get("port")
	// uploaded := r.URL.Query().Get("uploaded")
	// downloaded := r.URL.Query().Get("downloaded")
	left := r.URL.Query().Get("left")
	compact := r.URL.Query().Get("compact")
	// noPeerID := r.URL.Query().Get("no_peer_id")
	event := r.URL.Query().Get("event")
	numwant := r.URL.Query().Get("numwant")
	key := r.URL.Query().Get("key")
	// trackerID := r.URL.Query().Get("trackerid")
	ip := r.URL.Query().Get("ip")
	ipaddr := strings.Split(r.RemoteAddr, ":")[0] // Remove port ex: 127.0.0.1:9999

	// Check info hash
	if len(infoHash) != 20 {
		ThrowErr(w, "Invalid hash", http.StatusBadRequest)
		return
	}

	// Check peer id
	if len(peerID) != 20 {
		ThrowErr(w, "Invalid peer ID, must be 20 bytes.", http.StatusBadRequest)
		return
	}

	// Check valid port
	if port == "" {
		ThrowErr(w, "No port", http.StatusBadRequest)
		return
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		ThrowErr(w, "Port not int", http.StatusBadRequest)
		return
	}
	if portInt > 65535 || portInt < 1 {
		ThrowErr(w, "Port not uint16", http.StatusBadRequest)
		return
	}

	// Check valid event
	if event != "started" && event != "stopped" && event != "completed" {
		ThrowErr(w, "Invalid event", http.StatusBadRequest)
		return
	}

	// Check left is int
	if _, err := strconv.Atoi(left); err != nil {
		ThrowErr(w, "Left not int", http.StatusBadRequest)
		return
	}

	t, tError := NewTorrent(infoHash)
	if tError == Error {
		InternalError(w)
		return
	} else if tError == Banned {
		ThrowErr(w, "Banned torrent (hash)", http.StatusForbidden)
		return
	}

	if env == Prod {
		// in prod use remote addr
		if ip != "" && ip != ipaddr {
			ThrowErr(w, "IP address doesn't match", http.StatusBadRequest)
			return
		}
	} else {
		// In dev trust the client provided ip
		ipaddr = ip
	}

	// Check if valid IPv4 address
	validIP := net.ParseIP(ipaddr)
	if validIP == nil {
		ThrowErr(w, "Invalid IP address", http.StatusBadRequest)
		return
	}
	if strings.Contains(ipaddr, ":") { // We don't support ipv6
		ThrowErr(w, "IPv6 unsupported", http.StatusBadRequest)
		return
	}

	// If stopped remove the peer and return
	if event == "stopped" {
		if t.RemovePeer(peerID, key) != OK {
			InternalError(w)
			return
		}
		fmt.Fprint(w, "Goodbye")
		return
	}

	// qBittorrent doesn't send a completed on completion
	// Instead it sends a started but with left being 0
	complete := false
	if event == "completed" || (event == "started" && left == "0") {
		complete = true
	}

	if t.Peer(peerID, key, ipaddr, port, complete) != OK {
		InternalError(w)
		return
	}

	// Get number complete and incomplete
	c := t.Complete()
	if c == -1 {
		InternalError(w)
		return
	}
	i := t.Incomplete()
	if i == -1 {
		InternalError(w)
		return
	}

	// Encode and send the data
	d := bencoding.NewDict()
	// d.Add("tracker id", "ayy lmao") // Tracker id
	d.Add("interval", 60*5) // How often they should GET this
	d.Add("complete", c)    // Number of seeders
	d.Add("incomplete", i)  // Number of leeches

	// Get the peer list
	if compact == "1" {
		peerList, tErr := t.GetPeerListCompact(numwant)
		if tErr != OK {
			InternalError(w)
			return
		}
		d.Add("peers", peerList)
	} else {
		peerList, tErr := t.GetPeerList(numwant)
		if tErr != OK {
			InternalError(w)
			return
		}
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}
