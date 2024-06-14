package model

import (
	"fmt"

	"github.com/basecomplextech/spec/internal/lang/syntax"
)

type DefinitionType string

const (
	DefinitionUndefined DefinitionType = ""
	DefinitionEnum      DefinitionType = "enum"
	DefinitionMessage   DefinitionType = "message"
	DefinitionStruct    DefinitionType = "struct"
	DefinitionService   DefinitionType = "service"
)

func parseDefinitionType(ptype syntax.DefinitionType) (DefinitionType, error) {
	switch ptype {
	case syntax.DefinitionEnum:
		return DefinitionEnum, nil
	case syntax.DefinitionMessage:
		return DefinitionMessage, nil
	case syntax.DefinitionStruct:
		return DefinitionStruct, nil
	case syntax.DefinitionService:
		return DefinitionService, nil
	}
	return "", fmt.Errorf("unsupported syntax definition type %v", ptype)
}
