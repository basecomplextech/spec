package core

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
