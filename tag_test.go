// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"testing/quick"
	"time"
)

func TestReadTag(t *testing.T) {
	tagStr := "#foo bar"
	var tag Tag
	err := UnmarshalString(tagStr, &tag)
	if err != nil {
		t.Error("tag '#foo bar' failed, but expected success")
		t.Error(err.Error())
		t.FailNow()
	}
	if tag.Tagname != "foo" {
		t.Error("wrong tagname")
	}
	if tag.Value != Symbol("bar") {
		t.Error("wrong value")
	}
}

func TestReadNestedTag(t *testing.T) {
	tagStr := "#foo #bar baz"
	var tag Tag
	err := UnmarshalString(tagStr, &tag)
	if err != nil {
		t.Error("tag '#foo #bar baz' failed, but expected success")
		t.Error(err.Error())
		t.FailNow()
	}
	if tag.Tagname != "foo" {
		t.Error("wrong outer tagname")
	}
	switch val := tag.Value.(type) {
	case Tag:
		if val.Tagname != "bar" {
			t.Error("wrong inner tagname")
		}
		if val.Value != Symbol("baz") {
			t.Error("wrong inner value")
		}
	default:
		t.Errorf("Expected inner value to be Tag, but was %T", val)
	}
}

func TestReadStructWithTag(t *testing.T) {
	type T struct {
		Created Tag
		UUID    Tag
	}
	structStr := `{:created #inst "2015-08-29T21:28:34.311-00:00"
                 :uuid    #uuid "5c2d088b-bc77-47ec-8721-7fb78555ebaf"}`
	// These should NOT enable tag transformations at first level.
	var val T
	err := UnmarshalString(structStr, &val)
	expected := T{
		Created: Tag{"inst", "2015-08-29T21:28:34.311-00:00"},
		UUID:    Tag{"uuid", "5c2d088b-bc77-47ec-8721-7fb78555ebaf"},
	}
	if err != nil {
		t.Errorf("Couldn't unmarshal struct (T): %s", err.Error())
	} else if !reflect.DeepEqual(val, expected) {
		t.Error("Mismatch between the tags and the expected values")
	}
}

func TestReadTime(t *testing.T) {
	var v interface{}
	instStr := `#inst "2015-08-29T21:28:34.311-00:00"`
	inst, _ := time.Parse(time.RFC3339Nano, "2015-08-29T21:28:34.311-00:00")
	err := UnmarshalString(instStr, &v)
	if err != nil {
		t.Errorf("Couldn't unmarshal interface (time tag): %s", err.Error())
	} else {
		switch ednInst := v.(type) {
		case time.Time:
			if inst.UTC() != ednInst.UTC() {
				// TODO, I guess: I'm slightly surprised that I have to call UTC to
				// compare. I assumed the parse results would be identical.
				t.Error("Mismatch between time and the expected value")
				t.Logf("%s (expected) vs %s (actual)", inst, ednInst)
			}
		default:
			t.Errorf("Expected type to be time.Time, but was %T", ednInst)
		}
	}
}

func TestWriteTime(t *testing.T) {
	instStr := `#inst"2015-08-29T21:28:34.311Z"`
	inst, _ := time.Parse(time.RFC3339Nano, "2015-08-29T21:28:34.311Z")
	bs, err := Marshal(inst)
	if err != nil {
		t.Errorf("Expected no error when writing time, but got back %s", err.Error())
	} else if string(bs) != instStr {
		t.Errorf("Expected %s to be equal to %s, but wasn't", string(bs), instStr)
	}
}

func TestReadTimeVal(t *testing.T) {
	var ednInst time.Time
	instStr := `#inst "2015-08-29T21:28:34.311-00:00"`
	inst, _ := time.Parse(time.RFC3339, "2015-08-29T21:28:34.311-00:00")
	err := UnmarshalString(instStr, &ednInst)
	if err != nil {
		t.Errorf("Couldn't unmarshal interface (time tag): %s", err.Error())
	} else {
		if inst.UTC() != ednInst.UTC() {
			t.Error("Mismatch between time and the expected value")
			t.Logf("%s (expected) vs %s (actual)", inst, ednInst)
		}
	}
}

