/*
	All byte slice and string manipulation functions in this package use unsafe to modify the values. Use these functions with caution.
*/
package shared

import (
	"reflect"
	"unsafe"
)

// StringToBytes converts string to byte slice without escaping to the heap
func StringToBytes(s string) []byte {
	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&s))
	hdr.Cap = len(s)
	return *(*[]byte)(unsafe.Pointer(&hdr))
}

// StringToBytesFast converts string to byte slice without escaping to the heap but doesn't the cap properly
func StringToBytesFast(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

// SetSliceLen sets a byte slices length value.
// WARNING: You should use `bytes = bytes[:length]` instead, it's just as fast.
func SetSliceLen(s *[]byte, l int) (old int) {
	old = len(*s)
	header := (*reflect.SliceHeader)(unsafe.Pointer(s))
	header.Len = l
	return
}

// SetStringLen sets a strings length value
func SetStringLen(s *string, l int) int {
	oldlen := len(*s)
	header := (*reflect.StringHeader)(unsafe.Pointer(s))
	header.Len = l

	return oldlen
}
