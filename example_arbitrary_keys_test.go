package edn_test

import (
	"fmt"

	"olympos.io/encoding/edn"
)

type Point3 struct {
	X, Y, Z int64
}

type Unit struct {
	Type edn.Keyword
	HP   int
}

// EDN, in contrast to JSON, supports arbitrary values as keys.
func Example_arbitraryKeys() {
	input := `{{:x 1 :y 2 :z 3}    "Greybeard"
             {:y 10 :x 1 :z -10} "Blackwind"}`

	var locStarships map[Point3]string
	err := edn.UnmarshalString(input, &locStarships)
	if err != nil {
		panic(err)
	}

	p := Point3{1, 10, -10}

	fmt.Printf("Starship at location %v is %s\n", p, locStarships[p])

	input = `{[0 2]   {:type :scout :hp 55}
            [-3 10] {:type :villager :hp 25}
            [5 5]   {:type :bowman :hp 32}
            [5 6]   {:type :bowman :hp 29}}`

	var locUnits map[[2]int]Unit
	err = edn.UnmarshalString(input, &locUnits)
	if err != nil {
		panic(err)
	}

	loc := [2]int{5, 5}

	fmt.Printf("Unit at location %v is %+v\n", loc, locUnits[loc])

	// Output:
	// Starship at location {1 10 -10} is Blackwind
	// Unit at location [5 5] is {Type::bowman HP:32}
}