func TestAddTag(t *testing.T) {
	incer := func(val int) (int, error) {
		return val + 1, nil
	}
	d := NewDecoder(bytes.NewBufferString(`#inc 1 #inc #inc 1`))
	d.AddTagFn("inc", incer)
	var val int
	err := d.Decode(&val)
	if err != nil {
		t.Errorf("Couldn't unmarshal int: %s", err.Error())
	} else if val != 2 {
		t.Errorf("Expected value to be 2, was %d", val)
	}
	err = d.Decode(&val)
	if err != nil {
		t.Errorf("Couldn't unmarshal int: %s", err.Error())
	} else if val != 3 {
		t.Errorf("Expected value to be 3, was %d", val)
	}
}

func TestAddTagFnWithMissingError(t *testing.T) {
	missingErr := func(val int) int {
		return val + 1
	}
	d := NewDecoder(bytes.NewBufferString(``))
	if err := d.AddTagFn("inc", missingErr); err == nil {
		t.Error("Expected wrong arity tag function to cause error.")
	} else {
		t.Logf("Encountered error when adding tag function: %s", err)
	}
}

func TestMustAddTagFnWillPanic(t *testing.T) {
	defer func() { recover() }() // see [[https://stackoverflow.com/a/62028796/6247387][here]].

	missingErr := func(val int) int {
		return val + 1
	}
	d := NewDecoder(bytes.NewBufferString(``))
	d.MustAddTagFn("inc", missingErr)
	t.Error("Expected must add tag fn to panic but it didn't")
}

func TestAssignInterface(t *testing.T) {
	var v fmt.Stringer
	instStr := `#inst "2015-08-29T21:28:34.311-00:00"`
	err := UnmarshalString(instStr, &v)
	if err != nil {
		t.Errorf("Couldn't unmarshal time tag into stringer: %s", err.Error())
	}
}

type Colour interface {
	Space() string
}
type RGB struct {
	R uint8
	G uint8
	B uint8
}

func (_ RGB) Space() string { return "RGB" }

type YCbCr struct {
	Y  uint8
	Cb int8
	Cr int8
}

func (_ YCbCr) Space() string { return "YCbCr" }

func TestAssignMultiInterface(t *testing.T) {
	var colours []Colour
	j := `[#go-edn/ycbcr {:y 255 :cb 0 :cr -10}
         #go-edn/rgb {:r 98 :g 218 :b 255}]`
	d := NewDecoder(bytes.NewBufferString(j))
	d.AddTagFn("go-edn/rgb", func(r RGB) (RGB, error) { return r, nil })
	d.AddTagFn("go-edn/ycbcr", func(y YCbCr) (YCbCr, error) { return y, nil })
	err := d.Decode(&colours)
	if err != nil {
		t.Errorf("Couldn't unmarshal colours: %s", err.Error())
	} else {
		if colours[0].Space() != "YCbCr" {
			t.Errorf("Expected first colour to have space YCbCr, but was %s", colours[0].Space())
		}
		if colours[1].Space() != "RGB" {
			t.Errorf("Expected second colour to have space RGB, but was %s", colours[0].Space())
		}
	}
}

func TestAssignStruct(t *testing.T) {
	var val interface{}
	// reuse the rgb struct here
	j := `#go-edn/rgb {:r 98 :g 218 :b 255}`
	d := NewDecoder(bytes.NewBufferString(j))
	d.AddTagStruct("go-edn/rgb", RGB{})
	err := d.Decode(&val)
	if err != nil {
		t.Errorf("Couldn't unmarshal RGB struct tag: %s", err.Error())
	} else if rgb, ok := val.(RGB); ok {
		expected := RGB{98, 218, 255}
		if rgb != expected {
			t.Errorf("RGB struct tag had unexpected value: %q. Expected %q", rgb, expected)
		}
	} else { // if not  with rgb
		t.Errorf("RGB tag was unmarshalled into %T", val)
	}
}

func TestAssignType(t *testing.T) {
	type Nat int
	var val interface{}
	j := `#nat 10`
	d := NewDecoder(bytes.NewBufferString(j))
	d.AddTagStruct("nat", Nat(0))
	err := d.Decode(&val)
	if err != nil {
		t.Errorf("Couldn't unmarshal #nat type tag: %s", err.Error())
	} else if nat, ok := val.(Nat); ok {
		expected := Nat(10)
		if nat != expected {
			t.Errorf("nat type tag had unexpected value: %q. Expected %q", nat, expected)
		}
	} else {
		t.Errorf("nat tag was unmarshalled into %T", val)
	}
}

