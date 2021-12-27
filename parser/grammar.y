// --- Declarations ---
%{
package parser

import "fmt"
%}

// union defines yySymType body.
%union {
	// tokens
	ident   string
	integer int
	string  string

	// import
	import_ *Import
	imports []*Import

	// definition
	definition  *Definition
	definitions []*Definition

	// enum
	enum_value  *EnumValue
	enum_values []*EnumValue

	// message
	message_field  *MessageField
	message_fields []*MessageField

	// struct
	struct_field  *StructField
	struct_fields []*StructField

	// type
	type_ *Type
}

// keywords
%token ENUM
%token IMPORT
%token MESSAGE
%token STRUCT

// general
%token <ident>   IDENT
%token <integer> INTEGER
%token <string>  STRING

// import
%type <import_> import
%type <imports> import_list
%type <imports> imports

// definitions
%type <definition>  definition
%type <definitions> definitions

// enum
%type <definition>  enum
%type <enum_value>  enum_value
%type <enum_values> enum_values

// message
%type <definition>     message
%type <message_field>  message_field
%type <message_fields> message_fields

// struct
%type <definition>    struct
%type <struct_field>  struct_field
%type <struct_fields> struct_fields

// type
%type <type_> type

// start
%start file

%%

file: imports definitions
	{ 
		file := &File{
			Imports:     $1,
			Definitions: $2,
		}
		setLexerResult(yylex, file)
	}

// import

import:
	STRING
		{ 
			if debugParser {
				fmt.Println("import ", $1)
			}
			$$ = &Import{
				ID: trimString($1),
			}
		}
	| IDENT STRING
		{
			if debugParser {
				fmt.Println("import ", $1, $2)
			}
			$$ = &Import{
				Alias: $1,
				ID:    trimString($2),
			}
		}
import_list:
	// empty
		{ 
			$$ = nil
		}
	| import_list import
		{
			if debugParser {
				fmt.Println("import_list", $1, $2)
			}
			$$ = append($$, $2)
		}
imports:
	// empty
		{ 
			$$ = nil
		}
	| IMPORT '(' import_list ')'
		{
			if debugParser {
				fmt.Println("imports", $3)
			}
			$$ = append($$, $3...)
		}


// definition

definition: 
	enum 
	| message
	| struct
definitions:
	// empty
		{ 
			$$ = nil
		}
	| definitions definition
		{
			if debugParser {
				fmt.Println("definitions", $1, $2)
			}
			$$ = append($$, $2)
		}


// enum

enum: ENUM IDENT '{' enum_values '}'
	{
		if debugParser {
			fmt.Println("enum", $2, $4)
		}
		$$ = &Definition{
			Type: DefinitionEnum,
			Name: $2,

			Enum: &Enum{
				Values: $4,
			},
		}
	}
enum_value: IDENT '=' INTEGER ';'
	{
		if debugParser {
			fmt.Println("enum value", $1, $3)
		}
		$$ = &EnumValue{
			Name: $1,
			Value: $3,
		}
	}
enum_values:
	// empty
		{
			$$ = nil
		}
	| enum_values enum_value
		{
			if debugParser {
				fmt.Println("enum values", $1, $2)
			}
			$$ = append($$, $2)
		}


// message

message: MESSAGE IDENT '{' message_fields '}' 
	{ 
		if debugParser {
			fmt.Println("message", $2, $4)
		}
		$$ = &Definition{
			Type: DefinitionMessage,
			Name: $2,

			Message: &Message{
				Fields: $4,
			},
		}
	}
message_field: IDENT type INTEGER ';'
	{
		if debugParser {
			fmt.Println("message field", $1, $2, $3)
		}
		$$ = &MessageField{
			Name: $1,
			Type: $2,
			Tag: $3,
		}
	}
message_fields:
	// empty
		{ 
			$$ = nil
		}
	| message_fields message_field
		{
			if debugParser {
				fmt.Println("message fields", $1, $2)
			}
			$$ = append($$, $2)
		}

// struct

struct: STRUCT IDENT '{' struct_fields '}' 
	{ 
		if debugParser {
			fmt.Println("struct", $2, $4)
		}
		$$ = &Definition{
			Type: DefinitionStruct,
			Name: $2,

			Struct: &Struct{
				Fields: $4,
			},
		}
	}
struct_field: IDENT type ';'
	{
		if debugParser {
			fmt.Println("struct field", $1, $2)
		}
		$$ = &StructField{
			Name: $1,
			Type: $2,
		}
	}
struct_fields:
	// empty
		{ 
			$$ = nil
		}
	| struct_fields struct_field
		{
			if debugParser {
				fmt.Println("struct fields", $1, $2)
			}
			$$ = append($$, $2)
		}

// type

type:
	IDENT
		{
			if debugParser {
				fmt.Println("type", $1)
			}
			$$ = &Type{
				Kind: KindBase,
				Ident: $1,
			}
		}
	| '*' IDENT
		{
			if debugParser {
				fmt.Printf("type *%v\n", $2)
			}
			$$ = &Type{
				Kind: KindNullable,
				Ident: $2,
			}
		}
	| '[' ']' IDENT
		{
			if debugParser {
				fmt.Printf("type []%v\n", $3)
			}
			$$ = &Type{
				Kind: KindList,
				Ident: $3,
			}
		}
