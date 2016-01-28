package main

import (
	"flag"
	"fmt"
	"os"
)

const banner = `config - example of how to use EDN for configs
Usage: config --config conf-file

Read config.go for more information, and try it out with
'sample.config.edn'`

func main() {
	var confPath string
	flag.StringVar(&confPath, "config", "", "The path to the config file")
	flag.Parse()
	if confPath == "" {
		fmt.Println(banner)
		os.Exit(1)
	}
	c, err := ReadConf(confPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("The config in raw go format:")
	fmt.Printf("%#v\n", c)
}
