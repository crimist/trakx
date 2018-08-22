package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

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

	if ip != "" && ip != r.RemoteAddr {
		tracker.Error(w, "IP address doesn't match")
		return
	}

	ipaddr := strings.Split(r.RemoteAddr, ":")[0] // Remove port ex: 127.0.0.1:9999

	if event == "started" {
		err = t.NewPeer(peerID, key, ipaddr, port, false)
		if err != nil {
			fmt.Println(err)
		}
	} else if event == "stopped" {
		err = t.RemovePeer(peerID, key)
		if err != nil {
			fmt.Println(err)
		}
	} else if event == "completed" {
		err = t.UpdatePeer(peerID, key, ipaddr, port, true)
		if err != nil {
			fmt.Println(err)
		}
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
	prod := flag.Bool("x", true, "Production mode")
	port := flag.String("p", "8080", "HTTP port to serve")

	flag.Parse()

	if *prod == true {
		fmt.Println("Production")
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
