package generator

import (
	"testing"

	"github.com/complexl/spec/lang/compiler"
)

func TestGenerator_Golang__should_generate_go_package(t *testing.T) {
	opts := compiler.Options{ImportPath: []string{"../testdata"}}
	c, err := compiler.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	g := newGenerator()

	names := []string{"pkg1", "pkg2", "sub/pkg3"}
	for _, name := range names {
		pkg1, err := c.Compile("../testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}

		out := "../../testgen/golang/" + name
		if err := g.Golang(pkg1, out); err != nil {
			t.Fatal(err)
		}
	}
}
