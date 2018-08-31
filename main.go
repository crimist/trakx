package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/tracker"
	"github.com/davecgh/go-spew/spew"
)

var (
	testing = true
)

// Announce x
func Announce(w http.ResponseWriter, r *http.Request) {
	infoHash := r.URL.Query().Get("info_hash")
	peerID := r.URL.Query().Get("peer_id")
	port := r.URL.Query().Get("port")
	uploaded := r.URL.Query().Get("uploaded")
	downloaded := r.URL.Query().Get("downloaded")
	left := r.URL.Query().Get("left")
	compact := r.URL.Query().Get("compact")
	noPeerID := r.URL.Query().Get("no_peer_id")
	event := r.URL.Query().Get("event")
	numwant := r.URL.Query().Get("numwant")
	key := r.URL.Query().Get("key")
	trackerID := r.URL.Query().Get("trackerid")
	ip := r.URL.Query().Get("ip")
	ipaddr := strings.Split(r.RemoteAddr, ":")[0] // Remove port ex: 127.0.0.1:9999

	// So the vars are used
	spew.Sdump(infoHash, peerID, port, uploaded, downloaded, left, compact, noPeerID, event, ip, numwant, key, trackerID)

	if len(infoHash) != 20 {
		tracker.Error(w, "invalid hash")
		return
	}

	if port == "" {
		tracker.Error(w, "no port")
		return
	}

	if _, err := strconv.Atoi(port); err != nil {
		tracker.Error(w, "invalid port")
		return
	}

	t, err := tracker.NewTorrent(infoHash)
	if err != nil {
		tracker.Error(w, err.Error())
		return
	}

	if testing { // In test mode trust given ip
		ipaddr = ip
	} else { // In prod use real ip
		if ip != "" && ip != ipaddr {
			tracker.Error(w, "IP address doesn't match")
			return
		}
	}

	// Remove the peer and return
	if event == "stopped" {
		err = t.RemovePeer(peerID, key)
		if err != nil {
			tracker.Error(w, err.Error())
		}
		return
	}

	// qBittorrent doesn't send a completed on completion
	// Instead it sends a started but with left being 0
	complete := false
	if event == "completed" || (event == "started" && left == "0") {
		complete = true
	}
	err = t.Peer(peerID, key, ipaddr, port, complete)
	if err != nil {
		tracker.Error(w, err.Error())
	}

	// Get number complete and incomplete
	c, err := t.Complete()
	if err != nil {
		tracker.Error(w, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	i, err := t.Incomplete()
	if err != nil {
		tracker.Error(w, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Encode and send the data
	d := bencoding.NewDict()
	// d.Add("tracker id", "ayy lmao") // Tracker id
	d.Add("interval", 60*5) // How often they should GET this
	d.Add("complete", c)    // Number of seeders
	d.Add("incomplete", i)  // Number of leeches

	// Get the peer list
	if compact == "1" {
		peerList, err := t.GetPeerListCompact(numwant)
		if err != nil {
			tracker.Error(w, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		d.Add("peers", peerList)
	} else {
		peerList, err := t.GetPeerList(numwant)
		if err != nil {
			tracker.Error(w, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		d.Add("peers", peerList)
	}

	fmt.Fprint(w, d.Get())
}

func main() {
	prod := flag.Bool("x", false, "Production mode")
	port := flag.String("p", "8080", "HTTP port to serve")

	flag.Parse()

	if *prod == true {
		fmt.Println("Production")
		testing = false
	}

	fmt.Println("OSX:")
	fmt.Println("\tbrew services start mysql")
	fmt.Println("\tmysql -uroot")

	db, err := tracker.Init()
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
		fmt.Fprintf(w, "SoonTM")
	})

	http.HandleFunc("/announce", Announce)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Println(err)
	}
}
