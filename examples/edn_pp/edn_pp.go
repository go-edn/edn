package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"olympos.io/encoding/edn"
)

func main() {
	check_help()

	d := edn.NewDecoder(os.Stdin)
	e := edn.NewEncoder(os.Stdout)

	var err error
	for {
		var val interface{}
		err = d.Decode(&val)
		if err != nil {
			break
		}
		err = e.EncodePPrint(val, nil)
		if err != nil {
			break
		}
	}
	if err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Reader error: %s\n", err.Error())
		os.Exit(1)
	}
}

const banner = `edn_pp prettyprints EDN

Usage: edn_pp

edn_pp reads EDN-encoded input from stdin and prints EDN-encoded output
to stdout. For more information about EDN, see
https://github.com/edn-format/edn

edn_pp will not distinguish between lists and vectors, and will always
print out vectors.

To print this information, call edn_pp with --help, -h, --version or -v.`

func check_help() {
	helps := []*bool{
		flag.Bool("help", false, ""),
		flag.Bool("h", false, ""),
		flag.Bool("version", false, ""),
		flag.Bool("v", false, ""),
	}
	flag.Parse()
	help := false
	for _, v := range helps {
		help = help || *v
	}
	if help {
		fmt.Println(banner)
		os.Exit(0)
	}

}
