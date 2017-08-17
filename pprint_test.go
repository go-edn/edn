// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"bytes"
	"testing"
)

func TestPPrintSimple(t *testing.T) {
	inputs := []string{
		"{}",
		"[]",
		"{:a 42}",
	}

	for _, input := range inputs {
		buff := bytes.NewBuffer(nil)
		if err := PPrint(buff, []byte(input), &PPrintOpts{}); err != nil {
			t.Errorf(`PPrint("%s") failed, but expected success: %v`, input, err)
		}

		output := string(buff.Bytes())
		if output != input {
			t.Errorf(`Expected PPrint("%s") to be "%s"; was "%s"`, input, input, output)
		}
	}
}
