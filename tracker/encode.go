package tracker

import "encoding/base64"

// EncodeStr encodes the give string into values that can be stored in the db
func EncodeStr(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
