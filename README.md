# Go implementation of EDN, extensible data notation

go-edn is a Golang library to read and write EDN (extensible data notation), a
subset of Clojure used for transferring data between applications, much like
JSON, YAML, or XML.

This library is heavily influenced by the JSON library that ships with Go, and
people familiar with that package should know the basics of how this library
works. In fact, this should be close to a drop-in replacement for the `json`
package if you only use basic functionality.

This implementation is fully working and (presumably) stable.

## Quickstart

You can follow http://blog.golang.org/json-and-go and replace every occurence of
JSON with EDN (and the JSON data with EDN data), and the text makes almost
perfect sense. The only caveat is that, since EDN is more general than JSON, go-edn
stores arbitrary maps on the form `map[interface{}]interface{}`.

go-edn also ships with keywords, symbols and tags as types.

For a longer introduction on how to use the library, see
[introduction.md](docs/introduction.md).

## Example Usage

Say you want to describe your pet forum's users as EDN. They have the following
types:

```go
type Animal struct {
	Name string
    Type string `edn:"kind"`
}

type Person struct {
	Name      string
	Birthyear int `edn:"born"`
	Pets      []Animal
}
```

With go-edn, we can do as follows to read and write these types:

```go
func ReturnData() (Person, error) {
	data := `{:name "Hans",
              :born 1970,
              :pets [{:name "Cap'n Jack" :kind "Sparrow"}
                     {:name "Freddy" :kind "Cockatiel"}]}`
	var user Person
	err := edn.Unmarshal([]byte(data), &user)
	// user '==' Person{"Hans", 1970,
	//             []Animal{{"Cap'n Jack", "Sparrow"}, {"Freddy", "Cockatiel"}}}
	return user, err
}
```

If you want to write that user again, just `Marshal` it:

```go
	bs, err := edn.Marshal(user)
```

## License

Copyright © 2015 Jean Niklas L'orange

Distributed under the BSD 3-clause license, which is available in the file
COPYING.
