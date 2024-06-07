package syntax

type Field struct {
	Name string
	Type *Type
	Tag  int
}

type Fields []*Field
