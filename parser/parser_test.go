package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSpec(t *testing.T) string {
	b, err := os.ReadFile("test.spec")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// file

func TestParser_Parse__should_parse_file(t *testing.T) {
	p := newParser()
	s := testSpec(t)

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.NotNil(t, file)
}

func TestParser_Parse__should_parse_empty_file(t *testing.T) {
	p := newParser()

	file, err := p.Parse("")
	if err != nil {
		t.Fatal(err)
	}

	require.NotNil(t, file)
}

// imports

func TestParser_Parse__should_parse_single_import(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
import (
	"import1"
)`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Imports, 1)
	module := file.Imports[0]
	assert.Equal(t, "import1", module.ID)
	assert.Equal(t, "", module.Alias)
}

func TestParser_Parse__should_parse_import_alias(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
import (
	alias "import1"
)`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Imports, 1)
	module := file.Imports[0]

	assert.Equal(t, "import1", module.ID)
	assert.Equal(t, "alias", module.Alias)
}

func TestParser_Parse__should_parse_multiple_imports(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
import (
	"import1"
	"import2"
	alias3 "import3"
)`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Imports, 3)

	module0 := file.Imports[0]
	module1 := file.Imports[1]
	module2 := file.Imports[2]
	assert.Equal(t, "import1", module0.ID)
	assert.Equal(t, "import2", module1.ID)
	assert.Equal(t, "import3", module2.ID)
	assert.Equal(t, "alias3", module2.Alias)
}

func TestParser_Parse__should_parse_empty_imports(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`import ()`)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, file.Imports, 0)
}

// enum

func TestParser_Parse__should_parse_enum(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
enum TestEnum {
	UNDEFINED = 0;
	ONE = 1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	assert.Equal(t, "TestEnum", def.Name)
	require.Equal(t, DefinitionEnum, def.Type)
	require.Len(t, def.Enum.Values, 2)

	value0 := def.Enum.Values[0]
	value1 := def.Enum.Values[1]

	assert.Equal(t, "UNDEFINED", value0.Name)
	assert.Equal(t, 0, value0.Value)
	assert.Equal(t, "ONE", value1.Name)
	assert.Equal(t, 1, value1.Value)
}

func TestParser_Parse__should_parse_empty_enum(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`enum TestEnum {}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	assert.Equal(t, "TestEnum", def.Name)
	assert.Len(t, def.Enum.Values, 0)
}

// message

func TestParser_Parse__should_parse_message(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	int32	1;
	field2	string	2;
}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	assert.Equal(t, "TestMessage", def.Name)
	require.Equal(t, DefinitionMessage, def.Type)
	require.Len(t, def.Message.Fields, 2)

	field0 := def.Message.Fields[0]
	field1 := def.Message.Fields[1]

	assert.Equal(t, "field1", field0.Name)
	assert.Equal(t, "int32", field0.Type.Name)
	assert.Equal(t, 1, field0.Tag)

	assert.Equal(t, "field2", field1.Name)
	assert.Equal(t, "string", field1.Type.Name)
	assert.Equal(t, 2, field1.Tag)
}

func TestParser_Parse__should_parse_empty_message(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`message TestMessage {}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	assert.Equal(t, "TestMessage", def.Name)
	assert.Len(t, def.Message.Fields, 0)
}

// struct

func TestParser_Parse__should_parse_struct(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
struct TestStruct {
	field1	int32;
	field2	string;
}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	assert.Equal(t, "TestStruct", def.Name)
	require.Equal(t, DefinitionStruct, def.Type)
	require.Len(t, def.Struct.Fields, 2)

	field0 := def.Struct.Fields[0]
	field1 := def.Struct.Fields[1]

	assert.Equal(t, "field1", field0.Name)
	assert.Equal(t, "int32", field0.Type.Name)
	assert.Equal(t, "field2", field1.Name)
	assert.Equal(t, "string", field1.Type.Name)
}

func TestParser_Parse__should_parse_empty_struct(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`struct TestStruct {}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	str := file.Definitions[0]

	assert.Equal(t, "TestStruct", str.Name)
	assert.Len(t, str.Struct.Fields, 0)
}

// type

func TestParser_Parse__should_parse_base_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	int32	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, KindInt32, type_.Kind)
	assert.Equal(t, "int32", type_.Name)
}

func TestParser_Parse__should_parse_nullable_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	*int32	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, KindNullable, type_.Kind)
	require.NotNil(t, type_.Element)

	assert.Equal(t, KindInt32, type_.Element.Kind)
	assert.Equal(t, "int32", type_.Element.Name)
}

func TestParser_Parse__should_parse_nullable_imported_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	*pkg.Message	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, KindNullable, type_.Kind)
	require.NotNil(t, type_.Element)

	assert.Equal(t, KindReference, type_.Element.Kind)
	assert.Equal(t, "Message", type_.Element.Name)
	assert.Equal(t, "pkg", type_.Element.Import)
}

func TestParser_Parse__should_parse_list_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	[]int32	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, KindList, type_.Kind)
	require.NotNil(t, type_.Element)

	assert.Equal(t, KindInt32, type_.Element.Kind)
	assert.Equal(t, "int32", type_.Element.Name)
}

func TestParser_Parse__should_parse_imported_list_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	[]pkg.Message	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, KindList, type_.Kind)
	require.NotNil(t, type_.Element)

	assert.Equal(t, KindReference, type_.Element.Kind)
	assert.Equal(t, "Message", type_.Element.Name)
	assert.Equal(t, "pkg", type_.Element.Import)
}

func TestParser_Parse__should_parse_imported_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	pkg.Message	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, KindReference, type_.Kind)
	assert.Equal(t, "Message", type_.Name)
	assert.Equal(t, "pkg", type_.Import)
}
