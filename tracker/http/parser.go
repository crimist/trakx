package http

import (
	"bytes"
	"encoding/base64"

	"github.com/crimist/trakx/utils"
	"github.com/pkg/errors"
)

const (
	maxParameters = 45 // support scrapes with up to 40 `info_hash` params
)

var (
	invalidParse = errors.New("invalid parse")
)

type (
	parameters  [maxParameters][]byte
	requestData struct {
		Path       string
		Parameters parameters
		UrlEnd     int
		Method     string
		pathStart  int
		pathEnd    int
	}
)

// Custom HTTP parser only supports GET request and up to `maxparams` params but uses no heap memory
func parse(data []byte) (requestData, error) {
	// workaround: uTorrent sometimes encodes scrape requests in base64, `R0VU` = `GET `
	if bytes.HasPrefix(data, []byte("R0VU")) {
		decoded, err := base64.StdEncoding.Decode(data, data)
		if err != nil {
			return requestData{}, errors.Wrap(err, "failed to decode base64 encoded request")
		}
		data = data[:decoded]
	}

	p := requestData{
		UrlEnd:    bytes.Index(data, []byte(" HTTP/")),
		pathStart: bytes.Index(data, []byte("GET /")) + 4, // includes leading slash
		pathEnd:   bytes.Index(data, []byte("?")),
	}

	methodEnd := bytes.Index(data, []byte(" /"))
	if methodEnd == -1 {
		return requestData{}, errors.Wrap(invalidParse, "method end not found")
	}

	p.Method = utils.BytesToStringUnsafe(data[:methodEnd])

	if p.UrlEnd == -1 {
		return requestData{}, errors.Wrap(invalidParse, "url end not found")
	}

	// less than "GET / HTTP..."
	if p.UrlEnd < 5 {
		return requestData{}, errors.Wrap(invalidParse, "message too short to be valid")
	}

	// pathstart should come before URLend
	if p.pathStart > p.UrlEnd {
		return requestData{}, errors.Wrap(invalidParse, "path start after URL end")
	}

	// if the ? is part of a query then parse it
	if p.pathEnd != -1 && p.pathEnd < p.UrlEnd {
		if p.pathEnd < p.pathStart {
			return requestData{}, errors.Wrap(invalidParse, "path end before path start")
		}

		paramsBytes := data[p.pathEnd+1 : p.UrlEnd]

		var pos, pIndex int
		for i := 0; i < len(paramsBytes) && pIndex < maxParameters; i++ {
			if paramsBytes[i] == '&' {
				p.Parameters[pIndex] = paramsBytes[pos:i]
				pos = i + 1
				pIndex++
			} else if i == len(paramsBytes)-1 {
				p.Parameters[pIndex] = paramsBytes[pos : i+1]
			}
		}

		if pIndex == maxParameters {
			pIndex--
		}

		for i := 0; i <= pIndex; i++ {
			p.Parameters[i] = unescapeFast(p.Parameters[i])

			// nil if escape was invalid
			if p.Parameters[i] == nil {
				return requestData{}, errors.Wrap(invalidParse, "invalid url escape sequence")
			}
		}

		p.Path = utils.BytesToStringUnsafe(data[p.pathStart:p.pathEnd])
	} else {
		p.Path = utils.BytesToStringUnsafe(data[p.pathStart:p.UrlEnd])
	}

	return p, nil
}

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
func unescapeFast(msg []byte) []byte {
	l := len(msg)

	for i := 0; i < l; i++ {
		if msg[i] == '%' {
			// make sure there's 2 escape chars
			if i+2 >= l {
				return nil
			}

			// get hex chars
			a := fromHexChar(msg[i+1])
			b := fromHexChar(msg[i+2])
			// change percent to real byte
			msg[i] = (a << 4) | b

			// shift everything left by 2
			for x := i; x < len(msg)-3; x++ {
				msg[x+1] = msg[x+3]
			}

			// decrease slice length by 2
			l -= 2
		}
	}

	msg = msg[:l]
	return msg
}
