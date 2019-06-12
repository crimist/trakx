package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/Syc0x00/Trakx/tracker"
)

const indexStyle = "<style> body { background-color: black; font-family: arial; color: white; font-size: 25px; } </style>"
const dmcaStyle = "<style> body { background-color: black; font-family: arial; color: white; font-size: 200px; } </style>"

func index(w http.ResponseWriter, r *http.Request) {
	resp := indexStyle
	resp += "<p>This is an open p2p tracker. Feel free to use it :)</p>"
	resp += "<p>http://nibba.trade:1337/announce</p>"
	resp += "<p>Message me if you've got issues. Discord: <3#1527 / Email: tracker@nibba.trade</p>"
	resp += "<a href='/dmca'>DMCA?</a>"

	fmt.Fprintf(w, resp)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=BwSts2s4ba4", http.StatusMovedPermanently)
}

func scrape(w http.ResponseWriter, r *http.Request) {
	resp := "I'm a teapot\n             ;,'\n     _o_    ;:;'\n ,-.'---`.__ ;\n((j`=====',-'\n `-\\     /\n    `-=-'     This tracker doesn't support /scrape"

	w.WriteHeader(http.StatusTeapot)
	fmt.Fprintf(w, resp)
}

func main() {
	// Get flags
	prodFlag := flag.Bool("x", false, "Production mode")
	portFlag := flag.String("p", "1337", "HTTP port to serve")
	flag.Parse()

	// Init dbs ect.
	db, err := tracker.Init(*prodFlag)
	if err != nil {
		panic(err)
	}

	defer db.Close()
	go tracker.Clean()
	go tracker.Expvar()
	// stats := tracker.Stats{Directory: "/var/www/html/"}
	// go stats.Generator()

	// Handlers
	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)
	trackerMux.HandleFunc("/scrape", scrape)
	trackerMux.HandleFunc("/announce", tracker.Announce)

	// Serve
	if err := http.ListenAndServe(":"+*portFlag, trackerMux); err != nil {
		panic(err)
	}
}
