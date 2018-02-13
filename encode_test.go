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
	testEncode(t, 3.0, "3.")
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
}

func testEncode(t *testing.T, val interface{}, expects string) {
	bs, err := Marshal(val)
	if err != nil {
		t.Errorf("Unexpected error marshalling %q: %s", val, err.Error())
	} else if !bytes.Equal([]byte(expects), bs) {
		t.Errorf("Expected to see '%s', but got '%s' instead", expects, string(bs))
	}
}
