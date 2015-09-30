// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"testing"
)

func TestConvert(t *testing.T) {
	// basic
	checkConvert(t, "foo bar baz", "foo bar baz")
	// preserves correct spacing?
	checkConvert(t, "a,b\nc\td", "a,b\nc\td")
	// removes unnecessary spacing?
	checkConvert(t, "a                         b", "a b")
	// Compacts more complex stuff?
	checkConvert(t, `{:a "foo", :b zing ,:c 12.3e3}`, `{:a"foo":b zing,:c 12.3e3}`)
	// Doesn't compact away discards?
	checkConvert(t, `#_=> nil`, `#_=> nil`)
	// Removes comments?
	checkConvert(t, "; just a comment, I am ignored", "")
	checkConvert(t, "foo;; bar\nbaz", "foo\nbaz")
}

func checkConvert(t *testing.T, input, expected string) {
	var buf bytes.Buffer
	err := Compact(&buf, []byte(input))
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	} else if !bytes.Equal([]byte(expected), buf.Bytes()) {
		t.Errorf("Convert received '%s', expected '%s' back, was '%s'",
			input, expected, string(buf.Bytes()))
	}
	buf.Reset()
}
