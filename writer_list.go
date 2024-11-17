// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/writer"
)

type ListWriter = writer.ListWriter

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
