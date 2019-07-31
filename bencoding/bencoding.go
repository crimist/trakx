package bencoding

import (
	"errors"
	"fmt"
	"reflect"
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

type Dict struct {
	encoded string

	finished bool
}

// NewDict creates a new dictionary
func NewDict() Dict {
	dict := Dict{}
	dict.encoded += "d"
	return dict
}

// Add x
func (d *Dict) Add(key string, v interface{}) error {
	if d.finished {
		return errors.New("Add after Get")
	}

	// Add the key
	d.encoded += str(key)

	switch v := v.(type) {
	case string:
		d.encoded += str(v)
	case []byte:
		d.encoded += str(string(v))
	case []string:
		r := reflect.ValueOf(v)
		slice := make([]string, r.Len())
		for i := 0; i < r.Len(); i++ {
			slice[i] = r.Index(i).String()
		}
		d.encoded += list(slice...)
	case map[string]interface{}:
		dict := NewDict()
		for k, v := range v {
			dict.Add(k, v)
		}
		d.encoded += dict.Get()
	case map[string]map[string]int32:
		dict := NewDict()
		for k, v := range v {
			dict.Add(k, v)
		}
		d.encoded += dict.Get()
	case map[string]int32:
		dict := NewDict()
		for k, v := range v {
			dict.Add(k, v)
		}
		d.encoded += dict.Get()
	case int, int8, int16, int32, int64:
		d.encoded += integer(v)
	case uint, uint8, uint16, uint32, uint64:
		d.encoded += integer(v)
	default:
		return errors.New("Invalid type")
	}

	return nil
}

// Get ends the dicts and returns it as a string
func (d *Dict) Get() string {
	if !d.finished {
		d.encoded += "e"
		d.finished = true
	}
	return d.encoded
}
