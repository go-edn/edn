// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"unicode/utf8"
)

var (
	errInternal       = errors.New("Illegal internal parse state")
	errNoneLeft       = errors.New("No more tokens to read")
	errUnexpeced      = errors.New("Unexpected token")
	errNotImplemented = errors.New("Tags are not yet implemented")
	errIllegalRune    = errors.New("Illegal rune form")
)

func Unmarshal(data []byte, v interface{}) error {
	return newDecoder(bufio.NewReader(bytes.NewBuffer(data))).Decode(v)
}

func UnmarshalString(data string, v interface{}) error {
	return newDecoder(bufio.NewReader(bytes.NewBufferString(data))).Decode(v)
}

func NewDecoder(r io.Reader) *Decoder {
	return newDecoder(bufio.NewReader(r))
}

func (d *Decoder) Buffered() *bufio.Reader {
	return d.rd
}

func (d *Decoder) AddTagFn(tagname string, fn interface{}) (interface{}, error) {
	return d.tagmap.addTagFn(tagname, fn)
}

// Unmarshaler is the interface implemented by objects that can unmarshal an EDN
// description of themselves. The input can be assumed to be a valid encoding of
// an EDN value. UnmarshalEDN must copy the EDN data if it wishes to retain the
// data after returning.
//
// (This is not available yet, as it's not working properly.)
type Unmarshaler interface {
	UnmarshalEDN([]byte) error
}

type parseState int

const (
	parseToplevel = iota
	parseList
	parseVector
	parseMap
	parseSet
	parseTagged
	parseDiscard
)

type Decoder struct {
	lex        *lexer
	savedError error
	rd         *bufio.Reader
	tagmap     tagMap
	// parser-specific
	prevSlice []byte
	prevTtype tokenType
	undo      bool
	// if nextToken returned lexEndPrev, we must write the leftover value at
	// next call to nextToken
	hasLeftover bool
	leftover    rune
}

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "edn: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "edb: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "edn: Unmarshal(nil " + e.Type.String() + ")"
}

// An UnmarshalTypeError describes a EDN value that was
// not appropriate for a value of a specific Go type.
type UnmarshalTypeError struct {
	Value string       // description of EDN value - "bool", "array", "number -5"
	Type  reflect.Type // type of Go value it could not be assigned to
}

func (e *UnmarshalTypeError) Error() string {
	return "edn: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

func (d *Decoder) Decode(val interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(val)}
	}

	d.value(rv)

	return nil
}

func newDecoder(buf *bufio.Reader) *Decoder {
	lex := lexer{}
	lex.reset()
	return &Decoder{
		lex:         &lex,
		rd:          buf,
		hasLeftover: false,
		leftover:    '\uFFFD',
	}
}

func (d *Decoder) getTagFn(tagname string) *reflect.Value {
	d.tagmap.RLock()
	f, ok := d.tagmap.m[tagname]
	d.tagmap.RUnlock()
	if ok {
		return &f
	}
	globalTags.RLock()
	f, ok = globalTags.m[tagname]
	globalTags.RUnlock()
	if ok {
		return &f
	}
	return nil
}

func (d *Decoder) error(err error) {
	panic(err)
}

func (d *Decoder) doUndo(bs []byte, ttype tokenType) {
	if d.undo {
		d.error(errInternal) // this is LL(1), so this shouldn't happen
	}
	d.undo = true
	d.prevSlice = bs
	d.prevTtype = ttype
}

