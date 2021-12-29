package generator

import (
	"testing"

	"github.com/baseone-run/spec/compiler"
)

func TestGenerator_Go__should_generate_go_package(t *testing.T) {
	opts := compiler.Options{ImportPath: []string{"../compiler/testdata"}}
	c, err := compiler.New(opts)
	if err != nil {
		t.Fatal(err)
	}

	pkg1, err := c.Compile("../compiler/testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	out := "generated"
	g := newGenerator()
	if err := g.Golang(pkg1, out); err != nil {
		t.Fatal(err)
	}
}
