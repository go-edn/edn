package edn_test

import (
	"fmt"
	"io"
	"strings"

	"olympos.io/encoding/edn"
)

// Typically, you'd also include timestamps here. Imagine that they are here.
type LogEntry struct {
	Level       edn.Keyword
	Message     string `edn:"msg"`
	Environment string `edn:"env"`
	Service     string
}

// This example shows how one can do streaming with the decoder, and how to
// properly know when the stream has no elements left.
func Example_streaming() {
	const input = `
{:level :debug :msg "1 < 2 ? true" :env "dev" :service "comparer"}
{:level :warn :msg "slow response time from 127.0.1.39" :env "prod" :service "worker 10"}
{:level :warn :msg "worker 8 has been unavailable for 30s" :env "prod" :service "gateway"}
{:level :info :msg "new processing request: what.png" :env "prod" :service "gateway"}
{:level :debug :msg "1 < nil? error" :env "dev" :service "comparer"}
{:level :warn :msg "comparison failed: 1 < nil" :env "dev" :service "comparer"}
{:level :info :msg "received new processing request: what.png" :env "prod" :service "worker 3"}
{:level :warn :msg "bad configuration value :timeout, using 3h" :env "staging" :service "worker 3"}
`

	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)
	var err error
	for {
		var entry LogEntry
		err = dec.Decode(&entry)
		if err != nil {
			break
		}
		if entry.Level == edn.Keyword("warn") && entry.Environment != "dev" {
			fmt.Println(entry.Message)
		}
	}
	if err != nil && err != io.EOF {
		// Something bad happened to our reader
		fmt.Println(err)
		return
	}
	// If err == io.EOF then we've reached end of stream
	fmt.Println("End of stream!")
	// Output:
	// slow response time from 127.0.1.39
	// worker 8 has been unavailable for 30s
	// bad configuration value :timeout, using 3h
	// End of stream!
}
