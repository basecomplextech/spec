// --- Declarations ---
%{
package parser
%}

// union defines yySymType body.
%union {
	// file
	file *File

	// module
	module string

	// import
	import_        *Import
	import_module  *ImportModule
	import_modules []*ImportModule

	// definition
	definition  Definition
	definitions []Definition

	// enum
	enum_value  *EnumValue
	enum_values []*EnumValue

	// message
	message_field  *MessageField
	message_fields []*MessageField

	// struct
	struct_field  *StructField
	struct_fields []*StructField

	// general
	ident  string
	number int
	string string
}

// tokens
%token NEWLINE

// keywords
%token ENUM
%token IMPORT
%token MESSAGE
%token MODULE
%token STRUCT

// general
%token <ident>  IDENT
%token <number> NUMBER
%token <string> STRING

// module
%type <module> module

// import
%type <import_>         import
%type <import_module>  import_module
%type <import_modules> import_modules

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

%type <file> file

// start
%start file

%%

file: module import definitions
	{ 
		$$ = &File{
			Module: $1,
			Import: $2,
			Definitions: $3,
		}
	}

// module

module: MODULE IDENT ';'
	{ 
		$$ = $2
	}

// import

import: 
	// empty
		{
			$$ = &Import{}
		}
	| IMPORT '(' import_modules ')'
		{
			$$ = &Import{Modules: $3}
		}
import_module:
	STRING
		{ 
			$$ = &ImportModule{Name: $1}
		}
	| IDENT STRING
		{
			$$ = &ImportModule{
				Alias: $1,
				Name: $2,
			}
		}
import_modules:
	// empty
		{ 
			$$ = nil
		}
	| import_module
		{
			$$ = []*ImportModule{$1}
		}
	| import_modules NEWLINE import_module
		{
			$$ = append($$, $1...)
			$$ = append($$, $3)
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
			$$ = append($$, $1...)
			$$ = append($$, $2)
		}


// enum

enum: ENUM IDENT '{' enum_values '}'
	{
		$$ = &Enum{
			Name: $2,
			Values: $4,
		}
	}
enum_value: IDENT '=' NUMBER
	{
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
	| enum_value
		{
			$$ = []*EnumValue{$1}
		}
	| enum_values ';' enum_value
		{
			$$ = append($$, $1...)
			$$ = append($$, $3)
		}


// message

message: MESSAGE IDENT '{' message_fields '}' 
	{ 
		$$ = &Message{
			Name: $2,
			Fields: $4,
		}
	}
message_field: IDENT IDENT NUMBER 
	{ 
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
	| message_field
		{ 
			$$ = []*MessageField{$1}
		}
	| message_fields ';' message_field
		{
			$$ = append($$, $1...)
			$$ = append($$, $3)
		}

// struct

struct: STRUCT IDENT '{' struct_fields '}' 
	{ 
		$$ = &Struct{
			Name: $2,
			Fields: $4,
		}
	}
struct_field: IDENT IDENT
	{ 
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
	| struct_field
		{ 
			$$ = []*StructField{$1}
		}
	| struct_fields ';' struct_field
		{
			$$ = append($$, $1...)
			$$ = append($$, $3)
		}
