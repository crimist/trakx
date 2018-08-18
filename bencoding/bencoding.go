package bencoding

import (
	"fmt"
	"strconv"
	"strings"
)

// String encodes a string
func String(str string) string {
	return fmt.Sprintf("%d:%s", len(str), str)
}

// Integer encodes an int
func Integer(num int) string {
	return fmt.Sprintf("i%de", num)
}

// List encodes multiple strings into a list
func List(list ...string) string {
	encoded := "l"
	for _, str := range list {
		encoded += fmt.Sprintf("%d:%s", len(str), str)
	}
	encoded += "e"
	return encoded
}

// Dictionarie encodes a dict
// https://github.com/marksamman/bencode ?
func Dictionarie(dict ...string) string {
	encoded := "d"
	for _, part := range dict {
		parts := strings.Split(part, " ")
		// parts[0] key
		// parts[1] value
		encoded += String(parts[0]) // Add the key

		if len(parts) > 2 { // its a list
			// TODO check for dict in dict
			// remove index since weve already added it
			parts = append(parts[:0], parts[0+1:]...)
			encoded += List(parts...)
		} else { // string or int
			valInt, err := strconv.Atoi(parts[1])
			if err != nil {
				encoded += Integer(valInt)
			} else {
				encoded += String(parts[1])
			}
		}
	}
	encoded += "e"
	return encoded
}
