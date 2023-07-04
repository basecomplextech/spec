package types

import (
	"bytes"
	"strings"
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

// String is a spec string backed by a buffer.
// Clone it if you need to keep it around.
type String string

// Clone returns a string clone allocated on the heap.
func (s String) Clone() string {
	return strings.Clone(string(s))
}

// Unwrap returns a string.
func (s String) Unwrap() string {
	return string(s)
}
