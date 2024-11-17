// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

type MessageList[T any] struct {
	list List
	open func([]byte) (T, error)
}

// NewMessageList returns a new message list.
func NewMessageList[T any](list List, open func([]byte) (T, error)) MessageList[T] {
	return MessageList[T]{
		list: list,
		open: open,
	}
}

// OpenMessageList opens and returns a message list, or an empty list on error.
func OpenMessageList[T any](b []byte, open func([]byte) (T, error)) MessageList[T] {
	l := OpenList(b)

	return MessageList[T]{
		list: l,
		open: open,
	}
}

// OpenMessageListErr opens and returns a message list, or an error.
func OpenMessageListErr[T any](b []byte, open func([]byte) (T, error)) (_ MessageList[T], err error) {
	l, err := OpenListErr(b)
	if err != nil {
		return
	}

	l1 := MessageList[T]{
		list: l,
		open: open,
	}
	return l1, nil
}

// ParseMessageList decodes, recursively validates and returns a list.
func ParseMessageList[T any](b []byte, open func([]byte) (T, error)) (_ MessageList[T], size int, err error) {
	l, size, err := ParseList(b)
	if err != nil {
		return
	}

	list := MessageList[T]{
		list: l,
		open: open,
	}

	ln := l.Len()
	for i := 0; i < ln; i++ {
		b1 := l.GetBytes(i)
		if len(b1) == 0 {
			continue
		}

		if _, err = open(b1); err != nil {
			return
		}
	}
	return list, size, nil
}

// Len returns the number of elements in the list.
func (l MessageList[T]) Len() int {
	return l.list.Len()
}

// Raw returns the exact list bytes.
func (l MessageList[T]) Raw() []byte {
	return l.list.Raw()
}

// Empty returns true if bytes are empty or list has no elements.
func (l MessageList[T]) Empty() bool {
	return l.list.Empty()
}

// Get

// Get returns an open at index i, panics on out of range.
func (l MessageList[T]) Get(i int) T {
	b := l.list.GetBytes(i)
	elem, _ := l.open(b)
	return elem
}

// GetErr returns an open at index i or an error.
func (l MessageList[T]) GetErr(i int) (T, error) {
	b := l.list.GetBytes(i)
	return l.open(b)
}

// GetBytes returns open bytes at index i, panics on out of range.
func (l MessageList[T]) GetBytes(i int) []byte {
	return l.list.GetBytes(i)
}

// Values

// Values converts a list into a slice.
func (l MessageList[T]) Values() []T {
	result := make([]T, 0, l.list.Len())

	for i := 0; i < l.list.Len(); i++ {
		elem := l.Get(i)
		result = append(result, elem)
	}

	return result
}
