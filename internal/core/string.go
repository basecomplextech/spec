package core

import "strings"

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