// array consumes an array from d.data[d.off-1:], decoding into the value v.
// the first byte of the array ('[') has been read already.
func (d *Decoder) array(v reflect.Value, endType tokenType) {
	// Check for unmarshaler.
	u, pv := d.indirect(v, false)
	if u != nil {
		bs, err := d.nextValueBytes() // nextvaluebytes
		if err == nil {
			err = u.UnmarshalEDN(bs)
		}
		if err != nil {
			d.error(err)
		}
		return
	}
	v = pv

	// Check type of target.
	switch v.Kind() {
	case reflect.Interface:
		if v.NumMethod() == 0 {
			// Decoding into nil interface? Switch to non-reflect code.
			v.Set(reflect.ValueOf(d.arrayInterface(endType)))
			return
		}
		// Otherwise it's invalid.
		fallthrough
	default:
		d.error(&UnmarshalTypeError{"array", v.Type()})
		return
	case reflect.Array:
	case reflect.Slice:
		break
	}

	i := 0
	for {
		// Look ahead for ] - can only happen on first iteration.
		bs, ttype, err := d.nextToken()
		if err != nil {
			d.error(err)
		}
		if ttype == endType {
			break
		}
		d.doUndo(bs, ttype)

		// Get element of array, growing if necessary.
		if v.Kind() == reflect.Slice {
			// Grow slice if necessary
			if i >= v.Cap() {
				newcap := v.Cap() + v.Cap()/2
				if newcap < 4 {
					newcap = 4
				}
				newv := reflect.MakeSlice(v.Type(), v.Len(), newcap)
				reflect.Copy(newv, v)
				v.Set(newv)
			}
			if i >= v.Len() {
				v.SetLen(i + 1)
			}
		}

		if i < v.Len() {
			// Decode into element.
			d.value(v.Index(i))
		} else {
			// Ran out of fixed array: skip.
			d.value(reflect.Value{})
		}
		i++
	}

	if i < v.Len() {
		if v.Kind() == reflect.Array {
			// Array.  Zero the rest.
			z := reflect.Zero(v.Type().Elem())
			for ; i < v.Len(); i++ {
				v.Index(i).Set(z)
			}
		} else {
			v.SetLen(i)
		}
	}
	if i == 0 && v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}
}

func (d *Decoder) arrayInterface(endType tokenType) interface{} {
	var v = make([]interface{}, 0)
	for {
		// look out for endType
		bs, tt, err := d.nextToken()
		if err != nil {
			d.error(err)
			break
		}
		if tt == endType {
			break
		}
		d.doUndo(bs, tt)
		v = append(v, d.valueInterface())
	}
	return v
}

func (d *Decoder) value(v reflect.Value) {
	if !v.IsValid() {
		// read value and ignore it
		d.valueInterface()
		return
	}
	bs, ttype, err := d.nextToken()
	// check error first
	if err != nil {
		d.error(err)
		return
	}
	switch ttype {
	default:
		d.error(errUnexpeced)
	case tokenSymbol, tokenKeyword, tokenString, tokenInt, tokenFloat, tokenChar:
		d.literal(bs, ttype, v)
	case tokenTag:
		d.tag(bs, v)
	case tokenListStart:
		d.array(v, tokenListEnd)
	case tokenVectorStart:
		d.array(v, tokenVectorEnd)
	case tokenSetStart:
		d.set(v)
	case tokenMapStart:
		d.ednmap(v)
	}
}

func (d *Decoder) tag(tag []byte, v reflect.Value) {
	// Check for unmarshaler.
	u, pv := d.indirect(v, false)
	if u != nil {
		bs, err := d.nextValueBytes()
		if err == nil {
			err = u.UnmarshalEDN(append(append(tag, ' '), bs...))
		}
		if err != nil {
			d.error(err)
		}
		return
	}
	v = pv

	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		v.Set(reflect.ValueOf(d.tagInterface(tag)))
		return
	}

	fn := d.getTagFn(string(tag[1:]))
	if fn == nil {
		// So in theory we'd have to match against any interface that could be
		// assignable to the Tag type, to ensure we would decode whenever possible.
		// That is any interface that specifies any combination of the methods
		// MarshalEDN, UnmarshalEDN and String. I'm not sure if that makes sense
		// though, so I've punted this for now.
		d.error(errInternal)
		return
	} else {
		tfn := fn.Type()
		val := reflect.New(tfn.In(0))
		d.value(val)
		res := fn.Call([]reflect.Value{val.Elem()})
		if err, ok := res[1].Interface().(error); ok && err != nil {
			d.error(err)
		}
		r := res[0]
		// oh no, this is not going to be a fun exercise I think. what do we do with interfaces?
		if !r.Type().AssignableTo(v.Type()) {
			d.error(fmt.Errorf("Cannot assign %s to %s (tag issue?)", r.Type(), v.Type()))
		}
		v.Set(r)
	}
}

