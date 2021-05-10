package bencoding

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
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

// Dictionary holds the encoded key value pairs
type Dictionary struct {
	buf []byte
}

// NewDict returns a new initialized dictionary
func NewDict() *Dictionary {
	d := dictChan.Get()
	d.write("d")
	return d
}

func (d *Dictionary) write(s string) {
	d.buf = append(d.buf, s...)
}

func (d *Dictionary) writeBytes(b []byte) {
	d.buf = append(d.buf, b[:]...)
}

func (d *Dictionary) reset() {
	/* TODO: Consider implementing a maximum size check to prevent large allocations from permanently increasing memory
	if len(d.buf) > 10240 {
		d.buf = nil
	}
	*/
	d.buf = d.buf[:0]
}

// String writes a string to the dictionary
func (d *Dictionary) String(key string, v string) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + strconv.FormatInt(int64(len(v)), 10) + ":" + v)
}

// StringBytes writes a byte slice to the dictionary
func (d *Dictionary) StringBytes(key string, v []byte) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + strconv.FormatInt(int64(len(v)), 10) + ":")
	d.writeBytes(v)
}

// Int64 writes an int64 to the dictionary
func (d *Dictionary) Int64(key string, v int64) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + "i" + strconv.FormatInt(v, 10) + "e")
}

// Dictionary writes an encoded dictionary to the dictionary
func (d *Dictionary) Dictionary(key string, v string) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + v)
}

// StartDict starts an embeded dictionary with the given key.
// It must be followed by an EndDict() call otherwise the bencode will be invalid
func (d *Dictionary) StartDict(key string) *Dictionary {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key)
	d.write("d")
	return d
}

// EndDict ends the embeded dictionary. StartDict should be called before
func (d *Dictionary) EndDict() {
	d.write("e")
}

// Any attempts to decode all types and write them to the dictionary
//
// This function is far slower than the rest and should be avoided if possible
func (d *Dictionary) Any(key string, v interface{}) error {
	// Add the key
	d.write(str(key))

	switch v := v.(type) {
	case string:
		d.write(str(v))
	case []byte:
		d.write(str(string(v)))
	case []string:
		r := reflect.ValueOf(v)
		slice := make([]string, r.Len())
		for i := 0; i < r.Len(); i++ {
			slice[i] = r.Index(i).String()
		}
		d.write(list(slice...))
	case map[string]interface{}:
		dict := NewDict()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.write(dict.Get())
	case map[string]map[string]int32:
		dict := NewDict()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.write(dict.Get())
	case map[string]int32:
		dict := NewDict()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.write(dict.Get())
	case int, int8, int16, int32, int64:
		d.write(integer(v))
	case uint, uint8, uint16, uint32, uint64:
		d.write(integer(v))
	default:
		return errors.New("invalid type")
	}

	return nil
}

// Get returns the encoded dictionary as a string. The dictionary cannot be used after this is called.
func (d *Dictionary) Get() string {
	d.write("e")
	s := *(*string)(unsafe.Pointer(&d.buf))
	d.reset()

	dictChan.Put(d)
	return s
}

// GetBytes returns the encoded dictionary as a byte slice. The dictionary cannot be used after this is called.
func (d *Dictionary) GetBytes() []byte {
	d.write("e")
	b := d.buf
	d.reset()

	dictChan.Put(d)
	return b
}
