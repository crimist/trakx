package tracker

import (
	"fmt"
)

// EncodeInfoHash encodes the info hash of the torrent into a hex string for MySQL
func EncodeInfoHash(hash string) string {
	return fmt.Sprintf("Hash_%X", hash)
}
