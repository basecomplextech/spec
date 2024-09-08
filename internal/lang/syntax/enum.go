// Copyright 2023 Ivan Korobkov. All rights reserved.

package syntax

type Enum struct {
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}
