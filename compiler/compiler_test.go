package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCompiler(t *testing.T) *compiler {
	opts := Options{
		ImportPath: []string{"../testdata"},
	}
	c, err := newCompiler(opts)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// Package

func TestCompiler__should_compile_package(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "pkg1", pkg.Name)
	assert.Equal(t, "../testdata/pkg1", pkg.ID)
}

// File

func TestCompiler__should_compile_files(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Files, 2)

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Equal(t, "enum.spec", file0.Name)
	assert.Equal(t, "package.spec", file1.Name)
}

// Imports

func TestCompiler__should_compile_imports(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Len(t, file0.Imports, 0)
	assert.Len(t, file1.Imports, 1)
}

func TestCompiler__should_resolve_imports(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file := pkg.FileNames["package.spec"]
	require.NotNil(t, file)

	imp := file.Imports[0]
	assert.True(t, imp.Resolved)

	pkg2 := imp.Package
	require.NotNil(t, pkg2)

	assert.Equal(t, "pkg2", pkg2.ID)
}

func TestCompiler__should_recursively_resolve_imports(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file := pkg.FileNames["package.spec"]
	require.NotNil(t, file)

	imp := file.Imports[0]
	assert.True(t, imp.Resolved)

	pkg2 := imp.Package
	require.NotNil(t, pkg2)

	imp2 := pkg2.Files[0].Imports[0]
	require.NotNil(t, imp2)

	assert.Equal(t, "sub/pkg3", imp2.ID)
	assert.NotNil(t, imp2.Package)
	assert.True(t, imp2.Resolved)
}

// Options

func TestCompiler__should_compile_options(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]
	assert.Len(t, file0.Options, 0)
	assert.Len(t, file1.Options, 1)

	gopkg := file1.OptionMap["go_package"]
	require.NotNil(t, gopkg)
	assert.Equal(t, "github.com/baseone-run/spec/generated/golang/pkg2", gopkg.Value)
}

// Definitions

func TestCompiler__should_compile_file_definitions(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Len(t, file0.Definitions, 1)
	assert.Len(t, file1.Definitions, 3)
}

func TestCompiler__should_compile_package_definitions(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Definitions, 4)

	assert.Contains(t, pkg.DefinitionNames, "Enum")
	assert.Contains(t, pkg.DefinitionNames, "Message")
	assert.Contains(t, pkg.DefinitionNames, "Node")
	assert.Contains(t, pkg.DefinitionNames, "Struct")
}

// Enums

func TestCompiler__should_compile_enum(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].Definitions[0]
	assert.Equal(t, DefinitionEnum, def.Type)
	assert.NotNil(t, def.Enum)
	assert.Len(t, def.Enum.Values, 5)
}

func TestCompiler__should_compile_enum_values(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].Definitions[0]
	require.Equal(t, DefinitionEnum, def.Type)

	enum := def.Enum
	assert.Contains(t, enum.ValueNumbers, 0)
	assert.Contains(t, enum.ValueNumbers, 1)
	assert.Contains(t, enum.ValueNumbers, 2)
	assert.Contains(t, enum.ValueNumbers, 3)
	assert.Contains(t, enum.ValueNumbers, 10)
}

func TestCompiler__should_compile_enum_value_names(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].Definitions[0]
	require.Equal(t, DefinitionEnum, def.Type)

	enum := def.Enum
	assert.Contains(t, enum.ValueNames, "UNDEFINED")
	assert.Contains(t, enum.ValueNames, "ONE")
	assert.Contains(t, enum.ValueNames, "TWO")
	assert.Contains(t, enum.ValueNames, "THREE")
	assert.Contains(t, enum.ValueNames, "TEN")
}

// Messages

func TestCompiler__should_compile_message(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].Definitions[0]
	assert.Equal(t, DefinitionMessage, def.Type)
	assert.NotNil(t, def.Message)
	assert.Len(t, def.Message.Fields, 22)
}