func (d *Decoder) tagInterface(tag []byte) interface{} {
	fn := d.getTagFn(string(tag[1:]))
	if fn == nil {
		var t Tag
		t.Tagname = string(tag[1:])
		t.Value = d.valueInterface()
		return t
	} else {
		tfn := fn.Type()
		val := reflect.New(tfn.In(0))
		d.value(val)
		res := fn.Call([]reflect.Value{val.Elem()})
		if err, ok := res[1].Interface().(error); ok && err != nil {
			d.error(err)
		}
		return res[0].Interface()
	}
}

func (d *Decoder) valueInterface() interface{} {
	bs, ttype, err := d.nextToken()
	// check error first
	if err != nil {
		d.error(err)
		return nil /// won't get here
	}
	switch ttype {
	default:
		d.error(errUnexpeced)
		return nil
	case tokenSymbol, tokenKeyword, tokenString, tokenInt, tokenFloat, tokenChar:
		return d.literalInterface(bs, ttype)
	case tokenTag:
		d.error(errNotImplemented)
		return nil
	case tokenListStart:
		return d.arrayInterface(tokenListEnd)
	case tokenVectorStart:
		return d.arrayInterface(tokenVectorEnd)
	case tokenSetStart:
		return d.setInterface()
	case tokenMapStart:
		return d.ednmapInterface()
	}
	return nil
}

func (d *Decoder) ednmap(v reflect.Value) {
	// Check for unmarshaler.
	u, pv := d.indirect(v, false)
	if u != nil {
		bs, err := d.nextValueBytes()
		if err == nil {
			err = u.UnmarshalEDN(bs)
		}
		if err != nil {
			d.error(err)
		}
		return
	}
	v = pv

	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		v.Set(reflect.ValueOf(d.ednmapInterface()))
		return
	}

	var keyType reflect.Type
	var subv reflect.Value

	// Check type of target: Struct or map[T]U
	switch v.Kind() {
	case reflect.Map:
		t := v.Type()
		keyType = t.Key()
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}
	case reflect.Struct:

	default:
		d.error(&UnmarshalTypeError{"map", v.Type()})
	}

	// separate these to ease reading (theoretically fewer checks too)
	if v.Kind() == reflect.Struct {
		for {
			bs, tt, err := d.nextToken()
			if err != nil {
				d.error(err)
			}
			if tt == tokenSetEnd {
				break
			}
			skip := false
			var key []byte
			// The key can either be a symbol, a keyword or a string. We will skip
			// anything that is not any of these values.
			switch tt {
			case tokenSymbol:
				if bytes.Equal(bs, falseByte) || bytes.Equal(bs, trueByte) || bytes.Equal(bs, nilByte) {
					skip = true
				}
				key = bs
			case tokenKeyword:
				key = bs[1:]
			case tokenString:
				k, ok := unquoteBytes(bs)
				key = k
				if !ok {
					d.error(errInternal)
				}
			default:
				skip = true
			}

			if skip { // will panic if something bad happens, so this is fine
				d.valueInterface()
				continue
			}

			var f *field
			fields := cachedTypeFields(v.Type())
			for i := range fields {
				ff := &fields[i]
				if bytes.Equal(ff.nameBytes, key) {
					f = ff
					break
				}
				if f == nil && ff.equalFold(ff.nameBytes, key) {
					f = ff
				}
			}
			if f != nil {
				subv = v
				for _, i := range f.index {
					if subv.Kind() == reflect.Ptr {
						if subv.IsNil() {
							subv.Set(reflect.New(subv.Type().Elem()))
						}
						subv = subv.Elem()
					}
					subv = subv.Field(i)
				}
			}
			// If subv not set, value() will just skip.
			d.value(subv)
		}
		// if not struct, then it is a map
	} else if keyType.Kind() == reflect.Interface && keyType.NumMethod() == 0 {
		// special case for unhashable key types
		var mapElem reflect.Value
		for {
			bs, tt, err := d.nextToken()
			if err != nil {
				d.error(err)
			}
			if tt == tokenSetEnd {
				break
			}
			d.doUndo(bs, tt)

			key := d.valueInterface()
			elemType := v.Type().Elem()
			if !mapElem.IsValid() {
				mapElem = reflect.New(elemType).Elem()
			} else {
				mapElem.Set(reflect.Zero(elemType))
			}
			subv = mapElem
			d.value(subv)

			switch reflect.TypeOf(key).Kind() {
			case reflect.Slice, reflect.Map: // bypass issues with unhashable types
				v.SetMapIndex(reflect.ValueOf(&key), subv)
			default:
				v.SetMapIndex(reflect.ValueOf(key), subv)
			}
		}
	} else { // default map case
		var mapElem reflect.Value
		for {
			bs, tt, err := d.nextToken()
			if err != nil {
				d.error(err)
			}
			if tt == tokenSetEnd {
				break
			}
			d.doUndo(bs, tt)

			// should we do the same as with mapElem?
			key := reflect.New(keyType).Elem()
			d.value(key)

			elemType := v.Type().Elem()
			if !mapElem.IsValid() {
				mapElem = reflect.New(elemType).Elem()
			} else {
				mapElem.Set(reflect.Zero(elemType))
			}
			subv = mapElem
			d.value(subv)
			v.SetMapIndex(key, subv)
		}
	}
}

