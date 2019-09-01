package http

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/url"
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

// I wrote a shitty custom parser because the normal url.Parse().Values()
// creates a map of params which is very expensive with memory
func parse(data []byte) (parsed, error) {
	// uTorrent sometimes encodes the request in b64
	if bytes.HasSuffix(data, []byte("==")) {
		b := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		if _, err := base64.StdEncoding.Decode(b, data); err == nil {
			return parse(b)
		}
	}

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

		var err error
		for i := 0; i < len(p.Params); i++ {
			p.Params[i], err = url.QueryUnescape(p.Params[i])
			if err != nil {
				return p, errors.New(errorInvalid + ": " + err.Error())
			}
		}
		p.Path = string(data[p.pathstart:p.pathend])
	} else {
		p.Path = string(data[p.pathstart:p.URLend])
	}

	return p, nil
}
