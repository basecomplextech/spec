package ast

type Service struct {
	Sub     bool // Subservice
	Methods []*Method
}

type Method struct {
	Name    string
	Args    []*MethodField
	Results []*MethodField
}

type MethodField struct {
	Name string
	Type *Type
	Tag  int
}
