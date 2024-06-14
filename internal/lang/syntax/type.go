package syntax

type Type struct {
	Kind    Kind
	Name    string
	Import  string // package name in imported type, "pkg" in "pkg.Name"
	Element *Type  // element type in list and nullable types
}

func (t *Type) String() string {
	switch t.Kind {
	case KindReference:
		if t.Import != "" {
			return t.Import + "." + t.Name
		}
		return t.Name
	case KindList:
		return "[]" + t.Element.String()
	}
	return t.Kind.String()
}
