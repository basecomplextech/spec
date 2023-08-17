// --- Declarations ---
%{
package parser

import (
	"fmt"

	"github.com/basecomplextech/spec/lang/ast"
)
%}

// union defines yySymType body.
%union {
	// Tokens
	ident   string
	integer int
	string  string

    // Type
	type_ *ast.Type

	// Import
	import_ *ast.Import
	imports []*ast.Import

	// Option
	option  *ast.Option
	options []*ast.Option

	// Definition
	definition  *ast.Definition
	definitions []*ast.Definition

	// Enum
	enum_value  *ast.EnumValue
	enum_values []*ast.EnumValue

	// Message
	message_field  *ast.MessageField
	message_fields []*ast.MessageField

	// Struct
	struct_field  *ast.StructField
	struct_fields []*ast.StructField

    // Service
    service         *ast.Service
    method          *ast.Method
    methods         []*ast.Method
    method_arg      *ast.MethodArg
    method_args     []*ast.MethodArg
    method_result   *ast.MethodResult
    method_results  []*ast.MethodResult
}

// keywords
%token ANY
%token ENUM
%token IMPORT
%token MESSAGE
%token OPTIONS
%token STRUCT
%token SERVICE

// general
%token <ident>      IDENT
%token <integer>    INTEGER
%token <string>     STRING
%token <ident>      MESSAGE
%token <ident>      SERVICE
%type  <ident>      keyword
%type  <ident>      field_name

// import
%type <import_> import
%type <imports> import_list
%type <imports> imports

// option
%type <option>  option
%type <options> option_list
%type <options> options

// type
%type <type_> type
%type <type_> base_type

// definitions
%type <definition>  definition
%type <definitions> definitions

// enum
%type <definition>  enum
%type <enum_value>  enum_value
%type <enum_values> enum_values

// message
%type <definition>      message
%type <message_field>   message_field
%type <message_fields> 	message_fields

// struct
%type <definition>      struct
%type <struct_field>    struct_field
%type <struct_fields>   struct_fields

// service
%type <definition>      service
%type <methods>         methods
%type <method>          method
%type <method_args>     method_arg_list
%type <method_args>     method_args
%type <method_arg>      method_arg
%type <method_results>  method_result_list
%type <method_results>  method_results
%type <method_result>   method_result

// start
%start file

%%

// field_name is any field, method, argument or result name.
// it cannot appear at the top level as a definition name.
field_name:
    IDENT
    {
        $$ = $1
    }
	| keyword
	{
		$$ = $1
	};

keyword:
	ANY
	{
		$$ = "any"
	}
    | IMPORT
    {
        $$ = "import"
    }
	| MESSAGE
    {
        $$ = "message"
    }
	| OPTIONS
    {
        $$ = "options"
    }
	| STRUCT
    {
        $$ = "struct"
    }
	| SERVICE
	{
		$$ = "service"
	};

// file

file: imports options definitions
	{ 
		file := &ast.File{
			Imports:     $1,
			Options:     $2,
			Definitions: $3,
		}
		setLexerResult(yylex, file)
	};

// import

import:
	STRING
	{ 
		if debugParser {
			fmt.Println("import ", $1)
		}
		$$ = &ast.Import{
			ID: trimString($1),
		}
	}
	| IDENT STRING
	{
		if debugParser {
			fmt.Println("import ", $1, $2)
		}
		$$ = &ast.Import{
			Alias: $1,
			ID:    trimString($2),
		}
	};

import_list:
	// Empty
	{ 
		$$ = nil
	}
	| import_list import
	{
		if debugParser {
			fmt.Println("import_list", $1, $2)
		}
		$$ = append($$, $2)
	};

imports:
	// Empty
	{ 
		$$ = nil
	}
	| IMPORT '(' import_list ')'
	{
		if debugParser {
			fmt.Println("imports", $3)
		}
		$$ = append($$, $3...)
	};

// options

options:
	// Empty
	{ 
		$$ = nil
	}
	| OPTIONS '(' option_list ')'
	{
		if debugParser {
			fmt.Println("options", $3)
		}
		$$ = append($$, $3...)
	};

option_list:
	// Empty
	{ 
		$$ = nil
	}
	| option_list option
	{
		if debugParser {
			fmt.Println("option_list", $1, $2)
		}
		$$ = append($$, $2)
	};

option:
	IDENT '=' STRING
	{
		if debugParser {
			fmt.Println("option ", $1, $3)
		}
		$$ = &ast.Option{
			Name:  $1,
			Value: trimString($3),
		}
	};

// type

type:
	base_type
	{
		if debugParser {
			fmt.Printf("type *%v\n", $1)
		}
		$$ = $1
	};
	| '[' ']' base_type
	{
		if debugParser {
			fmt.Printf("type []%v\n", $3)
		}
		$$ = &ast.Type{
			Kind:    ast.KindList,
			Element: $3,
		}
	};

