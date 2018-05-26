package edn_test

import (
	"fmt"
	"strings"

	"olympos.io/encoding/edn"
)

type Notifiable interface {
	Notify()
}

type User struct {
	Username string
}

func (u User) MarshalEDN() ([]byte, error) {
	return edn.Marshal(edn.Tag{"myapp/user", u.Username})
}

func (u *User) Notify() {
	fmt.Printf("Notified user %s.\n", u.Username)
}

type Group struct {
	GroupID int
}

func (g Group) MarshalEDN() ([]byte, error) {
	return edn.Marshal(edn.Tag{"myapp/group", g.GroupID})
}

func (g *Group) Notify() {
	fmt.Printf("Notified group with id %d.\n", g.GroupID)
}

var notifyTagMap edn.TagMap

// We use a tagMap to avoid adding these values to the entire system.
func init() {
	err := notifyTagMap.AddTagFn("myapp/user", func(s string) (*User, error) {
		return &User{s}, nil
	})
	if err != nil {
		panic(err)
	}

	err = notifyTagMap.AddTagFn("myapp/group", func(id int) (*Group, error) {
		return &Group{id}, nil
	})
	if err != nil {
		panic(err)
	}
}

// This example shows how to read and write basic EDN tags, and how this can be
// utilised: In contrast to encoding/json, you can read in data where you only
// know that the input satisfies some sort of interface, provided the value is
// tagged.
func Example_polymorphicTags() {
	input := `[#myapp/user "eugeness"
             #myapp/group 10
             #myapp/user "jeannikl"
             #myapp/user "jeremiah"
             #myapp/group 100]`

	rdr := strings.NewReader(input)
	dec := edn.NewDecoder(rdr)
	dec.UseTagMap(&notifyTagMap)

	var toNotify []Notifiable
	err := dec.Decode(&toNotify)
	if err != nil {
		panic(err)
	}
	for _, notify := range toNotify {
		notify.Notify()
	}

	// Print out the values as well
	out, err := edn.Marshal(toNotify)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
	// Output:
	// Notified user eugeness.
	// Notified group with id 10.
	// Notified user jeannikl.
	// Notified user jeremiah.
	// Notified group with id 100.
	// [#myapp/user"eugeness" #myapp/group 10 #myapp/user"jeannikl" #myapp/user"jeremiah" #myapp/group 100]
}
