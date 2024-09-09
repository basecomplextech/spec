// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package syntax

type Field struct {
	Name string
	Type *Type
	Tag  int
}

type Fields []*Field
