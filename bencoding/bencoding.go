package bencoding

import (
	"errors"
	"fmt"
	"reflect"
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

// Dict x
type Dict struct {
	encoded string
}

// NewDict x
func NewDict() Dict {
	dict := Dict{}
	dict.encoded += "d"
	return dict
}

// Add x
func (d *Dict) Add(key string, v interface{}) error {
	d.encoded += String(key) // Add the key

	switch v := v.(type) {
	case string:
		d.encoded += String(v)
	case []string:
		r := reflect.ValueOf(v)
		slice := make([]string, r.Len())
		for i := 0; i < r.Len(); i++ {
			slice[i] = r.Index(i).String()
		}
		d.encoded += List(slice...)
	case map[string]interface{}:
		dict := NewDict()
		for k, v := range v {
			dict.Add(k, v)
		}
		d.encoded += dict.Get()
	case int, int8, int16, int32, int64:
		d.encoded += Integer(int(reflect.ValueOf(v).Int()))
	case uint, uint8, uint16, uint32, uint64:
		d.encoded += Integer(int(reflect.ValueOf(v).Uint()))
	default:
		return errors.New("Invalid type")
	}

	return nil
}

// Get ends the dicts and returns it as a string
func (d *Dict) Get() string {
	d.encoded += "e"
	return d.encoded
}
