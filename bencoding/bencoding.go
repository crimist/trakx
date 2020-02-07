package bencoding

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
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

type Dictionary struct {
	builder strings.Builder
}

var dictionaryPool = sync.Pool{New: func() interface{} { return new(Dictionary) }}

// NewDict creates a new dictionary
func NewDict() (d *Dictionary) {
	d = dictionaryPool.Get().(*Dictionary)
	d.builder.WriteString("d")
	return
}

func (d *Dictionary) String(key string, v string) {
	d.builder.WriteString(strconv.FormatInt(int64(len(key)), 10) + ":" + key + strconv.FormatInt(int64(len(v)), 10) + ":" + v)
}

func (d *Dictionary) Int64(key string, v int64) {
	d.builder.WriteString(strconv.FormatInt(int64(len(key)), 10) + ":" + key + "i" + strconv.FormatInt(v, 10) + "e")
}

func (d *Dictionary) Any(key string, v interface{}) error {
	// Add the key
	d.builder.WriteString(str(key))

	switch v := v.(type) {
	case string:
		d.builder.WriteString(str(v))
	case []byte:
		d.builder.WriteString(str(string(v)))
	case []string:
		r := reflect.ValueOf(v)
		slice := make([]string, r.Len())
		for i := 0; i < r.Len(); i++ {
			slice[i] = r.Index(i).String()
		}
		d.builder.WriteString(list(slice...))
	case map[string]interface{}:
		dict := NewDict()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.builder.WriteString(dict.Get())
	case map[string]map[string]int32:
		dict := NewDict()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.builder.WriteString(dict.Get())
	case map[string]int32:
		dict := NewDict()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.builder.WriteString(dict.Get())
	case int, int8, int16, int32, int64:
		d.builder.WriteString(integer(v))
	case uint, uint8, uint16, uint32, uint64:
		d.builder.WriteString(integer(v))
	default:
		return errors.New("Invalid type")
	}

	return nil
}

// Get ends the dicts and returns it as a string
func (d *Dictionary) Get() string {
	d.builder.WriteString("e")
	str := d.builder.String()

	d.builder.Reset()
	dictionaryPool.Put(d)
	return str
}
