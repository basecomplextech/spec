// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/writer"
)

type ValueWriter = writer.ValueWriter

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
