package edn

import (
	"fmt"
	"bytes"
)

func SafeUnmarshal(bs []byte, v interface{}) (err error) {
	defer func() {
	        if r := recover(); r != nil {
		err = fmt.Errorf("crash: %s", err)
		}
	}()
	return Unmarshal(bs, v)
}

func Fuzz(data []byte) int {
	var v interface{}
	if err := SafeUnmarshal(data, &v); err != nil {
		return 0
	}
	bs, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err = Compact(&buf, bs); err != nil {
		panic(err)
	}
	var vv interface{}
	if err = Unmarshal(buf.Bytes(), &vv); err != nil {
		panic(err)
	}
	return 1
}

func AllFuzz(data []byte) int {
	var v interface{}
	if err := Unmarshal(data, &v); err != nil {
		return 0
	}
	bs, err := Marshal(v)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err = Compact(&buf, bs); err != nil {
		panic(err)
	}
	var vv interface{}
	if err = Unmarshal(buf.Bytes(), &vv); err != nil {
		panic(err)
	}
	return 1
}
