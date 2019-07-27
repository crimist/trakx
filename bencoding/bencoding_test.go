package bencoding_test

import (
	"testing"

	"github.com/syc0x00/trakx/bencoding"
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

func TestDictionary(t *testing.T) {
	dictEncode := bencoding.Dictionary("cow moo", "spam eggs")
	if dictEncode != "d3:cow3:moo4:spam4:eggse" {
		t.Errorf("Expected d3:cow3:moo4:spam4:eggse got %s", dictEncode)
	}
	dictEncode2 := bencoding.Dictionary("spam a b")
	if dictEncode2 != "d4:spaml1:a1:bee" {
		t.Errorf("Expected d4:spaml1:a1:bee got %s", dictEncode2)
	}
	dictEncode3 := bencoding.Dictionary("publisher bob", "publisher-webpage www.example.com", "publisher.location home")
	if dictEncode3 != "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee" {
		t.Errorf("Expected d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee got %s", dictEncode3)
	}
}

func TestDict(t *testing.T) {
	d := bencoding.NewDict()
	d.Add("cow", "moo")
	d.Add("spam", "eggs")
	dictEncode := d.Get()
	if dictEncode != "d3:cow3:moo4:spam4:eggse" {
		t.Errorf("Expected d3:cow3:moo4:spam4:eggse got %s", dictEncode)
	}
	d = bencoding.NewDict()
	d.Add("spam", []string{"a", "b"})
	dictEncode = d.Get()
	if dictEncode != "d4:spaml1:a1:bee" {
		t.Errorf("Expected d4:spaml1:a1:bee got %s", dictEncode)
	}
	d = bencoding.NewDict()
	d.Add("publisher", "bob")
	d.Add("publisher-webpage", "www.example.com")
	d.Add("publisher.location", "home")
	dictEncode = d.Get()
	if dictEncode != "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee" {
		t.Errorf("Expected d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee got %s", dictEncode)
	}
}
