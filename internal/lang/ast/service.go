package ast

type Service struct {
	Sub     bool // Subservice
	Methods []*Method
}

type Method struct {
	Name   string
	Input  MethodInput
	Result MethodResult
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

// MethodResult is a union type for method results.
//
//	MethodResult:
//	| Reference
//	| MethodFields
//	| nil
type MethodResult interface{}
