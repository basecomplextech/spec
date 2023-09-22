package compiler

import (
	"testing"

	"github.com/basecomplextech/spec/internal/lang/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCompiler(t *testing.T) *compiler {
	opts := Options{
		ImportPath: []string{"../../tests"},
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

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "pkg1", pkg.Name)
	assert.Equal(t, "../../tests/pkg1", pkg.ID)
}

// File

func TestCompiler__should_compile_files(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Files, 2)

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]

	assert.Equal(t, "enum.spec", file0.Name)
	assert.Equal(t, "pkg1.spec", file1.Name)
}

// Imports

func TestCompiler__should_compile_imports(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
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

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file := pkg.FileNames["pkg1.spec"]
	require.NotNil(t, file)

	imp := file.Imports[0]
	assert.True(t, imp.Resolved)

	pkg2 := imp.Package
	require.NotNil(t, pkg2)

	assert.Equal(t, "pkg2", pkg2.ID)
}

func TestCompiler__should_recursively_resolve_imports(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file := pkg.FileNames["pkg1.spec"]
	require.NotNil(t, file)

	imp := file.Imports[0]
	assert.True(t, imp.Resolved)

	pkg2 := imp.Package
	require.NotNil(t, pkg2)

	imp2 := pkg2.Files[0].Imports[0]
	require.NotNil(t, imp2)

	assert.Equal(t, "pkg3/pkg3a", imp2.ID)
	assert.NotNil(t, imp2.Package)
	assert.True(t, imp2.Resolved)
}

// Options

func TestCompiler__should_compile_options(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	file0 := pkg.Files[0]
	file1 := pkg.Files[1]
	assert.Len(t, file0.Options, 0)
	assert.Len(t, file1.Options, 1)

	gopkg := file1.OptionMap["go_package"]
	require.NotNil(t, gopkg)
	assert.Equal(t, "github.com/basecomplextech/spec/internal/tests/pkg1", gopkg.Value)
}

// model.Definitions

func TestCompiler__should_compile_file_definitions(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
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

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, pkg.Definitions, 4)

	assert.Contains(t, pkg.DefinitionNames, "Enum")
	assert.Contains(t, pkg.DefinitionNames, "Message")
	assert.Contains(t, pkg.DefinitionNames, "Submessage")
	assert.Contains(t, pkg.DefinitionNames, "Struct")
}

// Enums

func TestCompiler__should_compile_enum(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].Definitions[0]
	assert.Equal(t, model.DefinitionEnum, def.Type)
	assert.NotNil(t, def.Enum)
	assert.Len(t, def.Enum.Values, 5)
}

func TestCompiler__should_compile_enum_values(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].Definitions[0]
	require.Equal(t, model.DefinitionEnum, def.Type)

	enum := def.Enum
	assert.Contains(t, enum.ValueNumbers, 0)
	assert.Contains(t, enum.ValueNumbers, 1)
	assert.Contains(t, enum.ValueNumbers, 2)
	assert.Contains(t, enum.ValueNumbers, 3)
	assert.Contains(t, enum.ValueNumbers, 10)
}

func TestCompiler__should_compile_enum_value_names(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].Definitions[0]
	require.Equal(t, model.DefinitionEnum, def.Type)

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

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].Definitions[0]
	assert.Equal(t, model.DefinitionMessage, def.Type)
	assert.NotNil(t, def.Message)
	assert.Len(t, def.Message.Fields.List, 26)
}

func TestCompiler__should_compile_message_field_names(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].Definitions[0]
	require.Equal(t, model.DefinitionMessage, def.Type)

	msg := def.Message
	require.Len(t, def.Message.Fields.List, 26)
	assert.Contains(t, msg.Fields.Names, "bool")
	assert.Contains(t, msg.Fields.Names, "enum1")
	assert.Contains(t, msg.Fields.Names, "byte")
}

func TestCompiler__should_compile_message_field_tags(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].Definitions[0]
	require.Equal(t, model.DefinitionMessage, def.Type)

	msg := def.Message
	require.Len(t, def.Message.Fields.Tags, 26)
	assert.Contains(t, msg.Fields.Tags, 1)
	assert.Contains(t, msg.Fields.Tags, 2)
	assert.Contains(t, msg.Fields.Tags, 10)
}

// Structs

func TestCompiler__should_compile_struct(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Struct"]
	assert.Equal(t, model.DefinitionStruct, def.Type)
	assert.NotNil(t, def.Struct)
	assert.Equal(t, def.Struct.Fields.Len(), 2)
}

func TestCompiler__should_compile_struct_field_names(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Struct"]
	assert.Equal(t, model.DefinitionStruct, def.Type)

	str := def.Struct
	require.NotNil(t, str)
	require.Equal(t, str.Fields.Len(), 2)

	assert.True(t, str.Fields.Contains("key"))
	assert.True(t, str.Fields.Contains("value"))
}