func (d *Decoder) ednmapInterface() interface{} {
	theMap := make(map[interface{}]interface{}, 0)
	for {
		bs, tt, err := d.nextToken()
		if err != nil {
			d.error(err)
		}
		if tt == tokenMapEnd {
			break
		}
		d.doUndo(bs, tt)
		key := d.valueInterface()
		value := d.valueInterface()
		switch reflect.TypeOf(key).Kind() {
		case reflect.Slice, reflect.Map: // bypass issues with unhashable types
			theMap[&key] = value
		default:
			theMap[key] = value
		}
	}
	return theMap
}

func (d *Decoder) set(v reflect.Value) {
	// Check for unmarshaler.
	u, pv := d.indirect(v, false)
	if u != nil {
		bs, err := d.nextValueBytes()
		if err == nil {
			err = u.UnmarshalEDN(bs)
		}
		if err != nil {
			d.error(err)
		}
		return
	}
	v = pv

	var setValue reflect.Value
	var keyType reflect.Type

	// Check type of target.
	// TODO: accept option structs? -- i.e. structs where all fields are bools
	// TODO: Also accept slices
	switch v.Kind() {
	case reflect.Map:
		// map must have bool or struct{} value type
		t := v.Type()
		keyType = t.Key()
		valKind := t.Elem().Kind()
		switch valKind {
		case reflect.Bool:
			setValue = reflect.ValueOf(true)
		case reflect.Struct:
			// check if struct, and if so, ensure it has 0 fields
			if t.Elem().NumField() != 0 {
				d.error(&UnmarshalTypeError{"set", v.Type()})
			}
			setValue = reflect.Zero(t.Elem())
		default:
			d.error(&UnmarshalTypeError{"set", v.Type()})
		}
		if v.IsNil() {
			v.Set(reflect.MakeMap(t))
		}
	case reflect.Interface:
		if v.NumMethod() == 0 {
			// break out and use setInterface
			v.Set(reflect.ValueOf(d.setInterface()))
			return
		} else {
			d.error(&UnmarshalTypeError{"set", v.Type()})
		}

	default:
		d.error(&UnmarshalTypeError{"set", v.Type()})
	}
	// special case here, to avoid panics when we have slices and maps as keys.
	// Split out from code below to improve perf
	if keyType.Kind() == reflect.Interface && keyType.NumMethod() == 0 {
		for {
			bs, tt, err := d.nextToken()
			if err != nil {
				d.error(err)
			}
			if tt == tokenSetEnd {
				break
			}
			d.doUndo(bs, tt)
			key := d.valueInterface()
			switch reflect.TypeOf(key).Kind() {
			case reflect.Slice, reflect.Map: // bypass issues with unhashable types
				v.SetMapIndex(reflect.ValueOf(&key), setValue)
			default:
				v.SetMapIndex(reflect.ValueOf(key), setValue)
			}
		}
	} else {
		for {
			bs, tt, err := d.nextToken()
			if err != nil {
				d.error(err)
			}
			if tt == tokenSetEnd {
				break
			}
			d.doUndo(bs, tt)

			key := reflect.New(keyType).Elem()
			d.value(key)
			v.SetMapIndex(key, setValue)
		}
	}
}

