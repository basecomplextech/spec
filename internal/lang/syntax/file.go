package syntax

type File struct {
	Path        string
	Imports     []*Import
	Options     []*Option
	Definitions []*Definition
}

// Import

type Import struct {
	ID    string
	Alias string
}

// Option

type Option struct {
	Name  string
	Value string
}
