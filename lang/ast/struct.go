package ast

type Struct struct {
	Fields []*StructField
}

type StructField struct {
	Name string
	Type *Type
}
