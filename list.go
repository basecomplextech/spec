// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/spec/internal/types"
)

// List is a raw list of elements.
type List = types.List

// NewList returns a new list from bytes or an empty list when not a list.
func NewList(b []byte) List {
	return types.NewList(b)
}

// NewListErr returns a new list from bytes or an error when not a list.
func NewListErr(b []byte) (List, error) {
	return types.NewListErr(b)
}

// ParseList recursively parses and returns a list.
func ParseList(b []byte) (l List, size int, err error) {
	return types.ParseList(b)
}