// Types

func TestCompiler__should_compile_builtin_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.Fields.Names["bool"]
	require.NotNil(t, field)

	type_ := field.Type
	assert.Equal(t, "bool", type_.Name)
	assert.Equal(t, model.KindBool, type_.Kind)
}

func TestCompiler__should_compile_reference_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.Fields.Names["submessage"]
	require.NotNil(t, field)

	// Resolved
	type_ := field.Type
	assert.Equal(t, "Submessage", type_.Name)
	assert.Equal(t, model.KindMessage, type_.Kind)
	assert.NotNil(t, type_.Ref)
	assert.Equal(t, "Submessage", type_.Ref.Name)
}

func TestCompiler__should_compile_imported_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.Fields.Names["submessage1"]
	require.NotNil(t, field)

	// Resolved
	type_ := field.Type
	assert.Equal(t, "Submessage", type_.Name)
	assert.Equal(t, "pkg2", type_.ImportName)
	assert.Equal(t, model.KindMessage, type_.Kind)
	assert.NotNil(t, type_.Ref)
	assert.NotNil(t, type_.Import)
}

func TestCompiler__should_compile_list_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.Fields.Names["structs"]
	require.NotNil(t, field)

	// List
	type_ := field.Type
	assert.Equal(t, model.KindList, type_.Kind)

	// Element
	elem := type_.Element
	require.NotNil(t, elem)
	assert.Equal(t, model.KindStruct, elem.Kind)
}

func TestCompiler__should_compile_list_reference_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.Fields.Names["submessages"]
	require.NotNil(t, field)

	// List
	type_ := field.Type
	assert.Equal(t, model.KindList, type_.Kind)

	// Element
	elem := type_.Element
	require.NotNil(t, elem)
	assert.Equal(t, model.KindMessage, elem.Kind)
	assert.Equal(t, "Submessage", elem.Name)
	assert.NotNil(t, elem.Ref)
}

func TestCompiler__should_compile_list_imported_type(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg1")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[1].DefinitionNames["Message"]
	require.NotNil(t, def.Message)

	field := def.Message.Fields.Names["submessages1"]
	require.NotNil(t, field)

	// List
	type_ := field.Type
	assert.Equal(t, model.KindList, type_.Kind)
	require.NotNil(t, type_.Element)

	// Element
	elem := type_.Element
	assert.Equal(t, model.KindMessage, elem.Kind)
	assert.Equal(t, "Submessage", elem.Name)
	assert.Equal(t, "pkg2", elem.ImportName)
	assert.NotNil(t, elem.Ref)
	assert.NotNil(t, elem.Import)
}

// Service

func TestCompiler__should_compile_service(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg4")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].DefinitionNames["Service"]
	require.NotNil(t, def.Service)

	m0 := def.Service.MethodNames["method"]
	require.NotNil(t, m0)
}

func TestCompiler__should_not_generate_request_from_primitive_fields(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg4")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].DefinitionNames["Service"]
	require.NotNil(t, def.Service)

	m := def.Service.MethodNames["method2"]
	require.NotNil(t, m)

	assert.Nil(t, m.Input)
	assert.NotNil(t, m.InputFields)
}

func TestCompiler__should_generate_request_from_complex_fields(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg4")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].DefinitionNames["Service"]
	require.NotNil(t, def.Service)

	m := def.Service.MethodNames["method4"]
	require.NotNil(t, m)

	in := m.Input
	require.NotNil(t, in)
	assert.Nil(t, m.InputFields)

	assert.Equal(t, "ServiceMethod4Request", in.Name)
	assert.True(t, in.Ref.Message.Generated)
}

func TestCompiler__should_not_generate_response_from_primitive_fields(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg4")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].DefinitionNames["Service"]
	require.NotNil(t, def.Service)

	m := def.Service.MethodNames["method10"]
	require.NotNil(t, m)

	assert.Nil(t, m.Output)
	assert.NotNil(t, m.OutputFields)
}

func TestCompiler__should_generate_response_from_fields(t *testing.T) {
	c := testCompiler(t)

	pkg, err := c.Compile("../../tests/pkg4")
	if err != nil {
		t.Fatal(err)
	}

	def := pkg.Files[0].DefinitionNames["Service"]
	require.NotNil(t, def.Service)

	m := def.Service.MethodNames["method11"]
	require.NotNil(t, m)

	out := m.Output
	require.NotNil(t, out)
	assert.Nil(t, m.OutputFields)

	assert.Equal(t, "ServiceMethod11Response", out.Name)
	assert.True(t, out.Ref.Message.Generated)
}
