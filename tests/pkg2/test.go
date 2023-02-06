package pkg2

import (
	"github.com/complex1tech/baselibrary/tests"
	"github.com/complex1tech/spec/tests/pkg3/pkg3a"
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
