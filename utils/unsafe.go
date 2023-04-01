package utils

import (
	"reflect"
	"unsafe"
)

/*
Docs regarding unsafe non copy []byte <-> string conversion:
* https://stackoverflow.com/a/69231355/6389542
* https://go.dev/src/strings/builder.go#L45
*/

// StringToBytesUnsafe converts String to []byte without copying
func StringToBytesUnsafe(s string) []byte {
	const MaxInt32 = 1<<31 - 1
	return (*[MaxInt32]byte)(unsafe.Pointer((*reflect.StringHeader)(
		unsafe.Pointer(&s)).Data))[: len(s)&MaxInt32 : len(s)&MaxInt32]
}

// BytesToStringUnsafe converts []byte to String without copying
func ByteToStringUnsafe(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
