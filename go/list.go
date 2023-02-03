package spec

import (
	"fmt"

	"github.com/complex1tech/spec/go/encoding"
)

type List[T any] struct {
	meta   encoding.ListMeta
	bytes  []byte
	decode func(b []byte) (T, int, error)
}

// NewList decodes and returns a list without recursive validation, or an empty list on error.
func NewList[T any](b []byte, decode func([]byte) (T, int, error)) List[T] {
	meta, n, err := encoding.DecodeListMeta(b)
	if err != nil {
		return List[T]{}
	}
	bytes := b[len(b)-n:]

	l := List[T]{
		meta:   meta,
		bytes:  bytes,
		decode: decode,
	}
	return l
}

// DecodeList decodes, recursively validates and returns a list.
func DecodeList[T any](b []byte, decode func([]byte) (T, int, error)) (_ List[T], size int, err error) {
	meta, size, err := encoding.DecodeListMeta(b)
	if err != nil {
		return
	}
	bytes := b[len(b)-size:]

	l := List[T]{
		meta:   meta,
		bytes:  bytes,
		decode: decode,
	}

	ln := l.Len()
	for i := 0; i < ln; i++ {
		elem := l.GetBytes(i)
		if len(elem) == 0 {
			continue
		}
		if _, _, err = decode(elem); err != nil {
			return
		}
	}
	return l, size, nil
}

// Len returns the number of elements in the list.
func (l List[T]) Len() int {
	return l.meta.Len()
}

// Bytes returns the exact list bytes.
func (l List[T]) Bytes() []byte {
	return l.bytes
}

// Get returns an element by index or panics on out of range.
func (l List[T]) Get(i int) (result T) {
	start, end := l.meta.Offset(i)
	size := l.meta.Data()

	if start < 0 || end > size {
		panic(fmt.Sprintf("index out out range: %d", i))
	}

	b := l.bytes[start:end]
	result, _, _ = l.decode(b)
	return result
}

// GetBytes returns raw element bytes or panics on out of range.
func (l List[T]) GetBytes(i int) []byte {
	start, end := l.meta.Offset(i)
	size := l.meta.Data()

	if start < 0 || end > size {
		panic(fmt.Sprintf("index out out range: %d", i))
	}

	return l.bytes[start:end]
}

// Values converts a list into a slice.
func (l List[T]) Values() []T {
	result := make([]T, 0, l.meta.Len())
	for i := 0; i < l.meta.Len(); i++ {
		elem := l.Get(i)
		result = append(result, elem)
	}
	return result
}
