// Copyright 2021 Ivan Korobkov. All rights reserved.

package generator

import (
	"testing"

	"github.com/basecomplextech/spec/internal/lang/compiler"
)

func TestGenerator_Package__should_generate_go_package(t *testing.T) {
	opts := compiler.Options{ImportPath: []string{"../../tests"}}
	c, err := compiler.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	g := newGenerator(false /* do not skip rpc */)

	names := []string{"pkg1", "pkg2", "pkg3/pkg3a", "pkg4"}
	for _, name := range names {
		pkg1, err := c.Compile("../../tests/" + name)
		if err != nil {
			t.Fatal(err)
		}

		out := "../../tests/" + name
		if err := g.Package(pkg1, out); err != nil {
			t.Fatal(err)
		}
	}
}
