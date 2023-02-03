package writer

import (
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/go/encoding"
)

// FieldWriter writes a message field.
type FieldWriter struct {
	w     *Writer
	field uint16
}

func newField(w *Writer, field uint16) FieldWriter {
	return FieldWriter{
		w:     w,
		field: field,
	}
}

func (f FieldWriter) Any(b []byte) error {
	_, _, err := encoding.DecodeType(b)
	if err != nil {
		return err
	}
	return f.w.FieldBytes(f.field, b)
}

func (f FieldWriter) Bool(v bool) error {
	if err := f.w.Bool(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Byte(v byte) error {
	if err := f.w.Byte(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Int32(v int32) error {
	if err := f.w.Int32(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Int64(v int64) error {
	if err := f.w.Int64(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Uint32(v uint32) error {
	if err := f.w.Uint32(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Uint64(v uint64) error {
	if err := f.w.Uint64(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Float32(v float32) error {
	if err := f.w.Float32(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Float64(v float64) error {
	if err := f.w.Float64(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Bin64(v types.Bin64) error {
	if err := f.w.Bin64(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Bin128(v types.Bin128) error {
	if err := f.w.Bin128(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Bin256(v types.Bin256) error {
	if err := f.w.Bin256(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) Bytes(v []byte) error {
	if err := f.w.Bytes(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) String(v string) error {
	if err := f.w.String(v); err != nil {
		return err
	}
	return f.w.Field(f.field)
}

func (f FieldWriter) List() ListWriter {
	f.w.BeginField(f.field)
	return f.w.List()
}

func (f FieldWriter) Message() MessageWriter {
	f.w.BeginField(f.field)
	return f.w.Message()
}
