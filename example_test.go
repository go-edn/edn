package edn_test

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"olympos.io/encoding/edn"
)

func ExampleDecoder_AddTagFn_duration() {
	input := `#com.myapp/duration "2h30m"`

	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)
	dec.AddTagFn("com.myapp/duration", time.ParseDuration)

	var d time.Duration
	dec.Decode(&d)
	fmt.Println(d)

	input = `#com.myapp/duration "1moment"`
	rdr = strings.NewReader(input)
	dec = edn.NewDecoder(rdr)
	dec.AddTagFn("com.myapp/duration", time.ParseDuration)
	err := dec.Decode(&d)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// 2h30m0s
	// time: unknown unit moment in duration 1moment
}

func ExampleDecoder_AddTagFn_complex() {
	input := `#complex [14.5 15.5]`

	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)

	intoComplex := func(v [2]float64) (complex128, error) {
		return complex(v[0], v[1]), nil
	}
	err := dec.AddTagFn("complex", intoComplex)
	if err != nil {
		panic(err)
	}

	var cmplx complex128
	err = dec.Decode(&cmplx)
	if err != nil {
		panic(err)
	}
	fmt.Println(cmplx)
	// Output: (14.5+15.5i)
}

type Length interface {
	ToMetres() float64
}

type Foot float64

func (f Foot) ToMetres() float64 {
	return float64(f) * 0.3048
}

type Yard float64

func (y Yard) ToMetres() float64 {
	return float64(y) * 0.9144
}

type Metre float64

func (m Metre) ToMetres() float64 {
	return float64(m)
}

func ExampleDecoder_AddTagStruct_intoInterface() {
	// We can insert things into interfaces with tagged literals.
	// Let's assume we have
	// type Length interface { ToMetres() float64 }
	// and Foot, Yard and Metre which satisfy the Length interface

	input := `[#foot 14.5, #yard 2, #metre 3.0]`
	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)
	dec.AddTagStruct("foot", Foot(0))
	dec.AddTagStruct("yard", Yard(0))
	dec.AddTagStruct("metre", Metre(0))

	var lengths []Length
	dec.Decode(&lengths)
	for _, len := range lengths {
		fmt.Printf("%.2f\n", len.ToMetres())
	}
	// Output:
	// 4.42
	// 1.83
	// 3.00
}

func ExampleDecoder_AddTagStruct_nested() {
	// Tag structs and tag functions can nest arbitrarily.
	type Node struct {
		Left  *Node
		Val   int
		Right *Node
	}

	// function for finding the total sum of a tree
	var sumTree func(n Node) int
	sumTree = func(root Node) (val int) {
		if root.Left != nil {
			val += sumTree(*root.Left)
		}
		val += root.Val
		if root.Right != nil {
			val += sumTree(*root.Right)
		}
		return
	}

	input := `#node {:left #node {:val 1}
                   :val 2
                   :right #node {:left #node {:val 5}
                                 :val 8
                                 :right #node {:val 12}}}`
	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)
	dec.AddTagStruct("node", Node{})
	var node Node
	dec.Decode(&node)
	fmt.Println(sumTree(node))
	// Output: 28
}

func ExampleMathContext_global() {
	input := "1.12345678901234567890123456789012345678901234567890M"
	var val *big.Float
	edn.UnmarshalString(input, &val)
	fmt.Printf("%.50f\n", val)

	// override default precision
	mathContext := edn.GlobalMathContext
	edn.GlobalMathContext.Precision = 30
	edn.UnmarshalString(input, &val)
	fmt.Printf("%.50f\n", val)

	// revert it back to original values
	edn.GlobalMathContext = mathContext

	// Output:
	// 1.12345678901234567890123456789012345678901234567890
	// 1.12345678918063640594482421875000000000000000000000
}

func ExampleDecoder_UseMathContext() {
	input := "3.14159265358979323846264338327950288419716939937510M"

	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)

	mathContext := edn.GlobalMathContext
	// use global math context (does nothing)
	dec.UseMathContext(mathContext)

	var val *big.Float
	dec.Decode(&val)
	fmt.Printf("%.50f\n", val)

	// reread with smaller precision and rounding towards zero
	rdr = strings.NewReader(input)
	dec = edn.NewDecoder(rdr)

	mathContext.Precision = 30
	mathContext.Mode = big.ToZero
	dec.UseMathContext(mathContext)

	dec.Decode(&val)
	fmt.Printf("%.50f\n", val)

	// Output:
	// 3.14159265358979323846264338327950288419716939937510
	// 3.14159265160560607910156250000000000000000000000000
}

func ExampleKeyword() {
	const Friday = edn.Keyword("friday")
	fmt.Println(Friday)

	input := `:friday`
	var weekday edn.Keyword
	edn.UnmarshalString(input, &weekday)

	if weekday == Friday {
		fmt.Println("It is friday!")
	}
	// Output:
	// :friday
	// It is friday!
}

func ExampleRune() {
	runeSlice := []edn.Rune{'a', 'b', 'c', ',', ' ', '\n'}

	bs, _ := edn.Marshal(runeSlice)

	fmt.Println(string(bs))
	// Output: [\a \b \c \u002c \space \newline]
}

func ExampleTag_reading() {
	input := "#unknown ???"

	var tag edn.Tag
	edn.UnmarshalString(input, &tag)

	fmt.Printf("Tag with name %s and value %q of type %T\n", tag.Tagname, tag.Value, tag.Value)
	// Output: Tag with name unknown and value "???" of type edn.Symbol
}

// TODO: Lots and lots of unmarshalstring samples, refer to them in unmarshal

func ExampleUnmarshal_set() {
	// map[T]bool is considered as sets as well as maps
	var val map[int]bool

	edn.UnmarshalString("#{1 -5 42}", &val)
	fmt.Println(val[42], val[123]) // => true false

	edn.UnmarshalString("{1 false 2 true}", &val)
	fmt.Println(val[1], val[2]) // => false true

	// Output:
	// true false
	// false true
}

func ExampleMarshal_set() {
	// values of type map[T]bool and map[T]struct{} are encoded as EDN sets by
	// default
	val := map[int]bool{42: true}

	bs, _ := edn.Marshal(val)
	fmt.Println(string(bs)) // => #{42}

	val2 := map[string]struct{}{"hiccup": {}}
	bs, _ = edn.Marshal(val2)
	fmt.Println(string(bs)) // => #{"hiccup"}

	// Output:
	// #{42}
	// #{"hiccup"}
}

func ExampleMarshal_setOverride() {
	// You can specify that map[T]bool/struct{} are printed as EDN maps by using
	// the `map` keyword in the EDN struct tag:
	type Value struct {
		BoolMap   map[int]bool        `edn:"bool-map,map,omitempty"`
		StructMap map[string]struct{} `edn:"struct-map,map,omitempty"`
	}

	var val Value
	val.BoolMap = map[int]bool{2: false}
	bs, _ := edn.Marshal(val)
	fmt.Println(string(bs)) // => {:bool-map{2 false}}

	val.BoolMap = nil
	val.StructMap = map[string]struct{}{"foo": {}}
	bs, _ = edn.Marshal(val)
	fmt.Println(string(bs)) // =>  {:struct-map{"foo"{}}}

	// Output:
	// {:bool-map{2 false}}
	// {:struct-map{"foo"{}}}
}
