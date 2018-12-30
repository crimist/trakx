package tracker

import "encoding/base64"

// EncodeHash encodes the hash into something we can actually store in the db
func EncodeHash(hash string) string {
	return base64.StdEncoding.EncodeToString([]byte(hash))
}
