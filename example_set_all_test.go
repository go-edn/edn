package edn_test

import (
	"fmt"

	"olympos.io/encoding/edn"
)

type UserOption edn.Keyword

type UnknownUserOptionError UserOption

func (err UnknownUserOptionError) Error() string {
	return fmt.Sprintf("Unknown user option %s", edn.Keyword(err))
}

const (
	UnknownOption = UserOption("")
	ShowEmail     = UserOption("show-email")
	Notifications = UserOption("notifications")
	DailyEmail    = UserOption("daily-email")
	RememberMe    = UserOption("remember-me")
)

func ListUserOptions() []UserOption {
	return []UserOption{
		ShowEmail,
		Notifications,
		DailyEmail,
		RememberMe,
	}
}

func (uo *UserOption) UnmarshalEDN(bs []byte) error {
	var kw edn.Keyword
	err := edn.Unmarshal(bs, &kw)
	if err != nil {
		return err
	}
	opt := UserOption(kw)
	switch opt {
	case ShowEmail, Notifications, DailyEmail, RememberMe:
		*uo = opt
		return nil
	default:
		return UnknownUserOptionError(opt)
	}
}

type UserOptions map[UserOption]bool

func (opts *UserOptions) UnmarshalEDN(bs []byte) error {
	var kw edn.Keyword
	// try to decode into keyword first
	err := edn.Unmarshal(bs, &kw)
	if err == nil && kw == edn.Keyword("all") {
		// Put all options into the user option map
		*opts = UserOptions(make(map[UserOption]bool))
		for _, opt := range ListUserOptions() {
			(*opts)[opt] = true
		}
		return nil
	}
	// then try to decode into user map
	var rawOpts map[UserOption]bool
	err = edn.Unmarshal(bs, &rawOpts)
	*opts = UserOptions(rawOpts)
	return err
}

// This example shows how one can implement enums and sets, and how to support
// multiple different forms for a specific value type. The set implemented here
// supports the notation `:all` for all values.
func Example_enumsAndSets() {
	inputs := []string{
		"#{:show-email :notifications}",
		"#{:notifications :show-email :remember-me}",
		":all",
		"#{:doot-doot}",
		":none",
		"#{} ;; no options",
	}
	for _, input := range inputs {
		var opts UserOptions
		err := edn.UnmarshalString(input, &opts)
		if err != nil {
			fmt.Println(err)
			// Do proper error handling here if something fails
			continue
		}
		// Cannot print out a map, as its ordering is nondeterministic.
		fmt.Printf("show email? %t, notifications? %t, daily email? %t, remember me? %t\n",
			opts[ShowEmail], opts[Notifications], opts[DailyEmail], opts[RememberMe])
	}

	// Output:
	// show email? true, notifications? true, daily email? false, remember me? false
	// show email? true, notifications? true, daily email? false, remember me? true
	// show email? true, notifications? true, daily email? true, remember me? true
	// Unknown user option :doot-doot
	// edn: cannot unmarshal keyword into Go value of type map[edn_test.UserOption]bool
	// show email? false, notifications? false, daily email? false, remember me? false
}
