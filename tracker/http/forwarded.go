// +build !heroku

package http

func getForwarded(data []byte) (bool, []byte) { return false, nil }
