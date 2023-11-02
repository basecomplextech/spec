package rpc

import (
	"sync"

	"github.com/basecomplextech/baselibrary/alloc"
	"github.com/basecomplextech/spec"
)

// NewBuffer returns a new alloc.Buffer.
// The method is used in generated code.
func NewBuffer() *alloc.Buffer {
	return alloc.NewBuffer()
}

// buffer pool

var bufferPool = &sync.Pool{}

func acquireBuffer() *alloc.Buffer {
	v := bufferPool.Get()
	if v == nil {
		return alloc.NewBuffer()
	}
	return v.(*alloc.Buffer)
}

func releaseBuffer(buf *alloc.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

// buffer writer

type bufferWriter struct {
	buf    *alloc.Buffer
	writer spec.Writer
}

func newBufferWriter() *bufferWriter {
	buf := alloc.NewBuffer()
	return &bufferWriter{
		buf:    buf,
		writer: spec.NewWriterBuffer(buf),
	}
}

func (w *bufferWriter) Free() {
	releaseBufferWriter(w)
}

func (w *bufferWriter) reset() {
	w.buf.Reset()
	w.writer.Reset(w.buf)
}

// buffer writer pool

var bufferWriterPool = &sync.Pool{}

func acquireBufferWriter() *bufferWriter {
	v := bufferWriterPool.Get()
	if v == nil {
		return newBufferWriter()
	}
	return v.(*bufferWriter)
}

func releaseBufferWriter(w *bufferWriter) {
	w.reset()
	bufferWriterPool.Put(w)
}
