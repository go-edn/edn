// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"reflect"
	"testing"
)

// your basic unit tests.. unfinished, probably.

func TestIntReading(t *testing.T) {
	intStrs := [...]string{"0", "1", "+100", "-982", "8223372036854775808", "-5N", "-0N"}
	ints := [...]int64{0, 1, 100, -982, 8223372036854775808, -5, 0}
	for i, istr := range intStrs {
		var n int64
		err := UnmarshalString(istr, &n)
		if err != nil {
			t.Errorf("int64 '%s' failed, but expected success", istr)
		} else if n != ints[i] {
			t.Errorf("int64 '%s' was decoded to %d, but expected %d", istr, n, ints[i])
		}
	}
}

func TestArray(t *testing.T) {
	stringArray := `("foo" "bar" "baz")`
	stringExpected := [...]string{"foo", "bar", "baz"}
	var sa [3]string
	err := UnmarshalString(stringArray, &sa)
	if err != nil {
		t.Error(`expected '("foo" "bar" "baz")' to decode fine, but didn't`)
	} else {
		for i, expected := range stringExpected {
			if expected != sa[i] {
				t.Errorf(`Element %d in '("foo" "bar" "baz")' (%q) was encoded to %q`,
					i, expected, sa[i])
			}
		}
	}
}

func TestStruct(t *testing.T) {
	type Animal struct {
		Name string
		Type string `edn:"kind"`
	}
	type Person struct {
		Name      string
		Birthyear int `edn:"born"`
		Pets      []Animal
	}
	hans := `{:name "Hans",
            :born 1970,
            :pets [{:name "Cap'n Jack" :kind "Sparrow"}
                   {:name "Freddy" :kind "Cockatiel"}]}`
	goHans := Person{"Hans", 1970,
		[]Animal{{"Cap'n Jack", "Sparrow"}, {"Freddy", "Cockatiel"}}}
	var ednHans Person
	err := UnmarshalString(hans, &ednHans)
	if err != nil {
		t.Error("Error when decoding Hans")
	} else if !reflect.DeepEqual(goHans, ednHans) {
		t.Error("EDN Hans is not equal to Go hans")
	}
}

func TestRec(t *testing.T) {
	type Node struct {
		Left  *Node
		Val   int
		Right *Node
	}
	// here we're using symbols
	tree := "{left {left {val 3} val 5 right nil} val 10 right {val 15 right {val 17}}}"
	goTree := Node{Left: &Node{Left: &Node{Val: 3}, Val: 5, Right: nil},
		Val: 10, Right: &Node{Val: 15, Right: &Node{Val: 17}}}
	var ednTree Node
	err := UnmarshalString(tree, &ednTree)
	if err != nil {
		t.Errorf("Couldn't unmarshal tree: %s", err.Error())
	} else if !reflect.DeepEqual(goTree, ednTree) {
		t.Error("Mismatch between the Go tree and the tree encoded as EDN")
	}
}

func TestDiscard(t *testing.T) {
	var s Symbol
	discarding := "#_ #zap #_ xyz foo bar"
	expected := Symbol("bar")
	err := UnmarshalString(discarding, &s)
	if err != nil {
		t.Errorf("Expected '#_ #zap #_ xyz foo bar' to successfully read")
		t.Log(err.Error())
	} else if expected != s {
		t.Error("Mismatch between the Go symbol and the symbol encoded as EDN")
	}

	discarding = "#_ #foo #foo #foo #_#_bar baz zip quux"
	expected = Symbol("quux")
	err = UnmarshalString(discarding, &s)
	if err != nil {
		t.Errorf("Expected '#_ #foo #foo #foo #_#_bar baz zip quux' to successfully read")
		t.Log(err.Error())
	} else if expected != s {
		t.Error("Mismatch between the Go symbol and the symbol encoded as EDN")
	}
}
