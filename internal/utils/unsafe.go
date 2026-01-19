package utils

import "unsafe"

// StringToBytesUnsafe converts String to []byte without copying
func StringToBytesUnsafe(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToStringUnsafe converts []byte to String without copying
func BytesToStringUnsafe(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
