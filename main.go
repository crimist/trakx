package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"syscall"
	"time"

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

	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		panic(err)
	}

	if runtime.GOOS == "darwin" {
		limit.Cur = 24576
	} else {
		limit.Cur = limit.Max
	}

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		panic(err)
	} else {
		fmt.Printf("Limit: %v\n", limit.Cur)
	}

	err := tracker.Init(*prodFlag)
	if err != nil {
		panic(err)
	}

	go tracker.Cleaner()
	go tracker.Expvar()

	// Handlers
	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)
	trackerMux.HandleFunc("/scrape", scrape)
	trackerMux.HandleFunc("/announce", tracker.Announce)

	// Server
	server := http.Server {
		Addr: ":"+*portFlag,
		Handler: trackerMux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// Serve
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
