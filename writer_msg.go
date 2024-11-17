// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/writer"
)

type (
	FieldWriter   = writer.FieldWriter
	MessageWriter = writer.MessageWriter
)

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
