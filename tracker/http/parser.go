package http

import (
	"bytes"
	"strings"
)

type parsed struct {
	Path   string
	Params []string
	URLend int

	pathstart int
	pathend   int
}

func parse(data *[]byte, length int) parsed {
	p := parsed{
		URLend:    bytes.Index((*data)[:length], []byte(" HTTP/")),
		pathstart: bytes.Index((*data)[:length], []byte("GET /")) + 4, // includes leading slash
		pathend:   bytes.Index((*data)[:length], []byte("?")),
	}

	if p.pathend != -1 && p.pathend < p.URLend { // if the ? is part of a query then parse it
		p.Params = strings.Split(string((*data)[p.pathend+1:p.URLend]), "&")
		p.Path = string((*data)[p.pathstart:p.pathend])
	} else {
		p.Path = string((*data)[p.pathstart:p.URLend])
	}

	return p
}
