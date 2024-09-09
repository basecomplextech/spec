// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package syntax

type Enum struct {
	Values []*EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}