func (d *Decoder) setInterface() interface{} {
	theSet := make(map[interface{}]bool, 0)
	for {
		bs, tt, err := d.nextToken()
		if err != nil {
			d.error(err)
		}
		if tt == tokenSetEnd {
			break
		}
		d.doUndo(bs, tt)
		key := d.valueInterface()
		switch reflect.TypeOf(key).Kind() {
		case reflect.Slice, reflect.Map: // bypass issues with unhashable types
			theSet[&key] = true
		default:
			theSet[key] = true
		}
	}
	return theSet
}

var nilByte = []byte(`nil`)
var trueByte = []byte(`true`)
var falseByte = []byte(`false`)

var symbolType = reflect.TypeOf(Symbol(""))
var keywordType = reflect.TypeOf(Keyword(""))
var byteSliceType = reflect.TypeOf([]byte(nil))

func (d *Decoder) literal(bs []byte, ttype tokenType, v reflect.Value) {
	wantptr := ttype == tokenSymbol && bytes.Equal(nilByte, bs)
	u, pv := d.indirect(v, wantptr)
	if u != nil {
		err := u.UnmarshalEDN(bs)
		if err != nil {
			d.error(err)
		}
		return
	}
	v = pv
	switch ttype {
	case tokenSymbol:
		if wantptr { // nil
			switch v.Kind() {
			case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
				v.Set(reflect.Zero(v.Type()))
				// else ignore
			}
		} else if bytes.Equal(trueByte, bs) || bytes.Equal(falseByte, bs) { // true|false
			value := bs[0] == 't'
			switch v.Kind() {
			default:
				d.error(&UnmarshalTypeError{"bool", v.Type()})
			case reflect.Bool:
				v.SetBool(value)
			case reflect.Interface:
				if v.NumMethod() == 0 {
					v.Set(reflect.ValueOf(value))
				} else {
					d.error(&UnmarshalTypeError{"bool", v.Type()})
				}
			}
		} else if v.Kind() == reflect.String && v.Type() == symbolType { // "actual" symbols
			v.SetString(string(bs))
		} else if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(Symbol(string(bs))))
		} else {
			d.error(&UnmarshalTypeError{"symbol", v.Type()})
		}
	case tokenKeyword:
		if v.Kind() == reflect.String && v.Type() == keywordType { // "actual" keywords
			v.SetString(string(bs[1:]))
		} else if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
			v.Set(reflect.ValueOf(Keyword(string(bs[1:]))))
		} else {
			d.error(&UnmarshalTypeError{"keyword", v.Type()})
		}
	case tokenInt:
		var s string
		if bs[len(bs)-1] == 'N' { // can end with N, which we promptly ignore
			// TODO: If the user expects a float and receives what is perceived as an
			// int (ends with N), what is the sensible thing to do?
			s = string(bs[:len(bs)-1])
		} else {
			s = string(bs)
		}
		switch v.Kind() {
		default:
			d.error(&UnmarshalTypeError{"int", v.Type()})
		case reflect.Interface:
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				d.error(&UnmarshalTypeError{"int " + s, reflect.TypeOf(int64(0))})
			}
			if v.NumMethod() != 0 {
				d.error(&UnmarshalTypeError{"int", v.Type()})
			}
			v.Set(reflect.ValueOf(n))

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil || v.OverflowInt(n) {
				d.error(&UnmarshalTypeError{"int " + s, v.Type()})
			}
			v.SetInt(n)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n, err := strconv.ParseUint(s, 10, 64)
			if err != nil || v.OverflowUint(n) {
				d.error(&UnmarshalTypeError{"int " + s, v.Type()})
			}
			v.SetUint(n)

		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(s, v.Type().Bits())
			if err != nil || v.OverflowFloat(n) {
				d.error(&UnmarshalTypeError{"int " + s, v.Type()})
			}
			v.SetFloat(n)
		}

	case tokenFloat:
		var s string
		if bs[len(bs)-1] == 'M' { // can end with M, which we promptly ignore
			s = string(bs[:len(bs)-1])
		} else {
			s = string(bs)
		}
		switch v.Kind() {
		default:
			d.error(&UnmarshalTypeError{"float", v.Type()})
		case reflect.Interface:
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				d.error(&UnmarshalTypeError{"float " + s, reflect.TypeOf(float64(0))})
			}
			if v.NumMethod() != 0 {
				d.error(&UnmarshalTypeError{"float", v.Type()})
			}
			v.Set(reflect.ValueOf(n))

		case reflect.Float32, reflect.Float64:
			n, err := strconv.ParseFloat(s, v.Type().Bits())
			if err != nil || v.OverflowFloat(n) {
				d.error(&UnmarshalTypeError{"float " + s, v.Type()})
			}
			v.SetFloat(n)
		}
	case tokenChar:
		r, err := toRune(bs)
		if err != nil {
			d.error(err)
		}
		switch v.Kind() {
		default:
			d.error(&UnmarshalTypeError{"rune", v.Type()})
		case reflect.Interface:
			if v.NumMethod() != 0 {
				d.error(&UnmarshalTypeError{"rune", v.Type()})
			}
			v.Set(reflect.ValueOf(r))
		case reflect.Int32: // rune is an alias for int32
			v.Set(reflect.ValueOf(r))
		}
	case tokenString:
		s, ok := unquoteBytes(bs)
		if !ok {
			d.error(errInternal)
		}
		switch v.Kind() {
		default:
			d.error(&UnmarshalTypeError{"string", v.Type()})
		case reflect.Slice:
			// TODO: bytes
			if v.Type() != byteSliceType {
				d.error(&UnmarshalTypeError{"string", v.Type()})
			}
			b := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
			n, err := base64.StdEncoding.Decode(b, s)
			if err != nil {
				d.error(err)
			}
			v.Set(reflect.ValueOf(b[0:n]))
		case reflect.String:
			v.SetString(string(s))
		case reflect.Interface:
			if v.NumMethod() == 0 {
				v.Set(reflect.ValueOf(string(s)))
			} else {
				d.error(&UnmarshalTypeError{"string", v.Type()})
			}
		}
	default:
		d.error(errInternal)
	}
}

