package writer

import "github.com/basecomplextech/baselibrary/bin"

// MessageWriter writes a message.
type MessageWriter struct {
	w *writer
}

// Field returns a field writer.
func (m MessageWriter) Field(field uint16) FieldWriter {
	return newField(m.w, field)
}

// HasField returns true if the message has the given field.
// The method is only valid when there is no pending field.
func (m MessageWriter) HasField(field uint16) bool {
	return m.w.hasField(field)
}

// Build ends the message and returns its bytes.
func (m MessageWriter) Build() ([]byte, error) {
	return m.w.end()
}

// End ends the message.
func (m MessageWriter) End() error {
	_, err := m.w.end()
	return err
}

// WriteField writes a generic field using the given write function.
func WriteField[T any](w FieldWriter, value T, write WriteFunc[T]) error {
	if err := WriteValue(w.w, value, write); err != nil {
		return err
	}
	return w.w.field(w.tag)
}

// Field

// FieldWriter writes a message field.
type FieldWriter struct {
	w   *writer
	tag uint16
}

func newField(w *writer, tag uint16) FieldWriter {
	return FieldWriter{
		w:   w,
		tag: tag,
	}
}

// Any writes a field with any valid spec object.
func (f FieldWriter) Any(b []byte) error {
	return f.w.fieldAny(f.tag, b)
}

func (f FieldWriter) Bool(v bool) error {
	if err := f.w.Value().Bool(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Byte(v byte) error {
	if err := f.w.Value().Byte(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

// Int

func (f FieldWriter) Int16(v int16) error {
	if err := f.w.Value().Int16(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Int32(v int32) error {
	if err := f.w.Value().Int32(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Int64(v int64) error {
	if err := f.w.Value().Int64(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

// Uint

func (f FieldWriter) Uint16(v uint16) error {
	if err := f.w.Value().Uint16(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Uint32(v uint32) error {
	if err := f.w.Value().Uint32(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Uint64(v uint64) error {
	if err := f.w.Value().Uint64(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

// Float

func (f FieldWriter) Float32(v float32) error {
	if err := f.w.Value().Float32(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Float64(v float64) error {
	if err := f.w.Value().Float64(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

// Bin

func (f FieldWriter) Bin64(v bin.Bin64) error {
	if err := f.w.Value().Bin64(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Bin128(v bin.Bin128) error {
	if err := f.w.Value().Bin128(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) Bin256(v bin.Bin256) error {
	if err := f.w.Value().Bin256(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

// Bytes/string

func (f FieldWriter) Bytes(v []byte) error {
	if err := f.w.Value().Bytes(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

func (f FieldWriter) String(v string) error {
	if err := f.w.Value().String(v); err != nil {
		return err
	}
	return f.w.field(f.tag)
}

// List/message

func (f FieldWriter) List() ListWriter {
	f.w.beginField(f.tag)
	return f.w.List()
}

func (f FieldWriter) Message() MessageWriter {
	f.w.beginField(f.tag)
	return f.w.Message()
}
