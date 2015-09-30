# Why should I (not) use EDN?

If you wonder whether EDN is a suitable for your use case, then these
bulletpoints may help you decide.

## Pros

* It is easier to interact with Clojure programs. Note, however, that this
  doesn't necessarily make sense for ClojureScript. It may be better to use
  [transit](https://github.com/cognitect/transit-format) for that task.
* More datatypes: Sets, (true) integers, big integers and big floats,
  characters, maps, keyword, symbols. Through tags, you also get: Timestamps,
  UUIDs and your own defined types.
* [tagged elements](https://github.com/edn-format/edn#tagged-elements).
* Comments and discard tokens: Excellent for configuration files that may need
  documentation.
* As readable as JSON.
* Corollary: Better to write and interact with for humans.
* Is a drop-in replacement for `encoding/json` if you use basic functionality.
  (go-edn specific)
* Is not very strict on whether you use symbols, keywords or strings as keys
  for structs. (go-edn specific)

## Cons

* I haven't sacrificed goats to make its performance super good. It's based upon
  the JSON library, so its performance should in general be good, but the
  underlying implementation performs allocations which could be removed. So if
  you need extremely high performance, this might not be the data format to use
  (the same applies to `encoding/json`)
* EDN isn't the best format to use for interacting with JavaScript and the web.
* The quality of an EDN library in $LANG may vary, and may not even exist in the
  first place.
* Currently not very helpful error messaging. (go-edn specific)
* Does not check/validate equality. (go-edn specific)
* Is not very strict on whether you use symbols, keywords or strings as keys
  for structs. (go-edn specific)
