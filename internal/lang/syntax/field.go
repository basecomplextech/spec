// Copyright 2023 Ivan Korobkov. All rights reserved.

package syntax

type Field struct {
	Name string
	Type *Type
	Tag  int
}

type Fields []*Field
