//go:build !heroku
// +build !heroku

package http

func parseForwarded(data []byte) (forwarded bool, ip []byte) { return false, nil }
