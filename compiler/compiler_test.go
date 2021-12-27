package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCompiler(t *testing.T) *compiler {
	opts := Options{
		ImportPath: []string{"testdata"},
	}
	c, err := newCompiler(opts)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestCompiler_Compile__should_parse_package_files(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Files, 2)
}
