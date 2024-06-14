// --- Declarations ---
%{
package parser

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)
%}

// union defines yySymType body.
%union {
	// Tokens
	ident   string
	bool	bool
	integer int
	string  string

    // Type
	type_ *syntax.Type

	// Import
	import_ *syntax.Import
	imports []*syntax.Import

	// Option
	option  *syntax.Option
	options []*syntax.Option

	// Definition
	definition  *syntax.Definition
	definitions []*syntax.Definition

	// Enum
	enum_value  *syntax.EnumValue
	enum_values []*syntax.EnumValue

	// Field
	field  *syntax.Field
	fields syntax.Fields

	// Struct
	struct_field  *syntax.StructField
	struct_fields []*syntax.StructField

    // Service
    service         *syntax.Service
    method          *syntax.Method
    methods         []*syntax.Method
	method_input	syntax.MethodInput
	method_output	syntax.MethodOutput
	method_channel	*syntax.MethodChannel
	method_field	*syntax.Field
	method_fields	syntax.Fields
}

// keywords
%token ANY
%token ENUM
%token IMPORT
%token MESSAGE
%token ONEWAY
%token OPTIONS
%token STRUCT
%token SERVICE
%token SUBSERVICE

// general
%token <ident>      IDENT
%token <integer>    INTEGER
%token <string>     STRING
%token <ident>      MESSAGE
%token <ident>      ONEWAY
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
%type <definition>	message

// field
%type <field>	field
%type <fields> 	fields

// struct
%type <definition>      struct
%type <struct_field>    struct_field
%type <struct_fields>   struct_fields

// service
%type <definition>      service
%type <definition>      subservice
%type <methods>         methods
%type <method>          method
%type <method_input>    method_input
%type <bool>			method_oneway
%type <method_output>   method_output
%type <method_channel>  method_channel
%type <type_>  			method_channel_in
%type <type_>  			method_channel_out
%type <field>			method_field
%type <fields>			method_fields
%type <fields>			method_field_list

%left METHOD_OUTPUT

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
	}
	| SUBSERVICE
	{
		$$ = "subservice"
	};

// file

