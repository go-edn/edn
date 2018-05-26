package main

import (
	"fmt"
	"os"

	"olympos.io/encoding/edn"
)

type PermissionSet int

const (
	execKey  = edn.Keyword("exec")
	readKey  = edn.Keyword("read")
	writeKey = edn.Keyword("write")
)

func (p *PermissionSet) UnmarshalEDN(bs []byte) error {
	*p = 0
	var keywordSet map[edn.Keyword]bool
	err := edn.Unmarshal(bs, &keywordSet)
	if err != nil {
		return err
	}
	for key, _ := range keywordSet {
		switch key {
		case execKey:
			*p |= 1
		case writeKey:
			*p |= 2
		case readKey:
			*p |= 4
		default:
			return fmt.Errorf("Unknown permission type: %s", key)
		}
	}
	return nil
}

func (p PermissionSet) AsMap() map[edn.Keyword]bool {
	keywordSet := make(map[edn.Keyword]bool)
	if p&1 == 1 {
		keywordSet[execKey] = true
	}
	if p&2 == 2 {
		keywordSet[writeKey] = true
	}
	if p&4 == 4 {
		keywordSet[readKey] = true
	}
	return keywordSet
}

func (p PermissionSet) MarshalEDN() ([]byte, error) {
	return edn.Marshal(p.AsMap())
}

type Permission int

// SymbolicNotation returns the symbolic notation of p
func (p Permission) SymbolicNotation() string {
	perms := []byte("-rwxrwxrwx")
	for i, j := 1, uint(8); i < 10; i, j = i+1, j-1 {
		if (p>>j)&1 == 0 {
			perms[i] = '-'
		}
	}
	return string(perms)
}

// NumericNotation returns the numeric notation of p
func (p Permission) NumericNotation() string {
	return fmt.Sprintf("0%o%o%o", (p>>6)&7, (p>>3)&7, p&7)
}

func (p Permission) MarshalEDN() ([]byte, error) {
	var sets struct {
		User  map[edn.Keyword]bool `edn:",omitempty"`
		Group map[edn.Keyword]bool `edn:",omitempty"`
		Other map[edn.Keyword]bool `edn:",omitempty"`
	}
	sets.User = PermissionSet((p >> 6) & 7).AsMap()
	sets.Group = PermissionSet((p >> 3) & 7).AsMap()
	sets.Other = PermissionSet(p & 7).AsMap()
	return edn.Marshal(sets)
}

func (p *Permission) UnmarshalEDN(bs []byte) error {
	var sets struct {
		User  PermissionSet
		Group PermissionSet
		Other PermissionSet
	}
	err := edn.Unmarshal(bs, &sets)
	if err != nil {
		return err
	}
	*p |= Permission(sets.User) << 6
	*p |= Permission(sets.Group) << 3
	*p |= Permission(sets.Other)
	return nil
}

func main() {
	fmt.Println("Hello! If you give me a map on the following shape:")
	fmt.Println("{:user #{:read :write} :group #{:exec} :other #{}}")
	fmt.Println("I will print out its symbolic notation *and* numeric notation!")
	fmt.Println("(I will also print out a minimal map representation for you too)")
	d := edn.NewDecoder(os.Stdin)
	var perm Permission
	err := d.Decode(&perm)
	if err != nil {
		fmt.Println("Oops, an error occurred:")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("---")
	fmt.Println("Numeric :", perm.NumericNotation())
	fmt.Println("Symbolic:", perm.SymbolicNotation())
	bs, _ := edn.Marshal(perm)
	fmt.Println("Map     :", string(bs))
}
