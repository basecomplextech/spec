// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

type ValueList[T any] struct {
	list   List
	decode func([]byte) (T, int, error)
}

// NewValueList returns a new value list.
func NewValueList[T any](list List, decode func([]byte) (T, int, error)) ValueList[T] {
	return ValueList[T]{
		list:   list,
		decode: decode,
	}
}

// OpenValueList opens and returns a value list, or an empty list on error.
func OpenValueList[T any](b []byte, decode func([]byte) (T, int, error)) ValueList[T] {
	l := OpenList(b)

	return ValueList[T]{
		list:   l,
		decode: decode,
	}
}

// OpenValueListErr opens and returns a value list, or an error.
func OpenValueListErr[T any](b []byte, decode func([]byte) (T, int, error)) (_ ValueList[T], err error) {
	l, err := OpenListErr(b)
	if err != nil {
		return
	}

	l1 := ValueList[T]{
		list:   l,
		decode: decode,
	}
	return l1, nil
}

// ParseValueList decodes, recursively validates and returns a list.
func ParseValueList[T any](b []byte, decode func([]byte) (T, int, error)) (_ ValueList[T], size int, err error) {
	l, size, err := ParseList(b)
	if err != nil {
		return
	}

	list := ValueList[T]{
		list:   l,
		decode: decode,
	}

	ln := l.Len()
	for i := 0; i < ln; i++ {
		b1 := l.GetBytes(i)
		if len(b1) == 0 {
			continue
		}

		if _, _, err = decode(b1); err != nil {
			return
		}
	}
	return list, size, nil
}

// Len returns the number of elements in the list.
func (l ValueList[T]) Len() int {
	return l.list.Len()
}

// Raw returns the exact list bytes.
func (l ValueList[T]) Raw() []byte {
	return l.list.Raw()
}

// Empty returns true if bytes are empty or list has no elements.
func (l ValueList[T]) Empty() bool {
	return l.list.Empty()
}

// Get returns an decode at index i, panics on out of range.
func (l ValueList[T]) Get(i int) T {
	b := l.list.GetBytes(i)
	elem, _, _ := l.decode(b)
	return elem
}

// GetErr returns an decode at index i or an error.
func (l ValueList[T]) GetErr(i int) (T, error) {
	b := l.list.GetBytes(i)
	elem, _, err := l.decode(b)
	return elem, err
}

// GetBytes returns decode bytes at index i, panics on out of range.
func (l ValueList[T]) GetBytes(i int) []byte {
	return l.list.GetBytes(i)
}

// Values converts a list into a slice.
func (l ValueList[T]) Values() []T {
	result := make([]T, 0, l.list.Len())

	for i := 0; i < l.list.Len(); i++ {
		elem := l.Get(i)
		result = append(result, elem)
	}

	return result
}
