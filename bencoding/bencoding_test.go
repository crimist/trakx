package bencoding_test

import (
	"testing"

	"github.com/crimist/trakx/bencoding"
)

func TestBencodingInt64(t *testing.T) {
	var cases = []struct {
		key    string
		val    int64
		result string
	}{
		{"small", 11, "d5:smalli11ee"},
		{"big", 0xFFFFFFFFFFF, "d3:bigi17592186044415ee"},
		{"negative", -11, "d8:negativei-11ee"},
	}

	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			d := bencoding.NewDict()
			d.Int64(c.key, c.val)
			if val := d.Get(); val != c.result {
				t.Errorf("Bad encode: '%s' should be '%s'", val, c.result)
			}
		})
	}
}

func TestBencodingString(t *testing.T) {
	var cases = []struct {
		key    string
		val    string
		result string
	}{
		{"short", "hello", "d5:short5:helloe"},
		{"long", "really_long_string_that_has_lots_of_shit_in_it", "d4:long46:really_long_string_that_has_lots_of_shit_in_ite"},
		{"specialchars", "this_has controlchars\n", "d12:specialchars22:this_has controlchars\ne"}, // TODO: real special chars test
	}

	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			d := bencoding.NewDict()
			d.String(c.key, c.val)
			if val := d.Get(); val != c.result {
				t.Errorf("Bad encode: '%s' should be '%s'", val, c.result)
			}
		})
	}
}

func TestBencodingAny(t *testing.T) {
	type kvpair struct {
		key string
		val interface{}
	}

	var cases = []struct {
		name   string
		pairs  []kvpair
		result string
	}{
		{"strings", []kvpair{{"cow", "moo"}, {"spam", "eggs"}}, "d3:cow3:moo4:spam4:eggse"},
		{"string array", []kvpair{{"spam", []string{"a", "b"}}}, "d4:spaml1:a1:bee"},
		{"mixed", []kvpair{{"strkey", "strval"}, {"strarray", []string{"arr1", "arr2"}}, {"intval", 123456}}, "d6:strkey6:strval8:strarrayl4:arr14:arr2e6:intvali123456ee"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := bencoding.NewDict()
			for _, pair := range c.pairs {
				d.Any(pair.key, pair.val)
			}
			if val := d.Get(); val != c.result {
				t.Errorf("Bad encode: '%s' should be '%s'", val, c.result)
			}
		})
	}
}

func BenchmarkBencodeAnnounceBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		d.Int64("interval", int64(1))
		d.Int64("complete", int64(1))
		d.Int64("incomplete", int64(1))
		d.String("peers", "\x01\x02\x03\x04\x05")
		_ = d.Get()
	}
}

func BenchmarkBencodeString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		d.String("key", "value")
		_ = d.Get()
	}
}

func BenchmarkBencodeInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		d.Int64("key", 0x1337)
		_ = d.Get()
	}
}

func BenchmarkBencodeStringGetBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		d.String("key", "value")
		_ = d.GetBytes()
	}
}

func BenchmarkBencodeInt64GetBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		d.Int64("key", 0x1337)
		_ = d.GetBytes()
	}
}

func BenchmarkBencodeAnyString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		d.Any("key", "value")
		_ = d.Get()
	}
}

func BenchmarkBencodeDict(b *testing.B) {
	for i := 0; i < b.N; i++ {
		d := bencoding.NewDict()
		embedDict := bencoding.NewDict()
		embedDict.String("key", "value")
		d.Dictionary("dict", embedDict.Get())
		_ = d.Get()
	}
}
