package main

import (
	"fmt"
	"os"
	"time"

	"github.com/satori/go.uuid"
	"olympos.io/encoding/edn"
)

type Config struct {
	Db struct {
		User     string
		Password string `edn:"pwd"`
		Host     string
		Port     uint16
	}
	Port          uint16
	Id            uuid.UUID
	MaxChildren   uint            `edn:"max-children"`
	ShutdownAfter time.Duration   `edn:"shutdown-after"`
	RootUsers     map[string]bool `edn:"root-users"`
	Environment   string          `edn:"env"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Specify the file to open as the first argument!")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	dec := edn.NewDecoder(f)
	dec.AddTagFn("uuid", uuid.FromString)
	dec.AddTagFn("duration", time.ParseDuration)

	var c Config

	err = dec.Decode(&c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Config (raw go):")
	fmt.Printf("%+v\n", c)
}
