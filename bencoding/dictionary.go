package bencoding

import (
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

// default Dictionary internal buf length
const bufLen = 64

// Dictionary holds the encoded key value pairs
type Dictionary struct {
	buf []byte
}

// NewDictionary returns a new initialized Dictionary
func NewDictionary() *Dictionary {
	var d Dictionary
	d.buf = make([]byte, 0, bufLen)
	d.write("d")
	return &d
}

func (d *Dictionary) write(s string) {
	d.buf = append(d.buf, s...)
}

func (d *Dictionary) writeBytes(b []byte) {
	d.buf = append(d.buf, b[:]...)
}

// Reset resets the Dictionary
func (d *Dictionary) Reset() {
	/* TODO: Consider implementing a maximum size check to prevent large allocations from permanently increasing memory
	if len(d.buf) > 10240 {
		d.buf = nil
	}
	*/
	d.buf = d.buf[:0]
	d.write("d")
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

// StartDict starts an embedded dictionary with the given string key.
// It must be followed by an EndDict() call otherwise the bencode will be invalid
func (d *Dictionary) StartDictionary(key string) {
	d.startDictionary(len(key))
	d.write(key)
	d.write("d")
}

// StartDictBytes starts an embedded dictionary with the given byte slice key.
// It must be followed by an EndDict() call otherwise the bencode will be invalid
func (d *Dictionary) StartDictionaryBytes(key []byte) {
	d.startDictionary(len(key))
	d.writeBytes(key)
	d.write("d")
}

func (d *Dictionary) startDictionary(len int) {
	d.write(strconv.FormatInt(int64(len), 10) + ":")
}

// EndDict ends the embedded dictionary. StartDict should be called before
func (d *Dictionary) EndDictionary() {
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
		dict := GetDictionary()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.write(dict.Get())
	case map[string]map[string]int32:
		dict := GetDictionary()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.write(dict.Get())
	case map[string]int32:
		dict := GetDictionary()
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

	return s
}

// GetBytes returns the encoded dictionary as a byte slice. The dictionary cannot be used after this is called.
func (d *Dictionary) GetBytes() []byte {
	d.write("e")
	b := d.buf

	return b
}
