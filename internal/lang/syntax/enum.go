package syntax

type Enum struct {
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}
