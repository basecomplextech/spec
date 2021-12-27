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

// Compile

func TestCompiler_Compile__should_create_package(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "pkg1", pkg.Name)
	assert.Equal(t, "testdata/pkg1", pkg.ID)
}

func TestCompiler_Compile__should_create_files(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Files, 2)

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Equal(t, "enum.spec", file0.Name)
	assert.Equal(t, "message.spec", file1.Name)
}

func TestCompiler_Compile__should_create_imports(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Len(t, file0.Imports, 0)
	assert.Len(t, file1.Imports, 1)
}

func TestCompiler_Compile__should_create_file_definitions(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Len(t, file0.Definitions, 1)
	assert.Len(t, file1.Definitions, 2)
}

func TestCompiler_Compile__should_create_package_definitions(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Definitions, 3)
}
