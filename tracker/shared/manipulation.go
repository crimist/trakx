package shared

import (
	"reflect"
	"unsafe"
)

// StringToBytes converts string to byte slice without escaping to the heap
func StringToBytes(s string) []byte {
	sh := reflect.SliceHeader{Data: 0, Len: len(s), Cap: len(s)}
	sh = *(*reflect.SliceHeader)(unsafe.Pointer(&s))
	sh.Cap = len(s)
	return *(*[]byte)(unsafe.Pointer(&sh))
}

// StringToBytesFast converts string to byte slice without escaping to the heap but isn't guaranteed to set the cap properly
func StringToBytesFast(s *string) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(unsafe.Pointer(s))))
}

func SetSliceLen(s *[]byte, l int) int {
	oldlen := len(*s)
	header := (*reflect.SliceHeader)(unsafe.Pointer(s))
	header.Len = l

	return oldlen
}

func SetStringLen(s *string, l int) int {
	oldlen := len(*s)
	header := (*reflect.StringHeader)(unsafe.Pointer(s))
	header.Len = l

	return oldlen
}
