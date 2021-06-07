package bencoding

import (
	"fmt"
	"strconv"
	"strings"
)

func str(str string) string {
	return strconv.Itoa(len(str)) + ":" + str
}

func integer(num interface{}) (s string) {
	/* // 69.12 ns/op 31 B/op
	num64 := func(n interface{}) interface{} {
		switch n := n.(type) {
		case int:
			return int64(n)
		case int8:
			return int64(n)
		case int16:
			return int64(n)
		case int32:
			return int64(n)
		case int64:
			return int64(n)
		case uint:
			return uint64(n)
		case uintptr:
			return uint64(n)
		case uint8:
			return uint64(n)
		case uint16:
			return uint64(n)
		case uint32:
			return uint64(n)
		case uint64:
			return uint64(n)
		}
		return nil
	}

	switch i := num64(num); i.(type) {
	case int64:
		s = "i" + strconv.FormatInt(i.(int64), 10) + "e"
	case uint64:
		s = "i" + strconv.FormatUint(i.(uint64), 10) + "e"
	}
	*/

	/* // 69.08 ns/op 31 B/op
	switch t := num.(type) {
	case int, int8, int16, int32, int64:
		s = "i" + strconv.FormatInt(reflect.ValueOf(t).Int(), 10) + "e"
	case uint, uint8, uint16, uint32, uint64:
		s = "i" + strconv.FormatUint(reflect.ValueOf(t).Uint(), 10) + "e"
	}
	*/

	// 88.22 ns/op 23 B/op
	s = fmt.Sprintf("i%de", num)
	return s
}

func list(list ...string) string {
	encoded := "l"
	for _, s := range list {
		encoded += str(s)
	}
	encoded += "e"
	return encoded
}

func dict(dict ...string) string {
	encoded := "d"
	for _, part := range dict {
		parts := strings.Split(part, " ")
		// parts[0] key
		// parts[1] value
		encoded += str(parts[0]) // Add the key

		if len(parts) > 2 { // its a list
			// remove index since weve already added it
			parts = append(parts[:0], parts[0+1:]...)
			encoded += list(parts...)
		} else { // string or int
			valInt, err := strconv.Atoi(parts[1])
			if err != nil {
				encoded += integer(valInt)
			} else {
				encoded += str(parts[1])
			}
		}
	}
	encoded += "e"
	return encoded
}
