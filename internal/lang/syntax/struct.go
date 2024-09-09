// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package syntax

type Struct struct {
	Fields []*StructField
}

type StructField struct {
	Name string
	Type *Type
}
