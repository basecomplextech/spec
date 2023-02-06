package spec

import (
	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/spec/writer"
)

type (
	Writer        = writer.Writer
	ListWriter    = writer.ListWriter
	MessageWriter = writer.MessageWriter
)

// NewWriter returns a new writer with a new empty buffer.
// The writer must be freed manually.
func NewWriter() Writer {
	return writer.New(false /* no auto release */)
}

// NewWriterBuffer returns a new writer with the given buffer.
// The writer must be freed manually.
func NewWriterBuffer(buf buffer.Buffer) Writer {
	return writer.NewBuffer(buf, false /* no auto release */)
}

// List writer

// NewListWriter returns a new list writer with a new empty buffer.
// The writer is autoreleased on end.
func NewListWriter() ListWriter {
	w := writer.New(true /* auto release */)
	return w.List()
}

// NewListWriterBuffer returns a new list writer with the given buffer.
// The writer is autoreleased on end.
func NewListWriterBuffer(buf buffer.Buffer) ListWriter {
	w := writer.NewBuffer(buf, true /* auto release */)
	return w.List()
}

// Message writer

// NewMessageWriter returns a new message writer with a new empty buffer.
// The writer is autoreleased on end.
func NewMessageWriter() MessageWriter {
	w := writer.New(true /* auto release */)
	return w.Message()
}

// NewMessageWriterBuffer returns a new message writer with the given buffer.
// The writer is autoreleased on end.
func NewMessageWriterBuffer(buf buffer.Buffer) MessageWriter {
	w := writer.NewBuffer(buf, true /* auto release */)
	return w.Message()
}
