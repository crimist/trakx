// +build !darwin

package shared

import (
	"bytes"
	"syscall"
)

func ulimitBugged() bool {
	var uname syscall.Utsname
	syscall.Uname(&uname)
	release := make([]byte, len(uname.Release))
	for i := range uname.Release {
		release[i] = byte(uname.Release[i])
	}
	return bytes.Contains(release, []byte("Microsoft"))
}
