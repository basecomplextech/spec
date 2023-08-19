package generator

import (
	"testing"

	"github.com/basecomplextech/spec/lang/compiler"
)

func TestGenerator_Golang__should_generate_go_package(t *testing.T) {
	opts := compiler.Options{ImportPath: []string{"../../internal/tests"}}
	c, err := compiler.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	g := newGenerator()

	names := []string{"pkg1", "pkg2", "pkg3/pkg3a", "pkg4"}
	for _, name := range names {
		pkg1, err := c.Compile("../../internal/tests/" + name)
		if err != nil {
			t.Fatal(err)
		}

		out := "../../internal/tests/" + name
		if err := g.Golang(pkg1, out); err != nil {
			t.Fatal(err)
		}
	}
}
