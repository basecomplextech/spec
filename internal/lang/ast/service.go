package ast

type Service struct {
	Sub     bool // Subservice
	Methods []*Method
}

type Method struct {
	Name   string
	Input  MethodInput
	Output MethodOutput
}

type MethodField struct {
	Name string
	Type *Type
	Tag  int
}

type MethodFields []*MethodField

// MethodInput is a union type for method inputs.
//
//	MethodInput:
//	| Reference
//	| MethodFields
//	| nil
type MethodInput interface{}

// MethodOutput is a union type for method outputs.
//
//	MethodOutput:
//	| Reference
//	| MethodFields
//	| nil
type MethodOutput interface{}
