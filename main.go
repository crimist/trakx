package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/tracker"
)

var db *sql.DB

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

	t, err := tracker.NewTorrent(infoHash)
	if err != nil {
		tracker.Error(w, err.Error())
	}

	if ip != r.RemoteAddr {
		tracker.Error(w, "IP address doesn't match")
	}

	if event == "started" {
		t.NewPeer(peerID, key, r.RemoteAddr, port, false)
	} else if event == "stopped" {
		t.RemovePeer(key)
	} else if event == "completed" {
		t.UpdatePeer(peerID, key, r.RemoteAddr, port, true)
	}

	fmt.Println(infoHash, peerID, port, uploaded, downloaded, left, compact, noPeerID, event, ip, numwant, key, trackerID)

	peerList, err := t.GetPeerList(numwant)
	if err != nil {
		tracker.Error(w, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	c, err := t.Complete()
	if err != nil {
		tracker.Error(w, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	i, err := t.Complete()
	if err != nil {
		tracker.Error(w, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	d := bencoding.NewDict()
	// d.Add("tracker id", "ayy lmao") // Tracker id
	d.Add("interval", 60)  // How often they should GET this
	d.Add("complete", c)   // Number of seeders
	d.Add("incomplete", i) // Number of leeches
	d.Add("peers", peerList)

	fmt.Fprint(w, d.Get())
}

func main() {
	tracker.Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Alive")
	})

	http.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Unsupported")
	})

	http.HandleFunc("/announce", Announce)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
