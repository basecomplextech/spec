// Copyright 2023 Ivan Korobkov. All rights reserved.

package syntax

type Struct struct {
	Fields []*StructField
}

type StructField struct {
	Name string
	Type *Type
}
