package edn

import (
	"encoding/base64"
	"errors"
	"reflect"
	"sync"
	"time"
)

var (
	ErrNotFunc         = errors.New("Value is not a function")
	ErrMismatchArities = errors.New("Function does not have single argument in, two argument out")
	ErrNotConcrete     = errors.New("Value is not a concrete non-function type")
	ErrTagOverwritten  = errors.New("Previous tag implementation was overwritten")
)

var globalTags TagMap

type TagMap struct {
	sync.RWMutex
	m map[string]reflect.Value
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func (tm *TagMap) AddTagFn(name string, fn interface{}) error {
	// TODO: check name
	rfn := reflect.ValueOf(fn)
	rtyp := rfn.Type()
	if rtyp.Kind() != reflect.Func {
		return ErrNotFunc
	}
	if rtyp.NumIn() != 1 || rtyp.NumOut() != 2 || !rtyp.Out(1).Implements(errorType) {
		// ok to have variadic arity?
		return ErrMismatchArities
	}
	return tm.addVal(name, rfn)
}

func (tm *TagMap) addVal(name string, val reflect.Value) error {
	tm.Lock()
	if tm.m == nil {
		tm.m = map[string]reflect.Value{}
	}
	_, ok := tm.m[name]
	tm.m[name] = val
	tm.Unlock()
	if ok {
		return ErrTagOverwritten
	} else {
		return nil
	}
}

func AddTagFn(name string, fn interface{}) error {
	return globalTags.AddTagFn(name, fn)
}

func (tm *TagMap) AddTagStruct(name string, val interface{}) error {
	rstruct := reflect.ValueOf(val)
	switch rstruct.Type().Kind() {
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return ErrNotConcrete
	}
	return tm.addVal(name, rstruct)
}

func AddTagStruct(name string, val interface{}) error {
	return globalTags.AddTagStruct(name, val)
}

func init() {
	err := AddTagFn("inst", func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339, s)
	})
	if err != nil {
		panic(err)
	}
	err = AddTagFn("base64", func(s string) ([]byte, error) {
		return base64.StdEncoding.DecodeString(s)
	})
	if err != nil {
		panic(err)
	}
}
