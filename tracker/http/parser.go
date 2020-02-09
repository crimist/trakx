package http

import (
	"bytes"
	"encoding/base64"
	"net/url"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

type parsedCode uint8

const (
	parseOk      parsedCode = iota
	parseInvalid parsedCode = iota
)

type parsed struct {
	Path      string
	Params    []string
	URLend    int
	Method    string
	pathstart int
	pathend   int
}

// I wrote a shitty custom parser because the normal url.Parse().Values() reates a map of params which is very expensive with memory
// parse returns nil on an invalid request and error when there was an internal error
func parse(data []byte) (parsed, parsedCode, error) {
	// uTorrent sometimes encodes scrape req in b64
	if bytes.HasPrefix(data, []byte("R0VU")) { // R0VUIC9zY3JhcGU/aW5mb19oYXNoPS = GET /scrape?info_hash=
		decoded, err := base64.StdEncoding.Decode(data, data)
		if err != nil {
			return parsed{}, parseOk, errors.New("Failed to decode base64 encoded payload")
		}
		data = data[:decoded]
	}

	p := parsed{
		URLend:    bytes.Index(data, []byte(" HTTP/")),
		pathstart: bytes.Index(data, []byte("GET /")) + 4, // includes leading slash
		pathend:   bytes.Index(data, []byte("?")),
	}

	methodend := bytes.Index(data, []byte(" /"))
	if methodend == -1 {
		return parsed{}, parseInvalid, nil
	}

	tmp := data[:methodend]
	p.Method = *(*string)(unsafe.Pointer(&tmp))

	if p.URLend == -1 {
		return parsed{}, parseInvalid, nil
	}

	if p.URLend < 5 || p.Method != "GET" { // less than "GET / HTTP..."
		return parsed{}, parseInvalid, nil
	}

	if p.pathend != -1 && p.pathend < p.URLend { // if the ? is part of a query then parse it
		if p.pathend < p.pathstart {
			return parsed{}, parseInvalid, nil
		}

		tmp = data[p.pathend+1 : p.URLend]
		p.Params = strings.Split(*(*string)(unsafe.Pointer(&tmp)), "&")

		var err error
		for i := 0; i < len(p.Params); i++ {
			p.Params[i], err = url.QueryUnescape(p.Params[i])
			if err != nil {
				return parsed{}, parseInvalid, nil // failed to escape a param
			}
		}

		tmp = data[p.pathstart:p.pathend]
		p.Path = *(*string)(unsafe.Pointer(&tmp))
	} else {
		tmp = data[p.pathstart:p.URLend]
		p.Path = *(*string)(unsafe.Pointer(&tmp))
	}

	return p, parseOk, nil
}
