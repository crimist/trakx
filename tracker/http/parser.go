package http

import (
	"bytes"
	"encoding/base64"
	"unsafe"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/pkg/errors"
)

const (
	parseOk      parsedCode = iota
	parseInvalid parsedCode = iota
	maxparams               = 45 // support scrapes with up to 40 `info_hash` params
)

type (
	parsedCode uint8
	params     [maxparams][]byte
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

		var err error
		for i := 0; i <= pIndex; i++ {
			p.Params[i] = unescapeFast(p.Params[i])
			// Leaving this dead error check in *somehow* makes the benchmark faster so I'm going to leave it for now
			// TODO: check generated asm cause this makes no sense
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

// unescapes url escaped string
// very fast but no checks
func unescapeFast(bs []byte) []byte {
	l := len(bs)

	for i := 0; i < l; i++ {
		// match escape
		if bs[i] == '%' {
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

		// fmt.Println(string(bs))
	}

	shared.SetSliceLen(&bs, l)
	// fmt.Println(hex.Dump(bs))

	return bs
}
