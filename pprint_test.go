// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"testing"
)

func TestPPrint(t *testing.T) {
	inputs := map[string]string{
		"{}":          "{}",
		"[]":          "[]",
		"{:a 42}":     "{:a 42}",
		"{:a 1 :b 2}": "{:a 1,\n :b 2}",
	}

	for input, expected := range inputs {
		buff := bytes.NewBuffer(nil)
		if err := PPrint(buff, []byte(input), &PPrintOpts{}); err != nil {
			t.Errorf(`PPrint(%q) failed, but expected success: %v`, input, err)
		}

		output := string(buff.Bytes())
		if output != expected {
			t.Errorf(`Expected PPrint(%q) to be %q; was %q`, input, expected, output)
		}
	}
}