func (d *Decoder) literalInterface(bs []byte, ttype tokenType) interface{} {
	switch ttype {
	case tokenSymbol:
		if bytes.Equal(nilByte, bs) {
			return nil
		}
		if bytes.Equal(trueByte, bs) {
			return true
		}
		if bytes.Equal(falseByte, bs) {
			return false
		}
		return Symbol(string(bs))
	case tokenKeyword:
		return Keyword(string(bs[1:]))
	case tokenInt:
		var s string
		if bs[len(bs)-1] == 'N' { // can end with N, which we promptly ignore
			s = string(bs[:len(bs)-1])
		} else {
			s = string(bs)
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			d.error(err)
		}
		return n
	case tokenFloat:
		var s string
		if bs[len(bs)-1] == 'M' { // can end with M, which we promptly ignore
			s = string(bs[:len(bs)-1])
		} else {
			s = string(bs)
		}
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			d.error(err)
		}
		return n
	case tokenChar:
		r, err := toRune(bs)
		if err != nil {
			d.error(err)
		}
		return r
	case tokenString:
		t, ok := unquote(bs)
		if !ok {
			d.error(errInternal)
		}
		return t
	default:
		d.error(errInternal)
		return nil
	}
}

var (
	newlineBytes  = []byte(`\newline`)
	returnBytes   = []byte(`\return`)
	spaceBytes    = []byte(`\space`)
	tabBytes      = []byte(`\tab`)
	formfeedBytes = []byte(`\formfeed`)
)

func toRune(bs []byte) (rune, error) {
	// handle special cases first:
	switch {
	case bytes.Equal(bs, newlineBytes):
		return '\n', nil
	case bytes.Equal(bs, returnBytes):
		return '\r', nil
	case bytes.Equal(bs, spaceBytes):
		return ' ', nil
	case bytes.Equal(bs, tabBytes):
		return '\t', nil
	case bytes.Equal(bs, formfeedBytes):
		return '\f', nil
	case len(bs) == 6 && bs[1] == 'u': // I don't think unicode chars could be 5 bytes long?
		return getu4(bs), nil
	default:
		r, size := utf8.DecodeRune(bs[1:])
		if r == utf8.RuneError && size == 1 {
			return r, errIllegalRune
		}
		return r, nil
	}
}

// nextToken handles #_
func (d *Decoder) nextToken() ([]byte, tokenType, error) {
	bs, tt, err := d.rawToken()
	if err != nil {
		return bs, tt, err
	}
	switch tt {
	case tokenDiscard:
		err := d.traverseValue()
		if err != nil {
			return nil, tokenError, err
		}
		return d.nextToken() // again for discards
	default:
		return bs, tt, err
	}
}

