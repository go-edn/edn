package edn

import (
	"reflect"
	"testing"
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
