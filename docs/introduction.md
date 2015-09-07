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

## MarshalEDN and UnmarshalEDN



## EDN Tags
