package tracker

import (
	"fmt"
	"net/http"
)

const trackerBase = "http://nibba.trade:1337"
const netdataBase = "https://nibba.trade/netdata"

const indexHTML = `
<head>
	<title>Trakx</title>
</head>
<style>
body {
	background-color: black;
	font-family: arial;
	color: #9b9ea3;
	font-size: 25px;
	text-align: center;
	align-items: center;
	display: flex;
	justify-content: center;
}
</style>
<div>
	<p>Trakx is an open p2p tracker. Feel free to use it :)</p>
	<p>Add <span style="background-color: #1dc135; color: black;">` + trackerBase + `/announce</span></p>
	<embed src="` + netdataBase + `/api/v1/badge.svg?chart=go_expvar_Trakx.scrapes_sec&alarm=trakx_scrapes&refresh=auto" type="image/svg+xml" height="20"/>
	<embed src="` + netdataBase + `/api/v1/badge.svg?chart=go_expvar_Trakx.announces_sec&alarm=trakx_announces&refresh=auto" type="image/svg+xml" height="20"/>
	<embed src="` + netdataBase + `/api/v1/badge.svg?chart=go_expvar_Trakx.announces_sec&alarm=trakx_announces_avg_5min&refresh=auto" type="image/svg+xml" height="20"/>
	<embed src="` + netdataBase + `/api/v1/badge.svg?chart=go_expvar_Trakx.announces_sec&alarm=trakx_announces_avg_1hour&refresh=auto" type="image/svg+xml" height="20"/>
	<embed src="` + netdataBase + `/api/v1/badge.svg?chart=go_expvar_Trakx.errors_sec&alarm=trakx_errors&refresh=auto" type="image/svg+xml" height="20"/>
	<p>Discord: <3#1527 / Email: tracker@nibba.trade</p>
	<a href='/dmca'>DMCA?</a>
</div>
`

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, indexHTML)
}

func dmca(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://www.youtube.com/watch?v=BwSts2s4ba4", http.StatusMovedPermanently)
}
