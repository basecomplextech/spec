// Copyright 2022 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package spec

import (
	"github.com/basecomplextech/baselibrary/buffer"
	"github.com/basecomplextech/spec/internal/writer"
)

type Writer = writer.Writer

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
