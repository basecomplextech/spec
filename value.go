// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/spec/internal/types"
)

// Value is a raw value.
type Value = types.Value

// Value

// NewValue returns a new value from bytes or nil when not valid.
func NewValue(b []byte) Value {
	return types.NewValue(b)
}

// NewValueErr returns a new value from bytes or an error when not valid.
func NewValueErr(b []byte) (Value, error) {
	return types.NewValueErr(b)
}

// ParseValue recursively parses and returns a value.
func ParseValue(b []byte) (_ Value, n int, err error) {
	return types.ParseValue(b)
}