file: imports options definitions
	{ 
		file := &syntax.File{
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
		$$ = &syntax.Import{
			ID: trimString($1),
		}
	}
	| IDENT STRING
	{
		if debugParser {
			fmt.Println("import ", $1, $2)
		}
		$$ = &syntax.Import{
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
		$$ = &syntax.Option{
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
		$$ = &syntax.Type{
			Kind:    syntax.KindList,
			Element: $3,
		}
	};

base_type:
	IDENT
	{
		if debugParser {
			fmt.Println("base type", $1)
		}
		$$ = &syntax.Type{
			Kind: syntax.GetKind($1),
			Name: $1,
		}
	}
	| IDENT '.' IDENT
	{
		if debugParser {
			fmt.Printf("base type %v.%v\n", $1, $3)
		}
		$$ = &syntax.Type{
			Kind:   syntax.KindReference,
			Name:   $3,
			Import: $1,
		}
	}
	| ANY
	{
		if debugParser {
			fmt.Println("base type", "any")
		}
		$$ = &syntax.Type{
			Kind: syntax.KindAny,
			Name: "any",
		}
	}
	| MESSAGE
	{
		if debugParser {
			fmt.Println("base type", "message")
		}
		$$ = &syntax.Type{
			Kind: syntax.KindAnyMessage,
			Name: "message",
		}
	};

// definition

definition: 
	enum 
	| message
	| struct
    | service
	| subservice
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
		$$ = &syntax.Definition{
			Type: syntax.DefinitionEnum,
			Name: $2,

			Enum: &syntax.Enum{
				Values: $4,
			},
		}
	};

enum_value: field_name '=' INTEGER ';'
	{
		if debugParser {
			fmt.Println("enum value", $1, $3)
		}
		$$ = &syntax.EnumValue{
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

message: MESSAGE IDENT '{' fields semi_opt '}' 
	{ 
		if debugParser {
			fmt.Println("message", $2, $4)
		}
		$$ = &syntax.Definition{
			Type: syntax.DefinitionMessage,
			Name: $2,

			Message: &syntax.Message{
				Fields: $4,
			},
		}
	};

field: field_name type INTEGER
	{
		if debugParser {
			fmt.Println("message field", $1, $2, $3)
		}
		$$ = &syntax.Field{
			Name: $1,
			Type: $2,
			Tag: $3,
		}
	};

fields:
	// Empty
	{
		$$ = nil
	}
	| field
	{
		if debugParser {
			fmt.Println("message fields", $1)
		}
		$$ = []*syntax.Field{$1}
	}
	| fields ';' field
	{
		if debugParser {
			fmt.Println("message fields", $1, $3)
		}
		$$ = append($$, $3)
	};


// struct

struct: STRUCT IDENT '{' struct_fields '}' 
	{ 
		if debugParser {
			fmt.Println("struct", $2, $4)
		}
		$$ = &syntax.Definition{
			Type: syntax.DefinitionStruct,
			Name: $2,

			Struct: &syntax.Struct{
				Fields: $4,
			},
		}
	};

struct_field: field_name type ';'
	{
		if debugParser {
			fmt.Println("struct field", $1, $2)
		}
		$$ = &syntax.StructField{
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
		$$ = &syntax.Definition{
			Type: syntax.DefinitionService,
			Name: $2,

			Service: &syntax.Service{
				Methods: $4,
			},
		}
	}
	;

subservice: SUBSERVICE IDENT '{' methods '}'
	{
		if debugParser {
			fmt.Println("subservice", $2, $4)
		}
		$$ = &syntax.Definition{
			Type: syntax.DefinitionService,
			Name: $2,

			Service: &syntax.Service{
				Sub: true,
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

method:
	field_name method_input ';'
	{
		if debugParser {
			fmt.Println("method", $1, $2)
		}
		$$ = &syntax.Method{
			Name: $1,
			Input: $2,
		}
	}
	| field_name method_input method_oneway ';'
	{
		if debugParser {
			fmt.Println("method", $1, $2, $3)
		}
		$$ = &syntax.Method{
			Name: $1,
			Input: $2,
			Oneway: true,
		}
	}
	| field_name method_input method_output ';'
	{
		if debugParser {
			fmt.Println("method", $1, $2, $3)
		}
		$$ = &syntax.Method{
			Name: $1,
			Input: $2,
			Output: $3,
		}
	}
	| field_name method_input method_channel ';'
	{
		if debugParser {
			fmt.Println("method", $1, $2, $3)
		}
		$$ = &syntax.Method{
			Name: $1,
			Input: $2,
			Channel: $3,
		}
	}
	| field_name method_input method_channel method_output ';'
	{
		if debugParser {
			fmt.Println("method", $1, $2, $3, $4)
		}
		$$ = &syntax.Method{
			Name: $1,
			Input: $2,
			Channel: $3,
			Output: $4,
		}
	};

method_input:
	'(' base_type ')'
	{
		if debugParser {
			fmt.Println("method input", $2)
		}
		$$ = $2
	}
	| '(' method_field_list ')'
	{
		if debugParser {
			fmt.Println("method input", $2)
		}
		$$ = $2
	};

method_oneway: ONEWAY
	{
		if debugParser {
			fmt.Println("method oneway", $1)
		}
		$$ = true
	};

method_output:
	base_type
	{
		if debugParser {
			fmt.Println("method output", $1)
		}
		$$ = $1
	}
	| '(' method_field_list ')'
	{
		if debugParser {
			fmt.Println("method output", $2)
		}
		$$ = $2
	};

method_channel:
	'(' method_channel_in ')'
	{
		if debugParser {
			fmt.Println("method channel", $2)
		}

		$$ = &syntax.MethodChannel{
			In: $2,
		}
	}
	| '(' method_channel_out ')'
	{
		if debugParser {
			fmt.Println("method channel", $2)
		}

		$$ = &syntax.MethodChannel{
			Out: $2,
		}
	}
	| '(' method_channel_in method_channel_out ')'
	{
		if debugParser {
			fmt.Println("method channel", $2, $3)
		}

		$$ = &syntax.MethodChannel{
			In: $2,
			Out: $3,
		}
	};

method_channel_in:
	'<' '-' type
	{
		$$ = $3
	};

method_channel_out:
	'-' '>' type
	{
		$$ = $3
	};

method_field_list:
	method_fields comma_opt
	{
		if debugParser {
			fmt.Println("method field list", $1)
		}
		$$ = $1
	};

method_fields:
	// Empty
	{
		$$ = nil
	}
	| method_field
	{
		if debugParser {
			fmt.Println("method fields", $1)
		}
		$$ = []*syntax.Field{$1}
	}
	| method_fields ',' method_field
	{
		if debugParser {
			fmt.Println("method fields", $1, $3)
		}
		$$ = append($1, $3)
	};

method_field:
	field_name type INTEGER
	{
		if debugParser {
			fmt.Println("method field", $1, $2, $3)
		}
		$$ = &syntax.Field{
			Name: $1,
			Type: $2,
			Tag: $3,
		}
	};


// util

comma_opt:
	// Empty
	{}
	| ','
	{}
	;

semi_opt:
	// Empty
	{}
	| ';'
	{}
	;
