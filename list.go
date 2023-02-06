package spec

import (
	"fmt"

	"github.com/complex1tech/spec/encoding"
)

// List is a raw list of elements.
type List struct {
	meta  encoding.ListMeta
	bytes []byte
}

// NewList returns a new list from bytes or an empty list when not a list.
func NewList(b []byte) List {
	meta, n, err := encoding.DecodeListMeta(b)
	if err != nil {
		return List{}
	}
	bytes := b[len(b)-n:]

	return List{
		meta:  meta,
		bytes: bytes,
	}
}

// ParseList recursively parses and returns a list.
func ParseList(b []byte) (_ List, size int, err error) {
	meta, n, err := encoding.DecodeListMeta(b)
	if err != nil {
		return List{}, 0, err
	}
	bytes := b[len(b)-n:]

	l := List{
		meta:  meta,
		bytes: bytes,
	}

	ln := l.Len()
	for i := 0; i < ln; i++ {
		elem := l.GetBytes(i)
		if len(elem) == 0 {
			continue
		}
		if _, _, err = ParseValue(elem); err != nil {
			return
		}
	}
	return l, n, nil
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return l.meta.Len()
}

// Bytes returns the exact list bytes.
func (l List) Bytes() []byte {
	return l.bytes
}

// Empty returns true if bytes are empty or list has no elements.
func (l List) Empty() bool {
	return len(l.bytes) == 0 || l.meta.Len() == 0
}

// Elements

// Get returns an element at index i, panics on out of range.
func (l List) Get(i int) Value {
	start, end := l.meta.Offset(i)
	if start < 0 {
		panic(fmt.Sprintf("index out of range: %d", i))
	}

	size := l.meta.DataSize()
	if end > int(size) {
		return Value{}
	}

	b := l.bytes[start:end]
	return NewValue(b)
}

// GetBytes returns element bytes at index i, panics on out of range.
func (l List) GetBytes(i int) []byte {
	start, end := l.meta.Offset(i)
	if start < 0 {
		panic(fmt.Sprintf("index out of range: %d", i))
	}

	size := l.meta.DataSize()
	if end > int(size) {
		return nil
	}

	return l.bytes[start:end]
}

// Clone

// List returns a list clone.
func (l List) Clone() List {
	b := make([]byte, len(l.bytes))
	copy(b, l.bytes)
	return NewList(b)
}

// CloneTo clones a list into a byte slice.
func (l List) CloneTo(b []byte) List {
	ln := len(l.bytes)
	if cap(b) < ln {
		b = make([]byte, ln)
	}
	b = b[:ln]

	copy(b, l.bytes)
	return NewList(b)
}
