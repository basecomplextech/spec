// --- Declarations ---
%{
package parser
%}

// union defines yySymType body.
%union {
	// module
	module      *Module
	module_name string

	// definition
	definition  Definition
	definitions []Definition

	// message
	message_field  *MessageField
	message_fields []*MessageField

	// enum
	enum_value  *EnumValue
	enum_values []*EnumValue

	// general
	ident  string
	number int
}

// tokens
%token NEWLINE

// keywords
%token ENUM
%token MODULE
%token MESSAGE


// general
%token <ident>  IDENT
%token <number> NUMBER

// module
%type <module>      module
%type <module_name> module_name

// definitions
%type <definition>  definition
%type <definitions> definitions

// message
%type <definition>     message
%type <message_field>  message_field
%type <message_fields> message_fields

// enum
%type <definition>  enum
%type <enum_value>  enum_value
%type <enum_values> enum_values

// start
%start file

%%

file: module definitions

// module

module: MODULE module_name ';'
	{ 
		$$ = &Module{Name: $2}
	}
module_name: IDENT
	{ $$ = $1 }

// definition

definition: 
	message | enum

definitions:
	// empty
		{ 
			$$ = nil
		}
	| definition
		{
			$$ = []Definition{$1}
		}
	| definitions NEWLINE definition
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
