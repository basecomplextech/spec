package ast

type Service struct {
	Sub     bool // Subservice
	Methods []*Method
}

type Method struct {
	Name    string
	Args    []*MethodArg
	Results []*MethodResult
}

type MethodArg struct {
	Name string
	Type *Type
}

type MethodResult struct {
	Name string
	Type *Type
}
