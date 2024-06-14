package parser

import (
	"os"
	"testing"

	"github.com/basecomplextech/spec/internal/lang/syntax"
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

func testServiceFile(t *testing.T) string {
	b, err := os.ReadFile("test_service.spec")
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
	def := file.Definitions[0]

	assert.Equal(t, "TestEnum", def.Name)
	require.Equal(t, syntax.DefinitionEnum, def.Type)
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
	require.Equal(t, syntax.DefinitionMessage, def.Type)
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
	require.Equal(t, syntax.DefinitionStruct, def.Type)
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

	assert.Equal(t, syntax.KindInt32, type_.Kind)
	assert.Equal(t, "int32", type_.Name)
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

	assert.Equal(t, syntax.KindList, type_.Kind)
	require.NotNil(t, type_.Element)

	assert.Equal(t, syntax.KindInt32, type_.Element.Kind)
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

	assert.Equal(t, syntax.KindList, type_.Kind)
	require.NotNil(t, type_.Element)

	assert.Equal(t, syntax.KindReference, type_.Element.Kind)
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

	assert.Equal(t, syntax.KindReference, type_.Kind)
	assert.Equal(t, "Message", type_.Name)
	assert.Equal(t, "pkg", type_.Import)
}

func TestParser_Parse__should_parse_any_message_type(t *testing.T) {
	p := newParser()

	file, err := p.Parse(`
message TestMessage {
	field1	message	1;
}`)
	if err != nil {
		t.Fatal(err)
	}

	def := file.Definitions[0]
	type_ := def.Message.Fields[0].Type

	assert.Equal(t, syntax.KindAnyMessage, type_.Kind)
	assert.Equal(t, "message", type_.Name)
}

// service

func TestParser_Parse__should_parse_service_file(t *testing.T) {
	p := newParser()
	s := testServiceFile(t)

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.NotNil(t, file)
}

func TestParser_Parse__should_parse_empty_service(t *testing.T) {
	p := newParser()
	s := `service Service {}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	assert.Equal(t, "Service", def.Name)
	assert.Equal(t, syntax.DefinitionService, def.Type)
	assert.Len(t, def.Service.Methods, 0)
}

// method

func TestParser_Parse__should_parse_empty_method(t *testing.T) {
	p := newParser()
	s := `service Service {
		method0();
		method1() ();
	}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]
	srv := def.Service

	assert.Len(t, srv.Methods, 2)
}

func TestParser_Parse__should_parse_method_args(t *testing.T) {
	p := newParser()
	s := `service Service {
		method(a bool 1, b int64 2, c string 3);
	}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	srv := def.Service
	require.Len(t, srv.Methods, 1)

	method := srv.Methods[0]
	assert.Len(t, method.Input, 3)
}

func TestParser_Parse__should_parse_method_input_reference(t *testing.T) {
	p := newParser()
	s := `service Service {
		method(Request);
	}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	srv := def.Service
	require.Len(t, srv.Methods, 1)

	method := srv.Methods[0]
	input := method.Input
	assert.NotNil(t, input)
	assert.IsType(t, &syntax.Type{}, input)
}

func TestParser_Parse__should_parse_method_oneway(t *testing.T) {
	p := newParser()
	s := `service Service {
		method() oneway;
	}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	srv := def.Service
	require.Len(t, srv.Methods, 1)

	method := srv.Methods[0]
	assert.True(t, method.Oneway)
}

func TestParser_Parse__should_parse_method_output_fields(t *testing.T) {
	p := newParser()
	s := `service Service {
		method() (a bool 1, b int64 2, c string 3);
	}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	srv := def.Service
	require.Len(t, srv.Methods, 1)

	method := srv.Methods[0]
	assert.Len(t, method.Output, 3)
}

func TestParser_Parse__should_parse_method_channel(t *testing.T) {
	p := newParser()
	s := `service Service {
		method0() ();
		method1() (<-In) ();
		method2() (Out->) ();
		method3() (<-In, Out->) ();
	}`

	file, err := p.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, file.Definitions, 1)
	def := file.Definitions[0]

	srv := def.Service
	require.Len(t, srv.Methods, 4)

	method0 := srv.Methods[0]
	assert.Nil(t, method0.Channel)

	method1 := srv.Methods[1]
	require.NotNil(t, method1.Channel)
	assert.NotNil(t, method1.Channel.In)
	assert.Nil(t, method1.Channel.Out)

	method2 := srv.Methods[2]
	require.NotNil(t, method2.Channel)
	assert.Nil(t, method2.Channel.In)
	assert.NotNil(t, method2.Channel.Out)

	method3 := srv.Methods[3]
	require.NotNil(t, method3.Channel)
	assert.NotNil(t, method3.Channel.In)
	assert.NotNil(t, method3.Channel.Out)
}

func TestParser_Parse__should_return_error_when_invalid_channel_syntax(t *testing.T) {
	p := newParser()
	s := `service Service {
		method() (Out->, <-In) ();
	}`

	_, err := p.Parse(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid channel syntax, expected (<-In, Out->), got (Out->, <-In)`)

	s = `service Service {
		method() (Msg<-) ();
	}`

	_, err = p.Parse(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid channel receive syntax, expected <-Msg, got Msg<-`)

	s = `service Service {
		method() (->Msg) ();
	}`

	_, err = p.Parse(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid channel send syntax, expected Msg->, got ->Msg`)
}
