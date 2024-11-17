// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package format

import (
	"bytes"
)

// Bytes is a spec byte slice backed by a buffer.
// Clone it if you need to keep it around.
type Bytes []byte

// Clone returns a bytes clone allocated on the heap.
func (b Bytes) Clone() []byte {
	return bytes.Clone(b)
}

// Unwrap returns a byte slice.
func (b Bytes) Unwrap() []byte {
	return []byte(b)
}
