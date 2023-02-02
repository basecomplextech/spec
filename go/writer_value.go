package spec

import (
	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/types"
)

// ValueWriter writes a value.
type ValueWriter struct {
	e *Writer
}

// NewValueWriterBuffer returns a new value writer with the given buffer.
func NewValueWriterBuffer(buf buffer.Buffer) ValueWriter {
	e := NewWriterBuffer(buf)
	return ValueWriter{e}
}

func newValueWriter(e *Writer) ValueWriter {
	return ValueWriter{e: e}
}

func (w ValueWriter) Bool(v bool) error {
	return w.e.Bool(v)
}

func (w ValueWriter) Byte(v byte) error {
	return w.e.Byte(v)
}

func (w ValueWriter) Int32(v int32) error {
	return w.e.Int32(v)
}

func (w ValueWriter) Int64(v int64) error {
	return w.e.Int64(v)
}

func (w ValueWriter) Uint32(v uint32) error {
	return w.e.Uint32(v)
}

func (w ValueWriter) Uint64(v uint64) error {
	return w.e.Uint64(v)
}

func (w ValueWriter) Float32(v float32) error {
	return w.e.Float32(v)
}

func (w ValueWriter) Float64(v float64) error {
	return w.e.Float64(v)
}

func (w ValueWriter) Bin64(v types.Bin64) error {
	return w.e.Bin64(v)
}

func (w ValueWriter) Bin128(v types.Bin128) error {
	return w.e.Bin128(v)
}

func (w ValueWriter) Bin256(v types.Bin256) error {
	return w.e.Bin256(v)
}

func (w ValueWriter) Bytes(v []byte) error {
	return w.e.Bytes(v)
}

func (w ValueWriter) String(v string) error {
	return w.e.String(v)
}

func (w ValueWriter) List() ListWriter {
	w.e.BeginList()
	return newListWriter(w.e)
}

func (w ValueWriter) Message() MessageWriter {
	w.e.BeginMessage()
	return newMessageWriter(w.e)
}
