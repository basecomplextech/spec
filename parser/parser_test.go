package parser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFile(t *testing.T) string {
	b, err := os.ReadFile("test.spec")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// file

func TestParser_Parse__should_parse_file(t *testing.T) {
	p := newParser()
	s := testFile(t)

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
	enum := file.Definitions[0].(*Enum)

	assert.Equal(t, "TestEnum", enum.Name)
	require.Len(t, enum.Values, 2)

	value0 := enum.Values[0]
	value1 := enum.Values[1]

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
	enum := file.Definitions[0].(*Enum)

	assert.Equal(t, "TestEnum", enum.Name)
	assert.Len(t, enum.Values, 0)
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
	msg := file.Definitions[0].(*Message)

	assert.Equal(t, "TestMessage", msg.Name)
	assert.Len(t, msg.Fields, 2)

	field0 := msg.Fields[0]
	field1 := msg.Fields[1]

	assert.Equal(t, "field1", field0.Name)
	assert.Equal(t, "int32", field0.Type.Ident)
	assert.Equal(t, 1, field0.Tag)

	assert.Equal(t, "field2", field1.Name)
	assert.Equal(t, "string", field1.Type.Ident)
	assert.Equal(t, 2, field1.Tag)
}

func TestParser_Parse__should_parse_empty_message(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`message TestMessage {}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	msg := file.Definitions[0].(*Message)

	assert.Equal(t, "TestMessage", msg.Name)
	assert.Len(t, msg.Fields, 0)
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
	str := file.Definitions[0].(*Struct)

	assert.Equal(t, "TestStruct", str.Name)
	assert.Len(t, str.Fields, 2)

	field0 := str.Fields[0]
	field1 := str.Fields[1]

	assert.Equal(t, "field1", field0.Name)
	assert.Equal(t, "int32", field0.Type.Ident)
	assert.Equal(t, "field2", field1.Name)
	assert.Equal(t, "string", field1.Type.Ident)
}

func TestParser_Parse__should_parse_empty_struct(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`struct TestStruct {}`)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	msg := file.Definitions[0].(*Struct)

	assert.Equal(t, "TestStruct", msg.Name)
	assert.Len(t, msg.Fields, 0)
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

	msg := file.Definitions[0].(*Message)
	type_ := msg.Fields[0].Type

	assert.Equal(t, KindBase, type_.Kind)
	assert.Equal(t, "int32", type_.Ident)
}

func TestParser_Parse__should_parse_list_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	*int32	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	msg := file.Definitions[0].(*Message)
	type_ := msg.Fields[0].Type

	assert.Equal(t, KindNullable, type_.Kind)
	assert.Equal(t, "int32", type_.Ident)
}

func TestParser_Parse__should_parse_nullable_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	[]int32	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	msg := file.Definitions[0].(*Message)
	type_ := msg.Fields[0].Type

	assert.Equal(t, KindList, type_.Kind)
	assert.Equal(t, "int32", type_.Ident)
}
