package spec

import (
	"github.com/complex1tech/spec/encoding"
	"github.com/complex1tech/spec/writer"
)

// MessageListWriter writes a list of messages.
type MessageListWriter[T any] struct {
	w    ListWriter
	next func(MessageWriter) T
}

// NewMessageListWriter returns a new message list writer.
func NewMessageListWriter[T any](w ListWriter, next func(w MessageWriter) T) (_ MessageListWriter[T]) {
	return MessageListWriter[T]{
		w:    w,
		next: next,
	}
}

// Add adds and returns the next element.
func (b MessageListWriter[T]) Add() (_ T) {
	msg := b.w.Message()
	return b.next(msg)
}

// Len returns the number of written elements.
// The method is only valid when there is no pending element.
func (b MessageListWriter[T]) Len() int {
	return b.w.Len()
}

// Err returns the current build error.
func (b MessageListWriter[T]) Err() error {
	return b.w.Err()
}

// End ends the list.
func (b MessageListWriter[T]) End() error {
	return b.w.End()
}

// Value list

// ValueListWriter writes a list of primitive values.
type ValueListWriter[T any] struct {
	w      ListWriter
	encode encoding.EncodeFunc[T]
}

// NewValueListWriter returns a new value list writer.
func NewValueListWriter[T any](w ListWriter, encode encoding.EncodeFunc[T]) (_ ValueListWriter[T]) {
	return ValueListWriter[T]{
		w:      w,
		encode: encode,
	}
}

// Add adds the next element.
func (b ValueListWriter[T]) Add(value T) error {
	return writer.WriteElement(b.w, value, b.encode)
}

// Len returns the number of written elements.
// The method is only valid when there is no pending element.
func (b ValueListWriter[T]) Len() int {
	return b.w.Len()
}

// End ends the list.
func (b ValueListWriter[T]) End() error {
	return b.w.End()
}
