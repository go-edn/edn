// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"fmt"
	"strings"
)

type Keyword string

func (k Keyword) String() string {
	return fmt.Sprintf(":%s", string(k))
}

func (k Keyword) MarshalEDN() ([]byte, error) {
	return []byte(k.String()), nil
}

type Symbol string

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) MarshalEDN() ([]byte, error) {
	return []byte(s), nil
}

type Vector []interface{}

func (v Vector) String() string {
	vals := []string{}
	for _, val := range v {
		vals = append(vals, fmt.Sprint(val))
	}
	return "[" + strings.Join(vals, " ") + "]"
}

type List []interface{}

func (l List) String() string {
	vals := []string{}
	for _, val := range l {
		vals = append(vals, fmt.Sprint(val))
	}
	return "(" + strings.Join(vals, " ") + ")"
}

type Tag struct {
	Tagname string
	Value   interface{}
}

func (t Tag) String() string {
	return fmt.Sprintf("#%s %s", t.Tagname, t.Value)
}

type rawTag struct {
	tagname   string
	value     []byte
	valueType tokenType
}

type Set map[interface{}]bool

func (s Set) String() string {
	keys := []string{}
	for k, _ := range s {
		keys = append(keys, fmt.Sprint(k))
	}
	return "#{" + strings.Join(keys, " ") + "}"
}

type Map map[interface{}]interface{}

func (m Map) String() string {
	kvs := []string{}
	for k, v := range m {
		kvs = append(kvs, fmt.Sprintf("%s %s", k, v))
	}
	return "{" + strings.Join(kvs, ", ") + "}"
}
