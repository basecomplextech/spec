package spec

import (
	"github.com/complex1tech/baselibrary/buffer"
	"github.com/complex1tech/baselibrary/types"
)

// MessageWriter writes a message.
type MessageWriter struct {
	e *Encoder
}

// NewMessageWriter returns a new message writer.
func NewMessageWriter() MessageWriter {
	e := NewEncoder()
	return MessageWriter{e}
}

// NewMessageWriterBuffer returns a new message writer with the given buffer.
func NewMessageWriterBuffer(buf buffer.Buffer) MessageWriter {
	e := NewEncoderBuffer(buf)
	return MessageWriter{e}
}

func newMessageWriter(e *Encoder) MessageWriter {
	w := MessageWriter{e: e}
	w.e.BeginMessage()
	return w
}

// Field returns a field writer.
func (w MessageWriter) Field(field uint16) FieldWriter {
	return newFieldWriter(w.e, field)
}

// Build ends the message and returns its bytes.
func (w MessageWriter) Build() ([]byte, error) {
	return w.e.End()
}

// End ends the message.
func (w MessageWriter) End() error {
	_, err := w.e.End()
	return err
}

// FieldWriter writes a message field.
type FieldWriter struct {
	e     *Encoder
	field uint16
}

func newFieldWriter(e *Encoder, field uint16) FieldWriter {
	return FieldWriter{
		e:     e,
		field: field,
	}
}

func (w FieldWriter) Any(b []byte) error {
	_, _, err := DecodeType(b)
	if err != nil {
		return err
	}
	return w.e.FieldBytes(w.field, b)
}

func (w FieldWriter) Bool(v bool) error {
	if err := w.e.Bool(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Byte(v byte) error {
	if err := w.e.Byte(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Int32(v int32) error {
	if err := w.e.Int32(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Int64(v int64) error {
	if err := w.e.Int64(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Uint32(v uint32) error {
	if err := w.e.Uint32(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Uint64(v uint64) error {
	if err := w.e.Uint64(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Float32(v float32) error {
	if err := w.e.Float32(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Float64(v float64) error {
	if err := w.e.Float64(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Bin64(v types.Bin64) error {
	if err := w.e.Bin64(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Bin128(v types.Bin128) error {
	if err := w.e.Bin128(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Bin256(v types.Bin256) error {
	if err := w.e.Bin256(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) Bytes(v []byte) error {
	if err := w.e.Bytes(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) String(v string) error {
	if err := w.e.String(v); err != nil {
		return err
	}
	return w.e.Field(w.field)
}

func (w FieldWriter) List() ListWriter {
	w.e.BeginField(w.field)
	return newListWriter(w.e)
}

func (w FieldWriter) Message() MessageWriter {
	w.e.BeginField(w.field)
	return newMessageWriter(w.e)
}
