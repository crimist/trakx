package bencoding

import (
	"fmt"
	"strconv"
	"strings"
)

func str(str string) string {
	return fmt.Sprintf("%d:%s", len(str), str)
}

func integer(num interface{}) string {
	return fmt.Sprintf("i%de", num)
}

func list(list ...string) string {
	encoded := "l"
	for _, str := range list {
		encoded += fmt.Sprintf("%d:%s", len(str), str)
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
