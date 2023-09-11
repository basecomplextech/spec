package ast

type Service struct {
	Sub     bool // Subservice
	Methods []*Method
}

type Method struct {
	Name    string
	Args    []*MethodField
	Results []*MethodResult
}

type MethodResult struct {
	Name string
	Type *Type
}

type MethodField struct {
	Name string
	Type *Type
	Tag  int
}
