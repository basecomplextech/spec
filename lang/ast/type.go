package ast

type Type struct {
	Kind    Kind
	Name    string
	Import  string // package name in imported type, "pkg" in "pkg.Name"
	Element *Type  // element type in list and nullable types
}
