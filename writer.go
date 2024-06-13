package spec

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/writer"
)

type (
	Writer        = writer.Writer
	ListWriter    = writer.ListWriter
	FieldWriter   = writer.FieldWriter
	MessageWriter = writer.MessageWriter
	ValueWriter   = writer.ValueWriter
)

// NewWriter returns a new writer with a new empty buffer.
//
// The writer must be freed manually.
func NewWriter() Writer {
	return writer.New(false /* no release */)
}

// NewWriterBuffer returns a new writer with the given buffer.
//
// The writer must be freed manually.
func NewWriterBuffer(buf buffer.Buffer) Writer {
	return writer.NewBuffer(buf, false /* no release */)
}

// List

// NewListWriter returns a new list writer with a new empty buffer.
//
// The writer is released on end.
func NewListWriter() ListWriter {
	w := writer.New(true /* release */)
	return w.List()
}

// NewListWriterBuffer returns a new list writer with the given buffer.
//
// The writer is freed on end.
func NewListWriterBuffer(buf buffer.Buffer) ListWriter {
	w := writer.Acquire(buf)
	return w.List()
}

// Message

// NewMessageWriter returns a new message writer with a new empty buffer.
//
// The writer is released on end.
func NewMessageWriter() MessageWriter {
	w := writer.New(true /* release */)
	return w.Message()
}

// NewMessageWriterBuffer returns a new message writer with the given buffer.
//
// The writer is freed on end.
func NewMessageWriterBuffer(buf buffer.Buffer) MessageWriter {
	w := writer.Acquire(buf)
	return w.Message()
}

// WriteField writes a generic field using the given encode function.
func WriteField[T any](w writer.FieldWriter, value T, write writer.WriteFunc[T]) error {
	return writer.WriteField(w, value, write)
}

// Value

// NewValueWriter returns a new value writer with a new empty buffer.
//
// The writer is released on end.
func NewValueWriter() ValueWriter {
	w := writer.New(true /* release */)
	return w.Value()
}

// NewValueWriterBuffer returns a new value writer with the given buffer.
//
// The writer is freed on end.
func NewValueWriterBuffer(buf buffer.Buffer) ValueWriter {
	w := writer.Acquire(buf)
	return w.Value()
}
