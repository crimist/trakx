package http

import (
	"bytes"
	"encoding/base64"
	"net/url"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	parseOk      parsedCode = iota
	parseInvalid parsedCode = iota
	maxparams               = 50
)

type (
	parsedCode uint8
	params     [maxparams]string
	parsed     struct {
		Path      string
		Params    params
		URLend    int
		Method    string
		pathstart int
		pathend   int
	}
)

// Custom HTTP parser - only supports GET request and up to `maxparams` params but uses no heap memory
func parse(data []byte, size int) (parsed, parsedCode, error) {
	// uTorrent sometimes encodes scrape req in b64
	if bytes.HasPrefix(data, []byte("R0VU")) { // R0VUIC9zY3JhcGU/aW5mb19oYXNoPS = GET /scrape?info_hash=
		decoded, err := base64.StdEncoding.Decode(data, data[:size])
		if err != nil {
			return parsed{}, parseOk, errors.New("Failed to decode base64 encoded payload")
		}
		data = data[:decoded] // this only modifies the local copy since `len int` vs `data uintptr`
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

	if p.URLend < 5 { // less than "GET / HTTP..."
		return parsed{}, parseInvalid, nil
	}

	if p.pathend != -1 && p.pathend < p.URLend { // if the ? is part of a query then parse it
		if p.pathend < p.pathstart {
			return parsed{}, parseInvalid, nil
		}

		tmp = data[p.pathend+1 : p.URLend]
		params := *(*string)(unsafe.Pointer(&tmp))

		var pos, pIndex int
		for i := 0; i < len(params) && pIndex < maxparams; i++ {
			if params[i] == '&' {
				p.Params[pIndex] = params[pos:i]
				pos = i + 1
				pIndex++
			} else if i == len(params)-1 {
				p.Params[pIndex] = params[pos : i+1]
			}
		}

		if pIndex == maxparams {
			pIndex--
		}

		var err error
		for i := 0; i <= pIndex; i++ {
			p.Params[i], err = unescape(p.Params[i])
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

func unescape(s string) (string, error) {
	return url.QueryUnescape(s)
}

// func unescape(s string) (string, error) {
// 	for _, c := range s {

// 	}
// }
