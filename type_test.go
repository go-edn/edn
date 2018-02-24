// Copyright 2015-2017 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"testing"
	"testing/quick"
)

func TestRunes(t *testing.T) {
	rStr := `[\a \b \c \newline \space \tab \ŋ \' \" \u002c \u002C]`
	var runes []Rune
	err := UnmarshalString(rStr, &runes)
	if err != nil {
		t.Errorf("Reading `%s` failed with following error:", rStr)
		t.Error(err.Error())
	} else {
		actualRunes := []Rune{
			'a', 'b', 'c', '\n', ' ', '\t', 'ŋ', '\'', '"', ',', ',',
		}
		for i := range actualRunes {
			if actualRunes[i] != runes[i] {
				t.Errorf("Expected rune at position %d to be %q, but was %q", i, actualRunes[i], runes[i])
			}
		}
	}
}

func TestQuickRunes(t *testing.T) {
	f := func(s string) bool {
		good := true
		for _, r := range []rune(s) {
			bs, err := Marshal(Rune(r))
			if err != nil {
				t.Log(err)
				good = false
				continue
			}
			var res Rune
			err = Unmarshal(bs, &res)
			if err != nil {
				t.Log(err)
				good = false
				continue
			}
			if rune(res) != r {
				good = false
			}
		}
		return good
	}
	conf := quick.Config{MaxCountScale: 100}
	if testing.Short() {
		conf.MaxCountScale = 5
	}
	if err := quick.Check(f, &conf); err != nil {
		t.Error(err)
	}
}

func TestTagRunes(t *testing.T) {
	type Foo struct {
		Value rune `edn:",rune"`
	}
	f := Foo{Value: ' '}
	bs, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `{:value\space}` {
		t.Errorf("Expected result to be `{:value\\space}`, but was `%s`", string(bs))
	}
	f.Value = '\''
	bs, err = Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `{:value\'}` {
		t.Errorf("Expected result to be `{:value\\'}`, but was `%s`", string(bs))
	}
}

func TestSpacing(t *testing.T) {
	type Foo struct {
		Value Rune `edn:",sym"`
		Data  Rune `edn:",sym"`
	}
	f := Foo{Value: Rune('a'), Data: Rune('b')}
	bs, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `{value \a data \b}` {
		t.Errorf("Expected result to be `{value \\a data \\b}`, but was `%s`", string(bs))
	}
}

func TestMarshalRawMessageValue(t *testing.T) {
	type Foo struct {
		SomeVal   string `edn:"some-val"`
		Leftovers RawMessage
		OtherVal  string `edn:"other-val"`
	}

	f := Foo{
		SomeVal:   "egg",
		Leftovers: []byte(`[\space #foo bar :baz 100 {#{} 1.0 "zap" nil}]`),
		OtherVal:  "spam",
	}
	bs, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != `{:some-val"egg":leftovers [\space #foo bar :baz 100{#{}1.0"zap"nil}] :other-val"spam"}` {
		t.Errorf("Expected result to be `{:some-val\"egg\":leftovers [\\space #foo bar :baz 100{#{}1.0\"zap\"nil}] :other-val\"spam\"}`, but was `%s`", string(bs))
	}
}

func TestUnmarshalRawMessageValue(t *testing.T) {
	type Foo struct {
		SomeVal   string `edn:"some-val"`
		Leftovers RawMessage
		OtherVal  string `edn:"other-val"`
	}
	const raw = `{
  :some-val"egg"
  :leftovers [\space #foo bar :baz 100{#{} 1.0 "zap" nil}]
  :other-val"spam"
}`
	var f Foo
	err := UnmarshalString(raw, &f)
	if err != nil {
		t.Fatal(err)
	}
	if string(f.Leftovers) != `[\space #foo bar :baz 100{#{} 1.0 "zap" nil}]` {
		t.Errorf("Expected result to be `[\\space #foo bar :baz 100{#{} 1.0 \"zap\" nil}]`, but was `%s`", string(f.Leftovers))
	}
}
