package writer

import (
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/encoding"
)

// ValueWriter writes spec values.
type ValueWriter struct {
	w *writer
}

// Build ends the root value and returns its bytes.
// The method returns an error if the value is not root.
func (w ValueWriter) Build() ([]byte, error) {
	return w.w.end()
}

// Values

func (w ValueWriter) Bool(v bool) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeBool(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Byte(v byte) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeByte(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

// Int

func (w ValueWriter) Int16(v int16) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeInt16(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Int32(v int32) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeInt32(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Int64(v int64) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeInt64(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

// Uint

func (w ValueWriter) Uint16(v uint16) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeUint16(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Uint32(v uint32) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeUint32(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Uint64(v uint64) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeUint64(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

// Bin64/128/256

func (w ValueWriter) Bin64(v types.Bin64) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeBin64(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Bin128(v types.Bin128) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeBin128(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Bin256(v types.Bin256) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeBin256(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

// Float

func (w ValueWriter) Float32(v float32) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeFloat32(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) Float64(v float64) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	encoding.EncodeFloat64(w.w.buf, v)
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

// Bytes/string

func (w ValueWriter) Bytes(v []byte) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	if _, err := encoding.EncodeBytes(w.w.buf, v); err != nil {
		return w.w.fail(err)
	}
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

func (w ValueWriter) String(v string) error {
	if w.w.err != nil {
		return w.w.err
	}

	start := w.w.buf.Len()
	if _, err := encoding.EncodeString(w.w.buf, v); err != nil {
		return w.w.fail(err)
	}
	end := w.w.buf.Len()

	return w.w.pushData(start, end)
}

// List/Message

func (w ValueWriter) List() ListWriter {
	return w.w.List()
}

func (w ValueWriter) Message() MessageWriter {
	return w.w.Message()
}
