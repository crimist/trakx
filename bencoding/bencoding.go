package bencoding

import "fmt"

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
		// str
	}
	encoded += "e"
	return encoded
}
