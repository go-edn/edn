package edn

import (
	"bytes"
	"testing"
)

func TestEncoding(t *testing.T) {
	testEncode(t, Keyword("foo"), ":foo")
	testEncode(t, Symbol("foo"), "foo")
	testEncode(t, "foo", `"foo"`)

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
}

func testEncode(t *testing.T, val interface{}, expects string) {
	bs, err := Marshal(val)
	if err != nil {
		t.Errorf("Unexpected error marshalling %q: %s", val, err.Error())
	} else if !bytes.Equal([]byte(expects), bs) {
		t.Errorf("Expected to see '%s', but got '%s' instead", expects, string(bs))
	}
}
