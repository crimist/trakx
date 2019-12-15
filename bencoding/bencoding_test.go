package bencoding_test

import (
	"testing"

	"github.com/crimist/trakx/bencoding"
)

func TestDict(t *testing.T) {
	d := bencoding.NewDict()
	d.Any("cow", "moo")
	d.Any("spam", "eggs")
	dictEncode := d.Get()
	if dictEncode != "d3:cow3:moo4:spam4:eggse" {
		t.Errorf("Expected d3:cow3:moo4:spam4:eggse got %s", dictEncode)
	}

	d = bencoding.NewDict()
	d.Any("spam", []string{"a", "b"})
	dictEncode = d.Get()
	if dictEncode != "d4:spaml1:a1:bee" {
		t.Errorf("Expected d4:spaml1:a1:bee got %s", dictEncode)
	}

	d = bencoding.NewDict()
	d.Any("publisher", "bob")
	d.Any("publisher-webpage", "www.example.com")
	d.Any("publisher.location", "home")
	dictEncode = d.Get()
	if dictEncode != "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee" {
		t.Errorf("Expected d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee got %s", dictEncode)
	}
}

func TestDictNew(t *testing.T) {
	d := bencoding.NewDict()
	d.String("cow", "moo")
	d.String("spam", "eggs")
	dictEncode := d.Get()
	if dictEncode != "d3:cow3:moo4:spam4:eggse" {
		t.Errorf("Expected d3:cow3:moo4:spam4:eggse got %s", dictEncode)
	}

	d = bencoding.NewDict()
	d.String("publisher", "bob")
	d.String("publisher-webpage", "www.example.com")
	d.String("publisher.location", "home")
	dictEncode = d.Get()
	if dictEncode != "d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee" {
		t.Errorf("Expected d9:publisher3:bob17:publisher-webpage15:www.example.com18:publisher.location4:homee got %s", dictEncode)
	}
}
