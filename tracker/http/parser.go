package http

import (
	"bytes"
	"errors"
	"strings"
)

const errorInvalid = "Invalid Request"

type parsed struct {
	Path   string
	Params []string
	URLend int
	Method string

	pathstart int
	pathend   int
}

func parse(data []byte) (parsed, error) {
	p := parsed{
		URLend:    bytes.Index(data, []byte(" HTTP/")),
		pathstart: bytes.Index(data, []byte("GET /")) + 4, // includes leading slash
		pathend:   bytes.Index(data, []byte("?")),
	}

	methodend := bytes.Index(data, []byte(" /"))
	if methodend == -1 {
		return p, errors.New(errorInvalid)
	}
	p.Method = string(data[:methodend])

	if p.URLend == -1 {
		return p, errors.New(errorInvalid)
	}

	if p.pathend != -1 && p.pathend < p.URLend { // if the ? is part of a query then parse it
		p.Params = strings.Split(string(data[p.pathend+1:p.URLend]), "&")
		p.Path = string(data[p.pathstart:p.pathend])
	} else {
		p.Path = string(data[p.pathstart:p.URLend])
	}

	return p, nil
}
