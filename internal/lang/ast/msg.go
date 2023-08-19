package ast

type Message struct {
	Fields []*MessageField
}

type MessageField struct {
	Name string
	Type *Type
	Tag  int
}