func (d *Decoder) rawToken() ([]byte, tokenType, error) {
	if d.undo {
		d.undo = false
		b := d.prevSlice
		tt := d.prevTtype
		d.prevSlice = nil
		d.prevTtype = tokenError
		return b, tt, nil
	}
	var val bytes.Buffer
	d.lex.reset()
	doIgnore := true
	if d.hasLeftover {
		d.hasLeftover = false
		d.lex.position++
		switch d.lex.state(d.leftover) {
		case lexCont:
			val.WriteRune(d.leftover)
			doIgnore = false
		case lexEnd:
			val.WriteRune(d.leftover)
			return val.Bytes(), d.lex.token, nil
		case lexEndPrev:
			return nil, tokenError, errInternal
		case lexError:
			return nil, tokenError, d.lex.err
		case lexIgnore:
			// just ignore
		}
	}
	if doIgnore { // ignore whitespace
	readWhitespace:
		for {
			r, _, err := d.rd.ReadRune()
			if err == io.EOF {
				return nil, tokenError, errNoneLeft
			}
			if err != nil {
				return nil, tokenError, err
			}
			d.lex.position++
			switch d.lex.state(r) {
			case lexCont: // got a value, so continue on past doIgnoring
				// TODO: This returns an error. Will it happen in practice? Probably?
				val.WriteRune(r)
				break readWhitespace
			case lexError:
				return nil, tokenError, d.lex.err
			case lexEnd:
				val.WriteRune(r)
				return val.Bytes(), d.lex.token, nil
			case lexEndPrev:
				return nil, tokenError, errInternal
			case lexIgnore:
				// keep on reading
			}
		}
	}
	for {
		r, _, err := d.rd.ReadRune()
		var ls lexState
		// this is not exactly perfect.
		switch {
		case err == io.EOF:
			ls = d.lex.eof()
		case err != nil:
			return nil, tokenError, err
		default:
			d.lex.position++
			ls = d.lex.state(r)
		}
		switch ls {
		case lexCont:
			val.WriteRune(r)
		case lexIgnore:
			if err != io.EOF {
				return nil, tokenError, errInternal
			} else {
				return nil, tokenError, errNoneLeft
			}
		case lexEnd:
			if err != io.EOF {
				val.WriteRune(r)
			}
			return val.Bytes(), d.lex.token, nil
		case lexEndPrev:
			d.hasLeftover = true
			d.leftover = r
			return val.Bytes(), d.lex.token, nil
		case lexError:
			return nil, tokenError, d.lex.err
		}
	}
}

// traverseValue reads a single value and skips it -- whether it is a list, map
// or a literal. Doesn't validate its state. skips over discard tokens as well.
func (d *Decoder) traverseValue() error {
	tstack := newTokenStack()
	for {
		_, tt, err := d.nextToken()
		if err != nil {
			return err
		}
		err = tstack.push(tt)
		if err != nil || tstack.done() {
			return err
		}
	}
}

type tokenStack struct {
	toks     []tokenType
	toplevel tokenType
}

func newTokenStack() *tokenStack {
	return &tokenStack{
		toks:     nil,
		toplevel: tokenError,
	}
}

func (t *tokenStack) done() bool {
	return len(t.toks) == 0 && t.toplevel != tokenDiscard
}

func (t *tokenStack) peek() tokenType {
	return t.toks[len(t.toks)-1]
}

func (t *tokenStack) pop() {
	t.toks = t.toks[:len(t.toks)-1]
}

func (t *tokenStack) push(tt tokenType) error {
	// retain toplevel value for done check
	if len(t.toks) == 0 {
		t.toplevel = tt
	}
	switch tt {
	case tokenMapStart, tokenVectorStart, tokenListStart, tokenSetStart, tokenDiscard, tokenTag:
		// append to toks, regardless
		t.toks = append(t.toks, tt)
		return nil
	case tokenMapEnd:
		if len(t.toks) == 0 || t.peek() != tokenMapStart || t.peek() != tokenSetStart {
			return errUnexpeced
		}
		t.pop()
	case tokenListEnd:
		if len(t.toks) == 0 || t.peek() != tokenListStart {
			return errUnexpeced
		}
		t.pop()
	case tokenVectorEnd:
		if len(t.toks) == 0 || t.peek() != tokenVectorEnd {
			return errUnexpeced
		}
		t.pop()
	default:
	}
	// popping of discards and tags
	for len(t.toks) > 0 && t.peek() == tokenTag {
		t.pop()
	}
	if len(t.toks) > 0 && t.peek() == tokenDiscard {
		t.pop()
	}
	return nil
}