base_type:
	IDENT
	{
		if debugParser {
			fmt.Println("base type", $1)
		}
		$$ = &ast.Type{
			Kind: ast.GetKind($1),
			Name: $1,
		}
	}
	| IDENT '.' IDENT
	{
		if debugParser {
			fmt.Printf("base type %v.%v\n", $1, $3)
		}
		$$ = &ast.Type{
			Kind:   ast.KindReference,
			Name:   $3,
			Import: $1,
		}
	}
	| ANY
	{
		if debugParser {
			fmt.Println("base type", "any")
		}
		$$ = &ast.Type{
			Kind: ast.KindAny,
			Name: "any",
		}
	}
	| MESSAGE
	{
		if debugParser {
			fmt.Println("base type", "message")
		}
		$$ = &ast.Type{
			Kind: ast.KindAnyMessage,
			Name: "message",
		}
	};

// definition

definition: 
	enum 
	| message
	| struct
    | service
    ;

definitions:
	// Empty
	{ 
		$$ = nil
	}
	| definitions definition
	{
		if debugParser {
			fmt.Println("definitions", $1, $2)
		}
		$$ = append($$, $2)
	};


// enum

enum: ENUM IDENT '{' enum_values '}'
	{
		if debugParser {
			fmt.Println("enum", $2, $4)
		}
		$$ = &ast.Definition{
			Type: ast.DefinitionEnum,
			Name: $2,

			Enum: &ast.Enum{
				Values: $4,
			},
		}
	};

enum_value: field_name '=' INTEGER ';'
	{
		if debugParser {
			fmt.Println("enum value", $1, $3)
		}
		$$ = &ast.EnumValue{
			Name: $1,
			Value: $3,
		}
	};

enum_values:
	// Empty
	{
		$$ = nil
	}
	| enum_values enum_value
	{
		if debugParser {
			fmt.Println("enum values", $1, $2)
		}
		$$ = append($$, $2)
	};


// message

message: MESSAGE IDENT '{' message_fields '}' 
	{ 
		if debugParser {
			fmt.Println("message", $2, $4)
		}
		$$ = &ast.Definition{
			Type: ast.DefinitionMessage,
			Name: $2,

			Message: &ast.Message{
				Fields: $4,
			},
		}
	};

message_field: field_name type INTEGER ';'
	{
		if debugParser {
			fmt.Println("message field", $1, $2, $3)
		}
		$$ = &ast.MessageField{
			Name: $1,
			Type: $2,
			Tag: $3,
		}
	};

message_fields:
	// Empty
	{
		$$ = nil
	}
	| message_fields message_field
	{
		if debugParser {
			fmt.Println("message fields", $1, $2)
		}
		$$ = append($$, $2)
	};

// struct

struct: STRUCT IDENT '{' struct_fields '}' 
	{ 
		if debugParser {
			fmt.Println("struct", $2, $4)
		}
		$$ = &ast.Definition{
			Type: ast.DefinitionStruct,
			Name: $2,

			Struct: &ast.Struct{
				Fields: $4,
			},
		}
	};

struct_field: field_name type ';'
	{
		if debugParser {
			fmt.Println("struct field", $1, $2)
		}
		$$ = &ast.StructField{
			Name: $1,
			Type: $2,
		}
	};

struct_fields:
	// Empty
	{ 
		$$ = nil
	}
	| struct_fields struct_field
	{
		if debugParser {
			fmt.Println("struct fields", $1, $2)
		}
		$$ = append($$, $2)
	};

// service

service: SERVICE IDENT '{' methods '}'
	{
		if debugParser {
			fmt.Println("service", $2, $4)
		}
		$$ = &ast.Definition{
			Type: ast.DefinitionService,
			Name: $2,

			Service: &ast.Service{
				Methods: $4,
			},
		}
	}
	;


// methods

methods:
	// Empty
	{
		$$ = nil
	}
	| methods method
	{
		$$ = append($1, $2)
	};

method: field_name method_arg_list method_result_list ';'
	{
		if debugParser {
			fmt.Println("method", $1, $2, $3)
		}
		$$ = &ast.Method{
			Name: $1,
			Args: $2,
			Results: $3,
		}
	};


// method args	

method_arg_list:
	'(' method_args ')'
	{
		if debugParser {
			fmt.Println("method_args", $2)
		}
		$$ = $2
	};

method_args:
	// Empty
	{
		$$ = nil
	}
	| method_arg
	{
		$$ = []*ast.MethodArg{$1}
	}
	| method_args ',' method_arg
	{
		$$ = append($1, $3)
	};

method_arg:
	field_name type
	{
		if debugParser {
			fmt.Println("method arg", $1, $2)
		}
		$$ = &ast.MethodArg {
			Name: $1,
			Type: $2,
		}
	};


// method results

method_result_list:
	// Empty
	{
		$$ = nil
	}
	| '(' method_results ')'
	{
		$$ = $2
	};

method_results:
	// Empty
	{
		$$ = nil
	}
	| method_result
	{
		$$ = []*ast.MethodResult{$1}
	}
	| method_results ',' method_result
	{
		$$ = append($1, $3)
	};

method_result:
	field_name type
	{
		if debugParser {
			fmt.Println("method result", $1, $2)
		}
		$$ = &ast.MethodResult{
			Name: $1,
			Type: $2,
		}
	};
