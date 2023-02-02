package spec

import (
	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/types"
)

// ListWriter writes a list of elements.
type ListWriter struct {
	e *Encoder
}

// NewListWriter returns a new list writer.
func NewListWriter() ListWriter {
	e := NewEncoder()
	return ListWriter{e}
}

// NewListWriterBuffer returns a new list writer with the given buffer.
func NewListWriterBuffer(buf buffer.Buffer) ListWriter {
	e := NewEncoderBuffer(buf)
	return ListWriter{e}
}

func newListWriter(e *Encoder) ListWriter {
	w := ListWriter{e: e}
	w.e.BeginList()
	return w
}

// Build ends the list and its bytes.
func (w ListWriter) Build() ([]byte, error) {
	return w.e.End()
}

// End ends the list.
func (w ListWriter) End() error {
	_, err := w.e.End()
	return err
}

// Elements

func (w ListWriter) Bool(v bool) error {
	if err := w.e.Bool(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Byte(v byte) error {
	if err := w.e.Byte(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Int32(v int32) error {
	if err := w.e.Int32(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Int64(v int64) error {
	if err := w.e.Int64(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Uint32(v uint32) error {
	if err := w.e.Uint32(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Uint64(v uint64) error {
	if err := w.e.Uint64(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Float32(v float32) error {
	if err := w.e.Float32(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Float64(v float64) error {
	if err := w.e.Float64(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Bin64(v types.Bin64) error {
	if err := w.e.Bin64(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Bin128(v types.Bin128) error {
	if err := w.e.Bin128(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Bin256(v types.Bin256) error {
	if err := w.e.Bin256(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) Bytes(v []byte) error {
	if err := w.e.Bytes(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) String(v string) error {
	if err := w.e.String(v); err != nil {
		return err
	}
	return w.e.Element()
}

func (w ListWriter) List() ListWriter {
	w.e.BeginElement()
	return newListWriter(w.e)
}

func (w ListWriter) Message() MessageWriter {
	w.e.BeginElement()
	return newMessageWriter(w.e)
}
