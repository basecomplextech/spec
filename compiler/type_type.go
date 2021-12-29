package compiler

import (
	"fmt"

	"github.com/baseone-run/spec/parser"
)

type Kind string

const (
	// builtin

	KindUndefined Kind = ""
	KindBool      Kind = "bool"

	KindInt8  Kind = "int8"
	KindInt16 Kind = "int16"
	KindInt32 Kind = "int32"
	KindInt64 Kind = "int64"

	KindUint8  Kind = "uint8"
	KindUint16 Kind = "uint16"
	KindUint32 Kind = "uint32"
	KindUint64 Kind = "uint64"

	KindFloat32 Kind = "float32"
	KindFloat64 Kind = "float64"

	KindBytes  Kind = "bytes"
	KindString Kind = "string"

	// references

	KindReference Kind = "reference"
	KindImport    Kind = "import"
	KindList      Kind = "list"
	KindNullable  Kind = "nullable"
)

var (
	builtin = map[string]*Type{
		string(KindBool): newBuiltinType(KindBool),

		string(KindInt8):  newBuiltinType(KindInt8),
		string(KindInt16): newBuiltinType(KindInt16),
		string(KindInt32): newBuiltinType(KindInt32),
		string(KindInt64): newBuiltinType(KindInt64),

		string(KindUint8):  newBuiltinType(KindUint8),
		string(KindUint16): newBuiltinType(KindUint16),
		string(KindUint32): newBuiltinType(KindUint32),
		string(KindUint64): newBuiltinType(KindUint64),

		string(KindFloat32): newBuiltinType(KindFloat32),
		string(KindFloat64): newBuiltinType(KindFloat64),

		string(KindBytes):  newBuiltinType(KindBytes),
		string(KindString): newBuiltinType(KindString),
	}

	numbers = map[Kind]bool{
		KindInt8:  true,
		KindInt16: true,
		KindInt32: true,
		KindInt64: true,

		KindUint8:  true,
		KindUint16: true,
		KindUint32: true,
		KindUint64: true,

		KindFloat32: true,
		KindFloat64: true,
	}
)

type Type struct {
	Kind       Kind
	Name       string
	Element    *Type  // element type in list, reference and nullable types
	ImportName string // imported package name, "pkg" in "pkg.Type"

	// Resolved
	Ref    *Definition
	Import *Import

	// Flags
	Builtin    bool
	Imported   bool
	Referenced bool
	Resolved   bool
}

func newType(ptype *parser.Type) (*Type, error) {
	switch ptype.Kind {
	case parser.KindBase:
		type_, ok := builtin[ptype.Name]
		if ok {
			return type_, nil
		}

		type_ = &Type{
			Kind: KindReference,
			Name: ptype.Name,

			Referenced: true,
		}
		return type_, nil

	case parser.KindImport:
		type_ := &Type{
			Kind:       KindImport,
			Name:       ptype.Name,
			ImportName: ptype.Import,

			Imported: true,
		}
		return type_, nil

	case parser.KindNullable:
		elem, err := newType(ptype.Element)
		if err != nil {
			return nil, err
		}

		type_ := &Type{
			Kind:    KindNullable,
			Name:    "*",
			Element: elem,
		}
		return type_, nil

	case parser.KindList:
		elem, err := newType(ptype.Element)
		if err != nil {
			return nil, err
		}

		type_ := &Type{
			Kind:    KindList,
			Name:    "[]",
			Element: elem,
		}
		return type_, nil
	}

	return nil, fmt.Errorf("unsupported type kind, kind=%v, name=%v", ptype.Kind, ptype.Name)
}

func newBuiltinType(kind Kind) *Type {
	return &Type{
		Kind:    kind,
		Name:    string(kind),
		Builtin: true,
	}
}

func (t *Type) resolveRef(def *Definition) {
	switch {
	case t.Kind != KindReference:
		panic("type not reference")
	case t.Resolved:
		panic("type already resolved")
	}

	t.Ref = def
	t.Resolved = true
}

func (t *Type) resolveImport(imp *Import, def *Definition) {
	switch {
	case t.Kind != KindImport:
		panic("type not import")
	case t.Resolved:
		panic("type already resolved")
	}

	t.Ref = def
	t.Import = imp
	t.Resolved = true
}

// Bool     bool
// 	Number   bool
// 	Bytes    bool
// 	String   bool
// 	List     bool
// 	Nullable bool

// 	Int8 bool
// 	Int16 bool
// 	Int32 bool

// 	Enum    bool
// 	Message bool
// 	Struct  bool

func (t *Type) Bool() bool   { return t.Kind == KindBool }
func (t *Type) Number() bool { return numbers[t.Kind] }

func (t *Type) Int8() bool  { return t.Kind == KindInt8 }
func (t *Type) Int16() bool { return t.Kind == KindInt16 }
func (t *Type) Int32() bool { return t.Kind == KindInt32 }
func (t *Type) Int64() bool { return t.Kind == KindInt64 }

func (t *Type) Uint8() bool  { return t.Kind == KindUint8 }
func (t *Type) Uint16() bool { return t.Kind == KindUint16 }
func (t *Type) Uint32() bool { return t.Kind == KindUint32 }
func (t *Type) Uint64() bool { return t.Kind == KindUint64 }

func (t *Type) Float32() bool { return t.Kind == KindFloat32 }
func (t *Type) Float64() bool { return t.Kind == KindFloat64 }

func (t *Type) Bytes() bool    { return t.Kind == KindBytes }
func (t *Type) String() bool   { return t.Kind == KindString }
func (t *Type) List() bool     { return t.Kind == KindList }
func (t *Type) Nullable() bool { return t.Kind == KindNullable }

func (t *Type) Enum() bool    { return t.Ref != nil && t.Ref.Type == DefinitionEnum }
func (t *Type) Message() bool { return t.Ref != nil && t.Ref.Type == DefinitionMessage }
func (t *Type) Struct() bool  { return t.Ref != nil && t.Ref.Type == DefinitionStruct }
