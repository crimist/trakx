package http

import (
	"bytes"
	"encoding/base64"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	maxparams = 45 // support scrapes with up to 40 `info_hash` params
)

var (
	invalidParse = errors.New("invalid parse")
)

type (
	params [maxparams][]byte
	parsed struct {
		Path      string
		Params    params
		URLend    int
		Method    string
		pathstart int
		pathend   int
	}
)

// Custom HTTP parser
// only supports GET request and up to `maxparams` params but uses no heap memory
func parse(data []byte, size int) (parsed, error) {
	// uTorrent sometimes encodes scrape req in b64
	if bytes.HasPrefix(data, []byte("R0VU")) { // R0VUIC9zY3JhcGU/aW5mb19oYXNoPS = GET /scrape?info_hash=
		decoded, err := base64.StdEncoding.Decode(data, data[:size])
		if err != nil {
			return parsed{}, errors.Wrap(err, "failed to decode base64 encoded request")
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
		return parsed{}, invalidParse
	}

	tmp := data[:methodend]
	p.Method = *(*string)(unsafe.Pointer(&tmp))

	if p.URLend == -1 {
		return parsed{}, invalidParse
	}

	if p.URLend < 5 { // less than "GET / HTTP..."
		return parsed{}, invalidParse
	}

	if p.pathend != -1 && p.pathend < p.URLend { // if the ? is part of a query then parse it
		if p.pathend < p.pathstart {
			return parsed{}, invalidParse
		}

		paramsBytes := data[p.pathend+1 : p.URLend]

		var pos, pIndex int
		for i := 0; i < len(paramsBytes) && pIndex < maxparams; i++ {
			if paramsBytes[i] == '&' {
				p.Params[pIndex] = paramsBytes[pos:i]
				pos = i + 1
				pIndex++
			} else if i == len(paramsBytes)-1 {
				p.Params[pIndex] = paramsBytes[pos : i+1]
			}
		}

		if pIndex == maxparams {
			pIndex--
		}

		for i := 0; i <= pIndex; i++ {
			p.Params[i] = unescapeFast(p.Params[i])

			// nil if escape was invalid
			if p.Params[i] == nil {
				return parsed{}, invalidParse
			}
		}

		tmp = data[p.pathstart:p.pathend]
		p.Path = *(*string)(unsafe.Pointer(&tmp))
	} else {
		tmp = data[p.pathstart:p.URLend]
		p.Path = *(*string)(unsafe.Pointer(&tmp))
	}

	return p, nil
}

// from stdlib
// https://github.com/golang/go/blob/2ebe77a2fda1ee9ff6fd9a3e08933ad1ebaea039/src/encoding/hex/hex.go#L84
func fromHexChar(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}

	return 0
}

// unescapeFast unescapes url encoded []byte
// returns nil if escape is invalid
func unescapeFast(bs []byte) []byte {
	l := len(bs)

	for i := 0; i < l; i++ {
		// match escape
		if bs[i] == '%' {
			// make sure there's 2 escape chars
			if i+2 >= l {
				return nil
			}

			// get hex chars
			a := fromHexChar(bs[i+1])
			b := fromHexChar(bs[i+2])
			// change percent to real byte
			bs[i] = (a << 4) | b

			// shift everything left by 2
			for x := i; x < len(bs)-3; x++ {
				bs[x+1] = bs[x+3]
			}

			// decrease slice length by 2
			l -= 2
		}
	}

	bs = bs[:l]
	return bs
}