func TestAssignPtrStruct(t *testing.T) {
	var val RGB
	j := `#go-edn/rgb {:r 0 :g 2 :b 3}`
	d := NewDecoder(bytes.NewBufferString(j))
	d.AddTagStruct("go-edn/rgb", (*RGB)(nil))
	err := d.Decode(&val)
	if err != nil {
		t.Errorf("Couldn't unmarshal RGB struct tag: %s", err.Error())
	} else {
		expected := RGB{0, 2, 3}
		if val != expected {
			t.Errorf("RGB struct tag had unexpected value: %q. Expected %q", val, expected)
		}
	}
}

func TestAssignFuncPtr(t *testing.T) {
	var ednInst time.Time
	inst, _ := time.Parse(time.RFC3339Nano, "2124-05-13T14:51:39.127-00:00")
	j := `#inst "2124-05-13T14:51:39.127-00:00"`
	d := NewDecoder(bytes.NewBufferString(j))
	// override original inst function
	d.AddTagFn("inst", func(s string) (*time.Time, error) {
		val, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return nil, err
		}
		return &val, nil
	})
	err := d.Decode(&ednInst)
	if err != nil {
		t.Errorf("Couldn't unmarshal interface (time tag): %s", err.Error())
	} else if inst.UTC() != ednInst.UTC() {
		// TODO: Still surprised by the UTC call here.
		t.Error("Mismatch between time and the expected value")
		t.Logf("%s (expected) vs %s (actual)", inst, ednInst)
	}
}

type Method interface {
	Call() string
}

type ValueMethod struct {
	Value int
}

func (v ValueMethod) Call() string {
	return fmt.Sprintf("%d", v.Value) // Won't bother with strconv here
}

func TestAssignIFaceVal(t *testing.T) {
	var ednVal Method
	d := NewDecoder(bytes.NewBufferString(`#go-edn/value-method {:value 10}`))
	d.AddTagStruct("go-edn/value-method", &ValueMethod{})

	err := d.Decode(&ednVal)
	if err != nil {
		t.Errorf("Couldn't unmarshal interface (ValueMethod): %s", err.Error())
	} else if ednVal.Call() != "10" {
		t.Errorf("Unexpected result of method call, value is %#v", ednVal)
	}
}

type PointerMethod struct {
	Value int
}

func (p PointerMethod) Call() string {
	return fmt.Sprintf("%d", p.Value)
}

func TestAssignIFacePtr(t *testing.T) {
	var ednVal Method
	d := NewDecoder(bytes.NewBufferString(`#go-edn/pointer-method {:value 10}`))
	d.AddTagStruct("go-edn/pointer-method", PointerMethod{})

	err := d.Decode(&ednVal)
	if err != nil {
		t.Errorf("Couldn't unmarshal interface (PointerMethod): %s", err.Error())
	} else if ednVal.Call() != "10" {
		t.Errorf("Unexpected result of method call, value is %#v", ednVal)
	}
}

func TestDecodeByteslice(t *testing.T) {
	var v interface{}
	err := UnmarshalString(`#base64 "SGVsbG8sIEVETiBmcmllbmRz"`, &v)
	if err != nil {
		t.Errorf("Couldn't unmarshal base64 into interface{}: %s", err.Error())
	} else {
		switch val := v.(type) {
		case []byte:
			if string(val) != "Hello, EDN friends" {
				t.Error(`The base64-encoding did not serialise into "Hello, EDN friends"`)
				t.Logf("Serialised to %q instead", string(val))
			}
		default:
			t.Errorf("Expected type to be []byte, but was %T", val)
		}
	}
}

func TestEncodeByteslice(t *testing.T) {
	bs, err := Marshal([]byte{0, 6, 7, 8, 9, 128})
	if err != nil {
		t.Error("Couldn't marshal byte string")
	} else if string(bs) != `#base64"AAYHCAmA"` {
		t.Errorf("Expected edn encoding to be %q, but was %q", `#base64"AAYHCAmA"`, string(bs))
	}
}

func TestQuickBytes(t *testing.T) {
	f := func(bs []byte) bool {
		enc, err := Marshal(bs)
		if err != nil {
			t.Log(err)
			return false
		}
		var dec []byte
		err = Unmarshal(enc, &dec)
		if err != nil {
			t.Log(err)
			return false
		}
		return bytes.Equal(bs, dec)
	}
	conf := quick.Config{MaxCountScale: 200}
	if testing.Short() {
		conf.MaxCountScale = 10
	}
	if err := quick.Check(f, &conf); err != nil {
		t.Error(err)
	}
}