// Oh, asking about why this is so similar to the part above, eh? Yes, I would
// also consider this a crime. At least I use the same lexer. This is probably
// next on the list when I have people complaining about perf issues.
func (d *Decoder) nextValueBytes() ([]byte, error) {
	// TODO: Ensure values inside maps come in pairs.
	tstack := newTokenStack()
	var val bytes.Buffer
	if d.undo {
		d.undo = false
		b := d.prevSlice
		tt := d.prevTtype
		d.prevSlice = nil
		d.prevTtype = tokenError
		if tt == tokenDiscard { // should be impossible to get a tokenDiscard here?
			return nil, errInternal
		}
		err := tstack.push(tt)
		if err != nil || tstack.done() {
			return val.Bytes(), err
		}
		val.Write(b)
	}
readElems:
	for {
		d.lex.reset()
		// Can't ignore whitespace in general. So I guess we just add it onto the buffer
		readWs := true
		if d.hasLeftover {
			// we can have leftover from previous iteration. e.g. "foo[bar]" will have
			// leftover "[" and "]"
			d.hasLeftover = false
			d.lex.position++
			val.WriteRune(d.leftover)
			switch d.lex.state(d.leftover) {
			case lexCont:
				readWs = false
			case lexEnd:
				err := tstack.push(d.lex.token)
				if err != nil || tstack.done() {
					return val.Bytes(), err
				}
				d.lex.reset()
			case lexEndPrev:
				return nil, errInternal
			case lexError:
				return nil, d.lex.err
			case lexIgnore:
				// just keep going
			}
		}
		if readWs {
		readWhitespace:
			// If we end up here, it means we expect at least one more token
			for {
				r, _, err := d.rd.ReadRune()
				if err == io.EOF {
					return nil, errNoneLeft
				}
				if err != nil {
					return nil, err
				}
				d.lex.position++
				val.WriteRune(r)
				switch d.lex.state(r) {
				case lexCont: // found something that looks like a value, so break out of whitespace loop
					break readWhitespace
				case lexError:
					return nil, d.lex.err
				case lexEnd:
					err := tstack.push(d.lex.token)
					if err != nil || tstack.done() {
						return val.Bytes(), err
					}
					// Here we'd usually continue on next iteration loop (which is safe
					// and valid), but since we know we don't have any leftovers, we can
					// just reset the lexer and keep attempting to read whitespace.
					d.lex.reset()
				case lexEndPrev:
					return nil, errInternal
				case lexIgnore:
					// keep on readin'
				}
			}
		}
		// read element
		for {
			r, rlength, err := d.rd.ReadRune()
			var ls lexState
			// ugh, this is not exactly perfect.
			switch {
			case err == io.EOF:
				ls = d.lex.eof()
			case err != nil:
				return nil, err
			default:
				d.lex.position++
				val.WriteRune(r)
				ls = d.lex.state(r)
			}
			switch ls {
			case lexCont:
				// keep going
			case lexIgnore:
				if err != io.EOF {
					return nil, errInternal
				} else {
					return nil, errNoneLeft
				}
			case lexEnd:
				ioErr := err
				err := tstack.push(d.lex.token)
				if err != nil || tstack.done() {
					return val.Bytes(), err
				}
				if ioErr == io.EOF /* && !tstack.done() */ {
					return nil, errNoneLeft
				}
				continue readElems
			case lexEndPrev: // if err == io.EOF then we cannot end up here. (Invariant forced by lexer)
				val.Truncate(val.Len() - rlength)
				d.hasLeftover = true
				d.leftover = r

				err := tstack.push(d.lex.token)
				if err != nil || tstack.done() {
					return val.Bytes(), err
				}
				continue readElems
			case lexError:
				return nil, d.lex.err
			}
		}
	}
}
