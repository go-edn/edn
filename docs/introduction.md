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
