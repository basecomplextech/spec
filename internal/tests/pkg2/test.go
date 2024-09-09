// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package pkg2

import (
	"github.com/basecomplextech/baselibrary/tests"
	"github.com/basecomplextech/spec/internal/tests/pkg3/pkg3a"
)

func TestSubmessage(t tests.T, w SubmessageWriter) Submessage {
	w.Key("submessage")
	w.Value(pkg3a.Value{
		X: 1,
		Y: -1,
	})

	m, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	return m
}
