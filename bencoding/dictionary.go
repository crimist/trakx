package bencoding

import (
	"errors"
	"reflect"
	"strconv"
)

// default Dictionary internal buf length
const bufLen = 32

// Dictionary holds the encoded key value pairs
type Dictionary struct {
	buf []byte
}

// NewDictionary creates and returns a new initialized Dictionary.
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

// Reset resets the Dictionary's underlying byte slice.
func (d *Dictionary) Reset() {
	d.buf = d.buf[:0]
	d.write("d")
}

// String writes a string to the dictionary.
func (d *Dictionary) String(key string, v string) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + strconv.FormatInt(int64(len(v)), 10) + ":" + v)
}

// StringBytes writes a byte slice to the dictionary.
func (d *Dictionary) StringBytes(key string, v []byte) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + strconv.FormatInt(int64(len(v)), 10) + ":")
	d.writeBytes(v)
}

// Int64 writes an int64 to the dictionary.
func (d *Dictionary) Int64(key string, v int64) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + "i" + strconv.FormatInt(v, 10) + "e")
}

// Dictionary writes an encoded dictionary to the dictionary.
func (d *Dictionary) Dictionary(key string, v string) {
	d.write(strconv.FormatInt(int64(len(key)), 10) + ":" + key + v)
}

// StartDictionary begins an embedded dictionary with the given string key.
// EndDictionary() must be called to complete the embedded dictionary before Get() is called.
func (d *Dictionary) StartDictionary(key string) {
	d.startDictionary(len(key))
	d.write(key)
	d.write("d")
}

// StartDictionary begins an embedded dictionary with the given byte slice key.
// EndDictionary() must be called to complete the embedded dictionary before Get() is called.
func (d *Dictionary) StartDictionaryBytes(key []byte) {
	d.startDictionary(len(key))
	d.writeBytes(key)
	d.write("d")
}

func (d *Dictionary) startDictionary(len int) {
	d.write(strconv.FormatInt(int64(len), 10) + ":")
}

// EndDictionary finishes the embedded dictionary. StartDictionary() should be called before this.
func (d *Dictionary) EndDictionary() {
	d.write("e")
}

// BytesliceSlice writes a list of form byte slice slice ([][]byte) to the dictionary.
func (d *Dictionary) BytesliceSlice(key string, slice [][]byte) {
	d.write(str(key) + "l")
	for _, b := range slice {
		d.write(strconv.Itoa(len(b)) + ":")
		d.writeBytes(b)
	}
	d.write("e")
}

// Any tries to write given type to dictionary. It returns an error if it is unabled to write the desired type to the dictionary.
// Any performs far worse than specific type encoding functions and should be avoided when possible.
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
		d.writeBytes(dict.GetBytes())
	case map[string]map[string]int32:
		dict := GetDictionary()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.writeBytes(dict.GetBytes())
	case map[string]int32:
		dict := GetDictionary()
		for k, v := range v {
			dict.Any(k, v)
		}
		d.writeBytes(dict.GetBytes())
	case int, int8, int16, int32, int64:
		d.write(integer(v))
	case uint, uint8, uint16, uint32, uint64:
		d.write(integer(v))
	default:
		return errors.New("failed to write value to dictionary: invalid type")
	}

	return nil
}

// Get returns the encoded dictionary as a string. The dictionary should not be used after.
func (d *Dictionary) Get() string {
	d.write("e")
	s := string(d.buf)

	return s
}

// Get returns the encoded dictionary as a byte slice. The dictionary should not be used after.
func (d *Dictionary) GetBytes() []byte {
	d.write("e")
	b := d.buf

	return b
}
