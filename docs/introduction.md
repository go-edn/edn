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



### Sets

## MarshalEDN and UnmarshalEDN

## EDN Tags