func TestCompiler__should_compile_message_field_names(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].Definitions[0]
	require.Equal(t, DefinitionMessage, def.Type)

	msg := def.Message
	require.Len(t, def.Message.FieldNames, 22)
	assert.Contains(t, msg.FieldNames, "field_bool")
	assert.Contains(t, msg.FieldNames, "field_enum")
	assert.Contains(t, msg.FieldNames, "field_int8")
}

func TestCompiler__should_compile_message_field_tags(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].Definitions[0]
	require.Equal(t, DefinitionMessage, def.Type)

	msg := def.Message
	require.Len(t, def.Message.FieldTags, 22)
	assert.Contains(t, msg.FieldTags, 1)
	assert.Contains(t, msg.FieldTags, 2)
	assert.Contains(t, msg.FieldTags, 10)
}

// Structs

func TestCompiler__should_compile_struct(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Struct"]
	assert.Equal(t, DefinitionStruct, def.Type)
	assert.NotNil(t, def.Struct)
	assert.Len(t, def.Struct.Fields, 2)
}

func TestCompiler__should_compile_struct_field_names(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Struct"]
	assert.Equal(t, DefinitionStruct, def.Type)

	str := def.Struct
	require.NotNil(t, str)
	require.Len(t, str.Fields, 2)

	assert.Contains(t, str.FieldNames, "key")
	assert.Contains(t, str.FieldNames, "value")
}

// Types

func TestCompiler__should_compile_builtin_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.FieldNames["field_bool"]
	require.NotNil(t, field)

	type_ := field.Type
	assert.Equal(t, "bool", type_.Name)
	assert.Equal(t, KindBool, type_.Kind)
}

func TestCompiler__should_compile_reference_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.FieldNames["msg"]
	require.NotNil(t, field)

	// resolved
	type_ := field.Type
	assert.Equal(t, "Node", type_.Name)
	assert.Equal(t, KindMessage, type_.Kind)
	assert.NotNil(t, type_.Ref)
	assert.Equal(t, "Node", type_.Ref.Name)
}

func TestCompiler__should_compile_imported_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.FieldNames["imported"]
	require.NotNil(t, field)

	// resolved
	type_ := field.Type
	assert.Equal(t, "SubMessage", type_.Name)
	assert.Equal(t, "pkg2", type_.ImportName)
	assert.Equal(t, KindMessage, type_.Kind)
	assert.NotNil(t, type_.Ref)
	assert.NotNil(t, type_.Import)
}

func TestCompiler__should_compile_list_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.FieldNames["list_values"]
	require.NotNil(t, field)

	// list
	type_ := field.Type
	assert.Equal(t, KindList, type_.Kind)

	// element
	elem := type_.Element
	require.NotNil(t, elem)
	assert.Equal(t, KindStruct, elem.Kind)
}

func TestCompiler__should_compile_list_reference_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.FieldNames["list_messages"]
	require.NotNil(t, field)

	// list
	type_ := field.Type
	assert.Equal(t, KindList, type_.Kind)

	// element
	elem := type_.Element
	require.NotNil(t, elem)
	assert.Equal(t, KindMessage, elem.Kind)
	assert.Equal(t, "Node", elem.Name)
	assert.NotNil(t, elem.Ref)
}

func TestCompiler__should_compile_list_imported_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../testdata/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.FieldNames["list_imported"]
	require.NotNil(t, field)

	// list
	type_ := field.Type
	assert.Equal(t, KindList, type_.Kind)
	require.NotNil(t, type_.Element)

	// element
	elem := type_.Element
	assert.Equal(t, KindMessage, elem.Kind)
	assert.Equal(t, "SubMessage", elem.Name)
	assert.Equal(t, "pkg2", elem.ImportName)
	assert.NotNil(t, elem.Ref)
	assert.NotNil(t, elem.Import)
}
