package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/tracker"
)

var (
	prod = false
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

	if len(infoHash) != 20 {
		tracker.Error(w, "Invalid hash", http.StatusBadRequest)
		return
	}

	if port == "" {
		tracker.Error(w, "No port", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(port); err != nil {
		tracker.Error(w, "Invalid port", http.StatusBadRequest)
		return
	}

	t, banErr := tracker.NewTorrent(infoHash)
	if banErr == tracker.Err {
		tracker.InternalError(w)
	} else if banErr == tracker.ErrBanned {
		tracker.Error(w, "Banned torrent (hash)", http.StatusForbidden)
	}

	if prod {
		// in prod use remote addr
		if ip != "" && ip != ipaddr {
			tracker.Error(w, "IP address doesn't match", http.StatusBadRequest)
			return
		}
	} else {
		// In testing trust the client provided ip
		ipaddr = ip
	}

	// Check if valid IP address
	validIP := net.ParseIP(ipaddr)
	if validIP == nil {
		tracker.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}

	// If stopped remove the peer and return
	if event == "stopped" {
		if t.RemovePeer(peerID, key) {
			tracker.InternalError(w)
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

	if t.Peer(peerID, key, ipaddr, port, complete) {
		tracker.InternalError(w)
		return
	}

	// Get number complete and incomplete
	c := t.Complete()
	if c == -1 {
		tracker.InternalError(w)
		return
	}
	i := t.Incomplete()
	if i == -1 {
		tracker.InternalError(w)
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
		peerList, failed := t.GetPeerListCompact(numwant)
		if failed {
			tracker.InternalError(w)
			return
		}
		d.Add("peers", peerList)
	} else {
		peerList, failed := t.GetPeerList(numwant)
		if failed {
			tracker.InternalError(w)
			return
		}
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}

func main() {
	prodFlag := flag.Bool("x", false, "Production mode")
	portFlag := flag.String("p", "1337", "HTTP port to serve")

	flag.Parse()

	prod = *prodFlag

	db, err := tracker.Init(prod)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close() // Cleanup

	go tracker.Clean()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Alive")
	})

	http.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		message := "I'm a teapot\n             ;,'\n     _o_    ;:;'\n ,-.'---`.__ ;\n((j`=====',-'\n `-\\     /\n    `-=-'     This tracker doesn't support /scrape\n"
		fmt.Fprintf(w, message)
	})

	http.HandleFunc("/announce", Announce)

	if err := http.ListenAndServe(":"+*portFlag, nil); err != nil {
		fmt.Println(err)
	}
}
