package spec

import (
	"github.com/complex1tech/spec/encoding"
)

type TypedList[T any] struct {
	meta   encoding.ListMeta
	bytes  []byte
	decode func(b []byte) (T, int, error)
}

// NewTypedList decodes and returns a list without recursive validation, or an empty list on error.
func NewTypedList[T any](b []byte, decode func([]byte) (T, int, error)) TypedList[T] {
	meta, n, err := encoding.DecodeListMeta(b)
	if err != nil {
		return TypedList[T]{}
	}
	bytes := b[len(b)-n:]

	l := TypedList[T]{
		meta:   meta,
		bytes:  bytes,
		decode: decode,
	}
	return l
}

// DecodeList decodes, recursively validates and returns a list.
func DecodeList[T any](b []byte, decode func([]byte) (T, int, error)) (_ TypedList[T], size int, err error) {
	meta, size, err := encoding.DecodeListMeta(b)
	if err != nil {
		return
	}
	bytes := b[len(b)-size:]

	l := TypedList[T]{
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
func (l TypedList[T]) Len() int {
	return l.meta.Len()
}

// Bytes returns the exact list bytes.
func (l TypedList[T]) Bytes() []byte {
	return l.bytes
}

// Get returns an element at index i or panics on out of range.
func (l TypedList[T]) Get(i int) (result T) {
	start, end := l.meta.Offset(i)
	size := l.meta.DataSize()

	// TODO: Or should be panic index out out range?
	switch {
	case start < 0:
		return
	case end > int(size):
		return
	}

	b := l.bytes[start:end]
	result, _, _ = l.decode(b)
	return result
}

// GetBytes returns raw element bytes or panics on out of range.
func (l TypedList[T]) GetBytes(i int) []byte {
	start, end := l.meta.Offset(i)
	size := l.meta.DataSize()

	// TODO: Or should be panic index out out range?
	switch {
	case start < 0:
		return nil
	case end > int(size):
		return nil
	}

	return l.bytes[start:end]
}

// Values converts a list into a slice.
func (l TypedList[T]) Values() []T {
	result := make([]T, 0, l.meta.Len())
	for i := 0; i < l.meta.Len(); i++ {
		elem := l.Get(i)
		result = append(result, elem)
	}
	return result
}
