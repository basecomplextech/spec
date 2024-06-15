package golang

import "github.com/basecomplextech/spec/internal/lang/model"

type Definition interface {
	Type
}

func newDefinition(pkg *Package, mdef *model.Definition) (Definition, error) {
	switch mdef.Type {
	case model.DefinitionEnum:
		return newEnum(pkg, mdef)
	case model.DefinitionMessage:
		return newMessage(pkg, mdef)
	case model.DefinitionStruct:
		return newStruct(pkg, mdef)
	case model.DefinitionService:
		return newService(pkg, mdef)
	}

	panic("unknown definition type")
}
