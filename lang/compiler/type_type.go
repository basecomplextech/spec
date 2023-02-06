package compiler

import (
	"fmt"

	"github.com/complex1tech/spec/lang/ast"
)

var builtin = map[Kind]*Type{
	KindBool: newBuiltinType(KindBool),
	KindByte: newBuiltinType(KindByte),

	KindInt16: newBuiltinType(KindInt16),
	KindInt32: newBuiltinType(KindInt32),
	KindInt64: newBuiltinType(KindInt64),

	KindUint16: newBuiltinType(KindUint16),
	KindUint32: newBuiltinType(KindUint32),
	KindUint64: newBuiltinType(KindUint64),

	KindBin64:  newBuiltinType(KindBin64),
	KindBin128: newBuiltinType(KindBin128),
	KindBin256: newBuiltinType(KindBin256),

	KindFloat32: newBuiltinType(KindFloat32),
	KindFloat64: newBuiltinType(KindFloat64),

	KindBytes:      newBuiltinType(KindBytes),
	KindString:     newBuiltinType(KindString),
	KindAnyMessage: newBuiltinType(KindAnyMessage),
}

type Type struct {
	Kind       Kind
	Name       string
	Element    *Type  // element type in list, reference and nullable types
	ImportName string // imported package name, "pkg" in "pkg.Type"

	// Resolved
	Ref    *Definition
	Import *Import
}

func newType(ptype *ast.Type) (*Type, error) {
	kind, err := parseKind(ptype.Kind)
	if err != nil {
		return nil, err
	}

	// builtin type
	t, ok := builtin[kind]
	if ok {
		return t, nil
	}

	switch kind {
	case KindList:
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

	case KindReference:
		type_ := &Type{
			Kind:       KindReference,
			Name:       ptype.Name,
			ImportName: ptype.Import,
		}
		return type_, nil
	}

	return nil, fmt.Errorf("unsupported type kind, kind=%v, name=%v", ptype.Kind, ptype.Name)
}

func newBuiltinType(kind Kind) *Type {
	return &Type{
		Kind: kind,
		Name: kind.String(),
	}
}

func (t *Type) builtin() bool {
	_, ok := builtin[t.Kind]
	return ok
}

func (t *Type) resolve(def *Definition, impOrNil *Import) {
	if t.Kind != KindReference {
		panic("type already resolved")
	}

	t.Ref = def
	t.Import = impOrNil

	switch def.Type {
	case DefinitionEnum:
		t.Kind = KindEnum
	case DefinitionMessage:
		t.Kind = KindMessage
	case DefinitionStruct:
		t.Kind = KindStruct
	}
}
