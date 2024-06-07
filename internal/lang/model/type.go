package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

var builtin = map[Kind]*Type{
	KindAny: newBuiltinType(KindAny),

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

var primitive = map[Kind]struct{}{
	KindBool: {},
	KindByte: {},

	KindInt16: {},
	KindInt32: {},
	KindInt64: {},

	KindUint16: {},
	KindUint32: {},
	KindUint64: {},

	KindBin64:  {},
	KindBin128: {},
	KindBin256: {},

	KindFloat32: {},
	KindFloat64: {},
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

func newType(ptype *syntax.Type) (*Type, error) {
	kind, err := parseKind(ptype.Kind)
	if err != nil {
		return nil, err
	}

	// Builtin type
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

func newTypeRef(def *Definition) *Type {
	t := &Type{
		Kind: KindReference,
		Name: def.Name,
		Ref:  def,
	}
	t._resolve(def, nil)
	return t
}

func (t *Type) builtin() bool {
	_, ok := builtin[t.Kind]
	return ok
}

// primitive returns true if the type is a primitive type.
// primitive types are not allocated on the heap.
func (t *Type) primitive() bool {
	_, ok := primitive[t.Kind]
	return ok
}

func (t *Type) resolve(file *File) error {
	switch t.Kind {
	case KindList:
		return t.Element.resolve(file)

	case KindReference:
		if t.ImportName == "" {
			// Local type

			pkg := file.Package
			def, ok := pkg.LookupType(t.Name)
			if !ok {
				return fmt.Errorf("type not found: %v", t.Name)
			}
			t._resolve(def, nil)

		} else {
			// Imported type

			imp, ok := file.LookupImport(t.ImportName)
			if !ok {
				return fmt.Errorf("type not found: %v.%v", t.ImportName, t.Name)
			}
			def, ok := imp.LookupType(t.Name)
			if !ok {
				return fmt.Errorf("type not found: %v.%v", t.ImportName, t.Name)
			}
			t._resolve(def, imp)
		}
	}
	return nil
}

func (t *Type) _resolve(def *Definition, impOrNil *Import) {
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
	case DefinitionService:
		t.Kind = KindService
	}
}
