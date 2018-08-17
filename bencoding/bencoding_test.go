package bencoding_test

import (
	"testing"

	"github.com/Syc0x00/Trakx/bencoding"
)

func TestString(t *testing.T) {
	strEncode := bencoding.String("test")
	if strEncode != "4:test" {
		t.Error("Failed to encode string")
	}

	str2Encode := bencoding.String("")
	if str2Encode != "0:" {
		t.Error("Failed to encode string")
	}
}

func TestInt(t *testing.T) {
	intEncode := bencoding.Integer(0)
	if intEncode != "i0e" {
		t.Error("Failed to encode integer")
	}

	int2Encode := bencoding.Integer(-69)
	if int2Encode != "i-69e" {
		t.Error("Failed to encode integer")
	}
}

func TestList(t *testing.T) {
	listEncode := bencoding.List("test", "test2")
	if listEncode != "l4:test5:test2e" {
		t.Error("Failed to encode list")
	}
	list2Encode := bencoding.List()
	if list2Encode != "le" {
		t.Error("Failed to encode list")
	}
}
