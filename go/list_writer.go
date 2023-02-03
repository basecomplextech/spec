package spec

import (
	"github.com/complex1tech/spec/go/encoding"
	"github.com/complex1tech/spec/go/writer"
)

// NewListBuilder begins and returns a new list.
func NewListBuilder[T any](e *Writer, next func(e *Writer) T) (_ ListBuilder[T]) {
	e.BeginList()
	return ListBuilder[T]{e: e, next: next}
}

// Add adds and returns the next element.
func (b ListBuilder[T]) Add() (_ T) {
	b.e.BeginElement()
	return b.next(b.e)
}

// Len returns the number of elements in the builder.
func (b ListBuilder[T]) Len() int {
	return b.e.ListLen()
}

// Err returns the current build error.
func (b ListBuilder[T]) Err() error {
	return b.e.Err()
}

// End ends the list.
func (b ListBuilder[T]) End() error {
	_, err := b.e.End()
	return err
}

// ValueListWriter

// ValueListBuilder builds a list of values.
type ValueListBuilder[T any] struct {
	e      *Writer
	encode encoding.EncodeFunc[T]
}

// NewValueListBuilder begins and returns a new value list builder.
func NewValueListBuilder[T any](e *Writer, encode encoding.EncodeFunc[T]) (_ ValueListBuilder[T]) {
	e.BeginList()
	return ValueListBuilder[T]{e: e, encode: encode}
}

// Add adds the next element.
func (b ValueListBuilder[T]) Add(value T) error {
	if err := writer.WriteValue(b.e, value, b.encode); err != nil {
		return err
	}
	return b.e.Element()
}

// Len returns the number of elements in the builder.
func (b ValueListBuilder[T]) Len() int {
	return b.e.ListLen()
}

// End ends the list.
func (b ValueListBuilder[T]) End() error {
	_, err := b.e.End()
	return err
}

// ListBuilder builds a list using nested element builder.
type ListBuilder[T any] struct {
	e    *Writer
	next func(e *Writer) T
}
