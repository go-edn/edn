// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"testing"
)

func TestEncoding(t *testing.T) {
	testEncode(t, Keyword("foo"), ":foo")
	testEncode(t, Symbol("foo"), "foo")
	testEncode(t, "foo", `"foo"`)

	testEncode(t, 1, "1")
	testEncode(t, -1.2, "-1.2")
	testEncode(t, 3.0, "3.0")
	testEncode(t, 10E12, "1e+13")

	var val interface{}

	val = struct {
		Foo int `edn:"quux"`
	}{10}
	testEncode(t, val, "{:quux 10}")

	val = struct {
		Bar [2]int `edn:",sym,list"`
	}{[2]int{3, 4}}
	testEncode(t, val, "{bar(3 4)}")

	val = struct {
		Comp string `edn:",str"`
	}{"ressed"}
	testEncode(t, val, `{"comp""ressed"}`)

	val = struct {
		TheSet [2]int   `edn:"the-set,set,sym"`
		Slice  []string `edn:",set"`
	}{
		[2]int{3, 4},
		[]string{"foo", "bar"},
	}
	testEncode(t, val, `{the-set #{3 4}:slice #{"foo""bar"}}`)

	val = map[float64]struct{}{
		1.0: {},
	}
	testEncode(t, val, `#{1.0}`)

	val = Tag{
		Tagname: "floatset",
		Value: map[float64]struct{}{
			1.0: {},
		},
	}
	testEncode(t, val, `#floatset #{1.0}`)

	val = Tag{
		Tagname: "some/tag",
		Value:   1,
	}
	testEncode(t, val, `#some/tag 1`)

	val = Tag{
		Tagname: "some/tag",
		Value:   struct{ X int }{1},
	}
	testEncode(t, val, `#some/tag{:x 1}`)

	val = Tag{
		Tagname: "a",
		Value: Tag{
			Tagname: "b",
			Value: Tag{
				Tagname: "c",
				Value:   nil,
			},
		},
	}
	testEncode(t, val, `#a #b #c nil`)

	val = Tag{
		Tagname: "a",
		Value: Tag{
			Tagname: "b",
			Value:   1,
		},
	}
	testEncode(t, val, `#a #b 1`)

	val = Tag{
		Tagname: "a",
		Value: Tag{
			Tagname: "b",
			Value:   "c",
		},
	}
	testEncode(t, val, `#a #b"c"`)

	val = Tag{
		Tagname: "a",
		Value:   []int{1},
	}
	testEncode(t, val, `#a[1]`)

	val = Tag{
		Tagname: "a",
		Value: Tag{
			Tagname: "b",
			Value:   []int{1},
		},
	}
	testEncode(t, val, `#a #b[1]`)

	val = Tag{
		Tagname: "some/tag",
		Value: Tag{
			Tagname: "inner",
			Value:   struct{ X int }{1},
		},
	}
	testEncode(t, val, `#some/tag #inner{:x 1}`)

	val = A{}
	testEncode(t, val, `#tag/a{:x 1}`)

	val = Tag{
		Tagname: "outer",
		Value:   A{},
	}
	testEncode(t, val, `#outer #tag/a{:x 1}`)
}

func testEncode(t *testing.T, val interface{}, expects string) {
	bs, err := Marshal(val)
	if err != nil {
		t.Errorf("Unexpected error marshalling %q: %s", val, err.Error())
	} else if !bytes.Equal([]byte(expects), bs) {
		t.Errorf("Expected to see '%s', but got '%s' instead", expects, string(bs))
	}
}

type A struct{}

func (a A) MarshalEDN() ([]byte, error) {
	t := Tag{
		Tagname: "tag/a",
		Value:   struct{ X int }{1},
	}
	return Marshal(t)
}

func TestJSONEncoding(t *testing.T) {
	jsonOnly := struct {
		Data string `json:"json"`
	}{Data: "hi"}
	jsonAndEdn := struct {
		Data string `json:"json" edn:"edn"`
	}{Data: "hi"}

	UseJSONAsFallback(false)
	testEncode(t, jsonOnly, `{:data"hi"}`)
	testEncode(t, jsonAndEdn, `{:edn"hi"}`)

	UseJSONAsFallback(true)
	testEncode(t, jsonOnly, `{:json"hi"}`)
	testEncode(t, jsonAndEdn, `{:edn"hi"}`)

	UseJSONAsFallback(false)
	testEncode(t, jsonOnly, `{:data"hi"}`)
	testEncode(t, jsonAndEdn, `{:edn"hi"}`)
}
