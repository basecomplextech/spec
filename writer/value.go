package writer

import (
	"github.com/complex1tech/baselibrary/types"
	"github.com/complex1tech/spec/encoding"
)

// ValueWriter writes spec values.
type ValueWriter struct {
	w *writer
}

func WriteValue[T any](e *writer, v T, encode encoding.EncodeFunc[T]) error {
	if e.err != nil {
		return e.err
	}

	start := e.buf.Len()
	if _, err := encode(e.buf, v); err != nil {
		return e.close(err)
	}
	end := e.buf.Len()

	return e.pushData(start, end)
}

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
		return w.w.close(err)
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
		return w.w.close(err)
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
