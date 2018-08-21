package main

import (
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/tracker"
	"github.com/davecgh/go-spew/spew"
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

	spew.Dump(r.URL.Query())
	spew.Sdump(infoHash, peerID, port, uploaded, downloaded, left, compact, noPeerID, event, ip, numwant, key, trackerID)

	if len(infoHash) != 20 {
		tracker.Error(w, "invalid hash")
		return
	}

	t, err := tracker.NewTorrent(infoHash)
	if err != nil {
		tracker.Error(w, err.Error())
		return
	}

	if ip != r.RemoteAddr {
		tracker.Error(w, "IP address doesn't match")
		return
	}

	if event == "started" {
		t.NewPeer(peerID, key, r.RemoteAddr, port, false)
	} else if event == "stopped" {
		t.RemovePeer(peerID, key)
	} else if event == "completed" {
		t.UpdatePeer(peerID, key, r.RemoteAddr, port, true)
	}

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

	// TODO comply with compact
	d := bencoding.NewDict()
	// d.Add("tracker id", "ayy lmao") // Tracker id
	d.Add("interval", 60)  // How often they should GET this
	d.Add("complete", c)   // Number of seeders
	d.Add("incomplete", i) // Number of leeches
	d.Add("peers", peerList)

	fmt.Fprint(w, d.Get())
}

func main() {
	fmt.Println("OSX:")
	fmt.Println("\tbrew services start mysql")
	fmt.Println("\tmysql -uroot")

	db := tracker.Init()
	defer db.Close() // Cleanup

	go tracker.Clean()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Alive")
	})

	http.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Unsupported")
	})

	http.HandleFunc("/announce", Announce)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}
}
