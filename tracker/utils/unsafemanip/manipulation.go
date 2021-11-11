package unsafemanip

import (
	"reflect"
	"unsafe"
)

// StringToBytes converts string to byte slice without escaping to the heap.
func StringToBytes(s string) []byte {
	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&s))
	hdr.Cap = len(s)
	return *(*[]byte)(unsafe.Pointer(&hdr))
}

// StringToBytesFast converts string to byte slice without escaping to the heap.
// NOTE: It is not guarenteed to set the cap correctly.
func StringToBytesFast(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

// SetSliceLen sets a byte slices length and returns its previous length.
// Deprecated: You should use `bytes = bytes[:length]` instead as performance is identical.
func SetSliceLen(s *[]byte, l int) (prevLen int) {
	prevLen = len(*s)
	header := (*reflect.SliceHeader)(unsafe.Pointer(s))
	header.Len = l

	return prevLen
}

// SetStringLen sets a strings length and returns its previous length.
func SetStringLen(s *string, l int) (prevLen int) {
	prevLen = len(*s)
	header := (*reflect.StringHeader)(unsafe.Pointer(s))
	header.Len = l

	return prevLen
}
