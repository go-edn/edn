# Go-EDN

[EDN](https://github.com/edn-format/edn) (Extensible Data Notation) is a subset
of Clojure syntax, generally used to store or transfer data between different
applications. It looks very similar to JSON, but has a couple of differences
that makes it better for some use cases, and worse for others. Generally
speaking, if you use JSON as a data interchange format and do not need to speak
with JavaScript programs, EDN should be a drop-in replacement and in many cases
a better fit.

The go-edn package is designed to be very similar to the
[json package](http://golang.org/pkg/encoding/json/) shipped with Go, so it
should be easy to convert from JSON to EDN. It should also be easy to get going
if you are already familiar with the JSON library.

## Encoding

Encoding data is generally done with the Marshal function.

```go
func Marshal(v interface{}) ([]byte, error)
```

As with the JSON package, go-edn will be able to serialise and deserialise most
structs. For example, given the struct User:

```go
type User struct {
	Username   string
	Email      string
	Registered int64
}
```

We can marshal users into EDN by using edn.Marshal:

```go
m := User{"alice", "alice@example.com", 1441576365}
bs, err := edn.Marshal(m)
```

If err is not nil, the output should look something like this:

```go
bs == []byte(`{:username"alice":email"alice@example.com":registered 1441576365}`)
```

This is somewhat hard to read, though. If you do not need space efficient
results, you can use a more debugging-friendly output. `edn.MarshalIndent` works
like json's MarshalIndent, and the call

```go
bs, err := edn.MarshalIndent(m, "", "  ")
```

should yield the bytes

```clojure
{
  :username "alice",
  :email "alice@example.com",
  :registered 1441576365
}
```

This is, however, not a very Clojure-like way to pretty-print EDN. If you want
a more `pprint`-like result, you can use `edn.MarshalPPrint`:

```go
bs, err := edn.MarshalPPrint(m, nil)
```

which would yield

```clojure
{:username "alice",
 :email "alice@example.com",
 :registered 1441576365}
```

Like the JSON package, only data that can be represented as EDN values can be
encoded. This means that channels, complex, and function types cannot be
encoded. As EDN convey values, you do not have reference types, and as such,
marshalling circular types will lead to an infinite loop.

*Unlike* the JSON package, map keys can be any legal EDN value, provided it is
not equal to any other key:

```go
playerLocations := map[[2]int]string{
	[2]int{0, 2}:  "Alice",
	[2]int{1, -3}: "Thao",
}
bs, _ := MarshalPPrint(playerLocations, nil)
```

will happily be encoded into

```clojure
{[0 2] "Alice",
 [1 -3] "Thao"}
```

## Decoding

Decoding is done using the Unmarshal function

```go
func Unmarshal(data []byte, v interface{}) error
```

As with the JSON package, you have to specify where to store the contents first,
then call `edn.Unmarshal` with the byte slice to decode along with a pointer to
the location you want to store the data:

```go
var u User
err := edn.Unmarshal(bs, &u)
```

If our content is any of the results from Marshal call shown earlier, then `u`
will contain contents as if assigned like this:

```go
u := User{
	Username: "alice",
	Email: "alice@example.com",
	Registered: 1441576365,
}
```

go-edn utilises reflection to detect which struct field, if any, it should
attach to a value in an EDN-map. The priority of which field a value should be
assigned to (if any), is as follows:

1. Fields with exported tags that matches the key exactly
2. Exported fields in lowercase that matches the key
3. Exported fields matching, case insensitive

go-edn will look for both keys, symbols and strings keys that match fields.
Currently, if multiple keys maps to the same value, the latest value read is
used.

If go-edn mathes a particular field with a value, the field is ignored. If the
key does not match a particular field, the value is ignored. (The same semantics
as the JSON package)

## Struct Tags

There are a lot of different ways to tag struct fields to enable different kinds
of semantics, most of it related to marshalling. Let's have a look at the
different options:

```go
type Organisation struct {
	OrgName int `edn:"org-name"`
	Type string `edn:",omitempty"`
    UserIds []int64 `edn:"user-ids,list"`
    Plugins map[string]bool
    InternalData []byte `edn:"-"`
}
```

By standard rules, the `OrgName` field would be matched against the keyword
`:orgname`, the symbol `orgname`, the string `"orgname"` or any case insensitive
match of the previous mentioned. The first argument to the `edn` field tag is
always the name of the key to match against. OrgName will in this case match
against `:org-name`, `org-name` or `"org-name"` instead of the standard rules.


```go
	Type string `edn:",omitempty"`
```

The next field, `Type`, has no special name tied to it, consequently we leave
the name argument empty. To specify more than one argument, we delimit them by
using commas. The remaining arguments can be in any order – the only requirement
is that the name argument is the first.

`omitempty` omits the field if the value it points to is the zero value of its
type. For strings, this means that the string equals `""`. For all numbers, it
means that the number is equal to `0`, and for maps, pointers and slices, if
they are `nil` they are considered empty.

```go
    UserIds []int64 `edn:"user-ids,list"`
```

The `list` argument specifies that the given slice or array is printed as an EDN
list, rather than an EDN vector. So, if `UserIds` is `[]int{1, 2, 3}`, it will
be encoded as `(1 2 3)` instead of the usual `[1 2 3]`.

You can also use the `edn.List` type to encode arbitrary slices and arrays as
lists instead of vectors.

```go
    Plugins map[string]bool
```

The next field does not contain any tags, but shows a feature of go-edn we
haven't looked at yet: Sets. Any map of type `map[T]bool` or `map[T]struct{}`
will by default encode into sets. So if the plugin value is
`map[string]bool{"foo": true, "bar", true}`, it will encode it as `#{"foo"
"bar"}` instead of `{"foo" true, "bar" true}`.

Sometimes, this may be not be what you want -- perhaps false values actually
mean something other then absence, in which case you can turn on default map
encoding by setting the tag `map`. In our case, this means attaching the string
`` `edn:",map"` `` to the `Plugins` field.

You can also enforce the set notation if you want to, by setting the option
`set`. This has currently no effect, but it's intended to convert slices and
structs with only boolean values to sets in the future.

```go
    InternalData []byte `edn:"-"`
```

The last field is intended to be unexported during edn encoding, and is
therefore the name is marked set to `-`. This clashes with the keyword, symbol
and string key `:-`, `-` and `"-"` respectively, but this is assumed to be a
rare case and can be bypassed with the MarshalEDN and UnmarshalEDN functions.

### Keys

By default, keys in structs will be encoded as keywords. However, you can also
emit symbols and strings by setting the respective tags `sym` and `str`. You can
also set the keyword tag `key`, although this has no effect as of this writing.

```go
type Data {
	Value string
}

func main() {
	bs, _ := edn.MarshalPPrint(Data{"foo"}, nil)
	fmt.Println(string(bs))
}
```

will emit the following EDN:

```clj
{:value "foo"}  ;; if no edn tag is specified or is `edn:",key"`
{value "foo"}   ;; if the edn tag is `edn:",sym"`
{"value" "foo"} ;; if the edn tag is `edn:",str"`
```

These options have no effect on decoding, although it is intended to make the
decoding rules more strict in the future.

### Characters/Runes

The Go programming language ships with the rune type, and as the EDN
specification specifies that characters are values, it seems tempting to try it
out:

```go
bs, err := edn.MarshallPPrint([]rune{'h', 'e', 'l', 'l', 'o'}, nil)
```

However, the result is slightly surprising:

```clj
[104 101 108 108 111]
```

What happened? The rune type is actually just an alias for int32, and so all
reflection will believe this is an int32 slice. It's easier to see if you use
the fmt package:

```go
fmt.Printf("%#v\n", []rune{'h', 'e', 'l', 'l', 'o'})
// prints []int32{104, 101, 108, 108, 111}
```

This is slightly bothersome, but you can bypass this in two ways. One option is
by specifying the `rune` option for rune fields in the edn tag options. The
other option is, as you may guess, by replacing the rune type in your
application by the `edn.Rune` type.

## MarshalEDN and UnmarshalEDN

If you want finer control of marshalling and unmarshalling, you can let your
types implement the `edn.Marshaler` or `edn.Unmarshaler` interfaces. The
Marshaler interface gives you a way to implement your own encoding of a type,
and the Unmarshaler interface gives you a way to decode the type yourself.

The Marshaler implementation must emit a single, legal EDN value. The
Unmarshaler implementation must be able to read a single, legal EDN value. The
value may or may not contain comments and discard values, so it's recommended
that you leave this work to `go-edn` where possible.

## EDN Tags

One of the characteristic features of EDN is
[tagged elements](https://github.com/edn-format/edn#tagged-elements). Tagged
elements are fully supported in go-edn.

One of the built-in tagged elements are timestamps, which use the tag `#inst`:

```clj
#inst "2015-09-28T13:20:19.570-00:00"
```

In our original user example, we used `int64`s to represent the time when a user
was registered. We could have just as easily used `time.Time` for that.

```go
type User struct {
	Username   string
	Email      string
	Registered time.Time
}
```

We can marshal users in the exact same fashion.

```go
m := User{"alice", "alice@example.com", time.Now()}
bs, err := edn.MarshalPPrint(m, nil)
```

which would yield

```clj
{:username "Alice",
 :email "alice@example.com",
 :registered #inst "2015-09-06T21:52:45Z"}
```

(Where the timestamp will be your current time)

### Reading Tags

The only tags that are provided by go-edn by default are `#inst` and `#base64`.
If you want to add more ways to read tags, then it can be done in one out of two
ways:

- Providing a function that converts a specific type to some other type
- Providing the structure of the tagged element

As a general rule, you should _namespace_ your tags, or ensure that the tag is
unique in the context you use it.

#### Function-Based Conversion

Providing a function is easy: If you have a type T and want to convert it to U,
provide `edn.AddTagFn` with a function from type T to U, with an additional
error value if something went wrong:

```go
intoComplex := func(v [2]float64) (complex128, error) {
	return complex(v[0], v[1]), nil
}
err := edn.AddTagFn("complex", intoComplex)
// handle error
```

This will automatically turn values of shape `#complex [0.5, 0.6]` into complex
Go numbers.

Sometimes, libraries give you this function for free. For example, if you want
to add UUID support, you can use [go.uuid](https://github.com/satori/go.uuid)
and use the function `uuid.FromString` as argument:

```go
err := edn.AddTagFn("uuid", uuid.FromString)
// handle error
```

As a final example, let's have a look at how the internal init function that
adds the default tagged elements:

```go
func init() {
	err := AddTagFn("inst", func(s string) (time.Time, error) {
		return time.Parse(time.RFC3339Nano, s)
	})
	if err != nil {
		panic(err)
	}
	err = AddTagFn("base64", base64.StdEncoding.DecodeString)
	if err != nil {
		panic(err)
	}
}
``` 

#### Struct-Based Conversion

For convenience, there is a function `edn.AddTagStruct` that takes the tag name
and a struct:

```go
edn.AddTagStruct("mystruct", MyStruct{})
// is semantically equivalent to
edn.AddTagStruct("mystruct", func(s MyStruct) (MyStruct, error) { return s, nil })
```

Why would tagging a struct with the type it serialises to be useful? It ensures
that the type will only evaluate to its type, regardless of context: Any
`interface{}` this will be called with will be converted to MyStruct, instead of
a go map.

This can be useful if you do not know the shape of the input beforehand, but
still want to ensure it is of a type that satisfies an interface. For example,
consider these types and the interfaces they satsify.

```go
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
```

Now, if we attach tags that evaluate to their structs

```go
edn.AddTagFn("go-edn/rgb", func(r RGB) (RGB, error) { return r, nil })
edn.AddTagFn("go-edn/ycbcr", func(y YCbCr) (YCbCr, error) { return y, nil })
// or, more succinctly
edn.AddTagFn("go-edn/rgb", RGB{})
edn.AddTagFn("go-edn/ycbcr", YCbCr{})
```

We can now read a `[]Colour` without trouble

```go
s := `[#go-edn/ycbcr {:y 255 :cb 0 :cr -10}
       #go-edn/rgb {:r 98 :g 218 :b 255}]`
var colours []Colour
err := edn.Unmarshal([]byte(s), &colours)
// error handling..
for _, colour := range colours {
    fmt.Println(colour.Space())
}
```

It's recommended to provide these functions and structures to a `TagMap` instead
of directly manipulating global defaults if you don't have control of the entire
project, or if you want to change them safely later. See the documentation on
`TagMap` for more information.

#### Unknown Tags and Skipping Evaluation

If you receive a tag that go-edn does not know how to translate, it is returned
as an `edn.Tag`. The `edn.Tag` implements `MarshalEDN`, so you should be able to
pass it over to other services even though you don't know how to evaluate it.

There is no way to explicitly avoid evaluating tags yet, but if you do not want
to evaluate them and you know where they are located, you can set its type to
`edn.Tag`. When the type is `edn.Tag` (or any type that implements
UnmarshalEDN), it will not attempt to convert the instance.

### Writing Tags

There are no easy ways to "just write" tags yet. One option is to implement like
so:

```go
func (t MyVal) MarshalEDN() ([]byte, error) {
	return edn.Marshal(edn.Tag{"myapp/mytag", t.innerMarshal()})
}

func (t MyVal) innerMarshal() interface{} {
  return foo // to avoid infinite recursion
}
```

But it is somewhat clumsy and difficult to comprehend. A solution that should
solve this problem would be [#1](https://github.com/go-edn/edn/issues/1), but it
is not readily available yet.

## Big Numbers

go-edn supports big numbers out of the box. When numbers are appended with a
`N`, they are assumed to be of type `math/big.Int`, and when appended with `M`,
`math/big.Float`. The decoder will also attempt to coerce non-big types into big
ones if the type expected is big and vice versa.

Big floats do not have unlimited precision, but it can be configured globally or
on a decoder-per-decoder basis. In addition, the rounding mode can be set if
wanted. They exist in `edn.GlobalMathContext`, or you can make a
`edn.MatchContext` struct and pass it in to decoders using `UseMathContext`.
