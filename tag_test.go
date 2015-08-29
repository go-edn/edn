package edn

import (
	"bytes"
	"reflect"
	"testing"
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
	inst, _ := time.Parse(time.RFC3339, "2015-08-29T21:28:34.311-00:00")
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
