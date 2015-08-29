package edn

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

var (
	ErrNotFunc         = errors.New("Value is not a function")
	ErrMismatchArities = errors.New("Function does not have single argument in, two argument out")
)

var globalTags tagMap

type tagMap struct {
	sync.RWMutex
	m map[string]reflect.Value
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func (tm *tagMap) addTagFn(name string, fn interface{}) (interface{}, error) {
	// TODO: check name
	rfn := reflect.ValueOf(fn)
	rtyp := rfn.Type()
	if rtyp.Kind() != reflect.Func {
		return nil, ErrNotFunc
	}
	if rtyp.NumIn() != 1 || rtyp.NumOut() != 2 || !rtyp.Out(1).Implements(errorType) {
		// ok to have variadic arity?
		return nil, ErrMismatchArities
	}

	tm.Lock()
	if tm.m == nil {
		tm.m = map[string]reflect.Value{}
	}
	f := tm.m[name]
	tm.m[name] = rfn
	tm.Unlock()
	return f, nil
}

func AddTagFn(name string, fn interface{}) (interface{}, error) {
	return globalTags.addTagFn(name, fn)
}

func init() {
	_, err := AddTagFn("inst", func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339, s)
	})
	if err != nil {
		panic(err)
	}
}
